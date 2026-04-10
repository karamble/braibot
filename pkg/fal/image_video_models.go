// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- veo2 ---

type veo2Model struct{}

func (m *veo2Model) Define() Model {
	return Model{
		Name:        "veo2",
		Description: "Creates videos from images with realistic motion using Google's Veo 2 model.",
		Type:        "image2video",
		Endpoint:    "/veo2/image-to-video",
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
		Description: "Convert images to video using Kling 2.0 Master.",
		Type:        "image2video",
		Endpoint:    "/kling-video/v2/master/image-to-video",
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
		Type:        "image2video",
		Endpoint:    "/minimax/video-01-subject-reference",
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
		Type:        "image2video",
		Endpoint:    "/minimax/video-01-live/image-to-video",
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
		Name:        "veo3",
		Description: "Google's Veo 3 - state-of-the-art video generation with audio support",
		Type:        "image2video",
		Endpoint:    "/veo3/image-to-video",
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
		Name:        "veo31fast",
		Description: "Google's Veo 3.1 Fast - fast video generation with audio support",
		Type:        "image2video",
		Endpoint:    "/veo3.1/fast/image-to-video",
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
		Name:        "kling-video-v25-image",
		Description: "Kling 2.5 Turbo Pro Image-to-Video - High quality video from images",
		Type:        "image2video",
		Endpoint:    "/kling-video/v2.5/turbo-pro/image-to-video",
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
		Type:        "image2video",
		Endpoint:    "/luma-dream-machine",
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
		Type:        "image2video",
		Endpoint:    "/ltx-video-13b-distilled/multiconditioning",
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

// --- grok-imagine-video ---

type grokImagineVideoModel struct{}

func (m *grokImagineVideoModel) Define() Model {
	defaultOpts := &GrokImagineVideoOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "grok-imagine-video",
		Description: "Grok Imagine Video - Image-to-video generation by xAI",
		Type:        "image2video",
		Endpoint:    "https://queue.fal.run/xai/grok-imagine-video/image-to-video",
		Options: &GrokImagineVideoOptions{
			Duration:    defaults["duration"].(int),
			AspectRatio: defaults["aspect_ratio"].(string),
			Resolution:  defaults["resolution"].(string),
		},
	}
}

// --- kling-video-v3-image ---

type klingVideoV3ImageModel struct{}

func (m *klingVideoV3ImageModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:        "kling-video-v3-image",
		Description: "Kling 3.0 Standard Image-to-Video - High quality video from images with audio",
		Type:        "image2video",
		Endpoint:    "/kling-video/v3/standard/image-to-video",
		Options: &KlingVideoV3Options{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
			GenerateAudio:  &defaultAudio,
		},
	}
}

// --- kling-video-v3-pro-image ---

type klingVideoV3ProImageModel struct{}

func (m *klingVideoV3ProImageModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:        "kling-video-v3-pro-image",
		Description: "Kling 3.0 Pro Image-to-Video - Premium quality video from images with audio",
		Type:        "image2video",
		Endpoint:    "/kling-video/v3/pro/image-to-video",
		Options: &KlingVideoV3Options{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
			GenerateAudio:  &defaultAudio,
		},
	}
}

// --- seedance-2.0-image ---

type seedanceImageModel struct{}

func (m *seedanceImageModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:        "seedance-2.0-image",
		Description: "ByteDance Seedance 2.0 Image-to-Video - Realistic motion with native audio generation",
		Type:        "image2video",
		Endpoint:    "https://queue.fal.run/bytedance/seedance-2.0/image-to-video",
		Options: &SeedanceOptions{
			Duration:      "5",
			AspectRatio:   "auto",
			Resolution:    "720p",
			GenerateAudio: &defaultAudio,
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
	registerModel(&grokImagineVideoModel{})
	registerModel(&klingVideoV3ImageModel{})
	registerModel(&klingVideoV3ProImageModel{})
	registerModel(&seedanceImageModel{})
}
