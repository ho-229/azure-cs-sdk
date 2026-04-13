package azure_cs_sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecognizeShortSimpleReturnsPrimaryLanguage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"RecognitionStatus":"Success","DisplayText":"hello","Offset":10,"Duration":20,"PrimaryLanguage":{"Language":"en-US","Confidence":"High"}}`)
	}))
	defer ts.Close()

	stt := &AzureCSSTT{
		speechToTextAPI: ts.URL,
		client: &AzureCS{
			accessToken: "token",
			httpClient:  http.DefaultClient,
		},
	}

	resp, err := stt.RecognizeShortSimple(strings.NewReader("audio"), RIFF16khz16bitMonoPCM, "en-US")
	require.NoError(t, err)
	assert.Equal(t, RecognitionStatusSuccess, resp.RecognitionStatus)
	assert.Equal(t, "hello", resp.DisplayText)
	require.NotNil(t, resp.PrimaryLanguage)
	assert.Equal(t, "en-US", resp.PrimaryLanguage.Language)
	assert.Equal(t, "High", resp.PrimaryLanguage.Confidence)
}
