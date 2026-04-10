// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"fmt"
)

// --- kling-video-text ---

type klingVideoTextModel struct{}

func (m *klingVideoTextModel) Define() Model {
	return Model{
		Name:        "kling-video-text",
		Description: "Generate videos from text using Kling 2.0 Master.",
		PriceUSD:    0.4,
		Type:        "text2video",
		Endpoint:    "/kling-video/v2/master/text-to-video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.40 per video.",
		Options: &KlingVideoOptions{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
		},
		PerSecondPricing: true,
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
		PriceUSD:    0.8, // Per second
		Type:        "text2video",
		Endpoint:    "/minimax/video-01-director",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.80 per video.",
		Options: &MinimaxDirectorOptions{
			PromptOptimizer: &defaultOptimizer,
		},
		PerSecondPricing: false,
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
		PriceUSD:    0.8, // Per second
		Type:        "text2video",
		Endpoint:    "/minimax/video-01",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.80 per video",
		Options: &MinimaxVideo01Options{
			PromptOptimizer: &defaultOptimizer,
		},
		PerSecondPricing: false,
	}
}

func init() {
	registerModel(&minimaxVideo01Model{})
}

// MiniMax Hailuo-02 Text To Video Model

type minimaxHailuo02Model struct{}

type MinimaxHailuo02Options struct {
	Duration        string `json:"duration,omitempty"` // "6" or "10"
	PromptOptimizer *bool  `json:"prompt_optimizer,omitempty"`
}

type VideoOptions = MinimaxHailuo02Options

func (o *MinimaxHailuo02Options) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"duration":         "6",
		"prompt_optimizer": &defaultOptimizer,
	}
}

func (o *MinimaxHailuo02Options) Validate() error {
	if o.Duration != "" && o.Duration != "6" && o.Duration != "10" {
		return fmt.Errorf("invalid duration: %s (must be 6 or 10)", o.Duration)
	}
	return nil
}

func (m *minimaxHailuo02Model) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:             "minimax/hailuo-02",
		Description:      "MiniMax Hailuo-02 Text To Video. Per-second pricing.",
		PriceUSD:         0.09,
		Type:             "text2video",
		Endpoint:         "/minimax/hailuo-02/standard/text-to-video",
		HelpDoc:          "Usage: !text2video [prompt] [--duration 6|10] [--prompt_optimizer true|false]\n\n💰 **Price: $0.10 per video second**\nExample: A 10-second video will cost $1.00.\nTotal cost = price per second × duration.",
		Options:          &MinimaxHailuo02Options{Duration: "6", PromptOptimizer: &defaultOptimizer},
		PerSecondPricing: true,
	}
}

// --- hunyuan-video ---

type hunyuanVideoModel struct{}

func (m *hunyuanVideoModel) Define() Model {
	defaultOpts := &HunyuanVideoOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "hunyuan-video",
		Description: "Tencent Hunyuan Video - High visual quality, motion diversity and text alignment",
		PriceUSD:    0.50,
		Type:        "text2video",
		Endpoint:    "/hunyuan-video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.50 per video\n\nParameters:\n• prompt: Text description (required)\n• --aspect_ratio: 16:9, 9:16, 4:3, 3:4, 1:1 (default: 16:9)\n• --resolution: 480p, 580p, 720p, 1080p (default: 720p)\n• --video_length: 5s, 10s (default: 5s)\n• --num_inference_steps: Number of steps (default: 50)\n• --seed: Specific seed (optional)\n• --enable_safety_checker: Enable safety filter (default: true)",
		Options: &HunyuanVideoOptions{
			AspectRatio:         defaults["aspect_ratio"].(string),
			Resolution:          defaults["resolution"].(string),
			VideoLength:         defaults["video_length"].(string),
			NumInferenceSteps:   defaults["num_inference_steps"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
		},
	}
}

// --- kling-video-v25-text ---

type klingVideoV25TextModel struct{}

func (m *klingVideoV25TextModel) Define() Model {
	defaultOpts := &KlingVideoV25Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:             "kling-video-v25-text",
		Description:      "Kling 2.5 Turbo Pro Text-to-Video - High quality video generation from text",
		PriceUSD:         0.32, // Per second
		Type:             "text2video",
		Endpoint:         "/kling-video/v2.5/turbo-pro/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.32 per second\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (5 or 10, default: 5)\n• --aspect_ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)",
		Options: &KlingVideoV25Options{
			Duration:       defaults["duration"].(string),
			AspectRatio:    defaults["aspect_ratio"].(string),
			NegativePrompt: defaults["negative_prompt"].(string),
			CFGScale:       defaults["cfg_scale"].(float64),
		},
	}
}

// --- kling-video-v3-text ---

type klingVideoV3TextModel struct{}

