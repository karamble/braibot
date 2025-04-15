package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// ListModelsCommand returns the listmodels command
func ListModelsCommand() Command {
	return Command{
		Name:        "listmodels",
		Description: "Lists available models for a specific command. Usage: !listmodels [command]",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) == 0 {
				return bot.SendPM(ctx, pm.Nick, "Please specify a command (text2image, text2speech, or image2image). Usage: !listmodels [command]")
			}

			commandName := strings.ToLower(args[0])

			var modelList string
			var models map[string]falapi.Model

			switch commandName {
			case "text2image":
				models = falapi.Text2ImageModels
				modelList = "Available models for text2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			case "text2speech":
				models = falapi.Text2SpeechModels
				modelList = "Available models for text2speech:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			case "image2image":
				models = falapi.Image2ImageModels
				modelList = "Available models for image2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			default:
				return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image', 'text2speech', or 'image2image'.")
			}

			for _, model := range models {
				modelList += fmt.Sprintf("| %s | %s | $%.2f |\n", model.Name, model.Description, model.Price)
			}

			return bot.SendPM(ctx, pm.Nick, modelList)
		},
	}
}

// SetModelCommand returns the setmodel command
func SetModelCommand(registry *Registry) Command {
	return Command{
		Name:        "setmodel",
		Description: "Sets the model to use for specified commands. Usage: !setmodel [command] [modelname]",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 2 {
				return bot.SendPM(ctx, pm.Nick, "Please specify a command and a model name. Usage: !setmodel [command] [modelname]")
			}
			commandName := args[0]
			modelName := args[1]

			// Check if the command is valid
			if _, exists := registry.Get(commandName); !exists {
				return bot.SendPM(ctx, pm.Nick, "Invalid command name. Use !listmodels to see available commands.")
			}

			// Check if the model is valid for the specific command
			var models map[string]falapi.Model
			switch commandName {
			case "text2image":
				models = falapi.Text2ImageModels
			case "text2speech":
				models = falapi.Text2SpeechModels
			case "image2image":
				models = falapi.Image2ImageModels
			default:
				return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image', 'text2speech', or 'image2image'.")
			}

			if _, exists := models[modelName]; exists {
				falapi.SetDefaultModel(commandName, modelName)
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Model for %s set to: %s", commandName, modelName))
			}

			return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Invalid model name for %s. Use !listmodels %s to see available models.", commandName, commandName))
		},
	}
}
