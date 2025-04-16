package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/pkg/fal"
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
				return bot.SendPM(ctx, pm.Nick, "Please specify a command (text2image, text2speech, image2image, or image2video). Usage: !listmodels [command]")
			}

			commandName := strings.ToLower(args[0])

			var modelList string
			var models map[string]fal.Model
			var exists bool

			switch commandName {
			case "text2image":
				models, exists = faladapter.GetModels("text2image")
				modelList = "Available models for text2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			case "text2speech":
				models, exists = faladapter.GetModels("text2speech")
				modelList = "Available models for text2speech:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			case "image2image":
				models, exists = faladapter.GetModels("image2image")
				modelList = "Available models for image2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			case "image2video":
				models, exists = faladapter.GetModels("image2video")
				modelList = "Available models for image2video:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
			default:
				return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image', 'text2speech', 'image2image', or 'image2video'.")
			}

			if !exists {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("No models found for command: %s", commandName))
			}

			for _, model := range models {
				modelList += fmt.Sprintf("| %s | %s | $%.2f |\n", model.Name, model.Description, model.PriceUSD)
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
			var models map[string]fal.Model
			var exists bool
			switch commandName {
			case "text2image":
				models, exists = faladapter.GetModels("text2image")
			case "text2speech":
				models, exists = faladapter.GetModels("text2speech")
			case "image2image":
				models, exists = faladapter.GetModels("image2image")
			default:
				return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image', 'text2speech', or 'image2image'.")
			}

			if !exists {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("No models found for command: %s", commandName))
			}

			if _, exists := models[modelName]; exists {
				err := faladapter.SetCurrentModel(commandName, modelName)
				if err != nil {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error setting model: %v", err))
				}
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Model for %s set to: %s", commandName, modelName))
			}

			return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Invalid model name for %s. Use !listmodels %s to see available models.", commandName, commandName))
		},
	}
}
