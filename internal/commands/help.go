package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// HelpCommand returns the help command
func HelpCommand(registry *Registry, dbManager *database.DBManager) Command {
	return Command{
		Name:        "help",
		Description: "Shows help message. Usage: **!help [command] [model]**",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// If no args, show general help with contextual information
			if len(args) == 0 {
				// Get user's balance for contextual information
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)
				userIDStr := userID.String()
				balance, err := dbManager.GetBalance(userIDStr)
				if err != nil {
					return fmt.Errorf("failed to get balance: %v", err)
				}
				balanceDCR := float64(balance) / 1e11

				// Get current exchange rate for USD value
				dcrPrice, _, err := GetDCRPrice()
				if err != nil {
					return fmt.Errorf("failed to get DCR price: %v", err)
				}
				usdValue := balanceDCR * dcrPrice

				// Create enhanced help message with user context
				helpMsg := fmt.Sprintf("🤖 **Welcome to BraiBot Help!**\n\n")
				helpMsg += fmt.Sprintf("💰 **Your Balance:** %.8f DCR ($%.2f USD)\n\n", balanceDCR, usdValue)

				// Get current model selections
				helpMsg += "🎯 **Your Current Model Selections:**\n"
				for _, cmdType := range []string{"text2image", "text2speech", "image2image", "image2video"} {
					if model, exists := faladapter.GetCurrentModel(cmdType); exists {
						helpMsg += fmt.Sprintf("• %s: %s ($%.2f USD)\n", cmdType, model.Name, model.PriceUSD)
					}
				}
				helpMsg += "\n"

				// Add command categories
				helpMsg += "## 🎯 Basic Commands\n"
				helpMsg += "| Command | Description | Usage |\n"
				helpMsg += "| ------- | ----------- | ----- |\n"
				for _, cmdName := range []string{"help", "balance", "rate"} {
					if cmd, exists := registry.Get(cmdName); exists {
						helpMsg += fmt.Sprintf("| !%s | %s | !%s |\n", cmd.Name, cmd.Description, cmd.Name)
					}
				}

				helpMsg += "\n## 🔧 Model Configuration\n"
				helpMsg += "| Command | Description | Usage |\n"
				helpMsg += "| ------- | ----------- | ----- |\n"
				for _, cmdName := range []string{"listmodels", "setmodel"} {
					if cmd, exists := registry.Get(cmdName); exists {
						usage := "!%s [command]"
						if cmdName == "setmodel" {
							usage = "!%s [command] [model]"
						}
						helpMsg += fmt.Sprintf("| !%s | %s | "+usage+" |\n", cmd.Name, cmd.Description, cmd.Name)
					}
				}

				helpMsg += "\n## 🎨 AI Generation\n"
				helpMsg += "| Command | Description | Starting Price |\n"
				helpMsg += "| ------- | ----------- | ------------- |\n"

				// Use generalized descriptions for AI commands
				aiCommands := map[string]string{
					"text2image":  "Generate images from text descriptions",
					"image2image": "Transform images using AI",
					"image2video": "Convert images to videos with AI",
					"text2speech": "Convert text to speech with AI",
				}

				for cmdName, description := range aiCommands {
					if _, exists := registry.Get(cmdName); exists {
						if model, exists := faladapter.GetCurrentModel(cmdName); exists {
							helpMsg += fmt.Sprintf("| !%s | %s | $%.2f |\n", cmdName, description, model.PriceUSD)
						} else {
							helpMsg += fmt.Sprintf("| !%s | %s | - |\n", cmdName, description)
						}
					}
				}

				helpMsg += "\n💡 **Tips:**\n"
				helpMsg += "• Use `!help [command]` for detailed command information\n"
				helpMsg += "• Use `!help [command] [model]` for model-specific details\n"
				helpMsg += "• Send tips through Bison Relay to add funds to your balance\n"

				return bot.SendPM(ctx, pm.Nick, helpMsg)
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

				// Get current model selection
				currentModel, hasCurrentModel := faladapter.GetCurrentModel(commandName)
				currentModelInfo := ""
				if hasCurrentModel {
					currentModelInfo = fmt.Sprintf("\n\n**Currently Selected Model:** %s ($%.2f USD)\n\n%s",
						currentModel.Name,
						currentModel.PriceUSD,
						currentModel.HelpDoc)
				}

				// Format command help with model list
				helpMsg := fmt.Sprintf("Command: !%s\nDescription: %s%s\n\nAvailable Models:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n",
					cmd.Name,
					cmd.Description,
					currentModelInfo)

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

				// Get user's balance for contextual information
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)
				userIDStr := userID.String()
				balance, err := dbManager.GetBalance(userIDStr)
				if err != nil {
					return fmt.Errorf("failed to get balance: %v", err)
				}
				balanceDCR := float64(balance) / 1e11

				// Convert model price to DCR
				dcrAmount, err := USDToDCR(model.PriceUSD)
				if err != nil {
					return fmt.Errorf("failed to convert USD to DCR: %v", err)
				}

				// Format model-specific help message with balance context
				helpMsg := fmt.Sprintf("Model: %s\nDescription: %s\nPrice: $%.2f USD (%.8f DCR)\n\nYour Balance: %.8f DCR\n\n%s",
					model.Name,
					model.Description,
					model.PriceUSD,
					dcrAmount,
					balanceDCR,
					model.HelpDoc)

				return bot.SendPM(ctx, pm.Nick, helpMsg)
			}

			// If more than two args, show error
			return bot.SendPM(ctx, pm.Nick, "Too many arguments. Usage: **!help [command] [model]**")
		},
	}
}
