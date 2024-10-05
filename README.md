# AzureTextToSpeech Client #

> This is a fork of https://github.com/gmaisto/azuretexttospeech.

![Execute Test Cases](https://github.com/ho-229/azuretexttospeech/workflows/Execute%20Test%20Cases/badge.svg)

This package provides a client for Azure's Cognitive Services (speech services) Text To Speech API. Users of the client
can specify the voice name (e.g. "en-US-AvaMultilingualNeural"), check [supported voice](https://learn.microsoft.com/en-us/azure/ai-services/speech-service/language-support?tabs=tts#prebuilt-neural-voices) for all available voices. The library fetches the audio rendered in the format of your choice (see `AudioOutput` types for supported formats).

API documents of interest

* Text to speech [Azure pricing details](https://azure.microsoft.com/en-us/pricing/details/cognitive-services/speech-services/). Note there is a *free* tier available.
* Text to speech, speech services [API specifications](https://docs.microsoft.com/en-us/azure/cognitive-services/speech-service/rest-apis#text-to-speech-api).

## Requirements ##

A Cognitive Services (kind=Speech Services) API key is required to access the URL. This service can be enabled at the Azure portal.

## Howto ##

The following will synthesize the string `64 BASIC BYTES FREE. READY.`, using the en-US locale, rending with `en-US-JennyNeural`. The output file format is a 16khz 32kbit single channel MP3 audio file.

```golang
import tts "github.com/ho-229/azuretexttospeech"
func main() {
    // See TextToSpeechAPI and TokenRefreshAPI types for list of endpoints and regions.
    azureSpeech, cleanup, _ := tts.New("YOUR-API-KEY", tts.RegionEastUS)
    defer cleanup()
    ctx := context.Background()
    payload, _ := az.SynthesizeWithContext(
        ctx,
        "64 BASIC BYTES FREE. READY.",
        "en-US-JennyNeural",             // voice name
        tts.Audio16khz32kbitrateMonoMp3) // AudioOutput type
    // the response `payload` is your byte array containing audio data.
}
```
