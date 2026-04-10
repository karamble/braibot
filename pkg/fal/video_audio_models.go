// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- mmaudio-v2 ---

type mmaudioV2Model struct{}

func (m *mmaudioV2Model) Define() Model {
	defaultOpts := &MMAudioV2Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "mmaudio-v2",
		Description: "MMAudio V2 - Generate synchronized audio from video and/or text",
		PriceUSD:    0.20,
		Type:        "video2audio",
		Endpoint:    "/mmaudio-v2",
		HelpDoc:     "Usage: !video2audio [video_url] [prompt] [options]\n\n💰 **Price: $0.20 per video\n\nParameters:\n• video_url: URL of the source video\n• prompt: Description of the desired audio (optional)\n• --duration: Output duration in seconds (default: video duration)\n• --num_inference_steps: Number of steps (default: 25)\n• --seed: Specific seed (optional)",
		Options: &MMAudioV2Options{
			NumInferenceSteps: defaults["num_inference_steps"].(int),
		},
	}
}

func init() {
	registerModel(&mmaudioV2Model{})
}
