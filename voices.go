package azure_cs_sdk

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
