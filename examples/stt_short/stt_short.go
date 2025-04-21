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

	resp, err := stt.RecognizeShortSimple(sampleFile, azure.RIFF16khz16bitMonoPCM, "zh-CN")
	if err != nil {
		exit(fmt.Errorf("failed to recognize short simple, received %v", err))
	}
	fmt.Printf("Status: %s Recognized text: %s\n", resp.RecognitionStatus, resp.DisplayText)
}
