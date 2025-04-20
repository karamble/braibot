package commands

import (
	"context"
	// "encoding/json"
	"fmt"
	// "net/http"
	// "time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/utils" // Added import for utils
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// RateCommand defines the !rate command
func RateCommand() Command {
	return Command{
		Name:        "rate",
		Description: "💲 Show the current DCR cost per second for AI model usage.",
		Category:    "🎯 Basic",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// Send a status message to indicate we're fetching rates
			bot.SendPM(ctx, pm.Nick, "Fetching current exchange rates...")

			// Get DCR prices using the utility function
			usdPrice, btcPrice, err := utils.GetDCRPrice()
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching rates: %v", err))
			}

			// Format the response with more detailed information
			message := fmt.Sprintf("Current DCR Exchange Rates:\n"+
				"USD: $%.2f\n"+
				"BTC: %.8f BTC\n"+
				"Source: CoinGecko",
				usdPrice,
				btcPrice)

			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
