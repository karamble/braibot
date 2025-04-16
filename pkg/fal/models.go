// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// Text2ImageModels contains all available text-to-image models
var Text2ImageModels = map[string]Model{
	"fast-sdxl": {
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly",
		PriceUSD:    0.02,
		Type:        "text2image",
	},
	"hidream-i1-full": {
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images",
		PriceUSD:    0.10,
		Type:        "text2image",
	},
	"hidream-i1-dev": {
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model",
		PriceUSD:    0.06,
		Type:        "text2image",
	},
	"hidream-i1-fast": {
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model",
		PriceUSD:    0.03,
		Type:        "text2image",
	},
	"flux-pro/v1.1": {
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation",
		PriceUSD:    0.08,
		Type:        "text2image",
	},
	"flux-pro/v1.1-ultra": {
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model",
		PriceUSD:    0.12,
		Type:        "text2image",
	},
	"flux/schnell": {
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation",
		PriceUSD:    0.02,
		Type:        "text2image",
	},
}

// Image2ImageModels contains all available image-to-image models
var Image2ImageModels = map[string]Model{
	"ghiblify": {
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
	},
	"cartoonify": {
		Name:        "cartoonify",
		Description: "Transforms images into cartoon style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
	},
}

// Text2SpeechModels contains all available text-to-speech models
var Text2SpeechModels = map[string]Model{
	"minimax-tts/text-to-speech": {
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio",
		PriceUSD:    0.10,
		Type:        "text2speech",
	},
}

// Image2VideoModels contains all available image-to-video models
var Image2VideoModels = map[string]Model{
	"kling-video": {
		Name:        "kling-video",
		Description: "Transform images into animated videos using Kling Video",
		PriceUSD:    0.05,
		Type:        "image2video",
	},
}

// DefaultModels specifies the default model for each command type
var DefaultModels = map[string]string{
	"text2image":  "fast-sdxl",
	"image2image": "ghiblify",
	"image2video": "kling-video",
	"text2speech": "minimax-tts/text-to-speech",
}

// GetModel returns a model by name and type
func GetModel(name, modelType string) (Model, bool) {
	var models map[string]Model
	switch modelType {
	case "text2image":
		models = Text2ImageModels
	case "text2speech":
		models = Text2SpeechModels
	case "image2image":
		models = Image2ImageModels
	case "image2video":
		models = Image2VideoModels
	default:
		return Model{}, false
	}

	model, exists := models[name]
	return model, exists
}

// GetModels returns all available models for a command type
func GetModels(commandType string) (map[string]Model, bool) {
	switch commandType {
	case "text2image":
		return Text2ImageModels, true
	case "text2speech":
		return Text2SpeechModels, true
	case "image2image":
		return Image2ImageModels, true
	case "image2video":
		return Image2VideoModels, true
	default:
		return nil, false
	}
}

// GetCurrentModel returns the current model for a command type
func GetCurrentModel(commandType string) (Model, bool) {
	modelName, exists := DefaultModels[commandType]
	if !exists {
		return Model{}, false
	}

	return GetModel(modelName, commandType)
}

// SetCurrentModel sets the current model for a command type
func SetCurrentModel(commandType, modelName string) error {
	if _, exists := GetModel(modelName, commandType); !exists {
		return &Error{
			Code:    "INVALID_MODEL",
			Message: "model not found: " + modelName,
		}
	}

	DefaultModels[commandType] = modelName
	return nil
}
