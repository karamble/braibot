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
		Description: "💲 Show the current DCR and BTC exchange rates.",
		Category:    "🎯 Basic",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// Send a status message to indicate we're fetching rates
			bot.SendPM(ctx, pm.Nick, "Fetching current exchange rates...")

			// Get DCR prices using the utility function
			dcrUsdPrice, dcrBtcPrice, err := utils.GetDCRPrice()
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching DCR rates: %v", err))
			}

			// Get BTC price using the utility function
			btcUsdPrice, err := utils.GetBTCPrice()
			if err != nil {
				// Optionally log the error but don't block the DCR rates from showing
				fmt.Printf("ERROR [RateCommand] Failed to get BTC/USD price: %v\n", err)
				btcUsdPrice = 0 // Set to 0 if fetch failed
			}

			// Format the response with more detailed information
			message := fmt.Sprintf("Current Exchange Rates (Source: CoinGecko):\n"+
				"1 DCR = $%s USD\n"+
				"1 DCR = %.8f BTC\n"+
				"1 BTC = $%s USD",
				utils.FormatThousands(dcrUsdPrice),
				dcrBtcPrice,
				utils.FormatThousands(btcUsdPrice))

			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
