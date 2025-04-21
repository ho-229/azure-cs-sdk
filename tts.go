package azure_cs_sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ho-229/azure-cs-sdk/ssml"
)

// synthesizeActionTimeout is the amount of time the http client will wait for a response during Synthesize request
const synthesizeActionTimeout = time.Second * 30

// AzureCSTTS stores configuration and state information for the TTS client.
type AzureCSTTS struct {
	regionVoiceMap      RegionVoiceMap
	textToSpeechURL     string
	voiceServiceListURL string
	client              *AzureCS
}

func (az *AzureCSTTS) GetVoicesMap() RegionVoiceMap {
	return az.regionVoiceMap
}

// Synthesize directs to SynthesizeWithContext. A new context.Withtimeout is created with the timeout as defined by synthesizeActionTimeout
func (az *AzureCSTTS) Synthesize(speechText string, voiceName string, audioOutput AudioType) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), synthesizeActionTimeout)
	defer cancel()
	return az.SynthesizeWithContext(ctx, speechText, voiceName, audioOutput)
}

// SynthesizeWithContext returns a bytestream of the rendered text-to-speech in the target audio format. `speechText` is the string of
// text in which a user wishes to Synthesize, `region` is the language/locale
// and `audioOutput` captures the audio format.
func (az *AzureCSTTS) SynthesizeWithContext(ctx context.Context, speechText string, voiceName string, audioOutput AudioType) ([]byte, error) {
	if _, ok := az.regionVoiceMap[voiceName]; !ok {
		return nil, fmt.Errorf("voice name %s is not found in the voice map", voiceName)
	}

	var escapedBuffer bytes.Buffer
	if err := xml.EscapeText(&escapedBuffer, []byte(speechText)); err != nil {
		return nil, err
	}

	voice := ssml.NewVoice(voiceName)
	voice.Child = escapedBuffer.String()

	return az.SynthesizeSsmlWithContext(ctx, voice, audioOutput)
}

// SynthesizeSsmlWithContext returns a bytestream of the rendered text-to-speech in the target audio format.
// `ctx` is the context in which the request is made, `elems` is the SSML payload, and `audioOutput` captures the audio format.
func (az *AzureCSTTS) SynthesizeSsmlWithContext(
	ctx context.Context,
	elems xml.Token,
	audioOutput AudioType,
) ([]byte, error) {
	doc := ssml.NewSpeak()
	doc.Child = elems

	reqBody, err := xml.Marshal(doc)
	if err != nil {
		return nil, err
	}

	return az.SynthesizeRawSsmlWithContext(ctx, string(reqBody), audioOutput)
}

// SynthesizeRawSsmlWithContext returns a bytestream of the rendered text-to-speech in the target audio format.
// `ctx` is the context in which the request is made, `ssml` is the SSML payload, and `audioOutput` captures the audio format.
func (az *AzureCSTTS) SynthesizeRawSsmlWithContext(
	ctx context.Context,
	ssml string,
	audioOutput AudioType,
) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, az.textToSpeechURL, strings.NewReader(ssml))
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Microsoft-OutputFormat", audioOutput.String())
	request.Header.Set("Content-Type", "application/ssml+xml")
	request.Header.Set("Authorization", "Bearer "+az.client.accessToken)
	request.Header.Set("User-Agent", "azuretts")

	response, err := az.client.httpClient.Do(request.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// list of acceptable response status codes
	// see: https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech#http-status-codes-1
	switch response.StatusCode {
	case http.StatusOK:
		// The request was successful; the response body is an audio file.
		return io.ReadAll(response.Body)
	case http.StatusBadRequest:
		return nil, fmt.Errorf("%d - A required parameter is missing, empty, or null. Or, the value passed to either a required or optional parameter is invalid. A common issue is a header that is too long", response.StatusCode)
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("%d - The request is not authorized. Check to make sure your subscription key or token is valid and in the correct region", response.StatusCode)
	case http.StatusRequestEntityTooLarge:
		return nil, fmt.Errorf("%d - The SSML input is longer than 1024 characters", response.StatusCode)
	case http.StatusUnsupportedMediaType:
		return nil, fmt.Errorf("%d - It's possible that the wrong Content-Type was provided. Content-Type should be set to application/ssml+xml", response.StatusCode)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("%d - You have exceeded the quota or rate of requests allowed for your subscription", response.StatusCode)
	case http.StatusBadGateway:
		return nil, fmt.Errorf("%d - Network or server-side issue. May also indicate invalid headers", response.StatusCode)
	}

	return nil, fmt.Errorf("%d - received unexpected HTTP status code", response.StatusCode)
}

func (az *AzureCSTTS) buildVoiceToRegionMap() (RegionVoiceMap, error) {

	v, err := az.fetchVoiceList()
	if err != nil {
		return nil, err
	}

	m := make(map[string]RegionVoice)
	for _, x := range v {
		m[x.ShortName] = x
	}
	return m, err
}

func (az *AzureCSTTS) fetchVoiceList() ([]RegionVoice, error) {
	req, _ := http.NewRequest(http.MethodGet, az.voiceServiceListURL, nil)
	req.Header.Set("Authorization", "Bearer "+az.client.accessToken)

	// Perform the request
	res, err := az.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		var r []RegionVoice

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			return nil, fmt.Errorf("unable to decode voice list response body, %v", err)
		}
		return r, nil
	case http.StatusBadRequest:
		return nil, fmt.Errorf("%d - A required parameter is missing, empty, or null. Or, the value passed to either a required or optional parameter is invalid. A common issue is a header that is too long", res.StatusCode)
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("%d - The request is not authorized. Check to make sure your subscription key or token is valid and in the correct region", res.StatusCode)
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("%d - You have exceeded the quota or rate of requests allowed for your subscription", res.StatusCode)
	case http.StatusBadGateway:
		return nil, fmt.Errorf("%d - Network or server-side issue. May also indicate invalid headers", res.StatusCode)
	}
	return nil, fmt.Errorf("%d - unexpected response code from voice list API", res.StatusCode)
}
