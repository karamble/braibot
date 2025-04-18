package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/speech"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(dbManager *database.DBManager, speechService *speech.SpeechService, debug bool) Command {
	// Get the current model to use its description and help doc
	model, exists := faladapter.GetCurrentModel("text2speech")
	if !exists {
		model = fal.Model{
			Name:        "text2speech",
			Description: "Converts text to speech using AI.",
			HelpDoc:     "Usage: !text2speech [voice_id] [text]\nDefault voice: Wise_Woman. Please provide text to convert.",
		}
	}

	description := model.Description // Use the model description
	help := model.HelpDoc            // Use the model help doc

	return Command{
		Name:        "text2speech",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				return bot.SendPM(ctx, pm.Nick, help) // Send detailed help doc
			}

			// Parse arguments using the helper
			text, voiceID, options, err := parseTextSpeechArgs(args)
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Argument error: %v", err))
			}

			// Get model configuration (required for PriceUSD)
			model, exists := faladapter.GetCurrentModel("text2speech")
			if !exists {
				return fmt.Errorf("no default model found for text2speech") // Should not happen if validation passed
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2speech")

			// Create internal speech request
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			req := &speech.SpeechRequest{
				Text:      text,
				VoiceID:   voiceID,
				ModelName: model.Name, // Use current default model
				Progress:  progress,
				UserNick:  pm.Nick,
				UserID:    userID,
				PriceUSD:  model.PriceUSD,
			}

			// Populate options from the parsed map
			if val, ok := options["speed"].(*float64); ok {
				req.Speed = val
			}
			if val, ok := options["vol"].(*float64); ok {
				req.Vol = val
			}
			if val, ok := options["pitch"].(*int); ok {
				req.Pitch = val
			}
			if val, ok := options["emotion"].(string); ok {
				req.Emotion = val
			}
			if val, ok := options["sample_rate"].(string); ok {
				req.SampleRate = val
			}
			if val, ok := options["bitrate"].(string); ok {
				req.Bitrate = val
			}
			if val, ok := options["format"].(string); ok {
				req.Format = val
			}
			if val, ok := options["channel"].(string); ok {
				req.Channel = val
			}

			// Generate speech using the service
			result, err := speechService.GenerateSpeech(ctx, req)
			if err != nil {
				// Service handles billing checks/refunds, just return the error
				return fmt.Errorf("failed to generate speech: %v", err)
			}

			if !result.Success {
				// If service handled billing/refunds, error might already be user-friendly
				errMsg := "speech generation failed"
				if result.Error != nil {
					errMsg += ": " + result.Error.Error()
				}
				return fmt.Errorf(errMsg)
			}

			// Success message is handled by the service (audio embed + billing info)
			return nil
		},
	}
}

// parseTextSpeechArgs parses the command arguments for text2speech.
// It identifies an optional voice_id as the first argument if it doesn't start with '--'
// and isn't likely part of a multi-word text when only one arg is given.
// Returns the text, voice_id (or default), and parsed options map, and error.
func parseTextSpeechArgs(args []string) (text, voiceID string, options map[string]interface{}, err error) {
	defaultVoiceID := "Wise_Woman"
	options = make(map[string]interface{})
	var promptParts []string

	if len(args) == 0 {
		err = fmt.Errorf("please provide text to convert to speech")
		return
	}

	// Identify voice ID
	firstArgIsLikelyVoiceID := false
	if len(args) > 1 && !strings.HasPrefix(args[0], "--") {
		if len(strings.Split(args[0], "_")) > 1 && len(args[0]) < 30 { // Basic heuristic
			firstArgIsLikelyVoiceID = true
		}
	}
	if len(args) == 1 && !strings.HasPrefix(args[0], "--") {
		if len(strings.Split(args[0], "_")) > 1 && len(args[0]) < 30 {
			err = fmt.Errorf("voice ID '%s' provided, but no text found", args[0])
			return
		}
	}

	// Separate potential voice ID from other args
	startIdx := 0
	if firstArgIsLikelyVoiceID {
		voiceID = args[0]
		startIdx = 1
	} else {
		voiceID = defaultVoiceID
	}

	// Parse remaining args for flags and text
	i := startIdx
	for i < len(args) {
		arg := args[i]
		argLower := strings.ToLower(arg)

		// Handle boolean flags like --flag=value (though none are bool here yet)
		var flagValue string
		if strings.Contains(argLower, "=") {
			parts := strings.SplitN(argLower, "=", 2)
			argLower = parts[0]
			if len(parts) > 1 {
				flagValue = parts[1]
			}
		}

		if strings.HasPrefix(argLower, "--") {
			flagName := strings.TrimPrefix(argLower, "--")
			var value string
			if flagValue != "" {
				value = flagValue
				i++ // Consume the flag=value arg
			} else if i+1 < len(args) {
				value = args[i+1]
				i += 2 // Consume flag and value
			} else {
				err = fmt.Errorf("missing value for argument: %s", arg)
				return
			}

			switch flagName {
			case "speed":
				fVal, parseErr := strconv.ParseFloat(value, 64)
				if parseErr != nil {
					err = fmt.Errorf("invalid value for --speed: %s", value)
					return
				}
				options["speed"] = &fVal
			case "vol":
				fVal, parseErr := strconv.ParseFloat(value, 64)
				if parseErr != nil {
					err = fmt.Errorf("invalid value for --vol: %s", value)
					return
				}
				options["vol"] = &fVal
			case "pitch":
				iVal, parseErr := strconv.Atoi(value)
				if parseErr != nil {
					err = fmt.Errorf("invalid value for --pitch: %s", value)
					return
				}
				options["pitch"] = &iVal
			case "emotion":
				options["emotion"] = value
			case "sample_rate":
				options["sample_rate"] = value
			case "bitrate":
				options["bitrate"] = value
			case "format":
				options["format"] = strings.ToLower(value)
			case "channel":
				options["channel"] = value
			default:
				err = fmt.Errorf("unknown argument: %s", arg)
				return
			}
		} else {
			// Assume it's part of the prompt
			promptParts = append(promptParts, arg)
			i++
		}
	}

	text = strings.Join(promptParts, " ")
	if text == "" {
		err = fmt.Errorf("please provide text to convert to speech")
		return
	}

	return text, voiceID, options, nil
}
