package main

import (
	"context"
	"fmt"
	"os"

	tts "github.com/gmaisto/azuretexttospeech"
)

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
func main() {
	// create a key for "Cognitive Services" (kind=SpeechServices). Once the key is available
	// in the Azure portal, push it into an environment variable (export AZUREKEY=SYS64738).
	// By default the free tier keys are served out of West US2
	var apiKey string
	var region tts.Region
	if apiKey = os.Getenv("AZUREKEY"); apiKey == "" {
		exit(fmt.Errorf("please set your AZUREKEY environment variable"))
	}
	var err error
	if region, err = tts.RegionString(os.Getenv("AZUREREGION")); err != nil {
		exit(fmt.Errorf("please set your AZUREREGION environment variable"))
	}

	az, cleanup, err := tts.New(apiKey, region)
	if err != nil {
		exit(fmt.Errorf("failed to create new client, received %v", err))
	}
	defer cleanup()

	// Digitize a text string using the enUS locale, female voice and specify the
	// audio format of a 16Khz, 32kbit mp3 file.
	ctx := context.Background()
	b, err := az.SynthesizeWithContext(
		ctx,
		"64 BASIC BYTES FREE. READY.",
		"en-US-JennyNeural",
		tts.AUDIO16khz32kbitrateMonoMP3)

	if err != nil {
		exit(fmt.Errorf("unable to synthesize, received: %v", err))
	}

	// send results to disk.
	err = os.WriteFile("audio.mp3", b, 0644)
	if err != nil {
		exit(fmt.Errorf("unable to write file, received %v", err))
	}
}
