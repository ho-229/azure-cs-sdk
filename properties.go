package azuretexttospeech

// AudioOutput types represent the supported audio encoding formats for the text-to-speech endpoint.
// This type is required when requesting to azuretexttospeech.Synthesize text-to-speed request.
// Each incorporates a bitrate and encoding type. The Speech service supports 24 kHz, 16 kHz, and 8 kHz audio outputs.
// See: https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-text-to-speech#audio-outputs

type AudioOutput int

const (
	RAW16khz16bitMonoPCM AudioOutput = iota
	RAW24khz16bitMonoPCM
	RAW48khz16bitMonoPCM
	RAW8khz8bitMonoMulaw
	RAW8khz8bitMonoAlaw
	AUDIO16khz32kbitrateMonoMP3
	AUDIO16khz128kbitrateMonoMP3
	AUDIO24khz96kbitrateMonoMP3
	AUDIO48khz96kbitrateMonoMP3
	RAW16khz16bitMonoTruesilk
	WEBM16khz16bitMonoOpus
	OGG16khz16bitMonoOpus
	OGG48khz16bitMonoOpus
	RIFF16khz16bitMonoPCM
	RIFF24khz16bitMonoPCM
	RIFF48khz16bitMonoPCM
	RIFF8khz8bitMonoMulaw
	RIFF8khz8bitMonoAlaw
	AUDIO16khz64kbitrateMonoMP3
	AUDIO24khz48kbitrateMonoMP3
	AUDIO24khz160kbitrateMonoMP3
	AUDIO48khz192kbitrateMonoMP3
	RAW24khz16bitMonoTruesilk
	WEBM24khz16bitMonoOpus
	OGG24khz16bitMonoOpus
)

var MapAudioFileExtensions = map[string]string{
	"RAW16khz16bitMonoPCM":         "raw",
	"RAW24khz16bitMonoPCM":         "raw",
	"RAW48khz16bitMonoPCM":         "raw",
	"RAW8khz8bitMonoMulaw":         "raw",
	"RAW8khz8bitMonoAlaw":          "raw",
	"AUDIO16khz32kbitrateMonoMP3":  "mp3",
	"AUDIO16khz128kbitrateMonoMP3": "mp3",
	"AUDIO24khz96kbitrateMonoMP3":  "mp3",
	"AUDIO48khz96kbitrateMonoMP3":  "mp3",
	"RAW16khz16bitMonoTruesilk":    "raw",
	"WEBM16khz16bitMonoOpus":       "opus",
	"OGG16khz16bitMonoOpus":        "opus",
	"OGG48khz16bitMonoOpus":        "opus",
	"RIFF16khz16bitMonoPCM":        "wav",
	"RIFF24khz16bitMonoPCM":        "wav",
	"RIFF48khz16bitMonoPCM":        "wav",
	"RIFF8khz8bitMonoMulaw":        "wav",
	"RIFF8khz8bitMonoAlaw":         "wav",
	"AUDIO16khz64kbitrateMonoMP3":  "mp3",
	"AUDIO24khz48kbitrateMonoMP3":  "mp3",
	"AUDIO24khz160kbitrateMonoMP3": "mp3",
	"AUDIO48khz192kbitrateMonoMP3": "mp3",
	"RAW24khz16bitMonoTruesilk":    "raw",
	"WEBM24khz16bitMonoOpus":       "webm",
	"OGG24khz16bitMonoOpus":        "ogg",
}

