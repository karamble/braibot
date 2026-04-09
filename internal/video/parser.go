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
// It returns the parsed values individually.
func (p *ArgumentParser) Parse(args []string, expectImageURL bool) (prompt, imageURL, duration, aspectRatio, negativePrompt string, cfgScale *float64, promptOptimizer *bool, resolution string, generateAudio *bool, endImageURL string, seed *int64, err error) {
	var promptParts []string
	// Set defaults
	duration = "5"
	aspectRatio = "16:9"
	negativePrompt = "blur, distort, and low quality"
	cfgScale = nil        // Default to nil, only set if parsed
	promptOptimizer = nil // Default to nil
	resolution = ""       // Default to empty, let model defaults apply
	seed = nil            // Default to nil, only set if parsed

	parsedArgs := make(map[int]bool) // Track indices consumed by flags
	currentIndex := 0

	// First pass: Handle image URL if expected
	if expectImageURL {
		if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
			imageURL = args[0]
			parsedArgs[0] = true
			currentIndex = 1
		} else {
			err = fmt.Errorf("image URL is required as the first argument for this command")
			return
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
				duration = strings.TrimSuffix(value, "s")
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--aspect":
			if value != "" {
				aspectRatio = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--negative_prompt", "--negative-prompt":
			if value != "" {
				negativePrompt = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--cfg_scale", "--cfg-scale":
			if value != "" {
				if scale, parseErr := strconv.ParseFloat(value, 64); parseErr == nil {
					cfgScale = &scale // Assign the pointer
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					err = fmt.Errorf("invalid value for %s: %s", flag, value)
					return
				}
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--prompt_optimizer", "--prompt-optimizer":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					promptOptimizer = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else if valStr == "false" {
					result := false
					promptOptimizer = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					err = fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
					return
				}
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--resolution":
			if value != "" {
				resolution = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--audio":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					generateAudio = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else if valStr == "false" {
					result := false
					generateAudio = &result
					parsedArgs[originalIndex] = true
					parsedArgs[originalIndex+1] = true
					i += 2
				} else {
					err = fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
					return
				}
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--end_image", "--end-image":
			if value != "" {
				endImageURL = value
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--seed":
			if value != "" {
				s, parseErr := strconv.ParseInt(value, 10, 64)
				if parseErr != nil {
					err = fmt.Errorf("invalid value for %s: %s (must be an integer)", flag, value)
					return
				}
				seed = &s
				parsedArgs[originalIndex] = true
				parsedArgs[originalIndex+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
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
	prompt = strings.Join(promptParts, " ")

	// No final validation here anymore - that will happen in the FAL layer

	return prompt, imageURL, duration, aspectRatio, negativePrompt, cfgScale, promptOptimizer, resolution, generateAudio, endImageURL, seed, nil
}

// ParseMulti2Video parses arguments for the multi2video (reference-to-video) command.
// Usage: !multi2video [prompt text] [--image1..9 url] [--video1..3 url] [--audio1..3 url] [--duration N] [--aspect auto] [--resolution 720p] [--audio true|false] [--seed N]
func (p *ArgumentParser) ParseMulti2Video(args []string) (prompt, duration, aspectRatio, resolution string, generateAudio *bool, seed *int64, imageURLs, videoURLs, audioURLs []string, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("prompt is required")
		return
	}

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
				imageURLs = append(imageURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--video1", "--video2", "--video3":
			if value != "" {
				videoURLs = append(videoURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--audio1", "--audio2", "--audio3":
			if value != "" {
				audioURLs = append(audioURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--duration":
			if value != "" {
				duration = strings.TrimSuffix(value, "s")
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--aspect":
			if value != "" {
				aspectRatio = value
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--resolution":
			if value != "" {
				resolution = value
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--audio":
			if value != "" {
				valStr := strings.ToLower(value)
				if valStr == "true" {
					result := true
					generateAudio = &result
				} else if valStr == "false" {
					result := false
					generateAudio = &result
				} else {
					err = fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
					return
				}
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--seed":
			if value != "" {
				s, parseErr := strconv.ParseInt(value, 10, 64)
				if parseErr != nil {
					err = fmt.Errorf("invalid value for %s: %s (must be an integer)", flag, value)
					return
				}
				seed = &s
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
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
	prompt = strings.Join(promptParts, " ")

	return
}

// ParseVideo2Video parses arguments for the video2video command.
// Usage: !video2video [video_url] [prompt text] [--keep_audio true|false] [--image1 url] ... [--image4 url] [--duration N]
func (p *ArgumentParser) ParseVideo2Video(args []string) (videoURL, prompt, duration string, keepAudio *bool, imageURLs []string, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("video URL is required as the first argument")
		return
	}

	// First non-flag arg is the video URL
	if strings.HasPrefix(args[0], "--") {
		err = fmt.Errorf("video URL is required as the first argument")
		return
	}
	videoURL = args[0]

	var promptParts []string
	parsedArgs := make(map[int]bool)
	parsedArgs[0] = true // video URL consumed
	duration = "5"       // Default duration for billing

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
					keepAudio = &result
				} else if valStr == "false" {
					result := false
					keepAudio = &result
				} else {
					err = fmt.Errorf("invalid value for %s: %s (must be true or false)", flag, value)
					return
				}
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--image1", "--image2", "--image3", "--image4":
			if value != "" {
				imageURLs = append(imageURLs, value)
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
			}
		case "--duration":
			if value != "" {
				duration = strings.TrimSuffix(value, "s")
				parsedArgs[i] = true
				parsedArgs[i+1] = true
				i += 2
			} else {
				err = fmt.Errorf("missing value for %s", flag)
				return
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
	prompt = strings.Join(promptParts, " ")

	return
}
