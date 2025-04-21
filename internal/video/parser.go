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
func (p *ArgumentParser) Parse(args []string, expectImageURL bool) (prompt, imageURL, duration, aspectRatio, negativePrompt string, cfgScale *float64, promptOptimizer *bool, err error) {
	var promptParts []string
	// Set defaults
	duration = "5"
	aspectRatio = "16:9"
	negativePrompt = "blur, distort, and low quality"
	cfgScale = nil        // Default to nil, only set if parsed
	promptOptimizer = nil // Default to nil

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

	return prompt, imageURL, duration, aspectRatio, negativePrompt, cfgScale, promptOptimizer, nil
}
