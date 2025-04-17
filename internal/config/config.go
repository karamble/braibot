package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vctt94/bisonbotkit/config"
)

// CheckAndUpdateConfig checks if required configuration settings are present
// and prompts the user to enter them if they're missing.
func CheckAndUpdateConfig(cfg *config.BotConfig, appRoot string) error {
	// Ensure the directory exists first
	if err := os.MkdirAll(appRoot, 0755); err != nil {
		return fmt.Errorf("error creating app root directory: %v", err)
	}

	// Check if falapikey exists in ExtraConfig
	if _, exists := cfg.ExtraConfig["falapikey"]; !exists {
		// Prompt for fal.ai API key
		fmt.Print("Enter your fal.ai API key: ")
		reader := bufio.NewReader(os.Stdin)
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read API key: %v", err)
		}
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		// Add falapikey to ExtraConfig
		cfg.ExtraConfig["falapikey"] = apiKey

		// Write to config file
		configPath := filepath.Join(appRoot, "braibot.conf")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open config file: %v", err)
		}
		defer f.Close()

		if _, err := f.WriteString(fmt.Sprintf("falapikey=%s\n", apiKey)); err != nil {
			return fmt.Errorf("failed to write to config file: %v", err)
		}
	}

	return nil
}

// TestAssetServerCredentials tests if the asset server credentials are valid
func TestAssetServerCredentials(serverURL, apiKey string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", serverURL+"/test", nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add API key header
	req.Header.Set("X-API-Key", apiKey)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error connecting to asset server: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("asset server returned error status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("asset server test failed: %s", result.Message)
	}

	return nil
}
