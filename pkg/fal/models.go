// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"fmt"
)

var (
	// allModels stores all registered models
	allModels = make(map[string]Model)
	// defaultModels stores the default model for each type
	defaultModels = make(map[string]string)
	// userModels stores per-user model settings
	userModels = make(map[string]map[string]string) // map[userID]map[modelType]modelName
)

// registerModel registers a model defined by a ModelDefinition
func registerModel(def ModelDefinition) {
	model := def.Define()
	if model.Name == "" {
		// Optional: Add some logging or error handling if a model definition is invalid
		return
	}
	allModels[model.Name] = model
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
func GetCurrentModel(commandType string, userID string) (Model, bool) {
	var modelName string
	var exists bool

	// First check user-specific model if userID is provided
	if userID != "" {
		if userModels, hasUserModels := userModels[userID]; hasUserModels {
			if modelName, exists = userModels[commandType]; exists {
				return GetModel(modelName, commandType)
			}
		}
	}

	// Fall back to global model
	modelName, exists = defaultModels[commandType]
	if !exists {
		return Model{}, false
	}
	return GetModel(modelName, commandType)
}

// SetCurrentModel sets the current model for a command type
func SetCurrentModel(commandType, modelName string, userID string) error {
	if _, exists := allModels[modelName]; !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	if userID != "" {
		// Set per-user model
		if _, exists := userModels[userID]; !exists {
			userModels[userID] = make(map[string]string)
		}
		userModels[userID][commandType] = modelName
	} else {
		// Set global model
		defaultModels[commandType] = modelName
	}
	return nil
}

func init() {
	// Registration now happens in the init() functions of individual model files.
	// We only set the defaults here.

	// Set default models
	setDefaultModel("text2image", "flux/schnell")
	setDefaultModel("image2image", "ghiblify")
	setDefaultModel("text2speech", "minimax-tts/text-to-speech")
	setDefaultModel("image2video", "veo2")
	setDefaultModel("text2video", "kling-video-text")
	setDefaultModel("audio2text", "elevenlabs/speech-to-text/scribe-v2")
}
