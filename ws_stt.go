package azure_cs_sdk

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const wavHeaderSize = 44
const wsAudioChunkSize = 4096
const defaultWAVByteRate = 32000

type wsTextMessage struct {
	headers map[string]string
	body    string
}

type wsSpeechConfig struct {
	Context wsSpeechConfigContext `json:"context"`
}

type wsSpeechConfigContext struct {
	System wsSpeechConfigSystem `json:"system"`
	Audio  wsSpeechConfigAudio  `json:"audio"`
}

type wsSpeechConfigSystem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Build   string `json:"build"`
	Lang    string `json:"lang"`
}

type wsSpeechConfigAudio struct {
	Source wsSpeechConfigAudioSource `json:"source"`
}

type wsSpeechConfigAudioSource struct {
	Bitspersample int    `json:"bitspersample"`
	Channelcount  int    `json:"channelcount"`
	Connectivity  string `json:"connectivity"`
	Manufacturer  string `json:"manufacturer"`
	Model         string `json:"model"`
	Samplerate    int    `json:"samplerate"`
	Type          string `json:"type"`
	Version       string `json:"version"`
}

type wsSpeechContext struct {
	LanguageID      wsLanguageIDContext      `json:"languageId"`
	PhraseDetection wsPhraseDetectionContext `json:"phraseDetection"`
	PhraseOutput    wsPhraseOutputContext    `json:"phraseOutput"`
}

type wsLanguageIDContext struct {
	Languages []string        `json:"languages"`
	Mode      string          `json:"mode"`
	OnSuccess wsActionContext `json:"onSuccess"`
	OnUnknown wsActionContext `json:"onUnknown"`
	Priority  string          `json:"priority"`
}

type wsPhraseDetectionContext struct {
	Mode string `json:"mode"`
}

type wsPhraseOutputContext struct {
	InterimResults wsPhraseOutputResultType `json:"interimResults"`
	PhraseResults  wsPhraseOutputResultType `json:"phraseResults"`
}

type wsPhraseOutputResultType struct {
	ResultType string `json:"resultType"`
}

type wsActionContext struct {
	Action string `json:"action"`
}

func normalizeCandidateLanguages(languages []string) ([]string, error) {
	if len(languages) == 0 {
		return nil, fmt.Errorf("at least one candidate language is required")
	}

	seen := make(map[string]struct{}, len(languages))
	result := make([]string, 0, len(languages))
	for _, language := range languages {
		language = strings.TrimSpace(language)
		if language == "" {
			return nil, fmt.Errorf("candidate languages must not be empty")
		}
		if _, ok := seen[language]; ok {
			continue
		}
		seen[language] = struct{}{}
		result = append(result, language)
	}
	return result, nil
}

type RecognizeEventType string

const (
	RecognizeEventPartial RecognizeEventType = "partial"
	RecognizeEventFinal   RecognizeEventType = "final"
	RecognizeEventError   RecognizeEventType = "error"
)

type RecognizeEvent struct {
	Type   RecognizeEventType
	Result *RecognizeSimpleResponse
	Err    error
}

type recognizeHypothesisResponse struct {
	Text            string
	Offset          uint64
	Duration        uint64
	PrimaryLanguage *PrimaryLanguage
}

func (r *recognizeHypothesisResponse) toSimpleResponse() *RecognizeSimpleResponse {
	resp := &RecognizeSimpleResponse{
		DisplayText:     r.Text,
		Offset:          r.Offset,
		Duration:        r.Duration,
		PrimaryLanguage: r.PrimaryLanguage,
	}
	return resp
}

func (az *AzureCSSTT) Recognize(
	reader io.Reader,
	audioType AudioType,
	languages []string,
	opts ...Option,
) (<-chan RecognizeEvent, error) {
	return az.RecognizeWithContext(context.Background(), reader, audioType, languages, opts...)
}

func (az *AzureCSSTT) RecognizeWithContext(
	ctx context.Context,
	reader io.Reader,
	audioType AudioType,
	languages []string,
	opts ...Option,
) (<-chan RecognizeEvent, error) {
	conn, requestID, err := az.openRecognizeConnection(ctx, audioType, languages, opts...)
	if err != nil {
		return nil, err
	}

	events := make(chan RecognizeEvent, 8)
	go az.runRecognizeStream(ctx, conn, requestID, reader, events)
	return events, nil
}

