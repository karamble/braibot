package speech

import (
	braibottypes "github.com/karamble/braibot/internal/types"
)

// SpeechRequest represents an internal request to generate speech
type SpeechRequest struct {
	braibottypes.GenerationRequest
	Text    string
	VoiceID string // Optional, specific voice
	// Parsed Options
	Speed      *float64
	Vol        *float64
	Pitch      *int
	Emotion    string
	SampleRate string
	Bitrate    string
	Format     string
	Channel    string
}

// SpeechResult represents the result of a speech generation
type SpeechResult struct {
	AudioURL string // URL of the generated audio
	Success  bool
	Error    error
}
