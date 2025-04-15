package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	lastRateUpdate time.Time
	dcrRate        float64
)

// USDToDCR converts USD amount to DCR using current exchange rate
func USDToDCR(usdAmount float64) (float64, error) {
	// Update rate if needed (every 5 minutes)
	if time.Since(lastRateUpdate) > 5*time.Minute {
		rate, err := getDCRRate()
		if err != nil {
			return 0, fmt.Errorf("failed to get DCR rate: %v", err)
		}
		dcrRate = rate
		lastRateUpdate = time.Now()
	}

	return usdAmount / dcrRate, nil
}

// getDCRRate fetches the current DCR/USD rate from CoinGecko
func getDCRRate() (float64, error) {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=decred&vs_currencies=usd")
	if err != nil {
		return 0, fmt.Errorf("failed to fetch DCR rate: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode DCR rate: %v", err)
	}

	rate, ok := result["decred"]["usd"]
	if !ok {
		return 0, fmt.Errorf("DCR rate not found in response")
	}

	return rate, nil
}
