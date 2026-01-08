// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- veo2 ---

type veo2Model struct{}

func (m *veo2Model) Define() Model {
	return Model{
		Name:        "veo2",
		Description: "Creates videos from images with realistic motion using Google's Veo 2 model. Base price: $2.50 for 5 seconds, $0.50 per additional second",
		PriceUSD:    3.50,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --aspect 16:9 --duration 5\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (16:9, 9:16, 1:1)\n• --duration: Video duration (5, 6, 7, 8)\n\nPricing:\n• Base price: $3.50 for 5 seconds\n• Additional cost: $0.50 per second beyond 5 seconds",
		Options: &Veo2Options{
			AspectRatio: "16:9",
			Duration:    "5",
		},
	}
}

// --- kling-video-image ---

type klingVideoImageModel struct{}

func (m *klingVideoImageModel) Define() Model {
	return Model{
		Name:        "kling-video-image",
		Description: "Convert images to video using Kling 2.0 Master. Base price: $2.0 for 5 seconds, $0.4 per additional second",
		PriceUSD:    2.0,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 10 --aspect 16:9\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --duration: Video duration in seconds (default: 5, min: 5)\n• --aspect: Aspect ratio (default: 16:9)\n• --negative-prompt: Text describing what to avoid (default: blur, distort, and low quality)\n• --cfg-scale: Configuration scale (default: 0.5)\n\nPricing:\n• Base price: $2.0 for 5 seconds\n• Additional cost: $0.4 per second beyond 5 seconds",
		Options: &KlingVideoOptions{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
		},
	}
}

// --- minimax/video-01-subject-reference ---

type minimaxSubjectReferenceModel struct{}

func (m *minimaxSubjectReferenceModel) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01-subject-reference",
		Description: "Generate video from a subject reference image.",
		PriceUSD:    0.8,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [subject_reference_image_url] [prompt] [options]\nExample: !image2video https://example.com/subject.jpg a person walking --prompt-optimizer false\n\nParameters:\n• subject_reference_image_url: URL of the image to use for consistent subject appearance.\n• prompt: Description of the desired video animation.\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)",
		Options: &MinimaxSubjectReferenceOptions{
			PromptOptimizer: &defaultOptimizer,
		},
	}
}

// --- minimax/video-01-live ---

type minimaxLiveModel struct{}

func (m *minimaxLiveModel) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01-live",
		Description: "Generate video from an image, specialized in bringing 2D illustrations to life.",
		PriceUSD:    0.8,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.png A character waving --prompt-optimizer true\n\nInfo: This model is specialized in bringing 2D illustrations to life.\n\nParameters:\n• image_url: URL of the image to animate.\n• prompt: Description of the desired video animation.\n• --prompt-optimizer: Whether to use the model's prompt optimizer (default: true)",
		Options: &MinimaxLiveOptions{
			PromptOptimizer: &defaultOptimizer,
		},
	}
}

// --- veo3 ---

type veo3Model struct{}

func (m *veo3Model) Define() Model {
	defaultAudio := true
	defaultAutoFix := false
	return Model{
		Name:             "veo3",
		Description:      "Google's Veo 3 - state-of-the-art video generation with audio support",
		PriceUSD:         0.45, // $0.45 per second
		Type:             "image2video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 8s --resolution 1080p --audio\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (auto, 16:9, 9:16). Default: 16:9\n• --duration: Video duration (4s, 6s, 8s). Default: 8s\n• --resolution: Video resolution (720p, 1080p). Default: 720p\n• --audio: Enable audio generation. Default: true\n• --auto-fix: Auto-fix failed prompts. Default: false\n\nPricing:\n• $0.45 per second of video generated",
		Options: &Veo3Options{
			AspectRatio:   "16:9",
			Duration:      "8s",
			Resolution:    "720p",
			GenerateAudio: &defaultAudio,
			AutoFix:       &defaultAutoFix,
		},
	}
}

// --- veo31fast ---

type veo31FastModel struct{}

