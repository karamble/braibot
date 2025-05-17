package commands

import (
	"context"
	// "encoding/json"
	"fmt"
	// "net/http"
	// "time"

	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils" // Added import for utils
)

// RateCommand returns the rate command
func RateCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "rate",
		Description: "Show current DCR and BTC exchange rates",
		Category:    "🎯 Basic",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			dcrPrice, btcPrice, err := utils.GetDCRPrice()
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to fetch exchange rates: %v", err))
			}
			msg := fmt.Sprintf("Current Exchange Rates:\n• DCR: $%.2f\n• BTC: $%.2f", dcrPrice, btcPrice)
			return sender.SendMessage(ctx, msgCtx, msg)
		}),
	}
}
