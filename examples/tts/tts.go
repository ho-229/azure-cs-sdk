package main

import (
	"context"
	"fmt"
	"os"

	azure "github.com/ho-229/azure-cs-sdk"
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
	var region azure.Region
	if apiKey = os.Getenv("AZUREKEY"); apiKey == "" {
		exit(fmt.Errorf("please set your AZUREKEY environment variable"))
	}
	var err error
	if region, err = azure.RegionString(os.Getenv("AZUREREGION")); err != nil {
		exit(fmt.Errorf("please set your AZUREREGION environment variable"))
	}

	az, cleanup, err := azure.New(apiKey, region)
	if err != nil {
		exit(fmt.Errorf("failed to create new client, received %v", err))
	}
	defer cleanup()
	tts, err := az.NewTTS()
	if err != nil {
		exit(fmt.Errorf("failed to create new TTS client, received %v", err))
	}

	// Digitize a text string using the enUS locale, female voice and specify the
	// audio format of a 16Khz, 32kbit mp3 file.
	ctx := context.Background()
	b, err := tts.SynthesizeWithContext(
		ctx,
		"64 BASIC BYTES FREE. READY.",
		"en-US-JennyNeural",
		azure.AUDIO16khz32kbitrateMonoMP3)

	if err != nil {
		exit(fmt.Errorf("unable to synthesize, received: %v", err))
	}

	// send results to disk.
	err = os.WriteFile("audio.mp3", b, 0644)
	if err != nil {
		exit(fmt.Errorf("unable to write file, received %v", err))
	}
}
