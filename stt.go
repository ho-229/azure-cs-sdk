package azure_cs_sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AzureCSSTT struct {
	speechToTextAPI string
	client          *AzureCS
}

type RecognitionStatus string

const (
	RecognitionStatusSuccess               RecognitionStatus = "Success"
	RecognitionStatusNoMatch               RecognitionStatus = "NoMatch"
	RecognitionStatusInitialSilenceTimeout RecognitionStatus = "InitialSilenceTimeout"
	RecognitionStatusBabbleTimeout         RecognitionStatus = "BabbleTimeout"
	RecognitionStatusError                 RecognitionStatus = "Error"
)

type RecognizeSimpleResponse struct {
	RecognitionStatus RecognitionStatus
	DisplayText       string
	Offset            uint64
	Duration          uint64
}

type Option func(*options)

type Profanity string

const (
	ProfanityMasked  Profanity = "masked"
	ProfanityRemoved Profanity = "removed"
	ProfanityRaw     Profanity = "raw"
)

type options struct {
	Profanity Profanity
	Cid       string
	Expect    uint32
}

func WithProfanity(p Profanity) Option {
	return func(o *options) {
		o.Profanity = p
	}
}

func WithCid(cid string) Option {
	return func(o *options) {
		o.Cid = cid
	}
}

func WithExpect(expect uint32) Option {
	return func(o *options) {
		o.Expect = expect
	}
}

func (az *AzureCSSTT) RecognizeShortSimple(
	reader io.Reader,
	audioType AudioType,
	language string,
	opts ...Option,
) (*RecognizeSimpleResponse, error) {
	if audioType != RIFF16khz16bitMonoPCM && audioType != OGG16khz16bitMonoOpus {
		return nil, fmt.Errorf("audio type %s is not supported", audioType)
	}

	params := options{
		Profanity: ProfanityMasked,
		Cid:       "",
		Expect:    100,
	}
	for _, opt := range opts {
		opt(&params)
	}

	query := fmt.Sprintf("?language=%s&format=simple&profanity=%s&cid=%s",
		url.QueryEscape(language),
		params.Profanity,
		url.QueryEscape(params.Cid),
	)
	req, err := http.NewRequest(http.MethodPost, az.speechToTextAPI+query, reader)
	if err != nil {
		return nil, err
	}
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Authorization", "Bearer "+az.client.accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Expect", fmt.Sprintf("%d-continue", params.Expect))
	switch audioType {
	case RIFF16khz16bitMonoPCM:
		req.Header.Set("Content-Type", "audio/wav; codecs=\"audio/pcm\"; samplerate=16000")
	case OGG16khz16bitMonoOpus:
		req.Header.Set("Content-Type", "audio/ogg; codecs=\"opus\"")
	}

	return doAndUnmarshal[RecognizeSimpleResponse](az.client.httpClient, req)
}

func doAndUnmarshal[T any](client *http.Client, req *http.Request) (*T, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	v := new(T)
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return v, nil
}
