package azuretexttospeech

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// The following are V1 endpoints for Cognitiveservices endpoints
const textToSpeechAPI = "https://%s.tts.speech.microsoft.com/cognitiveservices/v1"
const tokenRefreshAPI = "https://%s.api.cognitive.microsoft.com/sts/v1.0/issueToken"

// synthesizeActionTimeout is the amount of time the http client will wait for a response during Synthesize request
const synthesizeActionTimeout = time.Second * 30

// tokenRefreshTimeout is the amount of time the http client will wait during the token refresh action.
const tokenRefreshTimeout = time.Second * 10

// tokenRefreshInterval is the amount of time between token refreshes
const tokenRefreshInterval = time.Hour

// TTSApiXMLPayload templates the payload required for API.
// See: https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech#sample-request
// Example:
// <speak version='1.0' xml:lang='en-US'><voice xml:lang='en-US' xml:gender='Male'
//    name='en-US-ChristopherNeural'>
//        Microsoft Speech Service Text-to-Speech API
// </voice></speak>

const ttsApiXMLPayload = "<speak version='1.0' xml:lang='%s'><voice xml:lang='%s' xml:gender='%s' name='%s'>%s</voice></speak>"

// const ttsApiXMLPayload2 = "<speak xmlns='http://www.w3.org/2001/10/synthesis' xmlns:mstts='http://www.w3.org/2001/mstts' xmlns:emo='http://www.w3.org/2009/10/emotionml' version='1.0' xml:lang='%s'><voice name='%s'>%s</voice></speak>"

func (az *AzureCSTextToSpeech) GetVoicesMap() RegionVoiceMap {
	return az.regionVoiceMap
}

// SynthesizeWithContext returns a bytestream of the rendered text-to-speech in the target audio format. `speechText` is the string of
// text in which a user wishes to Synthesize, `region` is the language/locale, `gender` is the desired output voice
// and `audioOutput` captures the audio format.
func (az *AzureCSTextToSpeech) SynthesizeWithContext(ctx context.Context, speechText string, voicesName string, audioOutput AudioOutput) ([]byte, error) {
	vmap, ok := az.regionVoiceMap[voicesName]
	if !ok {
		return nil, fmt.Errorf("unable to locate RegionVoiceMap{voice.ShortName=%s}", voicesName)
	}

	v := voiceXML(speechText, voicesName, vmap.Locale, vmap.Gender)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, az.textToSpeechURL, bytes.NewBufferString(v))
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Microsoft-OutputFormat", fmt.Sprint(audioOutput))
	request.Header.Set("Content-Type", "application/ssml+xml")
	request.Header.Set("Authorization", "Bearer "+az.accessToken)
	request.Header.Set("User-Agent", "azuretts")

	response, err := az.client.Do(request.WithContext(ctx))
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

// Synthesize directs to SynthesizeWithContext. A new context.Withtimeout is created with the timeout as defined by synthesizeActionTimeout
func (az *AzureCSTextToSpeech) Synthesize(speechText string, voicesname string, audioOutput AudioOutput) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), synthesizeActionTimeout)
	defer cancel()
	return az.SynthesizeWithContext(ctx, speechText, voicesname, audioOutput)
}

// voiceXML renders the XML payload for the TTS api.
// For API reference see https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech#sample-request
func voiceXML(speechText, description string, locale string, gender Gender) string {
	return fmt.Sprintf(ttsApiXMLPayload, locale, locale, gender, description, speechText)
}

// refreshToken fetches an updated token from the Azure cognitive speech/text services, or an error if unable to retrive.
// Each token is valid for a maximum of 10 minutes. Details for auth tokens are referenced at
// https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-apis#authentication .
// Note: This does not need to be called by a client, since this automatically runs via a background go-routine (`startRefresher`)
func (az *AzureCSTextToSpeech) refreshToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), tokenRefreshTimeout)
	defer cancel()

	request, _ := http.NewRequestWithContext(ctx, http.MethodPost, az.tokenRefreshURL, nil)
	request.Header.Set("Ocp-Apim-Subscription-Key", az.subscriptionKey)

	res, err := az.client.Do(request)
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
func (az *AzureCSTextToSpeech) startRefresher() chan bool {
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

// AzureCSTextToSpeech stores configuration and state information for the TTS client.
type AzureCSTextToSpeech struct {
	accessToken         string // is the auth token received from `TokenRefreshAPI`. Used in the Authorization: Bearer header.
	regionVoiceMap      RegionVoiceMap
	subscriptionKey     string    // API key for Azure's Congnitive Speech services
	tokenRefreshDoneCh  chan bool // channel to stop the token refresh goroutine.
	tokenRefreshURL     string
	voiceServiceListURL string
	textToSpeechURL     string
	client              *http.Client
}

// New returns an AzureCSTextToSpeech object.
func New(subscriptionKey string, region Region) (*AzureCSTextToSpeech, func(), error) {
	return NewWithClient(http.DefaultClient, subscriptionKey, region)
}

func NewWithClient(client *http.Client, subscriptionKey string, region Region) (*AzureCSTextToSpeech, func(), error) {
	az := &AzureCSTextToSpeech{
		subscriptionKey: subscriptionKey,
		client:          client,
	}

	az.textToSpeechURL = fmt.Sprintf(textToSpeechAPI, region)
	az.tokenRefreshURL = fmt.Sprintf(tokenRefreshAPI, region)
	az.voiceServiceListURL = fmt.Sprintf(voiceListAPI, region)

	// api requires that the token is refreshed every 10 mintutes.
	// We will do this task in the background every ~9 minutes.
	if err := az.refreshToken(); err != nil {
		return nil, nil, fmt.Errorf("failed to fetch initial token, %v", err)
	}

	m, err := az.buildVoiceToRegionMap()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to fetch voice-map, %v", err)
	}

	az.regionVoiceMap = m

	az.tokenRefreshDoneCh = az.startRefresher()
	cleanup := func() {
		close(az.tokenRefreshDoneCh)
	}
	return az, cleanup, nil
}
