package azuretexttospeech

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// voiceListAPI is the source for supported voice list to region mapping
// See: https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech#regions-and-endpoints
const voiceListAPI = "https://%s.tts.speech.microsoft.com/cognitiveservices/voices/list"

//go:generate enumer -type=VoiceType -linecomment -json
type VoiceType int

const (
	VoiceStandard VoiceType = iota // Standard
	VoiceNeural                    // Neural
	VoiceNeuralHD                  // NeuralHD
	VoiceNeutral                   // Neutral
)

/*

{
    "Name": "Microsoft Server Speech Text to Speech Voice (it-IT, ElsaNeural)",
    "DisplayName": "Elsa",
    "LocalName": "Elsa",
    "ShortName": "it-IT-ElsaNeural",
    "Gender": "Female",
    "Locale": "it-IT",
    "LocaleName": "Italian (Italy)",
    "SampleRateHertz": "48000",
    "VoiceType": "Neural",
    "Status": "GA",
    "WordsPerMinute": "148"
  },

*/

type RegionVoice struct {
	Name                string    `json:"Name"`
	DisplayName         string    `json:"DisplayName"`
	LocalName           string    `json:"LocalName"`
	ShortName           string    `json:"ShortName"`
	Gender              Gender    `json:"Gender"`
	Locale              string    `json:"Locale"`
	SampleRateHertz     string    `json:"SampleRateHertz"`
	SecondaryLocaleList []string  `json:"SecondaryLocaleList"`
	VoiceType           VoiceType `json:"VoiceType"`
	RolePlayList        []string  `json:"RolePlayList"`
	WordsPerMinute      string    `json:"WordsPerMinute"`
}

type RegionVoiceMap map[string]RegionVoice

func (az *AzureCSTextToSpeech) buildVoiceToRegionMap() (RegionVoiceMap, error) {

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

func (az *AzureCSTextToSpeech) fetchVoiceList() ([]RegionVoice, error) {
	req, _ := http.NewRequest(http.MethodGet, az.voiceServiceListURL, nil)
	req.Header.Set("Authorization", "Bearer "+az.accessToken)

	// Perform the request
	res, err := az.client.Do(req)
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
