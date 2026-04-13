package azure_cs_sdk

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecognizeStream(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("lidEnabled"))
		assert.Equal(t, "detailed", r.URL.Query().Get("format"))
		assert.Equal(t, "", r.URL.Query().Get("language"))

		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		msgType, speechConfig, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.TextMessage, msgType)
		assert.Contains(t, string(speechConfig), "Path: speech.config")

		msgType, speechContext, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.TextMessage, msgType)
		assert.Contains(t, string(speechContext), "Path: speech.context")
		assert.Contains(t, string(speechContext), `"languages":["zh-CN","en-US"]`)

		msgType, wavHeader, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.BinaryMessage, msgType)
		headerLen := int(binary.BigEndian.Uint16(wavHeader[:2]))
		assert.Contains(t, string(wavHeader[2:2+headerLen]), "Content-Type: audio/x-wav")

		msgType, audioChunk, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.BinaryMessage, msgType)
		chunkHeaderLen := int(binary.BigEndian.Uint16(audioChunk[:2]))
		assert.Contains(t, string(audioChunk[2:2+chunkHeaderLen]), "Path: audio")

		msgType, finalAudio, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, websocket.BinaryMessage, msgType)
		finalHeaderLen := int(binary.BigEndian.Uint16(finalAudio[:2]))
		assert.Equal(t, 2+finalHeaderLen, len(finalAudio))

		hypothesis := buildWSTextFrame(
			"speech.hypothesis",
			"reqid",
			"application/json",
			`{"Text":"你","Offset":1,"Duration":2,"PrimaryLanguage":{"Language":"zh-CN","Confidence":"High"}}`,
		)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(hypothesis)))

		resp := buildWSTextFrame(
			"speech.phrase",
			"reqid",
			"application/json",
			`{"RecognitionStatus":"Success","Offset":10,"Duration":20,"PrimaryLanguage":{"Language":"zh-CN","Confidence":"High"},"NBest":[{"Confidence":0.98,"Display":"你好"}]}`,
		)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(resp)))
	}))
	defer server.Close()

	stt := &AzureCSSTT{
		speechToTextWSAPI: strings.Replace(server.URL, "http://", "ws://", 1),
		client: &AzureCS{
			accessToken: "token",
		},
	}

	wav := append(make([]byte, wavHeaderSize), []byte("pcmdata")...)
	events, err := stt.RecognizeWithContext(
		context.Background(),
		bytes.NewReader(wav),
		RIFF16khz16bitMonoPCM,
		[]string{"zh-CN", "en-US"},
	)
	require.NoError(t, err)

	partial, ok := <-events
	require.True(t, ok)
	assert.Equal(t, RecognizeEventPartial, partial.Type)
	require.NotNil(t, partial.Result)
	assert.Equal(t, "你", partial.Result.DisplayText)

	final, ok := <-events
	require.True(t, ok)
	assert.Equal(t, RecognizeEventFinal, final.Type)
	require.NotNil(t, final.Result)
	assert.Equal(t, RecognitionStatusSuccess, final.Result.RecognitionStatus)
	assert.Equal(t, "你好", final.Result.DisplayText)
	require.Len(t, final.Result.NBest, 1)
	assert.Equal(t, 0.98, final.Result.NBest[0].Confidence)
	assert.Equal(t, "你好", final.Result.NBest[0].Display)

	_, ok = <-events
	assert.False(t, ok)
}

func TestRecognizeFinalEvent(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		for i := 0; i < 5; i++ {
			_, _, err := conn.ReadMessage()
			require.NoError(t, err)
		}

		resp := buildWSTextFrame(
			"speech.phrase",
			"reqid",
			"application/json",
			`{"RecognitionStatus":"Success","Offset":10,"Duration":20,"PrimaryLanguage":{"Language":"zh-CN","Confidence":"High"},"NBest":[{"Confidence":0.98,"Display":"你好"}]}`,
		)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(resp)))
	}))
	defer server.Close()

	stt := &AzureCSSTT{
		speechToTextWSAPI: strings.Replace(server.URL, "http://", "ws://", 1),
		client:            &AzureCS{accessToken: "token"},
	}

	wav := append(make([]byte, wavHeaderSize), []byte("pcmdata")...)
	events, err := stt.RecognizeWithContext(
		context.Background(),
		bytes.NewReader(wav),
		RIFF16khz16bitMonoPCM,
		[]string{"zh-CN", "en-US"},
	)
	require.NoError(t, err)
	var final *RecognizeSimpleResponse
	for event := range events {
		require.NoError(t, event.Err)
		if event.Type == RecognizeEventFinal {
			final = event.Result
		}
	}
	require.NotNil(t, final)
	assert.Equal(t, RecognitionStatusSuccess, final.RecognitionStatus)
	assert.Equal(t, "你好", final.DisplayText)
	require.Len(t, final.NBest, 1)
	assert.Equal(t, "你好", final.NBest[0].Display)
}

func TestRecognizeRejectsUnsupportedAudioType(t *testing.T) {
	stt := &AzureCSSTT{client: &AzureCS{accessToken: "token"}}
	_, err := stt.Recognize(strings.NewReader("audio"), OGG16khz16bitMonoOpus, []string{"en-US"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("%s", RIFF16khz16bitMonoPCM))
}

func TestWAVByteRate(t *testing.T) {
	header := make([]byte, wavHeaderSize)
	binary.LittleEndian.PutUint32(header[28:32], 32000)
	assert.Equal(t, 32000, wavByteRate(header))
	assert.Equal(t, defaultWAVByteRate, wavByteRate(nil))
}

func TestWaitForWSAudioClockHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := waitForWSAudioClock(ctx, time.Now(), 32000, 32000)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}
