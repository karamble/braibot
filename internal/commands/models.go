package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
)

// ListModelsCommand returns the listmodels command
func ListModelsCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "listmodels",
		Description: "List available models for a given task (e.g., text2image, text2speech)",
		Category:    "🔧 Model Configuration",
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
		Description: "Set the default model for a given task (e.g., text2image, text2speech)",
		Category:    "🔧 Model Configuration",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			if len(args) < 2 {
				return sender.SendMessage(ctx, msgCtx, "Usage: !setmodel [task] [model]")
			}
			task := strings.ToLower(args[0])
			modelName := strings.ToLower(args[1])
			if err := faladapter.SetCurrentModel(task, modelName); err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to set model: %v", err))
			}
			return sender.SendMessage(ctx, msgCtx, fmt.Sprintf("Model for %s set to: %s", task, modelName))
		}),
	}
}
