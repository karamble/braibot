// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL = "https://queue.fal.run/fal-ai"
)

// Client represents a Fal.ai API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	debug      bool
}

// NewClient creates a new Fal.ai API client
func NewClient(apiKey string, debug bool) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: debug,
	}
}

// makeRequest makes an HTTP request to the Fal.ai API
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Create request body
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
	}

	// Create request
	fullURL := baseURL + path
	if c.debug {
		fmt.Printf("Making request to URL: %s\n", fullURL)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.apiKey)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	return resp, nil
}

// GetModels returns all available models for a command type
func (c *Client) GetModels(commandType string) (map[string]Model, error) {
	switch commandType {
	case "text2image":
		return Text2ImageModels, nil
	case "text2speech":
		return Text2SpeechModels, nil
	case "image2image":
		return Image2ImageModels, nil
	case "image2video":
		return Image2VideoModels, nil
	default:
		return nil, fmt.Errorf("unknown command type: %s", commandType)
	}
}

// GetCurrentModel returns the current model for a command type
func (c *Client) GetCurrentModel(commandType string) (Model, error) {
	modelName, exists := DefaultModels[commandType]
	if !exists {
		return Model{}, fmt.Errorf("no default model for command type: %s", commandType)
	}

	model, exists := GetModel(modelName, commandType)
	if !exists {
		return Model{}, fmt.Errorf("model not found: %s", modelName)
	}

	return model, nil
}

// SetCurrentModel sets the current model for a command type
func (c *Client) SetCurrentModel(commandType, modelName string) error {
	if _, exists := GetModel(modelName, commandType); !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	DefaultModels[commandType] = modelName
	return nil
}
