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

// HelpCommand returns the help command
func HelpCommand(registry *Registry) Command {
	return Command{
		Name:        "help",
		Description: "Shows help message. Usage: **!help [command] [model]**",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// If no args, show general help
			if len(args) == 0 {
				return bot.SendPM(ctx, pm.Nick, registry.FormatHelpMessage())
			}

			// If only one arg, show command-specific help with model list
			if len(args) == 1 {
				commandName := strings.ToLower(args[0])
				cmd, exists := registry.Get(commandName)
				if !exists {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Unknown command: %s. Use **!help** to see available commands.", commandName))
				}

				// Get models for this command
				var models map[string]fal.Model
				var modelExists bool
				switch commandName {
				case "text2image":
					models, modelExists = faladapter.GetModels("text2image")
				case "text2speech":
					models, modelExists = faladapter.GetModels("text2speech")
				case "image2image":
					models, modelExists = faladapter.GetModels("image2image")
				case "image2video":
					models, modelExists = faladapter.GetModels("image2video")
				default:
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Command: !%s\nDescription: %s", cmd.Name, cmd.Description))
				}

				if !modelExists {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("No models found for command: %s", commandName))
				}

				// Format command help with model list
				helpMsg := fmt.Sprintf("Command: !%s\nDescription: %s\n\nAvailable Models:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n",
					cmd.Name,
					cmd.Description)

				for _, model := range models {
					helpMsg += fmt.Sprintf("| %s | %s | $%.2f |\n", model.Name, model.Description, model.PriceUSD)
				}

				helpMsg += "\nUse !help " + commandName + " <model_name> for detailed information about a specific model."
				return bot.SendPM(ctx, pm.Nick, helpMsg)
			}

			// If two args, show model-specific help
			if len(args) == 2 {
				commandName := strings.ToLower(args[0])
				modelName := strings.ToLower(args[1])

				// Get the model information
				model, exists := faladapter.GetModel(modelName, commandName)
				if !exists {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Unknown model: %s for command: %s. Use !help %s to see available models.", modelName, commandName, commandName))
				}

				// Format model-specific help message
				helpMsg := fmt.Sprintf("Model: %s\nDescription: %s\nPrice: $%.2f USD\n\n%s",
					model.Name,
					model.Description,
					model.PriceUSD,
					model.HelpDoc)

				return bot.SendPM(ctx, pm.Nick, helpMsg)
			}

			// If more than two args, show error
			return bot.SendPM(ctx, pm.Nick, "Too many arguments. Usage: **!help [command] [model]**")
		},
	}
}
