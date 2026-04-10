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
		Type:        "text2video",
		Endpoint:    "/kling-video/v2/master/text-to-video",
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
		Type:        "text2video",
		Endpoint:    "/minimax/video-01-director",
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
		Type:        "text2video",
		Endpoint:    "/minimax/video-01",
		Options: &MinimaxVideo01Options{
			PromptOptimizer: &defaultOptimizer,
		},
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
		Name:        "minimax/hailuo-02",
		Description: "MiniMax Hailuo-02 Text To Video. Per-second pricing.",
		Type:        "text2video",
		Endpoint:    "/minimax/hailuo-02/standard/text-to-video",
		Options:     &MinimaxHailuo02Options{Duration: "6", PromptOptimizer: &defaultOptimizer},
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
		Type:        "text2video",
		Endpoint:    "/hunyuan-video",
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
		Name:        "kling-video-v25-text",
		Description: "Kling 2.5 Turbo Pro Text-to-Video - High quality video generation from text",
		Type:        "text2video",
		Endpoint:    "/kling-video/v2.5/turbo-pro/text-to-video",
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
		Name:        "kling-video-v3-text",
		Description: "Kling 3.0 Standard Text-to-Video - High quality video generation with audio",
		Type:        "text2video",
		Endpoint:    "/kling-video/v3/standard/text-to-video",
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
		Name:        "kling-video-v3-pro-text",
		Description: "Kling 3.0 Pro Text-to-Video - Premium quality video generation with audio",
		Type:        "text2video",
		Endpoint:    "/kling-video/v3/pro/text-to-video",
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
		Name:        "kling-video-o3-text",
		Description: "Kling O3 Standard Text-to-Video - Multi-scene consistency with audio",
		Type:        "text2video",
		Endpoint:    "/kling-video/o3/standard/text-to-video",
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
		Name:        "kling-video-o3-pro-text",
		Description: "Kling O3 Pro Text-to-Video - Premium multi-scene consistency with audio",
		Type:        "text2video",
		Endpoint:    "/kling-video/o3/pro/text-to-video",
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
		Name:        "seedance-2.0-text",
		Description: "ByteDance Seedance 2.0 Text-to-Video - Realistic motion with native audio generation",
		Type:        "text2video",
		Endpoint:    "https://queue.fal.run/bytedance/seedance-2.0/text-to-video",
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
		Name:        "grok-imagine-video-text",
		Description: "Grok Imagine Video - Text-to-video generation by xAI",
		Type:        "text2video",
		Endpoint:    "https://queue.fal.run/xai/grok-imagine-video/text-to-video",
		Options:     &GrokImagineVideoTextOptions{Duration: 6, AspectRatio: "16:9", Resolution: "720p"},
	}
}
