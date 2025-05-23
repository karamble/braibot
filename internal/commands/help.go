package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
)

// HelpCommand returns the help command
func HelpCommand(registry *Registry, dbManager braibottypes.DBManagerInterface) braibottypes.Command {
	return braibottypes.Command{
		Name:        "help",
		Description: "📚 Show this help message or details for a specific command (e.g., !help text2image)",
		Category:    "Basic",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Get user ID for PMs
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}

			// If no args, show general help with contextual information
			if len(args) == 0 {
				// Get user's balance for contextual information
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)
				userIDStr := userID.String()
				balance, err := db.GetBalance(userIDStr)
				if err != nil {
					return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to get balance: %v", err))
				}
				balanceDCR := float64(balance) / 1e11

				// Get current exchange rate for USD value using utils
				dcrPrice, _, err := utils.GetDCRPrice()
				if err != nil {
					// Log the error but continue, maybe showing balance without USD value
					fmt.Printf("ERROR [help] Failed to get DCR price: %v\n", err)
					dcrPrice = 0 // Set to 0 to avoid NaN issues if used later
				}
				usdValue := balanceDCR * dcrPrice

				// Create enhanced help message with user context
				helpMsg := fmt.Sprintf("🤖 **Welcome to BraiBot Help!**\n\n")
				if msgCtx.IsPM {
					helpMsg += fmt.Sprintf("💰 **Your Balance:** %.8f DCR ($%.2f USD)\n\n", balanceDCR, usdValue)
				} else {
					helpMsg += "💰 **Balance Command:** Only available in private messages\n\n"
				}

				// Add billing disabled message if applicable
				if !registry.GetBillingEnabled() {
					helpMsg += "🎉 **Happy Days!** All commands are free to use.\n\n"
				}

				// Get current model selections
				helpMsg += "🎯 **Your Current Model Selections:**\n"
				for _, cmdType := range []string{"text2image", "text2speech", "image2image", "image2video", "text2video"} {
					if model, exists := faladapter.GetCurrentModel(cmdType, userIDStr); exists {
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
						if cmd.Category == "Basic" {
							helpMsg += fmt.Sprintf("| !%s | %s | !%s |\n", cmd.Name, cmd.Description, cmd.Name)
						}
					}
				}

				helpMsg += "\n## 🔧 Model Configuration\n"
				helpMsg += "| Command | Description | Usage |\n"
				helpMsg += "| ------- | ----------- | ----- |\n"
				for _, cmdName := range []string{"listmodels", "setmodel"} {
					if cmd, exists := registry.Get(cmdName); exists {
						if cmd.Category == "Model Configuration" {
							usage := "!%s [task]"
							if cmdName == "setmodel" {
								usage = "!%s [task] [model]"
							}
							helpMsg += fmt.Sprintf("| !%s | %s | "+usage+" |\n", cmd.Name, cmd.Description, cmd.Name)
						}
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
					"text2video":  "Generate videos from text descriptions",
					"text2speech": "Convert text to speech with AI",
				}

				// Add !ai command with conditional display
				webhookEnabled, hasWebhookEnabled := registry.GetWebhookEnabled()
				if hasWebhookEnabled && webhookEnabled {
					aiCommands["ai"] = "Send a message to the AI for processing"
				} else {
					aiCommands["ai"] = "Send a message to the AI webhook for processing **disabled**"
				}

				for cmdName, description := range aiCommands {
					if _, exists := registry.Get(cmdName); exists {
						if model, exists := faladapter.GetCurrentModel(cmdName, userIDStr); exists {
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

				return sender.SendMessage(ctx, msgCtx, helpMsg)
			}

			// If only one arg, show command-specific help with model list
			if len(args) == 1 {
				commandName := strings.ToLower(args[0])
				cmd, exists := registry.Get(commandName)
				if !exists {
					return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Unknown command: %s. Use **!help** to see available commands.", commandName))
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
					return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Command: !%s\nDescription: %s", cmd.Name, cmd.Description))
				}

				if !modelExists {
					return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("No models found for command: %s", commandName))
				}

				// Get current model selection
				currentModel, hasCurrentModel := faladapter.GetCurrentModel(commandName, userIDStr)
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
				return sender.SendMessage(ctx, msgCtx, helpMsg)
			}

			// If two args, show model-specific help
			if len(args) == 2 {
				commandName := strings.ToLower(args[0])
				modelName := strings.ToLower(args[1])

				// Get the model information
				model, exists := faladapter.GetModel(modelName, commandName)
				if !exists {
					return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Unknown model: %s for command: %s. Use !help %s to see available models.", modelName, commandName, commandName))
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader(commandName, model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return sender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			return sender.SendMessage(ctx, msgCtx, "Invalid help command usage. Use !help for general help or !help [command] for specific command help.")
		}),
	}
}
