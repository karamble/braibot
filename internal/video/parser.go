package video

import (
	"fmt"
	"strconv"
	"strings"
)

// ArgumentParser parses command arguments for video generation
type ArgumentParser struct{}

// NewArgumentParser creates a new ArgumentParser
func NewArgumentParser() *ArgumentParser {
	return &ArgumentParser{}
}

// ParseDuration parses the duration argument
func (p *ArgumentParser) ParseDuration(args []string) string {
	for i, arg := range args {
		if arg == "--duration" && i+1 < len(args) {
			return strings.TrimSuffix(args[i+1], "s")
		}
	}
	return "5" // Default duration
}

// ParseAspectRatio parses the aspect ratio argument
func (p *ArgumentParser) ParseAspectRatio(args []string) string {
	for i, arg := range args {
		if arg == "--aspect" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "16:9" // Default aspect ratio
}

// ParseNegativePrompt parses the negative prompt argument
func (p *ArgumentParser) ParseNegativePrompt(args []string) string {
	for i, arg := range args {
		if arg == "--negative_prompt" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "blur, distort, and low quality" // Default negative prompt
}

// ParseCFGScale parses the CFG scale argument
func (p *ArgumentParser) ParseCFGScale(args []string) float64 {
	for i, arg := range args {
		if arg == "--cfg_scale" && i+1 < len(args) {
			if scale, err := strconv.ParseFloat(args[i+1], 64); err == nil {
				return scale
			}
		}
	}
	return 0.5 // Default CFG scale
}

// ValidateOptions validates the video options
func (p *ArgumentParser) ValidateOptions(opts *VideoOptions) error {
	// Validate aspect ratio
	validAspectRatios := map[string]bool{
		"16:9": true,
		"9:16": true,
		"1:1":  true,
	}
	if !validAspectRatios[opts.AspectRatio] {
		return fmt.Errorf("invalid aspect ratio: %s (must be one of: 16:9, 9:16, 1:1)", opts.AspectRatio)
	}

	// Validate duration
	if opts.Duration != "5" {
		return fmt.Errorf("invalid duration: %s (must be 5 seconds)", opts.Duration)
	}

	// Validate CFG scale
	if opts.CFGScale < 0 || opts.CFGScale > 1 {
		return fmt.Errorf("invalid cfg_scale: %f (must be between 0 and 1)", opts.CFGScale)
	}

	return nil
}
