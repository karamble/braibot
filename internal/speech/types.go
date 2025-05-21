package speech

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// SpeechRequest represents an internal request to generate speech
type SpeechRequest struct {
	Text      string
	VoiceID   string // Optional, specific voice
	ModelName string // Target model name
	Progress  fal.ProgressCallback
	UserNick  string
	UserID    zkidentity.ShortID
	PriceUSD  float64
	IsPM      bool   // Whether this is a private message
	GC        string // Group chat name if not PM
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
