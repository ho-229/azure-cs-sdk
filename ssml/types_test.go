package ssml_test

import (
	"encoding/xml"
	"testing"

	"github.com/ho-229/azuretexttospeech/ssml"
	"github.com/stretchr/testify/assert"
)

func Test_speak(t *testing.T) {
	b, err := xml.Marshal(ssml.NewSpeak())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" lang="en-US"></speak>`, string(b))
}

func Test_multipleExpressAs(t *testing.T) {
	speak := ssml.NewSpeak()
	speak.Child = []xml.Token{
		ssml.Voice{
			Name: "en-US-JennyNeural",
			Child: []any{
				ssml.ExpressAs{
					Style: "normal",
					Child: "hello",
				},
				ssml.ExpressAs{
					Style: "normal",
					Child: "world",
				},
			},
		},
	}
	b, err := xml.Marshal(speak)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="http://www.w3.org/2001/mstts" xml:lang="en-US"><voice name="en-US-JennyNeural"><mstts:express-as style="normal">hello</mstts:express-as><mstts:express-as style="normal">world</mstts:express-as></voice></speak>`, string(b))
}
