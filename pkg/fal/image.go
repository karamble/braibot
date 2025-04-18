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
// It accepts specific request types like *FastSDXLRequest or *GhiblifyRequest.
func (c *Client) GenerateImage(ctx context.Context, req interface{}) (*ImageResponse, error) {
	var modelName string
	var modelType string
	var endpoint string
	var reqBody map[string]interface{}
	var progress ProgressCallback
	var baseReq *BaseImageRequest

	// Extract progress callback and base request details
	if progressable, ok := req.(Progressable); ok {
		progress = progressable.GetProgress()
	}

	// Determine model name, endpoint and create request body based on request type
	switch r := req.(type) {
	case *FastSDXLRequest:
		modelName = "fast-sdxl"
		modelType = "text2image"
		endpoint = "/fast-sdxl"
		baseReq = &r.BaseImageRequest
		reqBody = map[string]interface{}{
			"prompt": r.Prompt,
		}
		if r.NumImages > 0 {
			reqBody["num_images"] = r.NumImages
		}
		r.Model = modelName // Set model name internally
	case *GhiblifyRequest:
		modelName = "ghiblify"
		modelType = "image2image"
		endpoint = "/ghiblify"
		baseReq = &r.BaseImageRequest
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s model", modelName)
		}
		reqBody = map[string]interface{}{
			"image_url": r.ImageURL,
		}
		// Ghiblify might optionally use prompt, add if present
		if r.Prompt != "" {
			reqBody["prompt"] = r.Prompt
		}
		r.Model = modelName // Set model name internally
	case *FluxSchnellRequest:
		modelName = "flux/schnell"
		modelType = "text2image"
		endpoint = "/flux/schnell"
		baseReq = &r.BaseImageRequest
		// Validate specific options
		opts := FluxSchnellOptions{
			ImageSize:           r.ImageSize,
			NumInferenceSteps:   r.NumInferenceSteps,
			Seed:                r.Seed,
			SyncMode:            r.SyncMode,
			NumImages:           r.NumImages,
			EnableSafetyChecker: r.EnableSafetyChecker,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body from the specific request struct
		reqBody = map[string]interface{}{
			"prompt": r.Prompt, // From BaseImageRequest
		}
		if r.ImageSize != "" {
			reqBody["image_size"] = r.ImageSize
		}
		if r.NumInferenceSteps > 0 { // Only include if non-default might be intended
			reqBody["num_inference_steps"] = r.NumInferenceSteps
		}
		if r.Seed != nil {
			reqBody["seed"] = *r.Seed
		}
		if r.SyncMode { // Only include if true
			reqBody["sync_mode"] = r.SyncMode
		}
		if r.NumImages > 0 { // Only include if non-default might be intended
			reqBody["num_images"] = r.NumImages
		}
		if r.EnableSafetyChecker != nil { // Include if explicitly set
			reqBody["enable_safety_checker"] = *r.EnableSafetyChecker
		}
		r.Model = modelName // Set model name internally
	case *FluxProV1_1Request:
		modelName = "flux-pro/v1.1"
		modelType = "text2image"
		endpoint = "/flux-pro/v1.1"
		baseReq = &r.BaseImageRequest
		// Validate specific options
		opts := FluxProV1_1Options{
			ImageSize:           r.ImageSize,
			Seed:                r.Seed,
			SyncMode:            r.SyncMode,
			NumImages:           r.NumImages,
			EnableSafetyChecker: r.EnableSafetyChecker,
			SafetyTolerance:     r.SafetyTolerance,
			OutputFormat:        r.OutputFormat,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body
		reqBody = map[string]interface{}{
			"prompt": r.Prompt,
		}
		if r.ImageSize != "" {
			reqBody["image_size"] = r.ImageSize
		}
		if r.Seed != nil {
			reqBody["seed"] = *r.Seed
		}
		if r.SyncMode {
			reqBody["sync_mode"] = r.SyncMode
		}
		if r.NumImages > 0 {
			reqBody["num_images"] = r.NumImages
		}
		if r.EnableSafetyChecker != nil {
			reqBody["enable_safety_checker"] = *r.EnableSafetyChecker
		}
		if r.SafetyTolerance != "" {
			reqBody["safety_tolerance"] = r.SafetyTolerance
		}
		if r.OutputFormat != "" {
			reqBody["output_format"] = r.OutputFormat
		}
		r.Model = modelName // Set model name internally
	// HiDream Models (assuming shared parameters)
	case *HiDreamI1FullRequest, *HiDreamI1DevRequest, *HiDreamI1FastRequest:
		// Determine modelName and get concrete request struct pointer
		var concreteReq *HiDreamI1FullRequest
		switch reqTyped := req.(type) { // Use new variable reqTyped
		case *HiDreamI1FullRequest:
			modelName = "hidream-i1-full"
			concreteReq = reqTyped
		case *HiDreamI1DevRequest:
			modelName = "hidream-i1-dev"
			concreteReq = &reqTyped.HiDreamI1FullRequest
		case *HiDreamI1FastRequest:
			modelName = "hidream-i1-fast"
			concreteReq = &reqTyped.HiDreamI1FullRequest
		default: // Should not happen due to outer case
			return nil, fmt.Errorf("unexpected type within HiDream case: %T", req)
		}
		modelType = "text2image"
		endpoint = "/" + modelName // Use model name as endpoint
		baseReq = &concreteReq.BaseImageRequest
		// Validate specific options
		opts := HiDreamOptions{
			NegativePrompt:      concreteReq.NegativePrompt,
			ImageSize:           concreteReq.ImageSize,
			NumInferenceSteps:   concreteReq.NumInferenceSteps,
			Seed:                concreteReq.Seed,
			GuidanceScale:       concreteReq.GuidanceScale,
			SyncMode:            concreteReq.SyncMode,
			NumImages:           concreteReq.NumImages,
			EnableSafetyChecker: concreteReq.EnableSafetyChecker,
			OutputFormat:        concreteReq.OutputFormat,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body
		reqBody = map[string]interface{}{"prompt": concreteReq.Prompt}
		if concreteReq.NegativePrompt != "" {
			reqBody["negative_prompt"] = concreteReq.NegativePrompt
		}
		if concreteReq.ImageSize != "" {
			reqBody["image_size"] = concreteReq.ImageSize
		}
		if concreteReq.NumInferenceSteps != nil {
			reqBody["num_inference_steps"] = *concreteReq.NumInferenceSteps
		}
		if concreteReq.Seed != nil {
			reqBody["seed"] = *concreteReq.Seed
		}
		// Only include guidance_scale for the 'full' model
		if modelName == "hidream-i1-full" && concreteReq.GuidanceScale != nil {
			reqBody["guidance_scale"] = *concreteReq.GuidanceScale
		}
		if concreteReq.SyncMode {
			reqBody["sync_mode"] = concreteReq.SyncMode
		}
		if concreteReq.NumImages > 0 {
			reqBody["num_images"] = concreteReq.NumImages
		}
		if concreteReq.EnableSafetyChecker != nil {
			reqBody["enable_safety_checker"] = *concreteReq.EnableSafetyChecker
		}
		if concreteReq.OutputFormat != "" {
			reqBody["output_format"] = concreteReq.OutputFormat
		}
		concreteReq.Model = modelName
	case *FluxProV1_1UltraRequest:
		modelName = "flux-pro/v1.1-ultra"
		modelType = "text2image"
		endpoint = "/" + modelName
		baseReq = &r.BaseImageRequest
		// Validate specific options
		opts := FluxProV1_1UltraOptions{
			Seed:                r.Seed,
			SyncMode:            r.SyncMode,
			NumImages:           r.NumImages,
			EnableSafetyChecker: r.EnableSafetyChecker,
			SafetyTolerance:     r.SafetyTolerance,
			OutputFormat:        r.OutputFormat,
			AspectRatio:         r.AspectRatio,
			Raw:                 r.Raw,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body
		reqBody = map[string]interface{}{"prompt": r.Prompt}
		if r.Seed != nil {
			reqBody["seed"] = *r.Seed
		}
		if r.SyncMode {
			reqBody["sync_mode"] = r.SyncMode
		}
		if r.NumImages > 0 {
			reqBody["num_images"] = r.NumImages
		}
		if r.EnableSafetyChecker != nil {
			reqBody["enable_safety_checker"] = *r.EnableSafetyChecker
		}
		if r.SafetyTolerance != "" {
			reqBody["safety_tolerance"] = r.SafetyTolerance
		}
		if r.OutputFormat != "" {
			reqBody["output_format"] = r.OutputFormat
		}
		if r.AspectRatio != "" {
			reqBody["aspect_ratio"] = r.AspectRatio
		}
		if r.Raw != nil {
			reqBody["raw"] = *r.Raw
		}
		r.Model = modelName
	// Image2Image Models
	case *CartoonifyRequest:
		modelName = "cartoonify"
		modelType = "image2image"
		endpoint = "/" + modelName
		baseReq = &r.BaseImageRequest
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url required for %s", modelName)
		}
		// Validate specific options (none currently)
		opts := CartoonifyOptions{}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body
		reqBody = map[string]interface{}{"image_url": r.ImageURL}
		if r.Prompt != "" {
			reqBody["prompt"] = r.Prompt
		} // Allow optional prompt
		r.Model = modelName
	case *StarVectorRequest:
		modelName = "star-vector"
		modelType = "image2image"
		endpoint = "/fal-ai/star-vector" // Assuming full path based on potential complexity
		baseReq = &r.BaseImageRequest
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url required for %s", modelName)
		}
		// Validate specific options (none currently)
		opts := StarVectorOptions{}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Build request body
		reqBody = map[string]interface{}{"image_url": r.ImageURL}
		r.Model = modelName
	// case *OtherImageRequest:
	// ...
	default:
		return nil, fmt.Errorf("unsupported image request type: %T", req)
	}

	// Validate the model existence (using inferred modelType)
	if _, exists := GetModel(modelName, modelType); !exists {
		return nil, &Error{
			Code:    "INVALID_MODEL",
			Message: fmt.Sprintf("model %s not found or not of expected type %s", modelName, modelType),
		}
	}

	// Add any additional generic options from the base request
	if baseReq != nil && baseReq.Options != nil {
		for k, v := range baseReq.Options {
			// Avoid overwriting fields already set by specific request types
			if _, exists := reqBody[k]; !exists {
				reqBody[k] = v
			}
		}
	}

	// Define the decoder for the final image response
	decodeFunc := func(data []byte) (interface{}, error) {
		var response ImageResponse

		// Try parsing as standard response (includes images array and top-level seed)
		if err := json.Unmarshal(data, &response); err == nil && len(response.Images) > 0 {
			// Seed is already captured in 'response' by the unmarshal
			return &response, nil
		}

		// If that fails, try parsing as response with single 'image' or 'svg' field
		// We also need to capture the top-level seed in this case.
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
			Seed int `json:"seed"` // Capture top-level seed here too
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
		response.Seed = singleImageResp.Seed // Assign the captured seed
		return &response, nil
	}

	// Execute the workflow
	result, err := c.executeAsyncWorkflow(ctx, endpoint, reqBody, progress, decodeFunc)
	if err != nil {
		return nil, err // Error already wrapped by executeAsyncWorkflow
	}

	return result.(*ImageResponse), nil
}
