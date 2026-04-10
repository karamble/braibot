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
		Name:        "minimax-music-v2",
		Description: "Minimax Music V2 - Advanced AI music composition",
		Type:        "text2music",
		Endpoint:    "/minimax-music/v2",
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
		Type:        "text2music",
		Endpoint:    "/stable-audio-25/text-to-audio",
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
