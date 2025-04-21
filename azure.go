package azure_cs_sdk

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// The following are V1 endpoints for Cognitive Services endpoints
const textToSpeechAPI = "https://%s.tts.speech.microsoft.com/cognitiveservices"
const speechToTextAPI = "https://%s.stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1"
const tokenRefreshAPI = "https://%s.api.cognitive.microsoft.com/sts/v1.0/issueToken"

// tokenRefreshTimeout is the amount of time the http client will wait during the token refresh action.
const tokenRefreshTimeout = time.Second * 10

// tokenRefreshInterval is the amount of time between token refreshes
// ref: https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech?tabs=streaming#how-to-use-an-access-token
const tokenRefreshInterval = time.Minute * 9

// AzureCS is the main object for the Azure Cognitive Services client.
type AzureCS struct {
	accessToken        string    // is the auth token received from `TokenRefreshAPI`. Used in the Authorization: Bearer header.
	subscriptionKey    string    // API key for Azure's Cognitive Speech services
	tokenRefreshDoneCh chan bool // channel to stop the token refresh goroutine.
	tokenRefreshURL    string
	region             Region
	httpClient         *http.Client
}

// New returns an AzureCS object.
func New(subscriptionKey string, region Region) (*AzureCS, func(), error) {
	return NewWithClient(http.DefaultClient, subscriptionKey, region)
}

// NewWithClient returns an AzureCS object with a custom http client.
func NewWithClient(client *http.Client, subscriptionKey string, region Region) (*AzureCS, func(), error) {
	az := &AzureCS{
		subscriptionKey: subscriptionKey,
		region:          region,
		httpClient:      client,
	}

	az.tokenRefreshURL = fmt.Sprintf(tokenRefreshAPI, region)

	// api requires that the token is refreshed every 10 mintutes.
	// We will do this task in the background every ~9 minutes.
	if err := az.refreshToken(); err != nil {
		return nil, nil, fmt.Errorf("failed to fetch initial token, %v", err)
	}

	az.tokenRefreshDoneCh = az.startRefresher()
	cleanup := func() {
		close(az.tokenRefreshDoneCh)
	}
	return az, cleanup, nil
}

// NewTTS returns a new TTS client for the AzureCS object. This is used to create a new TTS client
func (az *AzureCS) NewTTS() (*AzureCSTTS, error) {
	base := fmt.Sprintf(textToSpeechAPI, az.region)
	tts := &AzureCSTTS{
		textToSpeechURL:     base + "/v1",
		voiceServiceListURL: base + "/voices/list",
		client:              az,
	}
	m, err := tts.buildVoiceToRegionMap()
	if err != nil {
		return nil, fmt.Errorf("failed to build voice to region map, %v", err)
	}
	tts.regionVoiceMap = m
	return tts, nil
}

func (az *AzureCS) NewSTT() (*AzureCSSTT, error) {
	return &AzureCSSTT{
		speechToTextAPI: fmt.Sprintf(speechToTextAPI, az.region),
		client:          az,
	}, nil
}

// refreshToken fetches an updated token from the Azure cognitive speech/text services, or an error if unable to retrive.
// Each token is valid for a maximum of 10 minutes. Details for auth tokens are referenced at
// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-apis#authentication .
// Note: This does not need to be called by a client, since this automatically runs via a background go-routine (`startRefresher`)
func (az *AzureCS) refreshToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), tokenRefreshTimeout)
	defer cancel()

	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, az.tokenRefreshURL, nil)
	request.Header.Set("Ocp-Apim-Subscription-Key", az.subscriptionKey)

	res, err := az.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code; received http status=%s", res.Status)
	}

	body, _ := io.ReadAll(res.Body)
	az.accessToken = string(body)
	return nil
}

// startRefresher updates the authentication token on at a 9 minute interval. A channel is returned
// if the caller wishes to cancel the channel.
func (az *AzureCS) startRefresher() chan bool {
	done := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(tokenRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := az.refreshToken()
				if err != nil {
					log.Printf("failed to refresh token, %v", err)
				}
			case <-done:
				return
			}
		}
	}()
	return done
}
