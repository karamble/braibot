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
		Name:        "elevenlabs/speech-to-text/scribe-v2",
		Description: "Scribe V2 - Fast, accurate speech-to-text transcription with word-level timestamps",
		Type:        "audio2text",
		Endpoint:    "/elevenlabs/speech-to-text/scribe-v2",
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
