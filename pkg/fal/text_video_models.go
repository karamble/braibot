// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- kling-video-text ---

type klingVideoTextModel struct{}

func (m *klingVideoTextModel) Define() Model {
	return Model{
		Name:        "kling-video-text",
		Description: "Generate videos from text using Kling 2.0 Master. Base price: $2.0 for 5 seconds, $0.4 per additional second",
		PriceUSD:    2.0,
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\nExample: !text2video a beautiful animation --duration 10 --aspect 16:9\n\nParameters:\n• prompt: Description of the desired video\n• --duration: Video duration in seconds (default: 5, min: 5)\n• --aspect: Aspect ratio (default: 16:9)\n• --negative-prompt: Text describing what to avoid (default: blur, distort, and low quality)\n• --cfg-scale: Configuration scale (default: 0.5)\n\nPricing:\n• Base price: $2.0 for 5 seconds\n• Additional cost: $0.4 per second beyond 5 seconds",
		Options: &KlingVideoOptions{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
		},
	}
}

func init() {
	registerModel(&klingVideoTextModel{})
}
