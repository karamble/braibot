// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

var (
	// allModels stores all registered models
	allModels = make(map[string]Model)
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
