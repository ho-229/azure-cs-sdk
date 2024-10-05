package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"

	tts "github.com/ho-229/azuretexttospeech"
	"github.com/ho-229/azuretexttospeech/ssml"
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

	voice := ssml.NewVoice("zh-CN-XiaomoNeural")
	voice.Child = []xml.Token{
		ssml.ExpressAs{
			Style: "calm",
			Role:  "YoungAdultFemale",
			Child: "女儿看见父亲走了进来，问道：您来的挺快的，怎么过来的？父亲放下手提包，说。",
		},
		ssml.ExpressAs{
			Style: "calm",
			Role:  "OlderAdultMale",
			Child: "刚打车过来的，路上还挺顺畅。",
		},
	}

	ctx := context.Background()
	b, err := az.SynthesizeSsmlWithContext(
		ctx,
		voice,
		tts.AUDIO16khz32kbitrateMonoMP3,
	)

	if err != nil {
		exit(fmt.Errorf("unable to synthesize, received: %v", err))
	}

	// send results to disk.
	err = os.WriteFile("audio.mp3", b, 0644)
	if err != nil {
		exit(fmt.Errorf("unable to write file, received %v", err))
	}
}
