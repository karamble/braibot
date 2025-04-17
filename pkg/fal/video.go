package fal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GenerateVideo sends a request to the video model and returns the video URL
func (c *Client) GenerateVideo(ctx context.Context, req interface{}) (*VideoResponse, error) {
	var modelName string
	var reqBody map[string]interface{}

	// Determine model name and create request body based on request type
	switch r := req.(type) {
	case *Veo2Request:
		modelName = "veo2"
		// Get model options
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*Veo2Options)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}

		// Set default values from model options if not provided
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.Duration == "" {
			r.Duration = options.Duration
		}

		// Validate options
		if err := options.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options: %v", err)
		}

		reqBody = map[string]interface{}{
			"prompt":       r.Prompt,
			"image_url":    r.ImageURL,
			"aspect_ratio": r.AspectRatio,
			"duration":     r.Duration,
		}
	case *KlingVideoRequest:
		modelName = "kling-video"
		// Get model options
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*KlingVideoOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}

		// Set default values from model options if not provided
		if r.Duration == "" {
			r.Duration = options.Duration
		}
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.NegativePrompt == "" {
			r.NegativePrompt = options.NegativePrompt
		}
		if r.CFGScale == 0 {
			r.CFGScale = options.CFGScale
		}

		// Validate options
		if err := options.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options: %v", err)
		}

		reqBody = map[string]interface{}{
			"prompt":          r.Prompt,
			"image_url":       r.ImageURL,
			"duration":        r.Duration,
			"aspect_ratio":    r.AspectRatio,
			"negative_prompt": r.NegativePrompt,
			"cfg_scale":       r.CFGScale,
		}
	case *BaseVideoRequest:
		// Handle legacy VideoRequest type
		modelName = r.Model
		reqBody = map[string]interface{}{
			"prompt":    r.Prompt,
			"image_url": r.ImageURL,
		}
	default:
		return nil, fmt.Errorf("unsupported request type: %T", req)
	}

	// Get the model to determine which endpoint to use
	model, exists := GetModel(modelName, "image2video")
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelName)
	}

	var endpoint string
	switch model.Name {
	case "veo2":
		endpoint = "/veo2/image-to-video"
	case "kling-video":
		endpoint = "/kling-video/v2/master/image-to-video"
	default:
		return nil, fmt.Errorf("unsupported model: %s", model.Name)
	}

	// Add any additional options
	if baseReq, ok := req.(interface{ GetOptions() map[string]interface{} }); ok {
		for k, v := range baseReq.GetOptions() {
			reqBody[k] = v
		}
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the initial queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}

	if c.debug {
		fmt.Printf("DEBUG - Queue Response:\n")
		fmt.Printf("  Queue ID: %s\n", queueResp.QueueID)
		fmt.Printf("  Status: %s\n", queueResp.Status)
		fmt.Printf("  Position: %d\n", queueResp.Position)
		fmt.Printf("  ETA: %d seconds\n", queueResp.ETA)
		fmt.Printf("  Response URL: %s\n", queueResp.ResponseURL)
	}

	// Notify about queue position
	if progressable, ok := req.(Progressable); ok {
		c.notifyQueuePosition(ctx, queueResp, progressable.GetProgress())
	}

	// Poll for completion
	if progressable, ok := req.(Progressable); ok {
		_, err = c.pollQueueStatus(ctx, queueResp, progressable.GetProgress())
	} else {
		_, err = c.pollQueueStatus(ctx, queueResp, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to poll queue status: %v", err)
	}

	// Get final response using the response URL
	finalResp, err := c.makeRequest(ctx, "GET", queueResp.ResponseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %v", err)
	}
	defer finalResp.Body.Close()

	// Read final response body
	finalBytes, err := io.ReadAll(finalResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read final response: %v", err)
	}

	// Debug log the response
	if c.debug {
		fmt.Printf("DEBUG - Final response body: %s\n", string(finalBytes))
	}

	// Parse the response
	var videoResp VideoResponse
	if err := json.Unmarshal(finalBytes, &videoResp); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v, body: %s", err, string(finalBytes))
	}

	// Check if any of the video URL fields are populated
	if videoResp.Video.URL == "" && videoResp.URL == "" && videoResp.VideoURL == "" {
		return nil, fmt.Errorf("no video URL found in the response: %s", string(finalBytes))
	}

	return &videoResp, nil
}
