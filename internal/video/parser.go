package video

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseResult holds the result of parsing command arguments.
// All parse functions return this struct; unused fields are zero-valued.
type ParseResult struct {
	Prompt          string
	ImageURL        string   // image2video only
	VideoURL        string   // video2video only
	Duration        string
	AspectRatio     string
	NegativePrompt  string
	CFGScale        *float64
	PromptOptimizer *bool
	Resolution      string
	GenerateAudio   *bool
	KeepAudio       *bool    // video2video only
	EndImageURL     string
	Seed            *int64
	ImageURLs       []string // multi2video / video2video
	VideoURLs       []string // multi2video only
	AudioURLs       []string // multi2video only
}

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

// ParseCFGScale parses the CFG scale argument, returning a pointer or nil
func (p *ArgumentParser) ParseCFGScale(args []string) *float64 {
	for i, arg := range args {
		if arg == "--cfg_scale" && i+1 < len(args) {
			if scale, err := strconv.ParseFloat(args[i+1], 64); err == nil {
				return &scale
			}
		}
	}
	return nil // Return nil if not found or invalid
}

// ParsePromptOptimizer parses the prompt optimizer flag, returning a pointer or nil
func (p *ArgumentParser) ParsePromptOptimizer(args []string) *bool {
	for i, arg := range args {
		flag := strings.ToLower(arg)
		if (flag == "--prompt_optimizer" || flag == "--prompt-optimizer") && i+1 < len(args) {
			valStr := strings.ToLower(args[i+1])
			if valStr == "true" {
				result := true
				return &result
			} else if valStr == "false" {
				result := false
				return &result
			} else {
				// Invalid value, return nil or handle error as needed
				return nil
			}
		}
	}
	return nil // Flag not found
}

