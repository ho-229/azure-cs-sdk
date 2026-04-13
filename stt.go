package azure_cs_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AzureCSSTT struct {
	speechToTextAPI   string
	speechToTextWSAPI string
	client            *AzureCS
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
	Confidence        float64
	Offset            uint64
	Duration          uint64
	PrimaryLanguage   *PrimaryLanguage
	NBest             []RecognizeAlternative
}

type PrimaryLanguage struct {
	Language   string
	Confidence string
}

func (r *RecognizeSimpleResponse) normalize() {}

type recognizeDetailedResponse struct {
	RecognitionStatus RecognitionStatus
	Offset            uint64
	Duration          uint64
	NBest             []RecognizeAlternative
	PrimaryLanguage   *PrimaryLanguage
}

type RecognizeAlternative struct {
	Confidence float64
	Display    string
}

func (r *recognizeDetailedResponse) normalize() {}

func (r *recognizeDetailedResponse) toSimpleResponse() *RecognizeSimpleResponse {
	maxIndex := -1
	maxConfidence := .0
	for i := range r.NBest {
		if r.NBest[i].Confidence > maxConfidence {
			maxConfidence = r.NBest[i].Confidence
			maxIndex = i
		}
	}

	resp := &RecognizeSimpleResponse{
		RecognitionStatus: r.RecognitionStatus,
		Offset:            r.Offset,
		Duration:          r.Duration,
		PrimaryLanguage:   r.PrimaryLanguage,
		NBest:             r.NBest,
	}
	if maxIndex != -1 {
		resp.DisplayText = r.NBest[maxIndex].Display
		resp.Confidence = r.NBest[maxIndex].Confidence
	}
	return resp
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
	return az.RecognizeShortSimpleWithContext(
		context.Background(),
		reader,
		audioType,
		language,
		opts...,
	)
}

func (az *AzureCSSTT) RecognizeShortSimpleWithContext(
	ctx context.Context,
	reader io.Reader,
	audioType AudioType,
	language string,
	opts ...Option,
) (*RecognizeSimpleResponse, error) {
	req, err := az.newRecognizeShortRequest(ctx, reader, audioType, language, "simple", opts...)
	if err != nil {
		return nil, err
	}

	return doAndUnmarshal[RecognizeSimpleResponse](az.client.httpClient, req)
}

func (az *AzureCSSTT) newRecognizeShortRequest(
	ctx context.Context,
	reader io.Reader,
	audioType AudioType,
	language string,
	format string,
	opts ...Option,
) (*http.Request, error) {
	if audioType != RIFF16khz16bitMonoPCM && audioType != RAW16khz16bitMonoPCM && audioType != OGG16khz16bitMonoOpus {
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

	query := fmt.Sprintf("?language=%s&format=%s&profanity=%s&cid=%s",
		url.QueryEscape(language),
		url.QueryEscape(format),
		params.Profanity,
		url.QueryEscape(params.Cid),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, az.speechToTextAPI+query, reader)
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
	return req, nil
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
	if normalizer, ok := any(v).(interface{ normalize() }); ok {
		normalizer.normalize()
	}
	return v, nil
}
