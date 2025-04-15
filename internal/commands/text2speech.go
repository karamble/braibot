package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/audio"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(debug bool) Command {
	return Command{
		Name:        "text2speech",
		Description: "Converts text to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 2 {
				voiceList := "Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl"
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Please provide a voice ID and text. Usage: !text2speech [voice_id] [text]\n\n%s", voiceList))
			}

			voiceID := args[0]
			text := strings.Join(args[1:], " ")

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("text2speech")
			if !exists {
				return fmt.Errorf("no default model found for text2speech")
			}
			_, exists = falapi.GetModel(modelName, "text2speech")
			if !exists {
				return fmt.Errorf("model not found: %s", modelName)
			}

			// Generate speech
			audioResp, err := client.GenerateSpeech(ctx, text, voiceID, bot, pm.Nick)
			if err != nil {
				return err
			}

			// Fetch the audio data
			resp, err := http.Get(audioResp.Audio.URL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			audioData, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Convert PCM to Opus using the audio package
			oggData, err := audio.ConvertPCMToOpus(audioData)
			if err != nil {
				return err
			}

			// Encode the audio data to base64
			encodedAudio := base64.StdEncoding.EncodeToString(oggData)

			// Create the message with embedded audio
			message := fmt.Sprintf("--embed[alt=%s,type=audio/ogg,data=%s]--", url.QueryEscape(text), encodedAudio)
			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
