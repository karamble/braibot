// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

// Text2ImageModels contains all available text-to-image models
var Text2ImageModels = map[string]Model{
	"fast-sdxl": {
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly.",
		Price:       0.02,
	},
	"hidream-i1-full": {
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images.",
		Price:       0.10,
	},
	"hidream-i1-dev": {
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model.",
		Price:       0.06,
	},
	"hidream-i1-fast": {
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model.",
		Price:       0.03,
	},
	"flux-pro/v1.1": {
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation.",
		Price:       0.08,
	},
	"flux-pro/v1.1-ultra": {
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model.",
		Price:       0.12,
	},
	"flux/schnell": {
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation.",
		Price:       0.02,
	},
}

// Text2SpeechModels contains all available text-to-speech models
var Text2SpeechModels = map[string]Model{
	"minimax-tts/text-to-speech": {
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters.",
		Price:       0.10,
	},
}

// DefaultModels contains the default model for each command type
var DefaultModels = map[string]string{
	"text2image":  "fast-sdxl",                  // Default model for text2image
	"text2speech": "minimax-tts/text-to-speech", // Default model for text2speech
}

// GetModel returns the model configuration for a given model name and command type
func GetModel(modelName, commandType string) (Model, bool) {
	var models map[string]Model
	switch commandType {
	case "text2image":
		models = Text2ImageModels
	case "text2speech":
		models = Text2SpeechModels
	default:
		return Model{}, false
	}
	model, exists := models[modelName]
	return model, exists
}

// GetDefaultModel returns the default model name for a given command type
func GetDefaultModel(commandType string) (string, bool) {
	model, exists := DefaultModels[commandType]
	return model, exists
}

// SetDefaultModel sets the default model for a given command type
func SetDefaultModel(commandType, modelName string) bool {
	if _, exists := GetModel(modelName, commandType); !exists {
		return false
	}
	DefaultModels[commandType] = modelName
	return true
}
