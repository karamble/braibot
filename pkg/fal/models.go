// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"fmt"
)

// Text2ImageModels contains all available text-to-image models
var Text2ImageModels = map[string]Model{
	"fast-sdxl": {
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly",
		PriceUSD:    0.02,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-full": {
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images",
		PriceUSD:    0.10,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-dev": {
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model",
		PriceUSD:    0.06,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-fast": {
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model",
		PriceUSD:    0.03,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux-pro/v1.1": {
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation",
		PriceUSD:    0.08,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux-pro/v1.1-ultra": {
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model",
		PriceUSD:    0.12,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux/schnell": {
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation",
		PriceUSD:    0.02,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
}

// Image2ImageModels contains all available image-to-image models
var Image2ImageModels = map[string]Model{
	"ghiblify": {
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
	},
	"cartoonify": {
		Name:        "cartoonify",
		Description: "Transforms images into Pixar like 3d cartoon-style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
	},
	"star-vector": {
		Name:        "star-vector",
		Description: "Convert images to SVG using AI vectorization",
		PriceUSD:    1.0,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nTo use this model, first set it as the default model:\n!setmodel image2image star-vector\n\nParameters:\n• image_url: URL of the source image\n\nPricing:\n• Base price: $1.0 per image",
	},
}

// Text2SpeechModels contains all available text-to-speech models
var Text2SpeechModels = map[string]Model{
	"minimax-tts/text-to-speech": {
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters",
		PriceUSD:    0.10,
		Type:        "text2speech",
		HelpDoc:     "Usage: !text2speech [voice_id] [text]\nExample: !text2speech Wise_Woman Hello, how are you today?\n\nParameters:\n• voice_id: Optional voice ID (defaults to Wise_Woman)\n• text: Text to convert to speech\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl",
	},
}

// Image2VideoModels contains all available image-to-video models
var Image2VideoModels = map[string]Model{
	"veo2": {
		Name:        "veo2",
		Description: "Creates videos from images with realistic motion using Google's Veo 2 model. Base price: $2.50 for 5 seconds, $0.50 per additional second",
		PriceUSD:    3.50,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --aspect 16:9 --duration 5\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (16:9, 9:16, 1:1)\n• --duration: Video duration (5, 6, 7, 8)\n\nPricing:\n• Base price: $3.50 for 5 seconds\n• Additional cost: $0.50 per second beyond 5 seconds",
		Options: &Veo2Options{
			AspectRatio: "16:9",
			Duration:    "5",
		},
	},
	"kling-video-image": {
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
	},
}

// Text2VideoModels contains all available text-to-video models
var Text2VideoModels = map[string]Model{
	"kling-video-text": {
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
	},
}

var (
	// allModels stores all registered models
	allModels = make(map[string]Model)
	// defaultModels stores the default model for each type
	defaultModels = make(map[string]string)
)

// registerModel registers a model with the given name
func registerModel(name string, model Model) {
	allModels[name] = model
}

// setDefaultModel sets the default model for a given type
func setDefaultModel(modelType, modelName string) {
	defaultModels[modelType] = modelName
}

// GetModel returns a model by name and type
func GetModel(name, modelType string) (Model, bool) {
	model, exists := allModels[name]
	if !exists {
		return Model{}, false
	}
	if model.Type != modelType {
		return Model{}, false
	}
	return model, true
}

// GetModels returns all available models for a command type
func GetModels(commandType string) (map[string]Model, bool) {
	models := make(map[string]Model)
	for name, model := range allModels {
		if model.Type == commandType {
			models[name] = model
		}
	}
	return models, len(models) > 0
}

// GetCurrentModel returns the current model for a command type
func GetCurrentModel(commandType string) (Model, bool) {
	modelName, exists := defaultModels[commandType]
	if !exists {
		return Model{}, false
	}
	return GetModel(modelName, commandType)
}

// SetCurrentModel sets the current model for a command type
func SetCurrentModel(commandType, modelName string) error {
	if _, exists := allModels[modelName]; !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}
	defaultModels[commandType] = modelName
	return nil
}

func init() {
	// Register all models
	for name, model := range Text2ImageModels {
		registerModel(name, model)
	}
	for name, model := range Image2ImageModels {
		registerModel(name, model)
	}
	for name, model := range Text2SpeechModels {
		registerModel(name, model)
	}
	for name, model := range Image2VideoModels {
		registerModel(name, model)
	}
	for name, model := range Text2VideoModels {
		registerModel(name, model)
	}

	// Set default models
	setDefaultModel("text2image", "fast-sdxl")
	setDefaultModel("image2image", "ghiblify")
	setDefaultModel("text2speech", "minimax-tts/text-to-speech")
	setDefaultModel("image2video", "veo2")
	setDefaultModel("text2video", "kling-video-text")
}
