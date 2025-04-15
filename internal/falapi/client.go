// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	kit "github.com/vctt94/bisonbotkit"
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
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewBuffer(reqBody))
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

// GenerateImageFromImage generates an image from an input image using the specified model
func (c *Client) GenerateImageFromImage(ctx context.Context, prompt, imageURL, modelName string, bot *kit.Bot, userID string) (*GhiblifyResponse, error) {
	// Create request body based on API schema
	reqBody := map[string]interface{}{
		"image_url": imageURL,
	}

	// Make request to the queue API
	resp, err := c.makeRequest(ctx, "POST", "/"+modelName, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to %s API: %v", modelName, err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Try to read the error response body
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	// First, parse the queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}

	// Notify about queue position
	c.notifyQueuePosition(ctx, queueResp, bot, userID)

	// Poll for completion
	_, err = c.pollQueueStatus(ctx, queueResp, bot, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll queue status: %v", err)
	}

	// Get final response using the base request URL (without /status)
	responseURL := strings.TrimSuffix(queueResp.ResponseURL, "/status")
	req, err := http.NewRequestWithContext(ctx, "GET", responseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create final response request: %v", err)
	}
	req.Header.Set("Authorization", "Key "+c.apiKey)

	finalResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %v", err)
	}
	defer finalResp.Body.Close()

	// Read final response body
	finalBytes, err := io.ReadAll(finalResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read final response: %v", err)
	}

	// Debug: Print the raw response
	if c.debug {
		fmt.Printf("Raw %s response: %s\n", modelName, string(finalBytes))
	}

	// Create a response object based on the model type
	var ghiblifyResp GhiblifyResponse

	if modelName == "ghiblify" {
		// Parse Ghiblify response
		if err := json.Unmarshal(finalBytes, &ghiblifyResp); err != nil {
			return nil, fmt.Errorf("failed to decode Ghiblify response: %v", err)
		}

		// Debug: Print the parsed response
		if c.debug {
			fmt.Printf("Parsed Ghiblify response: %+v\n", ghiblifyResp)
		}

		// Validate response
		if ghiblifyResp.Image.URL == "" {
			return nil, fmt.Errorf("received empty image URL in Ghiblify response")
		}
	} else if modelName == "cartoonify" {
		// Parse Cartoonify response
		var cartoonifyResp CartoonifyResponse
		if err := json.Unmarshal(finalBytes, &cartoonifyResp); err != nil {
			return nil, fmt.Errorf("failed to decode Cartoonify response: %v", err)
		}

		// Debug: Print the parsed response
		if c.debug {
			fmt.Printf("Parsed Cartoonify response: %+v\n", cartoonifyResp)
		}

		// Validate response
		if len(cartoonifyResp.Images) == 0 || cartoonifyResp.Images[0].URL == "" {
			return nil, fmt.Errorf("received empty image URL in Cartoonify response")
		}

		// Convert Cartoonify response to Ghiblify response format
		ghiblifyResp.Image.URL = cartoonifyResp.Images[0].URL
		ghiblifyResp.Image.ContentType = cartoonifyResp.Images[0].ContentType
		ghiblifyResp.Image.Width = cartoonifyResp.Images[0].Width
		ghiblifyResp.Image.Height = cartoonifyResp.Images[0].Height
	} else {
		return nil, fmt.Errorf("unsupported model: %s", modelName)
	}

	return &ghiblifyResp, nil
}
