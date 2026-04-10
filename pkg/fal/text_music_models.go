// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- minimax-music-v2 ---

type minimaxMusicV2Model struct{}

func (m *minimaxMusicV2Model) Define() Model {
	defaultOpts := &MinimaxMusicV2Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:             "minimax-music-v2",
		Description:      "Minimax Music V2 - Advanced AI music composition",
		PriceUSD:         0.01, // Per second
		Type:             "text2music",
		Endpoint:         "/minimax-music/v2",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2music [prompt] [options]\n\n💰 **Price: $0.01 per second of music\n\nParameters:\n• prompt: Description of the music (required)\n• --duration: Duration in seconds 1-300 (default: 60)\n• --reference_audio_url: URL of reference audio (optional)",
		Options: &MinimaxMusicV2Options{
			Duration: defaults["duration"].(int),
		},
	}
}

// --- stable-audio-25 ---

type stableAudio25Model struct{}

func (m *stableAudio25Model) Define() Model {
	defaultOpts := &StableAudio25Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "stable-audio-25",
		Description: "Stable Audio 2.5 - High-quality audio generation from text",
		PriceUSD:    0.02, // Per second
		Type:        "text2music",
		Endpoint:    "/stable-audio-25/text-to-audio",
		PerSecondPricing: true,
		HelpDoc:     "Usage: !text2music [prompt] [options]\n\n💰 **Price: $0.02 per second of audio\n\nParameters:\n• prompt: Description of the audio (required)\n• --duration: Duration in seconds 1-180 (default: 30)\n• --sample_rate: Sample rate (default: 44100)\n• --output_format: wav, mp3, ogg (default: wav)\n• --seed: Specific seed (optional)",
		Options: &StableAudio25Options{
			Duration:     defaults["duration"].(float64),
			SampleRate:   defaults["sample_rate"].(int),
			OutputFormat: defaults["output_format"].(string),
		},
	}
}

func init() {
	registerModel(&minimaxMusicV2Model{})
	registerModel(&stableAudio25Model{})
}