func (az *AzureCSSTT) openRecognizeConnection(
	ctx context.Context,
	audioType AudioType,
	languages []string,
	opts ...Option,
) (*websocket.Conn, string, error) {
	if audioType != RIFF16khz16bitMonoPCM {
		return nil, "", fmt.Errorf("websocket recognize currently supports only %s", RIFF16khz16bitMonoPCM)
	}

	candidates, err := normalizeCandidateLanguages(languages)
	if err != nil {
		return nil, "", err
	}

	params := options{Profanity: ProfanityMasked, Cid: ""}
	for _, opt := range opts {
		opt(&params)
	}

	requestID, err := newWSRequestID()
	if err != nil {
		return nil, "", err
	}
	connectionID, err := newWSRequestID()
	if err != nil {
		return nil, "", err
	}

	endpoint, err := az.buildAutoDetectWSEndpoint(params)
	if err != nil {
		return nil, "", err
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+az.client.accessToken)
	headers.Set("X-ConnectionId", connectionID)

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.DialContext(ctx, endpoint, headers)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			if len(body) > 0 {
				return nil, "", fmt.Errorf("failed to connect to speech websocket: %w (status=%s body=%q)", err, resp.Status, string(body))
			}
			return nil, "", fmt.Errorf("failed to connect to speech websocket: %w (status=%s)", err, resp.Status)
		}
		return nil, "", fmt.Errorf("failed to connect to speech websocket: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	if err := writeWSTextFrame(conn, "speech.config", requestID, "application/json", buildWSSpeechConfig()); err != nil {
		_ = conn.Close()
		return nil, "", err
	}
	if err := writeWSTextFrame(conn, "speech.context", requestID, "application/json", buildWSSpeechContext(candidates)); err != nil {
		_ = conn.Close()
		return nil, "", err
	}

	return conn, requestID, nil
}

func (az *AzureCSSTT) runRecognizeStream(
	ctx context.Context,
	conn *websocket.Conn,
	requestID string,
	reader io.Reader,
	events chan<- RecognizeEvent,
) {
	defer close(events)
	defer conn.Close()

	sendErrCh := make(chan error, 1)
	go func() {
		err := streamWSWaveAudio(ctx, conn, requestID, reader)
		if err == nil {
			err = writeWSBinaryFrame(conn, "audio", requestID, "", nil)
		}
		if err != nil {
			_ = conn.Close()
		}
		sendErrCh <- err
	}()

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				events <- RecognizeEvent{Type: RecognizeEventError, Err: ctx.Err()}
				return
			}
			select {
			case sendErr := <-sendErrCh:
				if sendErr != nil {
					events <- RecognizeEvent{Type: RecognizeEventError, Err: sendErr}
					return
				}
			default:
			}
			events <- RecognizeEvent{Type: RecognizeEventError, Err: fmt.Errorf("failed to read speech websocket response: %w", err)}
			return
		}
		if messageType != websocket.TextMessage {
			continue
		}

		message, err := parseWSTextMessage(payload)
		if err != nil {
			events <- RecognizeEvent{Type: RecognizeEventError, Err: err}
			return
		}

		event, done, err := parseWSRecognizeEvent(message)
		if err != nil {
			events <- RecognizeEvent{Type: RecognizeEventError, Err: err}
			return
		}
		if event != nil {
			events <- *event
		}
		if done {
			return
		}

		select {
		case sendErr := <-sendErrCh:
			if sendErr != nil {
				events <- RecognizeEvent{Type: RecognizeEventError, Err: sendErr}
				return
			}
		default:
		}
	}
}

func parseWSRecognizeEvent(message *wsTextMessage) (*RecognizeEvent, bool, error) {
	switch strings.ToLower(message.headers["path"]) {
	case "speech.hypothesis":
		var hypothesis recognizeHypothesisResponse
		if err := json.Unmarshal([]byte(message.body), &hypothesis); err != nil {
			return nil, false, fmt.Errorf("failed to decode speech hypothesis: %w", err)
		}
		return &RecognizeEvent{
			Type:   RecognizeEventPartial,
			Result: hypothesis.toSimpleResponse(),
		}, false, nil
	case "speech.phrase":
		var detailed recognizeDetailedResponse
		if err := json.Unmarshal([]byte(message.body), &detailed); err == nil && detailed.RecognitionStatus != "" {
			return &RecognizeEvent{
				Type:   RecognizeEventFinal,
				Result: detailed.toSimpleResponse(),
			}, true, nil
		}

		var simple RecognizeSimpleResponse
		if err := json.Unmarshal([]byte(message.body), &simple); err != nil {
			return nil, false, fmt.Errorf("failed to decode speech phrase: %w", err)
		}
		return &RecognizeEvent{Type: RecognizeEventFinal, Result: &simple}, true, nil
	case "turn.end":
		return nil, true, fmt.Errorf("speech websocket turn ended before a final phrase was returned")
	default:
		return nil, false, nil
	}
}

