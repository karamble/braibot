package fal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GenerateVideo sends a request to the kling-video model and returns the video URL
func (c *Client) GenerateVideo(ctx context.Context, req VideoRequest) (*VideoResponse, error) {
	// Set default values if not provided
	if req.Duration == "" {
		req.Duration = "5"
	}
	if req.AspectRatio == "" {
		req.AspectRatio = "16:9"
	}
	if req.NegativePrompt == "" {
		req.NegativePrompt = "blur, distort, and low quality"
	}
	if req.CFGScale == 0 {
		req.CFGScale = 0.5
	}

	// Create request body
	reqBody := map[string]interface{}{
		"prompt":          req.Prompt,
		"image_url":       req.ImageURL,
		"duration":        req.Duration,
		"aspect_ratio":    req.AspectRatio,
		"negative_prompt": req.NegativePrompt,
		"cfg_scale":       req.CFGScale,
	}

	// Add any additional options
	for k, v := range req.Options {
		reqBody[k] = v
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/kling-video/v2/master/image-to-video", reqBody)
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
