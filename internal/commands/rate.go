package commands

import (
	"context"
	// "encoding/json"
	"fmt"
	"strings"

	// "net/http"
	// "time"

	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils" // Added import for utils
)

// formatNumber formats a float64 with thousands separators
func formatNumber(num float64, decimals int) string {
	// Split into integer and decimal parts
	intPart := int64(num)
	decPart := num - float64(intPart)

	// Format integer part with thousands separators
	intStr := fmt.Sprintf("%'d", intPart)

	// Format decimal part
	decStr := fmt.Sprintf("%.*f", decimals, decPart)
	decStr = strings.TrimPrefix(decStr, "0.")

	// Combine parts
	return intStr + "." + decStr
}

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
				utils.FormatThousands(dcrUsdPrice),
				utils.FormatThousands(dcrBtcPrice),
				utils.FormatThousands(btcUsdPrice))
			return sender.SendMessage(ctx, msgCtx, msg)
		}),
	}
}