func (az *AzureCSSTT) buildAutoDetectWSEndpoint(params options) (string, error) {
	baseURL, err := url.Parse(az.speechToTextWSAPI)
	if err != nil {
		return "", fmt.Errorf("failed to parse websocket endpoint: %w", err)
	}

	query := baseURL.Query()
	query.Set("format", "detailed")
	query.Set("lidEnabled", "true")
	query.Set("profanity", string(params.Profanity))
	if params.Cid != "" {
		query.Set("cid", params.Cid)
	}
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func buildWSSpeechConfig() string {
	payload := wsSpeechConfig{
		Context: wsSpeechConfigContext{
			System: wsSpeechConfigSystem{
				Name:    "azure-cs-sdk",
				Version: "0.0.0",
				Build:   "Go",
				Lang:    "Go",
			},
			Audio: wsSpeechConfigAudio{
				Source: wsSpeechConfigAudioSource{
					Bitspersample: 16,
					Channelcount:  1,
					Connectivity:  "Unknown",
					Manufacturer:  "unknown",
					Model:         "unknown",
					Samplerate:    16000,
					Type:          "File",
					Version:       "1.0.0",
				},
			},
		},
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func buildWSSpeechContext(languages []string) string {
	payload := wsSpeechContext{
		LanguageID: wsLanguageIDContext{
			Languages: languages,
			Mode:      "DetectAtAudioStart",
			OnSuccess: wsActionContext{Action: "Recognize"},
			OnUnknown: wsActionContext{Action: "None"},
			Priority:  "PrioritizeLatency",
		},
		PhraseDetection: wsPhraseDetectionContext{Mode: "Conversation"},
		PhraseOutput: wsPhraseOutputContext{
			InterimResults: wsPhraseOutputResultType{ResultType: "Auto"},
			PhraseResults:  wsPhraseOutputResultType{ResultType: "Always"},
		},
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func streamWSWaveAudio(ctx context.Context, conn *websocket.Conn, requestID string, reader io.Reader) error {
	header := make([]byte, wavHeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return fmt.Errorf("failed to read wav header: %w", err)
	}
	byteRate := wavByteRate(header)
	if err := writeWSBinaryFrame(conn, "audio", requestID, "audio/x-wav", header); err != nil {
		return err
	}

	buf := make([]byte, wsAudioChunkSize)
	start := time.Now()
	var bytesSent int64
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if err := waitForWSAudioClock(ctx, start, bytesSent, byteRate); err != nil {
				return err
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if err := writeWSBinaryFrame(conn, "audio", requestID, "", chunk); err != nil {
				return err
			}
			bytesSent += int64(n)
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to stream wav audio: %w", err)
		}
	}
}

func waitForWSAudioClock(ctx context.Context, start time.Time, bytesSent int64, byteRate int) error {
	if byteRate <= 0 || bytesSent <= 0 {
		return nil
	}
	target := start.Add(time.Duration(bytesSent) * time.Second / time.Duration(byteRate))
	wait := time.Until(target)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func wavByteRate(header []byte) int {
	if len(header) < 32 {
		return defaultWAVByteRate
	}
	byteRate := int(binary.LittleEndian.Uint32(header[28:32]))
	if byteRate <= 0 {
		return defaultWAVByteRate
	}
	return byteRate
}

func parseWSTextMessage(payload []byte) (*wsTextMessage, error) {
	parts := strings.SplitN(string(payload), "\r\n\r\n", 2)
	headers := make(map[string]string)
	for _, line := range strings.Split(parts[0], "\r\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		segments := strings.SplitN(line, ":", 2)
		if len(segments) != 2 {
			continue
		}
		headers[strings.ToLower(strings.TrimSpace(segments[0]))] = strings.TrimSpace(segments[1])
	}
	body := ""
	if len(parts) == 2 {
		body = parts[1]
	}
	return &wsTextMessage{headers: headers, body: body}, nil
}

func writeWSTextFrame(conn *websocket.Conn, path, requestID, contentType, body string) error {
	frame := buildWSTextFrame(path, requestID, contentType, body)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(frame)); err != nil {
		return fmt.Errorf("failed to write websocket text frame %s: %w", path, err)
	}
	return nil
}

func writeWSBinaryFrame(conn *websocket.Conn, path, requestID, contentType string, body []byte) error {
	frame := buildWSBinaryFrame(path, requestID, contentType, body)
	if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("failed to write websocket binary frame %s: %w", path, err)
	}
	return nil
}

func buildWSTextFrame(path, requestID, contentType, body string) string {
	headers := []string{
		fmt.Sprintf("Path: %s", path),
		fmt.Sprintf("X-RequestId: %s", requestID),
		fmt.Sprintf("X-Timestamp: %s", time.Now().UTC().Format(time.RFC3339Nano)),
	}
	if contentType != "" {
		headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

func buildWSBinaryFrame(path, requestID, contentType string, body []byte) []byte {
	headers := []string{
		fmt.Sprintf("Path: %s", path),
		fmt.Sprintf("X-RequestId: %s", requestID),
		fmt.Sprintf("X-Timestamp: %s", time.Now().UTC().Format(time.RFC3339Nano)),
	}
	if contentType != "" {
		headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))
	}
	headerBytes := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n")
	payload := make([]byte, 2+len(headerBytes)+len(body))
	binary.BigEndian.PutUint16(payload[:2], uint16(len(headerBytes)))
	copy(payload[2:], headerBytes)
	copy(payload[2+len(headerBytes):], body)
	return payload
}

func newWSRequestID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate request id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
