// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- minimax-tts/text-to-speech ---

type minimaxTTSModel struct{}

func (m *minimaxTTSModel) Define() Model {
	// Define default options
	defaultOpts := &MinimaxTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()
	// Extract potentially nil defaults safely
	var defaultSpeed *float64
	if v, ok := defaults["speed"].(*float64); ok {
		defaultSpeed = v
	}
	var defaultVol *float64
	if v, ok := defaults["vol"].(*float64); ok {
		defaultVol = v
	}

	return Model{
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters",
		Type:        "text2speech",
		Endpoint:    "/minimax-tts/text-to-speech",
		Options: &MinimaxTTSOptions{
			Speed:      defaultSpeed,
			Vol:        defaultVol,
			SampleRate: defaults["sample_rate"].(string),
			Bitrate:    defaults["bitrate"].(string),
			Format:     defaults["format"].(string),
			Channel:    defaults["channel"].(string),
			// Pitch and Emotion default to nil/""
		},
	}
}

// --- chatterbox-tts ---

type chatterboxTTSModel struct{}

func (m *chatterboxTTSModel) Define() Model {
	defaultOpts := &ChatterboxTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "chatterbox-tts",
		Description: "Chatterbox TTS Turbo - Fast, natural text-to-speech",
		Type:        "text2speech",
		Endpoint:    "/chatterbox/text-to-speech/turbo",
		Options: &ChatterboxTTSOptions{
			Exaggeration: defaults["exaggeration"].(float64),
			CFGWeight:    defaults["cfg_weight"].(float64),
		},
	}
}

// --- elevenlabs-dialog ---

type elevenlabsDialogModel struct{}

func (m *elevenlabsDialogModel) Define() Model {
	defaultOpts := &ElevenLabsDialogOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "elevenlabs-dialog",
		Description: "ElevenLabs Text-to-Dialogue V3 - Multi-speaker dialogue generation",
		Type:        "text2speech",
		Endpoint:    "/elevenlabs/text-to-dialogue/eleven-v3",
		Options: &ElevenLabsDialogOptions{
			VoiceID:         defaults["voice_id"].(string),
			OutputFormat:    defaults["output_format"].(string),
			Stability:       defaults["stability"].(float64),
			SimilarityBoost: defaults["similarity_boost"].(float64),
		},
	}
}

// --- elevenlabs-tts-turbo ---

type elevenlabsTTSTurboModel struct{}

func (m *elevenlabsTTSTurboModel) Define() Model {
	defaultOpts := &ElevenLabsTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()

	var defaultStability *float64
	if v, ok := defaults["stability"].(*float64); ok {
		defaultStability = v
	}
	var defaultSimilarity *float64
	if v, ok := defaults["similarity_boost"].(*float64); ok {
		defaultSimilarity = v
	}
	var defaultStyle *float64
	if v, ok := defaults["style"].(*float64); ok {
		defaultStyle = v
	}
	var defaultSpeed *float64
	if v, ok := defaults["speed"].(*float64); ok {
		defaultSpeed = v
	}
	var defaultTimestamps *bool
	if v, ok := defaults["timestamps"].(*bool); ok {
		defaultTimestamps = v
	}

	return Model{
		Name:        "elevenlabs/tts/turbo-v2.5",
		Description: "ElevenLabs TTS Turbo v2.5 - High-quality, low-latency text-to-speech",
		Type:        "text2speech",
		Endpoint:    "/elevenlabs/tts/turbo-v2.5",
		Options: &ElevenLabsTTSOptions{
			Voice:           defaults["voice"].(string),
			Stability:       defaultStability,
			SimilarityBoost: defaultSimilarity,
			Style:           defaultStyle,
			Speed:           defaultSpeed,
			Timestamps:      defaultTimestamps,
		},
	}
}

func init() {
	registerModel(&minimaxTTSModel{})
	registerModel(&chatterboxTTSModel{})
	registerModel(&elevenlabsDialogModel{})
	registerModel(&elevenlabsTTSTurboModel{})
}
