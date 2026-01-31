package fal

import (
	"context"
	"encoding/json"
	"fmt"
	// "io" // No longer needed directly here
	// "strings" // Removed, no longer needed
	// "time" // Removed, no longer needed
)

// GenerateVideo sends a request to the video model and returns the video URL
func (c *Client) GenerateVideo(ctx context.Context, req interface{}) (*VideoResponse, error) {
	var modelName string
	var endpoint string
	var reqBody map[string]interface{}
	var progress ProgressCallback
	var queueInfo QueueInfoCallback

	// Extract progress callback if available
	if progressable, ok := req.(Progressable); ok {
		progress = progressable.GetProgress()
	}

	// Extract queue info callback if available (for recovery)
	if queueInfoable, ok := req.(QueueInfoable); ok {
		queueInfo = queueInfoable.GetQueueInfo()
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
		// Validate options before proceeding
		if err := options.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values from model options if not provided in request
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.Duration == "" {
			r.Duration = options.Duration
		}
		reqBody = map[string]interface{}{
			"prompt":       r.Prompt,
			"image_url":    r.ImageURL,
			"aspect_ratio": r.AspectRatio,
			"duration":     r.Duration,
		}
	case *KlingVideoRequest:
		if r.BaseVideoRequest.Model == "" { // Determine model based on fields if not set
			if r.BaseVideoRequest.ImageURL != "" {
				r.BaseVideoRequest.Model = "kling-video-image"
			} else {
				r.BaseVideoRequest.Model = "kling-video-text"
			}
		}
		modelName = r.BaseVideoRequest.Model
		model, exists := GetModel(modelName, "text2video") // Check both types
		modelTypeForPath := "text-to-video"                // Default for path construction
		if !exists {
			model, exists = GetModel(modelName, "image2video")
			modelTypeForPath = "image-to-video" // Update path type if found as image2video
		}
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		// Use modelTypeForPath (image-to-video or text-to-video) for the endpoint path
		endpoint = "/kling-video/v2/master/" + modelTypeForPath

		// Get model options for validation and defaults
		options, ok := model.Options.(*KlingVideoOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate options before proceeding
		klingOpts := KlingVideoOptions{
			Duration:       r.Duration,
			AspectRatio:    r.AspectRatio,
			NegativePrompt: r.NegativePrompt,
			CFGScale:       r.CFGScale,
		}
		if err := klingOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}

		// Set default values if not provided
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
		// Remove empty fields only if ImageURL was expected but not provided
		// Use the original modelType for this check (image2video)
		var originalModelType string
		if r.BaseVideoRequest.ImageURL != "" {
			originalModelType = "image2video"
		} else {
			originalModelType = "text2video"
		}
		if originalModelType == "image2video" && r.ImageURL == "" {
			delete(reqBody, "image_url") // Should ideally be caught earlier
		}
	case *MinimaxDirectorRequest:
		modelName = "minimax/video-01-director"
		endpoint = "/minimax/video-01-director" // Corrected relative path
		model, exists := GetModel(modelName, "text2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*MinimaxDirectorOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate options
		miniOpts := MinimaxDirectorOptions{
			PromptOptimizer: r.PromptOptimizer,
		}
		if err := miniOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}

		// Set default values if not provided
		// Need pointer comparison for boolean option
		if r.PromptOptimizer == nil {
			r.PromptOptimizer = options.PromptOptimizer // Get default from model definition
		}

		reqBody = map[string]interface{}{
			"prompt":           r.Prompt,
			"prompt_optimizer": r.PromptOptimizer,
		}
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt cannot be empty for %s", modelName)
		}
	case *MinimaxSubjectReferenceRequest:
		modelName = "minimax/video-01-subject-reference"
		endpoint = "/minimax/video-01-subject-reference" // Correct relative path
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*MinimaxSubjectReferenceOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate options
		subRefOpts := MinimaxSubjectReferenceOptions{
			PromptOptimizer: r.PromptOptimizer,
		}
		if err := subRefOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values if not provided
		if r.PromptOptimizer == nil {
			r.PromptOptimizer = options.PromptOptimizer
		}
		reqBody = map[string]interface{}{
			"prompt":                      r.Prompt,
			"subject_reference_image_url": r.SubjectReferenceImageURL,
			"prompt_optimizer":            r.PromptOptimizer,
		}
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt cannot be empty for %s", modelName)
		}
		if r.SubjectReferenceImageURL == "" {
			return nil, fmt.Errorf("subject_reference_image_url cannot be empty for %s", modelName)
		}
	case *MinimaxLiveRequest:
		modelName = "minimax/video-01-live"
		endpoint = "/minimax/video-01-live/image-to-video" // Correct relative path
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*MinimaxLiveOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate options
		liveOpts := MinimaxLiveOptions{
			PromptOptimizer: r.PromptOptimizer,
		}
		if err := liveOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values if not provided
		if r.PromptOptimizer == nil {
			r.PromptOptimizer = options.PromptOptimizer
		}
		reqBody = map[string]interface{}{
			"prompt":           r.Prompt,
			"image_url":        r.ImageURL,
			"prompt_optimizer": r.PromptOptimizer,
		}
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt cannot be empty for %s", modelName)
		}
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url cannot be empty for %s", modelName)
		}
	case *MinimaxVideo01Request:
		modelName = "minimax/video-01"
		endpoint = "/minimax/video-01" // Correct relative path
		model, exists := GetModel(modelName, "text2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*MinimaxVideo01Options)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate options
		vidOpts := MinimaxVideo01Options{
			PromptOptimizer: r.PromptOptimizer,
		}
		if err := vidOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values if not provided
		if r.PromptOptimizer == nil {
			r.PromptOptimizer = options.PromptOptimizer
		}
		reqBody = map[string]interface{}{
			"prompt":           r.Prompt,
			"prompt_optimizer": r.PromptOptimizer,
		}
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt cannot be empty for %s", modelName)
		}
		// Ensure ImageURL is not included (should be empty in BaseVideoRequest for text2video)
		delete(reqBody, "image_url")
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
		case "veo3":
			endpoint = "/veo3/image-to-video"
		case "veo31fast":
			endpoint = "/veo3.1/fast/image-to-video"
		case "kling-video-image":
			endpoint = "/kling-video/v2/master/image-to-video"
		case "kling-video-text":
			endpoint = "/kling-video/v2/master/text-to-video"
		case "minimax/video-01-director":
			endpoint = "/minimax/video-01-director"
		case "minimax/video-01-subject-reference":
			endpoint = "/minimax/video-01-subject-reference"
		case "minimax/video-01-live":
			endpoint = "/minimax/video-01-live/image-to-video"
		case "minimax/video-01":
			endpoint = "/minimax/video-01"
		case "minimax/hailuo-02":
			endpoint = "/minimax/hailuo-02/standard/text-to-video"
		case "grok-imagine-video":
			endpoint = "https://queue.fal.run/xai/grok-imagine-video/image-to-video"
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
	case *MinimaxHailuo02Request:
		modelName = "minimax/hailuo-02"
		endpoint = "/minimax/hailuo-02/standard/text-to-video"
		model, exists := GetModel(modelName, "text2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*MinimaxHailuo02Options)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Set defaults from options if not provided
		defaults := options.GetDefaultValues()
		if r.Duration == "" {
			r.Duration = defaults["duration"].(string)
		}
		if r.PromptOptimizer == nil {
			r.PromptOptimizer = defaults["prompt_optimizer"].(*bool)
		}
		// Validate using the options struct
		opts := MinimaxHailuo02Options{
			Duration:        r.Duration,
			PromptOptimizer: r.PromptOptimizer,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		reqBody = map[string]interface{}{
			"prompt":           r.Prompt,
			"duration":         r.Duration,
			"prompt_optimizer": r.PromptOptimizer,
		}
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt cannot be empty for %s", modelName)
		}
	case *Veo3Request:
		modelName = "veo3"
		endpoint = "/veo3/image-to-video"
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*Veo3Options)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate request options
		veo3Opts := Veo3Options{
			AspectRatio:   r.AspectRatio,
			Duration:      r.Duration,
			Resolution:    r.Resolution,
			GenerateAudio: r.GenerateAudio,
			AutoFix:       r.AutoFix,
		}
		if err := veo3Opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values from model options if not provided in request
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.Duration == "" {
			r.Duration = options.Duration
		}
		if r.Resolution == "" {
			r.Resolution = options.Resolution
		}
		if r.GenerateAudio == nil {
			r.GenerateAudio = options.GenerateAudio
		}
		if r.AutoFix == nil {
			r.AutoFix = options.AutoFix
		}
		// Validate image_url is required
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s", modelName)
		}
		// Build request body
		reqBody = map[string]interface{}{
			"prompt":    r.Prompt,
			"image_url": r.ImageURL,
		}
		// Add optional fields only if they have values
		if r.AspectRatio != "" {
			reqBody["aspect_ratio"] = r.AspectRatio
		}
		if r.Duration != "" {
			reqBody["duration"] = r.Duration
		}
		if r.Resolution != "" {
			reqBody["resolution"] = r.Resolution
		}
		if r.GenerateAudio != nil {
			reqBody["generate_audio"] = *r.GenerateAudio
		}
		if r.AutoFix != nil {
			reqBody["auto_fix"] = *r.AutoFix
		}
	case *Veo31FastRequest:
		modelName = "veo31fast"
		endpoint = "/veo3.1/fast/image-to-video"
		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		options, ok := model.Options.(*Veo31FastOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}
		// Validate request options
		veo31FastOpts := Veo31FastOptions{
			AspectRatio:   r.AspectRatio,
			Duration:      r.Duration,
			Resolution:    r.Resolution,
			GenerateAudio: r.GenerateAudio,
			AutoFix:       r.AutoFix,
		}
		if err := veo31FastOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}
		// Set default values from model options if not provided in request
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.Duration == "" {
			r.Duration = options.Duration
		}
		if r.Resolution == "" {
			r.Resolution = options.Resolution
		}
		if r.GenerateAudio == nil {
			r.GenerateAudio = options.GenerateAudio
		}
		if r.AutoFix == nil {
			r.AutoFix = options.AutoFix
		}
		// Validate image_url is required
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s", modelName)
		}
		// Build request body
		reqBody = map[string]interface{}{
			"prompt":    r.Prompt,
			"image_url": r.ImageURL,
		}
		// Add optional fields only if they have values
		if r.AspectRatio != "" {
			reqBody["aspect_ratio"] = r.AspectRatio
		}
		if r.Duration != "" {
			reqBody["duration"] = r.Duration
		}
		if r.Resolution != "" {
			reqBody["resolution"] = r.Resolution
		}
		if r.GenerateAudio != nil {
			reqBody["generate_audio"] = *r.GenerateAudio
		}
		if r.AutoFix != nil {
			reqBody["auto_fix"] = *r.AutoFix
		}
	case *KlingVideoV26MotionControlRequest:
		modelName = "kling-video-v26-motion-control"
		endpoint = "/kling-video/v2.6/standard/motion-control"

		model, exists := GetModel(modelName, "video2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}

		options, ok := model.Options.(*KlingVideoV26MotionControlOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}

		// Validate required fields
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s", modelName)
		}
		if r.VideoURL == "" {
			return nil, fmt.Errorf("video_url is required for %s", modelName)
		}
		if r.CharacterOrientation == "" {
			return nil, fmt.Errorf("character_orientation is required for %s (must be 'image' or 'video')", modelName)
		}

		// Validate options
		opts := KlingVideoV26MotionControlOptions{
			CharacterOrientation: r.CharacterOrientation,
			KeepOriginalSound:    r.KeepOriginalSound,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}

		// Set defaults if not provided
		if r.KeepOriginalSound == nil {
			r.KeepOriginalSound = options.KeepOriginalSound
		}

		// Build request body
		reqBody = map[string]interface{}{
			"image_url":             r.ImageURL,
			"video_url":             r.VideoURL,
			"character_orientation": r.CharacterOrientation,
		}

		// Add optional fields
		if r.Prompt != "" {
			reqBody["prompt"] = r.Prompt
		}
		if r.KeepOriginalSound != nil {
			reqBody["keep_original_sound"] = *r.KeepOriginalSound
		}
	case *GrokImagineVideoRequest:
		modelName = "grok-imagine-video"
		endpoint = "https://queue.fal.run/xai/grok-imagine-video/image-to-video"

		model, exists := GetModel(modelName, "image2video")
		if !exists {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}

		options, ok := model.Options.(*GrokImagineVideoOptions)
		if !ok {
			return nil, fmt.Errorf("invalid options type for model %s", modelName)
		}

		// Validate required fields
		if r.Prompt == "" {
			return nil, fmt.Errorf("prompt is required for %s", modelName)
		}
		if r.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s", modelName)
		}

		// Validate options
		opts := GrokImagineVideoOptions{
			Duration:    r.Duration,
			AspectRatio: r.AspectRatio,
			Resolution:  r.Resolution,
		}
		if err := opts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}

		// Set defaults if not provided
		if r.Duration == 0 {
			r.Duration = options.Duration
		}
		if r.AspectRatio == "" {
			r.AspectRatio = options.AspectRatio
		}
		if r.Resolution == "" {
			r.Resolution = options.Resolution
		}

		// Build request body
		reqBody = map[string]interface{}{
			"prompt":       r.Prompt,
			"image_url":    r.ImageURL,
			"duration":     r.Duration,
			"aspect_ratio": r.AspectRatio,
			"resolution":   r.Resolution,
		}
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

	// Execute the workflow (with queue info callback for recovery support)
	result, err := c.executeAsyncWorkflowWithCallback(ctx, endpoint, reqBody, progress, decodeFunc, queueInfo)
	if err != nil {
		return nil, err // Error already wrapped
	}

	return result.(*VideoResponse), nil
}
