// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- elevenlabs-voice-changer ---

type elevenlabsVoiceChangerModel struct{}

func (m *elevenlabsVoiceChangerModel) Define() Model {
	defaultOpts := &ElevenLabsVoiceChangerOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "elevenlabs-voice-changer",
		Description: "ElevenLabs Voice Changer - Transform audio with different voices",
		Type:        "audio2audio",
		Endpoint:    "/elevenlabs/voice-changer",
		Options: &ElevenLabsVoiceChangerOptions{
			Voice:        defaults["voice"].(string),
			OutputFormat: defaults["output_format"].(string),
		},
	}
}

func init() {
	registerModel(&elevenlabsVoiceChangerModel{})
}