// var MapAudioToFormatid = map[string]AudioOutput{
// 	"RAW16khz16bitMonoPCM":         RAW16khz16bitMonoPCM,
// 	"RAW24khz16bitMonoPCM":         RAW24khz16bitMonoPCM,
// 	"RAW48khz16bitMonoPCM":         RAW48khz16bitMonoPCM,
// 	"RAW8khz8bitMonoMulaw":         RAW8khz8bitMonoMulaw,
// 	"RAW8khz8bitMonoAlaw":          RAW8khz8bitMonoAlaw,
// 	"AUDIO16khz32kbitrateMonoMP3":  AUDIO16khz32kbitrateMonoMP3,
// 	"AUDIO16khz128kbitrateMonoMP3": AUDIO16khz128kbitrateMonoMP3,
// 	"AUDIO24khz96kbitrateMonoMP3":  AUDIO24khz96kbitrateMonoMP3,
// 	"AUDIO48khz96kbitrateMonoMP3":  AUDIO48khz96kbitrateMonoMP3,
// 	"RAW16khz16bitMonoTruesilk":    RAW16khz16bitMonoTruesilk,
// 	"WEBM16khz16bitMonoOpus":       WEBM16khz16bitMonoOpus,
// 	"OGG16khz16bitMonoOpus":        OGG16khz16bitMonoOpus,
// 	"OGG48khz16bitMonoOpus":        OGG48khz16bitMonoOpus,
// 	"RIFF16khz16bitMonoPCM":        RIFF16khz16bitMonoPCM,
// 	"RIFF24khz16bitMonoPCM":        RIFF24khz16bitMonoPCM,
// 	"RIFF48khz16bitMonoPCM":        RIFF48khz16bitMonoPCM,
// 	"RIFF8khz8bitMonoMulaw":        RIFF8khz8bitMonoMulaw,
// 	"RIFF8khz8bitMonoAlaw":         RIFF8khz8bitMonoAlaw,
// 	"AUDIO16khz64kbitrateMonoMP3":  AUDIO16khz64kbitrateMonoMP3,
// 	"AUDIO24khz48kbitrateMonoMP3":  AUDIO24khz48kbitrateMonoMP3,
// 	"AUDIO24khz160kbitrateMonoMP3": AUDIO24khz160kbitrateMonoMP3,
// 	"AUDIO48khz192kbitrateMonoMP3": AUDIO48khz192kbitrateMonoMP3,
// 	"RAW24khz16bitMonoTruesilk":    RAW24khz16bitMonoTruesilk,
// 	"WEBM24khz16bitMonoOpus":       WEBM24khz16bitMonoOpus,
// 	"OGG24khz16bitMonoOpus":        OGG24khz16bitMonoOpus,
// }

func (a AudioOutput) String() string {
	return []string{
		"raw-16khz-16bit-mono-pcm",
		"raw-24khz-16bit-mono-pcm",
		"raw-48khz-16bit-mono-pcm",
		"raw-8khz-8bit-mono-mulaw",
		"raw-8khz-8bit-mono-alaw",
		"audio-16khz-32kbitrate-mono-mp3",
		"audio-16khz-128kbitrate-mono-mp3",
		"audio-24khz-96kbitrate-mono-mp3",
		"audio-48khz-96kbitrate-mono-mp3",
		"raw-16khz-16bit-mono-truesilk",
		"webm-16khz-16bit-mono-opus",
		"ogg-16khz-16bit-mono-opus",
		"ogg-48khz-16bit-mono-opus",
		"riff-16khz-16bit-mono-pcm",
		"riff-24khz-16bit-mono-pcm",
		"riff-48khz-16bit-mono-pcm",
		"riff-8khz-8bit-mono-mulaw",
		"riff-8khz-8bit-mono-alaw",
		"audio-16khz-64kbitrate-mono-mp3",
		"audio-24khz-48kbitrate-mono-mp3",
		"audio-24khz-160kbitrate-mono-mp3",
		"audio-48khz-192kbitrate-mono-mp3",
		"raw-24khz-16bit-mono-truesilk",
		"webm-24khz-16bit-mono-opus",
		"ogg-24khz-16bit-mono-opus",
	}[a]
}

// Gender type for the digitized language
//
//go:generate enumer -type=Gender -linecomment -json
type Gender int

const (
	// GenderMale , GenderFemale are the static Gender constants for digitized voices.
	// See Gender in https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support#standard-voices for breakdown
	GenderMale    Gender = iota // Male
	GenderFemale                // Female
	GenderNeutral               // Neutral
)

// Region references the locations of the availability of standard voices.
// See https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/regions#standard-voices
//
//go:generate enumer -type=Region -linecomment -json -trimprefix Region
type Region int

const (
	// Azure regions and their endpoints that support the Text To Speech service.
	RegionAustraliaEast Region = iota
	RegionBrazilSouth
	RegionCanadaCentral
	RegionCentralUS
	RegionEastAsia
	RegionEastUS
	RegionEastUS2
	RegionFranceCentral
	RegionIndiaCentral
	RegionJapanEast
	RegionJapanWest
	RegionKoreaCentral
	RegionNorthCentralUS
	RegionNorthEurope
	RegionSouthCentralUS
	RegionSoutheastAsia
	RegionUKSouth
	RegionWestEurope
	RegionWestUS
	RegionWestUS2
)
