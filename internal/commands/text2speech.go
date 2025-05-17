package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/speech"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(bot *kit.Bot, cfg *botconfig.BotConfig, speechService *speech.SpeechService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2speech")
	if !exists {
		model = fal.Model{
			Name:        "text2speech",
			Description: "Convert text to speech",
		}
	}
	description := fmt.Sprintf("%s. Usage: !text2speech [text]", model.Description)

	return braibottypes.Command{
		Name:        "text2speech",
		Description: description,
		Category:    "🎤 AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			if len(args) < 1 {
				return sender.SendMessage(ctx, msgCtx, "Please provide text to convert to speech. Usage: !text2speech [text]")
			}

			// Get the text from the arguments
			text := strings.Join(args, " ")

			// Create the speech request
			req := speech.SpeechRequest{
				Text: text,
				IsPM: msgCtx.IsPM,
				GC:   msgCtx.GC,
			}

			// Process the speech
			result, err := speechService.GenerateSpeech(ctx, &req)
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, err)
			}

			// Send the result
			return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Generated speech: %s", result.AudioURL))
		}),
	}
}

// parseTextSpeechArgs parses the command arguments for text2speech.
// It requires voice_id to be specified with --voice_id parameter.
// Returns the text, voice_id (or default), and parsed options map, and error.
func parseTextSpeechArgs(args []string) (text, voiceID string, options map[string]interface{}, err error) {
	defaultVoiceID := "Wise_Woman"
	options = make(map[string]interface{})
	var promptParts []string

	if len(args) == 0 {
		err = fmt.Errorf("please provide text to convert to speech")
		return
	}

	// Initialize with default voice
	voiceID = defaultVoiceID

	// Parse args for flags and text
	i := 0
	for i < len(args) {
		arg := args[i]
		argLower := strings.ToLower(arg)

		// Handle flags like --flag=value
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
			case "voice_id":
				voiceID = value
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
