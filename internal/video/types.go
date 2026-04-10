package video

import (
	braibottypes "github.com/karamble/braibot/internal/types"
)

// VideoRequest represents a request to generate a video
type VideoRequest struct {
	braibottypes.GenerationRequest
	Prompt                   string
	ImageURL                 string   // Optional, used by some image2video models (Veo2, Kling)
	SubjectReferenceImageURL string   // Optional, used by minimax-subject-reference
	Duration                 string   // Optional, defaults handled by FAL
	AspectRatio              string   // Optional, defaults handled by FAL
	Resolution               string   // Optional, defaults handled by FAL
	NegativePrompt           string   // Optional, defaults handled by FAL
	CFGScale                 *float64 // Optional, use pointer to track if set
	PromptOptimizer          *bool    // Optional, for minimax-director model
	GenerateAudio            *bool    // Optional, for Kling v3/O3 audio toggle
	EndImageURL              string   // Optional, for Kling v3 image2video end frame
	VideoURL                 string   // Required for video2video edit models
	KeepAudio                *bool    // Optional, for O3 edit (default: true)
	ImageURLs                []string // Optional, up to 4 reference images for O3 edit / up to 9 for Seedance multi2video
	VideoURLs                []string // Optional, up to 3 reference videos for Seedance multi2video
	AudioURLs                []string // Optional, up to 3 reference audio files for Seedance multi2video
	Seed                     *int64   // Optional, for reproducibility (Seedance 2.0)
}

// VideoResult represents the result of a video generation
type VideoResult struct {
	VideoURL string
	Success  bool
	Error    error
}
