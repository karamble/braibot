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
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-full": {
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images.",
		Price:       0.10,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-dev": {
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model.",
		Price:       0.06,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"hidream-i1-fast": {
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model.",
		Price:       0.03,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux-pro/v1.1": {
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation.",
		Price:       0.08,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux-pro/v1.1-ultra": {
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model.",
		Price:       0.12,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
	"flux/schnell": {
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation.",
		Price:       0.02,
		HelpDoc:     "Usage: !text2image <prompt>\nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	},
}

// Image2ImageModels contains all available image-to-image models
var Image2ImageModels = map[string]Model{
	"ghiblify": {
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork.",
		Price:       0.15,
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
	},
	"cartoonify": {
		Name:        "cartoonify",
		Description: "Transforms images into Pixar like 3d cartoon-style artwork.",
		Price:       0.15,
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
	},
}

// Text2SpeechModels contains all available text-to-speech models
var Text2SpeechModels = map[string]Model{
	"minimax-tts/text-to-speech": {
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters.",
		Price:       0.10,
		HelpDoc:     "Usage: !text2speech [voice_id] [text]\nExample: !text2speech Wise_Woman Hello, how are you today?\n\nParameters:\n• voice_id: Optional voice ID (defaults to Wise_Woman)\n• text: Text to convert to speech\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl",
	},
}

// Image2VideoModels contains all available image2video models
var Image2VideoModels = map[string]Model{
	"kling-video": {
		Name:        "kling-video",
		Description: "Convert images to video using Kling 2.0 Master. Base price: $2.0 for 5 seconds, $0.4 per additional second.",
		Price:       2.0,
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 10 --aspect 16:9\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --duration: Video duration in seconds (default: 5, min: 5)\n• --aspect: Aspect ratio (default: 16:9, 9:16, 1:1)\n• --negative-prompt: Text describing what to avoid (default: blur, distort, and low quality)\n• --cfg-scale: Configuration scale (default: 0.5, how much weight to give the prompt)\n\nPricing:\n• Base price: $2.0 for 5 seconds\n• Additional cost: $0.4 per second beyond 5 seconds",
	},
}

// DefaultModels contains the default model for each command type
var DefaultModels = map[string]string{
	"text2image":  "fast-sdxl",                  // Default model for text2image
	"text2speech": "minimax-tts/text-to-speech", // Default model for text2speech
	"image2image": "ghiblify",                   // Default model for image2image
	"image2video": "kling-video",                // Default model for image2video
}

// GetModel returns the model configuration for a given model name and command type
func GetModel(modelName, commandType string) (Model, bool) {
	var models map[string]Model
	switch commandType {
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
