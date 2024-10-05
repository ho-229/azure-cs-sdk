package ssml

import "encoding/xml"

type Speak struct {
	XMLName    xml.Name  `xml:"speak"`
	Version    string    `xml:"version,attr"`
	XMLNS      string    `xml:"xmlns,attr"`
	XMLNSMSTTS string    `xml:"xmlns:mstts,attr"`
	Lang       string    `xml:"xml:lang,attr"`
	Child      xml.Token `xml:",innerxml"`
}

func NewSpeak() Speak {
	return Speak{
		Version:    "1.0",
		XMLNS:      "http://www.w3.org/2001/10/synthesis",
		XMLNSMSTTS: "http://www.w3.org/2001/mstts",
		Lang:       "en-US",
	}
}

type VoiceEffect string

const (
	// Optimize the auditory experience when providing high-fidelity speech in cars, buses, and other enclosed automobiles.
	VoiceEffectCar VoiceEffect = "eq_car"
	// Optimize the auditory experience for narrowband speech in telecom or telephone scenarios. You should use a sampling rate of 8 kHz.
	// If the sample rate isn't 8 kHz, the auditory quality of the output speech isn't optimized.
	VoiceEffectTelecom VoiceEffect = "eq_telecomhp8k"
)

type Voice struct {
	XMLName xml.Name    `xml:"voice"`
	Name    string      `xml:"name,attr"`
	Effect  VoiceEffect `xml:"effect,attr,omitempty"`
	Child   xml.Token   `xml:",innerxml"`
}

func NewVoice(name string) Voice {
	return Voice{
		Name: name,
	}
}

type ExpressAs struct {
	XMLName     xml.Name  `xml:"mstts:express-as"`
	Role        string    `xml:"role,attr,omitempty"`
	Style       string    `xml:"style,attr"`
	StyleDegree string    `xml:"styledegree,attr,omitempty"`
	Child       xml.Token `xml:",innerxml"`
}

func NewExpressAs(style string) ExpressAs {
	return ExpressAs{
		Style: style,
	}
}

type Lang struct {
	XMLName xml.Name `xml:"lang"`
	Lang    string   `xml:"xml:lang,attr"`
	Text    string   `xml:",chardata"`
}

func NewLang(lang, text string) Lang {
	return Lang{
		Lang: lang,
		Text: text,
	}
}

type Prosody struct {
	XMLName xml.Name  `xml:"prosody"`
	Contour string    `xml:"contour,attr,omitempty"`
	Pitch   string    `xml:"pitch,attr,omitempty"`
	Rate    string    `xml:"rate,attr,omitempty"`
	Range   string    `xml:"range,attr,omitempty"`
	Child   xml.Token `xml:",innerxml"`
}

type EmphasisLevel string

const (
	EmphasisLevelReduced  EmphasisLevel = "reduced"
	EmphasisLevelNone     EmphasisLevel = "none"
	EmphasisLevelModerate EmphasisLevel = "moderate"
	EmphasisLevelStrong   EmphasisLevel = "strong"
)

type Emphasis struct {
	XMLName xml.Name      `xml:"emphasis"`
	Level   EmphasisLevel `xml:"level,attr,omitempty"`
	Child   xml.Token     `xml:",innerxml"`
}
