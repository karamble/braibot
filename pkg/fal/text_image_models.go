// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- fast-sdxl ---

type fastSDXLModel struct{}

func (m *fastSDXLModel) Define() Model {
	return Model{
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly",
		Type:        "text2image",
		Endpoint:    "/fast-sdxl",
	}
}

// --- hidream-i1-full ---

type hidreamI1FullModel struct{}

func (m *hidreamI1FullModel) Define() Model {
	defaultOpts := &HiDreamOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)
	defaultNumSteps := 50       // Correct default for full
	defaultGuidanceScale := 5.0 // Correct default for full

	return Model{
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images (HiDream I1 Full 17B)",
		Type:        "text2image",
		Endpoint:    "/hidream-i1-full",
		Options: &HiDreamOptions{
			NegativePrompt:      defaults["negative_prompt"].(string),
			ImageSize:           defaults["image_size"].(string),
			NumInferenceSteps:   &defaultNumSteps,      // Use correct default pointer
			GuidanceScale:       &defaultGuidanceScale, // Use correct default pointer
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			OutputFormat:        defaults["output_format"].(string),
		},
	}
}

// --- hidream-i1-dev ---

type hidreamI1DevModel struct{}

func (m *hidreamI1DevModel) Define() Model {
	// Get base defaults, then override for dev
	baseDefaults := (&HiDreamOptions{}).GetDefaultValues()
	devDefaultSteps := 28 // Default for dev
	devSafetyChecker := baseDefaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model",
		Type:        "text2image",
		Endpoint:    "/hidream-i1-dev",
		Options: &HiDreamOptions{
			NegativePrompt:      baseDefaults["negative_prompt"].(string),
			ImageSize:           baseDefaults["image_size"].(string),
			NumInferenceSteps:   &devDefaultSteps, // Use dev default
			GuidanceScale:       nil,              // Not applicable for dev
			NumImages:           baseDefaults["num_images"].(int),
			EnableSafetyChecker: devSafetyChecker,
			OutputFormat:        baseDefaults["output_format"].(string),
		},
	}
}

// --- hidream-i1-fast ---

type hidreamI1FastModel struct{}

func (m *hidreamI1FastModel) Define() Model {
	// Get base defaults, then override for fast
	baseDefaults := (&HiDreamOptions{}).GetDefaultValues()
	fastDefaultSteps := 16 // Default for fast
	fastSafetyChecker := baseDefaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model",
		Type:        "text2image",
		Endpoint:    "/hidream-i1-fast",
		Options: &HiDreamOptions{
			NegativePrompt:      baseDefaults["negative_prompt"].(string),
			ImageSize:           baseDefaults["image_size"].(string),
			NumInferenceSteps:   &fastDefaultSteps, // Use fast default
			GuidanceScale:       nil,               // Not applicable for fast
			NumImages:           baseDefaults["num_images"].(int),
			EnableSafetyChecker: fastSafetyChecker,
			OutputFormat:        baseDefaults["output_format"].(string),
		},
	}
}

// --- flux-pro/v1.1 ---

type fluxProV1_1Model struct{}

func (m *fluxProV1_1Model) Define() Model {
	// Define default options
	defaultOpts := &FluxProV1_1Options{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation (FLUX1.1 pro)",
		Type:        "text2image",
		Endpoint:    "/flux-pro/v1.1",
		Options: &FluxProV1_1Options{
			ImageSize:           defaults["image_size"].(string),
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			SafetyTolerance:     defaults["safety_tolerance"].(string),
			OutputFormat:        defaults["output_format"].(string),
			// Seed and SyncMode default to nil/false
		},
	}
}

// --- flux-pro/v1.1-ultra ---

type fluxProV1_1UltraModel struct{}

func (m *fluxProV1_1UltraModel) Define() Model {
	defaultOpts := &FluxProV1_1UltraOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)
	defaultRaw := defaults["raw"].(*bool)

	return Model{
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model (FLUX pro ultra)",
		Type:        "text2image",
		Endpoint:    "/flux-pro/v1.1-ultra",
		Options: &FluxProV1_1UltraOptions{
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			SafetyTolerance:     defaults["safety_tolerance"].(string),
			OutputFormat:        defaults["output_format"].(string),
			AspectRatio:         defaults["aspect_ratio"].(string),
			Raw:                 defaultRaw,
		},
	}
}

// --- flux/schnell ---

type fluxSchnellModel struct{}

func (m *fluxSchnellModel) Define() Model {
	// Define default options
	defaultOpts := &FluxSchnellOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation (FLUX.1 schnell)",
		Type:        "text2image",
		Endpoint:    "/flux/schnell",
		Options: &FluxSchnellOptions{
			ImageSize:           defaults["image_size"].(string),
			NumInferenceSteps:   defaults["num_inference_steps"].(int),
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			// Seed and SyncMode default to nil/false
		},
	}
}

// --- flux/dev ---

type fluxDevModel struct{}

func (m *fluxDevModel) Define() Model {
	defaultOpts := &FluxDevOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "flux/dev",
		Description: "FLUX.1 [dev] - 12B parameter flow transformer for high-quality image generation",
		Type:        "text2image",
		Endpoint:    "/flux/dev",
		Options: &FluxDevOptions{
			ImageSize:           defaults["image_size"].(string),
			NumInferenceSteps:   defaults["num_inference_steps"].(int),
			GuidanceScale:       defaults["guidance_scale"].(float64),
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			OutputFormat:        defaults["output_format"].(string),
		},
	}
}

// --- stable-diffusion-v35-large ---

type stableDiffusionV35LargeModel struct{}

func (m *stableDiffusionV35LargeModel) Define() Model {
	defaultOpts := &StableDiffusionV35LargeOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)
	defaultPromptExpansion := defaults["prompt_expansion"].(*bool)

	return Model{
		Name:        "stable-diffusion-v35-large",
		Description: "Stable Diffusion 3.5 Large - Image quality, typography, complex prompt understanding",
		Type:        "text2image",
		Endpoint:    "/stable-diffusion-v35-large",
		Options: &StableDiffusionV35LargeOptions{
			ImageSize:           defaults["image_size"].(string),
			NumInferenceSteps:   defaults["num_inference_steps"].(int),
			GuidanceScale:       defaults["guidance_scale"].(float64),
			NumImages:           defaults["num_images"].(int),
			EnableSafetyChecker: defaultSafetyChecker,
			OutputFormat:        defaults["output_format"].(string),
			NegativePrompt:      defaults["negative_prompt"].(string),
			PromptExpansion:     defaultPromptExpansion,
		},
	}
}

func init() {
	registerModel(&fastSDXLModel{})
	registerModel(&hidreamI1FullModel{})
	registerModel(&hidreamI1DevModel{})
	registerModel(&hidreamI1FastModel{})
	registerModel(&fluxProV1_1Model{})
	registerModel(&fluxProV1_1UltraModel{})
	registerModel(&fluxSchnellModel{})
	registerModel(&fluxDevModel{})
	registerModel(&stableDiffusionV35LargeModel{})
}
