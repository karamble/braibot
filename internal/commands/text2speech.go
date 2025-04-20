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
	"github.com/karamble/braibot/internal/utils"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(dbManager *database.DBManager, speechService *speech.SpeechService, debug bool) Command {
	// Outer model retrieval removed

	return Command{
		Name:        "text2speech",
		Description: "🗣️ Generate speech audio from text (e.g., !text2speech Hello world!)",
		Category:    "🎨 AI Generation",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model
				model, exists := faladapter.GetCurrentModel("text2speech")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Error: Default text2speech model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("text2speech", model, userID, dbManager)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !text2speech [text] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return bot.SendPM(ctx, pm.Nick, header+helpDoc)
			}

			// Parse arguments using the helper
			text, voiceID, options, err := parseTextSpeechArgs(args)
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Argument error: %v", err))
			}

			// Get model configuration (required for PriceUSD)
			model, exists := faladapter.GetCurrentModel("text2speech") // Restore model retrieval
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
				ModelName: model.Name, // Restore model usage
				Progress:  progress,
				UserNick:  pm.Nick,
				UserID:    userID,
				PriceUSD:  model.PriceUSD, // Restore model usage
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

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, pm, "text2speech", result, err); handleErr != nil {
				return handleErr // Propagate error if not handled by the utility function
			}

			// If we reach here, the operation was successful and errors were handled
			// Success message (audio embed) is handled by the service itself in this case
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
