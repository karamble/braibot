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

// --- minimax-video-01-director ---

type minimaxDirectorModel struct{}

func (m *minimaxDirectorModel) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01-director",
		Description: "Generate video clips with camera movement instructions.",
		PriceUSD:    0.8, // TODO: Update with actual price
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\nParameters:\n• prompt: Description of the desired video. Include camera movements in square brackets `[]`.\n  - Single movement: `[Push in] A cat walking.`\n  - Combined (up to 3): `[Truck left, Pan right, Zoom in] A busy street scene.`\n  - Available movements: `Truck left/right`, `Pan left/right`, `Push in/Pull out`, `Pedestal up/down`, `Tilt up/down`, `Zoom in/out`, `Shake`, `Tracking shot`, `Static shot`. \n  - More details: https://sixth-switch-2ac.notion.site/T2V-01-Director-Model-Tutorial-with-camera-movement-1886c20a98eb80f395b8e05291ad8645\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)",
		Options: &MinimaxDirectorOptions{
			PromptOptimizer: &defaultOptimizer,
		},
	}
}

func init() {
	registerModel(&minimaxDirectorModel{})
}

// --- minimax/video-01 ---

type minimaxVideo01Model struct{}

func (m *minimaxVideo01Model) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01",
		Description: "Native high-resolution, high-frame-rate video generation model.",
		PriceUSD:    0.8,
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\nExample: !text2video A futuristic cityscape --prompt-optimizer true\n\nParameters:\n• prompt: Description of the desired video.\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)",
		Options: &MinimaxVideo01Options{
			PromptOptimizer: &defaultOptimizer,
		},
	}
}

func init() {
	registerModel(&minimaxVideo01Model{})
}
