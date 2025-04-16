// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GenerateImage generates an image from a text prompt
func (c *Client) GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	// Validate model
	if _, exists := GetModel(req.Model, "text2image"); !exists {
		// Try image2image type if text2image not found
		if _, exists := GetModel(req.Model, "image2image"); !exists {
			return nil, &Error{
				Code:    "INVALID_MODEL",
				Message: "invalid model for text2image/image2image: " + req.Model,
			}
		}
	}

	// Create request body
	reqBody := map[string]interface{}{
		"prompt": req.Prompt,
	}

	// Add any additional options
	for k, v := range req.Options {
		reqBody[k] = v
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/"+req.Model, reqBody)
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
	c.notifyQueuePosition(ctx, queueResp, req.Progress)

	// Poll for completion
	_, err = c.pollQueueStatus(ctx, queueResp, req.Progress)
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
	var response ImageResponse

	// Try parsing as text2image response first (with images array)
	if err := json.Unmarshal(finalBytes, &response); err == nil && len(response.Images) > 0 {
		return &response, nil
	}

	// If that fails, try parsing as image2image response (with single image)
	var image2ImageResp struct {
		Image struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			FileName    string `json:"file_name"`
			FileSize    int    `json:"file_size"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
		} `json:"image"`
	}

	if err := json.Unmarshal(finalBytes, &image2ImageResp); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v", err)
	}

	// Convert image2image response to ImageResponse format
	response.Images = []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	}{
		{
			URL:         image2ImageResp.Image.URL,
			ContentType: image2ImageResp.Image.ContentType,
			Width:       image2ImageResp.Image.Width,
			Height:      image2ImageResp.Image.Height,
		},
	}

	if c.debug {
		fmt.Printf("DEBUG - Image Response:\n")
		fmt.Printf("  Number of images: %d\n", len(response.Images))
		for i, img := range response.Images {
			fmt.Printf("  Image %d:\n", i+1)
			fmt.Printf("    URL: %s\n", img.URL)
			fmt.Printf("    Content Type: %s\n", img.ContentType)
			fmt.Printf("    Width: %d\n", img.Width)
			fmt.Printf("    Height: %d\n", img.Height)
		}
	}

	return &response, nil
}