// Parse parses all arguments, separating prompt, image URL (optional), and options.
func (p *ArgumentParser) Parse(args []string, expectImageURL bool) (*ParseResult, error) {
	r := &ParseResult{
		Duration:       "5",
		AspectRatio:    "16:9",
		NegativePrompt: "blur, distort, and low quality",
	}
	var promptParts []string

	parsedArgs := make(map[int]bool) // Track indices consumed by flags
	currentIndex := 0

	// First pass: Handle image URL if expected
	if expectImageURL {
		if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
			r.ImageURL = args[0]
			parsedArgs[0] = true
			currentIndex = 1
		} else {
			return nil, fmt.Errorf("image URL is required as the first argument for this command")
		}
	}

	// Second pass: Parse flags
	localArgs := args[currentIndex:] // Only parse flags after potential image URL
	localIndexOffset := currentIndex
	i := 0
	for i < len(localArgs) {
		arg := localArgs[i]
		if !strings.HasPrefix(arg, "--") {
			i++
			continue // Skip non-flags in this pass
		}

		flag := strings.ToLower(arg)
		originalIndex := i + localIndexOffset

		if parsedArgs[originalIndex] {
			i++
			continue // Already processed (e.g., was image URL)
		}

		// Check for value
		var value string
		if i+1 < len(localArgs) {
			value = localArgs[i+1]
		}

		switch flag {
		case "--duration":
			if value != "" {
				r.Duration = strings.TrimSuffix(value, "s")
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--aspect":
			if value != "" {
				r.AspectRatio = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--negative_prompt", "--negative-prompt":
			if value != "" {
				r.NegativePrompt = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--cfg_scale", "--cfg-scale":
			if value != "" {
				if scale, parseErr := strconv.ParseFloat(value, 64); parseErr == nil {
					r.CFGScale = &scale
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					return nil, fmt.Errorf("invalid value for %s: %s", flag, value)
				}
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--prompt_optimizer", "--prompt-optimizer":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					r.PromptOptimizer = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else if valStr == "false" {
					result := false
					r.PromptOptimizer = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					return nil, fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
				}
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--resolution":
			if value != "" {
				r.Resolution = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--audio":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					r.GenerateAudio = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else if valStr == "false" {
					result := false
					r.GenerateAudio = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					return nil, fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
				}
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--end_image", "--end-image":
			if value != "" {
				r.EndImageURL = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--seed":
			if value != "" {
				s, parseErr := strconv.ParseInt(value, 10, 64)
				if parseErr != nil {
					return nil, fmt.Errorf("invalid value for %s: %s (must be an integer)", flag, value)
				}
				r.Seed = &s
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		default:
			// Unknown flag, treat as part of prompt later or ignore
			i++
		}
	}

	// Third pass: Collect prompt parts
	for i, arg := range args {
		if !parsedArgs[i] {
			promptParts = append(promptParts, arg)
		}
	}
	r.Prompt = strings.Join(promptParts, " ")

	return r, nil
}

// ParseMulti2Video parses arguments for the multi2video (reference-to-video) command.
// Usage: !multi2video [prompt text] [--image1..9 url] [--video1..3 url] [--audio1..3 url] [--duration N] [--aspect auto] [--resolution 720p] [--audio true|false] [--seed N]
func (p *ArgumentParser) ParseMulti2Video(args []string) (*ParseResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("prompt is required")
	}

	r := &ParseResult{}
	var promptParts []string
	parsedArgs := make(map[int]bool)

	// Parse flags
	i := 0
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			i++
			continue
		}

		flag := strings.ToLower(arg)

		// Check for value
		var value string
		if i+1 < len(args) {
			value = args[i+1]
		}

		switch flag {
		case "--image1", "--image2", "--image3", "--image4",
			"--image5", "--image6", "--image7", "--image8", "--image9":
			if value != "" {
				r.ImageURLs = append(r.ImageURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--video1", "--video2", "--video3":
			if value != "" {
				r.VideoURLs = append(r.VideoURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--audio1", "--audio2", "--audio3":
			if value != "" {
				r.AudioURLs = append(r.AudioURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--duration":
			if value != "" {
				r.Duration = strings.TrimSuffix(value, "s")
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--aspect":
			if value != "" {
				r.AspectRatio = value
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--resolution":
			if value != "" {
				r.Resolution = value
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--audio":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					r.GenerateAudio = &result
				} else if valStr == "false" {
					result := false
					r.GenerateAudio = &result
				} else {
					return nil, fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
				}
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--seed":
			if value != "" {
				s, parseErr := strconv.ParseInt(value, 10, 64)
				if parseErr != nil {
					return nil, fmt.Errorf("invalid value for %s: %s (must be an integer)", flag, value)
				}
				r.Seed = &s
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		default:
			// Unknown flag, treat as part of prompt
			i++
		}
	}

	// Collect prompt parts from remaining args
	for i, arg := range args {
		if !parsedArgs[i] {
			promptParts = append(promptParts, arg)
		}
	}
	r.Prompt = strings.Join(promptParts, " ")

	return r, nil
}

// ParseVideo2Video parses arguments for the video2video command.
// Usage: !video2video [video_url] [prompt text] [--keep_audio true|false] [--image1 url] ... [--image4 url] [--duration N]
func (p *ArgumentParser) ParseVideo2Video(args []string) (*ParseResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("video URL is required as the first argument")
	}

	// First non-flag arg is the video URL
	if strings.HasPrefix(args[0], "--") {
		return nil, fmt.Errorf("video URL is required as the first argument")
	}

	r := &ParseResult{
		VideoURL: args[0],
		Duration: "5", // Default duration for billing
	}

	var promptParts []string
	parsedArgs := make(map[int]bool)
	parsedArgs[0] = true // video URL consumed

	// Parse flags
	i := 1
	for i < len(args) {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			i++
			continue
		}

		flag := strings.ToLower(arg)

		// Check for value
		var value string
		if i+1 < len(args) {
			value = args[i+1]
		}

		switch flag {
		case "--keep_audio", "--keep-audio":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					r.KeepAudio = &result
				} else if valStr == "false" {
					result := false
					r.KeepAudio = &result
				} else {
					return nil, fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
				}
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--image1", "--image2", "--image3", "--image4":
			if value != "" {
				r.ImageURLs = append(r.ImageURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		case "--duration":
			if value != "" {
				r.Duration = strings.TrimSuffix(value, "s")
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				return nil, fmt.Errorf("missing value for %s", flag)
			}
		default:
			// Unknown flag, treat as part of prompt
			i++
		}
	}

	// Collect prompt parts from remaining args
	for i, arg := range args {
		if !parsedArgs[i] {
			promptParts = append(promptParts, arg)
		}
	}
	r.Prompt = strings.Join(promptParts, " ")

	return r, nil
}
