// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
	// "io" // No longer needed directly here
)

// GenerateImage generates an image from a text prompt or image url
func (c *Client) GenerateImage(ctx context.Context, req ImageRequest) (*ImageResponse, error) {
	// Validate model type (text2image or image2image)
	model, exists := GetModel(req.Model, "text2image")
	if !exists {
		model, exists = GetModel(req.Model, "image2image")
		if !exists {
			return nil, &Error{
				Code:    "INVALID_MODEL",
				Message: "invalid or unsupported model type for image generation: " + req.Model,
			}
		}
	}

	// Create request body
	reqBody := map[string]interface{}{}
	if model.Type == "text2image" {
		reqBody["prompt"] = req.Prompt
	} else { // image2image specific fields
		if imgURL, ok := req.Options["image_url"]; ok {
			reqBody["image_url"] = imgURL
		} else {
			return nil, fmt.Errorf("image_url is required for image2image models")
		}
		// Add prompt if provided for image2image (some models might use it)
		if req.Prompt != "" {
			reqBody["prompt"] = req.Prompt
		}
	}

	// Add any additional common options
	for k, v := range req.Options {
		// Avoid overwriting primary fields like prompt or image_url
		if k != "prompt" && k != "image_url" {
			reqBody[k] = v
		}
	}

	// Define the decoder for the final image response
	decodeFunc := func(data []byte) (interface{}, error) {
		var response ImageResponse

		// Try parsing as standard response (with images array)
		if err := json.Unmarshal(data, &response); err == nil && len(response.Images) > 0 {
			return &response, nil
		}

		// If that fails, try parsing as response with single 'image' or 'svg' field
		var singleImageResp struct {
			Image struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				Width       int    `json:"width"`
				Height      int    `json:"height"`
			} `json:"image"`
			SVG struct { // Handle star-vector potentially returning SVG
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
			} `json:"svg"`
		}

		if err := json.Unmarshal(data, &singleImageResp); err != nil {
			// If both attempts fail, return the original unmarshal error
			return nil, fmt.Errorf("failed to parse final response as known image format: %w. Body: %s", err, string(data))
		}

		// Convert single image/svg response to ImageResponse format
		imgURL := singleImageResp.Image.URL
		imgContentType := singleImageResp.Image.ContentType
		imgWidth := singleImageResp.Image.Width
		imgHeight := singleImageResp.Image.Height

		if singleImageResp.SVG.URL != "" { // Prioritize SVG if present
			imgURL = singleImageResp.SVG.URL
			imgContentType = singleImageResp.SVG.ContentType
			imgWidth = 0 // SVG doesn't have inherent width/height in this context
			imgHeight = 0
		}

		if imgURL == "" {
			return nil, fmt.Errorf("final response did not contain a valid image or svg URL. Body: %s", string(data))
		}

		response.Images = []struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
		}{
			{
				URL:         imgURL,
				ContentType: imgContentType,
				Width:       imgWidth,
				Height:      imgHeight,
			},
		}
		return &response, nil
	}

	// Execute the workflow
	result, err := c.executeAsyncWorkflow(ctx, "/"+req.Model, reqBody, req.Progress, decodeFunc)
	if err != nil {
		return nil, err // Error already wrapped by executeAsyncWorkflow
	}

	return result.(*ImageResponse), nil
}