func (m *klingVideoV3TextModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "kling-video-v3-text",
		Description:      "Kling 3.0 Standard Text-to-Video - High quality video generation with audio",
		PriceUSD:         0.30, // Per second
		Type:             "text2video",
		Endpoint:         "/kling-video/v3/standard/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.30 per second**\nExample: A 5-second video will cost $1.50.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)",
		Options: &KlingVideoV3Options{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
			GenerateAudio:  &defaultAudio,
		},
	}
}

// --- kling-video-v3-pro-text ---

type klingVideoV3ProTextModel struct{}

func (m *klingVideoV3ProTextModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "kling-video-v3-pro-text",
		Description:      "Kling 3.0 Pro Text-to-Video - Premium quality video generation with audio",
		PriceUSD:         0.39, // Per second
		Type:             "text2video",
		Endpoint:         "/kling-video/v3/pro/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.39 per second**\nExample: A 5-second video will cost $1.95.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)\n• --audio: Enable audio generation (default: true)",
		Options: &KlingVideoV3Options{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
			GenerateAudio:  &defaultAudio,
		},
	}
}

// --- kling-video-o3-text ---

type klingVideoO3TextModel struct{}

func (m *klingVideoO3TextModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "kling-video-o3-text",
		Description:      "Kling O3 Standard Text-to-Video - Multi-scene consistency with audio",
		PriceUSD:         0.28, // Per second
		Type:             "text2video",
		Endpoint:         "/kling-video/o3/standard/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.28 per second**\nExample: A 5-second video will cost $1.40.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --audio: Enable audio generation (default: true)",
		Options: &KlingVideoO3TextOptions{
			Duration:      "5",
			AspectRatio:   "16:9",
			GenerateAudio: &defaultAudio,
		},
	}
}

// --- kling-video-o3-pro-text ---

type klingVideoO3ProTextModel struct{}

func (m *klingVideoO3ProTextModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "kling-video-o3-pro-text",
		Description:      "Kling O3 Pro Text-to-Video - Premium multi-scene consistency with audio",
		PriceUSD:         0.33, // Per second
		Type:             "text2video",
		Endpoint:         "/kling-video/o3/pro/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.33 per second**\nExample: A 5-second video will cost $1.65.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description (required)\n• --duration: Video duration in seconds (3-15, default: 5)\n• --aspect: Aspect ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --audio: Enable audio generation (default: true)",
		Options: &KlingVideoO3TextOptions{
			Duration:      "5",
			AspectRatio:   "16:9",
			GenerateAudio: &defaultAudio,
		},
	}
}

// --- seedance-2.0-text ---

type seedanceTextModel struct{}

func (m *seedanceTextModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "seedance-2.0-text",
		Description:      "ByteDance Seedance 2.0 Text-to-Video - Realistic motion with native audio generation",
		PriceUSD:         0.35, // $0.35 per second (flat rate; fal charges $0.3034/s at 720p)
		Type:             "text2video",
		Endpoint:         "https://queue.fal.run/bytedance/seedance-2.0/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.35 per second**\nExample: A 5-second video will cost $1.75.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description of the desired video (required)\n• --duration: Video duration in seconds (4-15, default: 5)\n• --aspect: Aspect ratio (auto, 21:9, 16:9, 4:3, 1:1, 3:4, 9:16). Default: auto\n• --resolution: Video resolution (480p, 720p). Default: 720p\n• --audio: Enable audio generation (default: true)\n• --seed: Seed for reproducibility (optional)",
		Options: &SeedanceOptions{
			Duration:      "5",
			AspectRatio:   "auto",
			Resolution:    "720p",
			GenerateAudio: &defaultAudio,
		},
	}
}

func init() {
	registerModel(&minimaxHailuo02Model{})
	registerModel(&hunyuanVideoModel{})
	registerModel(&klingVideoV25TextModel{})
	registerModel(&grokImagineVideoTextModel{})
	registerModel(&klingVideoV3TextModel{})
	registerModel(&klingVideoV3ProTextModel{})
	registerModel(&klingVideoO3TextModel{})
	registerModel(&klingVideoO3ProTextModel{})
	registerModel(&seedanceTextModel{})
}

// --- grok-imagine-video-text ---

type grokImagineVideoTextModel struct{}

func (m *grokImagineVideoTextModel) Define() Model {
	return Model{
		Name:             "grok-imagine-video-text",
		Description:      "Grok Imagine Video - Text-to-video generation by xAI",
		PriceUSD:         0.08,
		Type:             "text2video",
		Endpoint:         "https://queue.fal.run/xai/grok-imagine-video/text-to-video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.08 per video second**\nExample: A 6-second video will cost $0.48.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description (required, max 4096 chars)\n• --duration: Video duration in seconds (1-15, default: 6)\n• --aspect: Aspect ratio: 16:9, 4:3, 3:2, 1:1, 2:3, 3:4, 9:16 (default: 16:9)\n• --resolution: 480p, 720p (default: 720p)",
		Options:          &GrokImagineVideoTextOptions{Duration: 6, AspectRatio: "16:9", Resolution: "720p"},
	}
}
