package fal

import (
	"context"
	"encoding/json"
	"fmt"
	// "io" // No longer needed directly here
)

// GenerateVideo sends a request to the video model and returns the video URL
func (c *Client) GenerateVideo(ctx context.Context, req interface{}) (*VideoResponse, error) {
	var modelName string
	var endpoint string
	var reqBody map[string]interface{}
	var progress ProgressCallback

	// Extract progress callback if available
	if progressable, ok := req.(Progressable); ok {
		progress = progressable.GetProgress()
	}

	// Determine model name, endpoint and create request body based on request type
	switch r := req.(type) {
	case *Veo2Request:
		modelName = "veo2"
		endpoint = "/veo2/image-to-video"
		// Get model options
		model, exists := GetModel(modelName, "image2video") // Veo2 is image2video
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
		// Kling can be image2video or text2video, need to check model registry
		// Let's assume the specific type (image or text) is handled by caller setting r.Model correctly if BaseVideoRequest is embedded
		if r.ImageURL != "" { // Likely Image-to-Video
			modelName = "kling-video-image"
			endpoint = "/kling-video/v2/master/image-to-video"
		} else if r.Prompt != "" { // Likely Text-to-Video
			modelName = "kling-video-text"
			endpoint = "/kling-video/v2/master/text-to-video"
		} else {
			return nil, fmt.Errorf("kling request must have either image_url or prompt")
		}

		// Get model options (assuming KlingVideoOptions is used for both)
		model, exists := GetModel(modelName, "image2video") // Check image2video first
		if !exists {
			model, exists = GetModel(modelName, "text2video") // Check text2video if not found
			if !exists {
				return nil, fmt.Errorf("model not found: %s", modelName)
			}
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
			"prompt":          r.Prompt,   // May be empty for image2video
			"image_url":       r.ImageURL, // May be empty for text2video
			"duration":        r.Duration,
			"aspect_ratio":    r.AspectRatio,
			"negative_prompt": r.NegativePrompt,
			"cfg_scale":       r.CFGScale,
		}
		// Remove empty fields that are not applicable
		if r.ImageURL == "" {
			delete(reqBody, "image_url")
		}
		if r.Prompt == "" {
			delete(reqBody, "prompt")
		}
	case *BaseVideoRequest: // Handle potentially ambiguous base request
		modelName = r.Model
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			model, exists = GetModel(modelName, "text2video")
			if !exists {
				return nil, fmt.Errorf("model not found: %s", modelName)
			}
		}
		// Determine endpoint based on retrieved model
		switch model.Name {
		case "veo2":
			endpoint = "/veo2/image-to-video"
		case "kling-video-image":
			endpoint = "/kling-video/v2/master/image-to-video"
		case "kling-video-text":
			endpoint = "/kling-video/v2/master/text-to-video"
		default:
			return nil, fmt.Errorf("unsupported model: %s", model.Name)
		}
		reqBody = map[string]interface{}{ // Assume base fields
			"prompt":    r.Prompt,
			"image_url": r.ImageURL,
		}
		// Remove empty fields
		if r.ImageURL == "" {
			delete(reqBody, "image_url")
		}
		if r.Prompt == "" {
			delete(reqBody, "prompt")
		}

	// Deprecated: TextToVideoRequest is handled by KlingVideoRequest case now
	// case *TextToVideoRequest: ...

	default:
		return nil, fmt.Errorf("unsupported request type: %T", req)
	}

	// Add any additional generic options from the request
	if optionsGetter, ok := req.(interface{ GetOptions() map[string]interface{} }); ok {
		for k, v := range optionsGetter.GetOptions() {
			// Avoid overwriting fields already set by specific request types
			if _, exists := reqBody[k]; !exists {
				reqBody[k] = v
			}
		}
	}

	// Define the decoder for the final video response
	decodeFunc := func(data []byte) (interface{}, error) {
		var videoResp VideoResponse
		if err := json.Unmarshal(data, &videoResp); err != nil {
			return nil, fmt.Errorf("failed to parse final video response: %w, body: %s", err, string(data))
		}
		// Check if any of the video URL fields are populated
		if videoResp.GetURL() == "" {
			return nil, fmt.Errorf("no video URL found in the response: %s", string(data))
		}
		return &videoResp, nil
	}

	// Execute the workflow
	result, err := c.executeAsyncWorkflow(ctx, endpoint, reqBody, progress, decodeFunc)
	if err != nil {
		return nil, err // Error already wrapped
	}

	return result.(*VideoResponse), nil
}
