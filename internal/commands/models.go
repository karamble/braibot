package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
)

// ListModelsCommand returns the listmodels command
func ListModelsCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "listmodels",
		Description: "📋 List available AI models for a specific task",
		Category:    "Model Configuration",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			if len(args) < 1 {
				return sender.SendMessage(ctx, msgCtx, "Usage: !listmodels [task]")
			}
			task := strings.ToLower(args[0])
			models, exists := faladapter.GetModels(task)
			if !exists || len(models) == 0 {
				return sender.SendMessage(ctx, msgCtx, "Invalid command or no models found for that task.")
			}
			msg := fmt.Sprintf("Available models for %s:\n", task)
			for _, model := range models {
				msg += fmt.Sprintf("• %s: %s ($%.2f USD)\n", model.Name, model.Description, model.PriceUSD)
			}
			return sender.SendMessage(ctx, msgCtx, msg)
		}),
	}
}

// SetModelCommand returns the setmodel command
func SetModelCommand(registry *Registry) braibottypes.Command {
	return braibottypes.Command{
		Name:        "setmodel",
		Description: "⚙️ Set the default AI model for a specific task",
		Category:    "Model Configuration",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			if len(args) < 2 {
				return sender.SendMessage(ctx, msgCtx, "Usage: !setmodel [task] [model]")
			}
			task := strings.ToLower(args[0])
			modelName := strings.ToLower(args[1])

			// Convert user ID to string for PMs
			var userID string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userID = uid.String()
			}

			if err := faladapter.SetCurrentModel(task, modelName, userID); err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to set model: %v", err))
			}

			// Different message based on context
			if msgCtx.IsPM {
				return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Your personal model for %s set to: %s", task, modelName))
			} else {
				return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Group chat model for %s set to: %s", task, modelName))
			}
		}),
	}
}
