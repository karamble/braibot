// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- ghiblify ---

type ghiblifyModel struct{}

func (m *ghiblifyModel) Define() Model {
	return Model{
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork",
		Type:        "image2image",
		Endpoint:    "/ghiblify",
	}
}

// --- cartoonify ---

type cartoonifyModel struct{}

func (m *cartoonifyModel) Define() Model {
	return Model{
		Name:        "cartoonify",
		Description: "Transforms images into Pixar like 3d cartoon-style artwork",
		Type:        "image2image",
		Endpoint:    "/cartoonify",
		Options:     &CartoonifyOptions{},
	}
}

// --- star-vector ---

type starVectorModel struct{}

func (m *starVectorModel) Define() Model {
	return Model{
		Name:        "star-vector",
		Description: "Convert images to SVG using AI vectorization",
		Type:        "image2image",
		Endpoint:    "/fal-ai/star-vector",
		Options:     &StarVectorOptions{},
	}
}

// --- flux-2-pro/edit ---

type flux2ProEditModel struct{}

func (m *flux2ProEditModel) Define() Model {
	defaultOpts := &Flux2ProEditOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)

	return Model{
		Name:        "flux-2-pro/edit",
		Description: "FLUX 2.0 Pro Edit - AI-powered image editing with text prompts",
		Type:        "image2image",
		Endpoint:    "/flux-2-pro/edit",
		Options: &Flux2ProEditOptions{
			ImageSize:           defaults["image_size"].(string),
			EnableSafetyChecker: defaultSafetyChecker,
			SafetyTolerance:     defaults["safety_tolerance"].(string),
			OutputFormat:        defaults["output_format"].(string),
		},
	}
}

func init() {
	registerModel(&ghiblifyModel{})
	registerModel(&cartoonifyModel{})
	registerModel(&starVectorModel{})
	registerModel(&flux2ProEditModel{})
	registerModel(&flux2EditModel{})
}

// --- flux-2/edit ---

type flux2EditModel struct{}

func (m *flux2EditModel) Define() Model {
	defaultOpts := &Flux2EditOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultSafetyChecker := defaults["enable_safety_checker"].(*bool)
	defaultPromptExpansion := defaults["enable_prompt_expansion"].(*bool)

	return Model{
		Name:        "flux-2/edit",
		Description: "FLUX 2.0 Edit - Affordable AI-powered image editing",
		Type:        "image2image",
		Endpoint:    "/flux-2/edit",
		Options: &Flux2EditOptions{
			ImageSize:             defaults["image_size"].(string),
			GuidanceScale:         defaults["guidance_scale"].(float64),
			NumInferenceSteps:     defaults["num_inference_steps"].(int),
			NumImages:             defaults["num_images"].(int),
			Acceleration:          defaults["acceleration"].(string),
			EnablePromptExpansion: defaultPromptExpansion,
			EnableSafetyChecker:   defaultSafetyChecker,
			OutputFormat:          defaults["output_format"].(string),
		},
	}
}
