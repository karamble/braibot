package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// RateCommand returns the rate command
func RateCommand() Command {
	return Command{
		Name:        "rate",
		Description: "Shows current DCR exchange rate in USD and BTC",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// Send a status message to indicate we're fetching rates
			bot.SendPM(ctx, pm.Nick, "Fetching current exchange rates...")

			// Create HTTP client with timeout
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			// Make request to dcrdata API
			resp, err := client.Get("https://explorer.dcrdata.org/api/exchangerate")
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching rates: %v", err))
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != http.StatusOK {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error: API returned status %d", resp.StatusCode))
			}

			var rates struct {
				DCRPrice float64 `json:"dcrPrice"`
				BTCPrice float64 `json:"btcPrice"`
				Time     int64   `json:"time"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error parsing rates: %v", err))
			}

			// Convert timestamp to human readable format
			timeStr := time.Unix(rates.Time, 0).Format(time.RFC1123)

			// Format the response with more detailed information
			message := fmt.Sprintf("Current DCR Exchange Rates (as of %s):\n"+
				"USD: $%.2f\n"+
				"BTC: %.8f BTC\n"+
				"Source: dcrdata",
				timeStr, rates.DCRPrice, rates.BTCPrice)

			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