func (m *veo31FastModel) Define() Model {
	defaultAudio := true
	defaultAutoFix := false
	return Model{
		Name:             "veo31fast",
		Description:      "Google's Veo 3.1 Fast - fast video generation with audio support",
		PriceUSD:         0.10, // $0.10 per second (no audio), $0.15 per second (with audio)
		Type:             "image2video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 8s --resolution 1080p --audio\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (auto, 16:9, 9:16). Default: auto\n• --duration: Video duration (4s, 6s, 8s). Default: 8s\n• --resolution: Video resolution (720p, 1080p). Default: 720p\n• --audio: Enable audio generation. Default: true\n• --auto-fix: Auto-fix failed prompts. Default: false\n\nPricing:\n• $0.10 per second (no audio)\n• $0.15 per second (with audio)",
		Options: &Veo31FastOptions{
			AspectRatio:   "auto",
			Duration:      "8s",
			Resolution:    "720p",
			GenerateAudio: &defaultAudio,
			AutoFix:       &defaultAutoFix,
		},
	}
}

// --- kling-video-v25-image ---

type klingVideoV25ImageModel struct{}

func (m *klingVideoV25ImageModel) Define() Model {
	defaultOpts := &KlingVideoV25Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:             "kling-video-v25-image",
		Description:      "Kling 2.5 Turbo Pro Image-to-Video - High quality video from images",
		PriceUSD:         0.32, // Per second
		Type:             "image2video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !image2video [image_url] [prompt] [options]\n\n💰 **Price: $0.32 per second\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired animation\n• --duration: Video duration in seconds (5 or 10, default: 5)\n• --aspect_ratio: 16:9, 9:16, 1:1 (default: 16:9)\n• --negative_prompt: Things to avoid (default: blur, distort, and low quality)\n• --cfg_scale: Configuration scale 0-1 (default: 0.5)",
		Options: &KlingVideoV25Options{
			Duration:       defaults["duration"].(string),
			AspectRatio:    defaults["aspect_ratio"].(string),
			NegativePrompt: defaults["negative_prompt"].(string),
			CFGScale:       defaults["cfg_scale"].(float64),
		},
	}
}

// --- luma-dream-machine ---

type lumaDreamMachineModel struct{}

func (m *lumaDreamMachineModel) Define() Model {
	defaultOpts := &LumaDreamMachineOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultLoop := defaults["loop"].(*bool)

	return Model{
		Name:        "luma-dream-machine",
		Description: "Luma Dream Machine 1.5 - High quality video generation from images",
		PriceUSD:    0.40,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\n\n💰 **Price: $0.40 per video\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired animation\n• --aspect_ratio: 16:9, 9:16, 4:3, 3:4, 21:9, 9:21, 1:1 (default: 16:9)\n• --loop: Create looping video (default: false)",
		Options: &LumaDreamMachineOptions{
			AspectRatio: defaults["aspect_ratio"].(string),
			Loop:        defaultLoop,
		},
	}
}

// --- ltx-video-13b ---

type ltxVideo13BModel struct{}

func (m *ltxVideo13BModel) Define() Model {
	defaultOpts := &LTXVideo13BOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "ltx-video-13b",
		Description: "LTX Video 13B Distilled - Generate videos from prompts and images",
		PriceUSD:    0.30,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\n\n💰 **Price: $0.30 per video\n\nParameters:\n• image_url: URL of the source image (for first/last frame)\n• prompt: Description of the desired animation\n• --num_frames: Number of frames (default: 97)\n• --frame_rate: Frame rate (default: 25)\n• --num_inference_steps: Number of steps (default: 30)\n• --guidance_scale: Prompt adherence (default: 3.0)\n• --negative_prompt: Things to avoid (optional)\n• --seed: Specific seed (optional)\n• --enable_safety_checker: Enable safety filter (default: true)",
		Options: &LTXVideo13BOptions{
			NumFrames:           defaults["num_frames"].(int),
			FrameRate:           defaults["frame_rate"].(int),
			NumInferenceSteps:   defaults["num_inference_steps"].(int),
			GuidanceScale:       defaults["guidance_scale"].(float64),
			NegativePrompt:      defaults["negative_prompt"].(string),
			EnableSafetyChecker: defaultSafetyChecker,
		},
	}
}

func init() {
	registerModel(&veo2Model{})
	registerModel(&klingVideoImageModel{})
	registerModel(&minimaxSubjectReferenceModel{})
	registerModel(&minimaxLiveModel{})
	registerModel(&veo3Model{})
	registerModel(&veo31FastModel{})
	registerModel(&klingVideoV25ImageModel{})
	registerModel(&lumaDreamMachineModel{})
	registerModel(&ltxVideo13BModel{})
}
