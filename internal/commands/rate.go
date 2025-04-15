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

// GetDCRPrice gets the current DCR price in USD and BTC from CoinGecko
func GetDCRPrice() (float64, float64, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request to CoinGecko API
	req, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/simple/price?ids=decred&vs_currencies=usd,btc", nil)
	if err != nil {
		return 0, 0, fmt.Errorf("error creating request: %v", err)
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("error fetching rates: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, fmt.Errorf("error parsing rates: %v", err)
	}

	dcrData, ok := result["decred"]
	if !ok {
		return 0, 0, fmt.Errorf("no data returned for DCR")
	}

	usdPrice, ok := dcrData["usd"]
	if !ok {
		return 0, 0, fmt.Errorf("no USD price found for DCR")
	}

	btcPrice, ok := dcrData["btc"]
	if !ok {
		return 0, 0, fmt.Errorf("no BTC price found for DCR")
	}

	return usdPrice, btcPrice, nil
}

// USDToDCR converts a USD amount to DCR using current exchange rate
func USDToDCR(usdAmount float64) (float64, error) {
	dcrPrice, _, err := GetDCRPrice()
	if err != nil {
		return 0, err
	}

	// Calculate DCR amount (USD amount / DCR price)
	dcrAmount := usdAmount / dcrPrice
	return dcrAmount, nil
}

// RateCommand returns the rate command
func RateCommand() Command {
	return Command{
		Name:        "rate",
		Description: "Shows current DCR exchange rate in USD and BTC",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// Send a status message to indicate we're fetching rates
			bot.SendPM(ctx, pm.Nick, "Fetching current exchange rates...")

			// Get DCR prices
			usdPrice, btcPrice, err := GetDCRPrice()
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
