package main

import (
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

	stt, err := az.NewSTT()
	if err != nil {
		exit(fmt.Errorf("failed to create new STT client, received %v", err))
	}

	sampleFile, err := os.Open("examples/stt_short/audio.wav")
	if err != nil {
		exit(fmt.Errorf("failed to open sample file, received %v", err))
	}
	defer sampleFile.Close()

	events, err := stt.Recognize(
		sampleFile,
		azure.RIFF16khz16bitMonoPCM,
		[]string{"ja-JP", "zh-CN", "ko-KR", "fr-FR"},
	)
	if err != nil {
		exit(fmt.Errorf("failed to start recognize stream, received %v", err))
	}

	for event := range events {
		if event.Err != nil {
			exit(fmt.Errorf("recognize stream failed, received %v", event.Err))
		}
		if event.Result == nil {
			continue
		}

		language := ""
		if event.Result.PrimaryLanguage != nil {
			language = event.Result.PrimaryLanguage.Language
		}

		switch event.Type {
		case azure.RecognizeEventPartial:
			fmt.Printf(
				"PARTIAL language=%s text=%s\n",
				language,
				event.Result.DisplayText,
			)
		case azure.RecognizeEventFinal:
			fmt.Printf(
				"FINAL status=%s language=%s detectionConfidence=%.2f text=%s nbest=%+v\n",
				event.Result.RecognitionStatus,
				language,
				event.Result.Confidence,
				event.Result.DisplayText,
				event.Result.NBest,
			)
			return
		}
	}
}
