package commands

import (
	"context"
	"fmt"

	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
)

// RateCommand returns the rate command
func RateCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "rate",
		Description: "💱 Show current DCR exchange rates",
		Category:    "Basic",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Get DCR prices in USD and BTC
			dcrUsdPrice, dcrBtcPrice, err := utils.GetDCRPrice()
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to fetch DCR rates: %v", err))
			}

			// Get BTC price in USD
			btcUsdPrice, err := utils.GetBTCPrice()
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to fetch BTC rate: %v", err))
			}

			msg := fmt.Sprintf("Current Exchange Rates:\n• DCR: $%s USD\n• DCR: %s BTC\n• BTC: $%s USD",
				utils.FormatUSDThousands(dcrUsdPrice),
				utils.FormatThousands(dcrBtcPrice),
				utils.FormatUSDThousands(btcUsdPrice))
			return sender.SendMessage(ctx, msgCtx, msg)
		}),
	}
}
