// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- elevenlabs/speech-to-text/scribe-v2 ---

type scribeV2Model struct{}

func (m *scribeV2Model) Define() Model {
	defaultOpts := &ScribeV2Options{}
	defaults := defaultOpts.GetDefaultValues()

	var defaultDiarize *bool
	if v, ok := defaults["diarize"].(*bool); ok {
		defaultDiarize = v
	}

	return Model{
		Name:             "elevenlabs/speech-to-text/scribe-v2",
		Description:      "Scribe V2 - Fast, accurate speech-to-text transcription with word-level timestamps",
		PriceUSD:         0.008, // Per minute of audio
		Type:             "audio2text",
		PerSecondPricing: true,
		HelpDoc: `Usage: Transcribe audio to text with word-level timestamps

Price: $0.008 per minute of audio (plus 30% for diarization)

Parameters:
- audio_url: URL to audio file (required)
- task: transcribe (default) or translate
- language: ISO 639-1 code (auto-detected if not specified)
- chunk_level: segment (default) or word
- diarize: Enable speaker diarization (default: true)
- num_speakers: Number of speakers (optional, 1-50)

Supported formats: mp3, wav, m4a, ogg, flac, webm`,
		Options: &ScribeV2Options{
			Task:       defaults["task"].(string),
			ChunkLevel: defaults["chunk_level"].(string),
			Diarize:    defaultDiarize,
		},
	}
}

func init() {
	registerModel(&scribeV2Model{})
}
