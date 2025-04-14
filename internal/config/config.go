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

	// Check for falapikey
	if _, exists := cfg.ExtraConfig["falapikey"]; !exists {
		fmt.Print("Enter your Fal.ai API key: ")
		reader := bufio.NewReader(os.Stdin)
		apiKey, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading API key: %v", err)
		}
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		// Update config in memory
		cfg.ExtraConfig["falapikey"] = apiKey

		// Append to config file
		configPath := filepath.Join(appRoot, "braibot.conf")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening config file: %v", err)
		}
		defer f.Close()

		if _, err := f.WriteString(fmt.Sprintf("falapikey=%s\n", apiKey)); err != nil {
			return fmt.Errorf("error writing to config file: %v", err)
		}
	}

	// Check for asset server settings
	if _, exists := cfg.ExtraConfig["use_assetserver"]; !exists {
		var useAssetServer string
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Do you want to use the asset server? (y/N): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading asset server preference: %v", err)
			}
			useAssetServer = strings.TrimSpace(strings.ToLower(input))
			if useAssetServer == "y" || useAssetServer == "n" || useAssetServer == "" {
				break
			}
			fmt.Println("Please enter 'y' or 'n' (or press Enter for no)")
		}

		// Convert y/n to true/false, default to false if empty
		useAssetServerValue := "false"
		if useAssetServer == "y" {
			useAssetServerValue = "true"
		}

		// Update config in memory
		cfg.ExtraConfig["use_assetserver"] = useAssetServerValue

		// Append to config file
		configPath := filepath.Join(appRoot, "braibot.conf")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening config file: %v", err)
		}
		defer f.Close()

		if _, err := f.WriteString(fmt.Sprintf("use_assetserver=%s\n", useAssetServerValue)); err != nil {
			return fmt.Errorf("error writing to config file: %v", err)
		}

		// If asset server is enabled, get additional settings
		if useAssetServerValue == "true" {
			// Get asset server URL
			fmt.Print("Enter the asset server URL (e.g., https://assets.example.com): ")
			assetServerURL, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading asset server URL: %v", err)
			}
			assetServerURL = strings.TrimSpace(assetServerURL)
			if assetServerURL == "" {
				return fmt.Errorf("asset server URL cannot be empty")
			}

			// Validate URL format
			if !strings.HasPrefix(assetServerURL, "https://") {
				return fmt.Errorf("asset server URL must start with https://")
			}

			// Update config in memory
			cfg.ExtraConfig["assetserver_url"] = assetServerURL

			// Append to config file
			if _, err := f.WriteString(fmt.Sprintf("assetserver_url=%s\n", assetServerURL)); err != nil {
				return fmt.Errorf("error writing to config file: %v", err)
			}

			// Get API secret
			fmt.Print("Enter the asset server API secret: ")
			apiSecret, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading API secret: %v", err)
			}
			apiSecret = strings.TrimSpace(apiSecret)
			if apiSecret == "" {
				return fmt.Errorf("API secret cannot be empty")
			}

			// Update config in memory
			cfg.ExtraConfig["assetserver_secret"] = apiSecret

			// Append to config file
			if _, err := f.WriteString(fmt.Sprintf("assetserver_secret=%s\n", apiSecret)); err != nil {
				return fmt.Errorf("error writing to config file: %v", err)
			}

			// Test the credentials
			fmt.Println("Testing asset server credentials...")
			if err := TestAssetServerCredentials(assetServerURL, apiSecret); err != nil {
				fmt.Printf("Asset server test failed: %v\n", err)
				fmt.Println("Asset server will be disabled. You can try again later by editing the config file.")

				// Set use_assetserver to false
				useAssetServerValue = "false"
				cfg.ExtraConfig["use_assetserver"] = "false"

				// Update the config file with the new value
				configPath := filepath.Join(appRoot, "braibot.conf")
				f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("error opening config file: %v", err)
				}
				defer f.Close()

				if _, err := f.WriteString(fmt.Sprintf("use_assetserver=false\n")); err != nil {
					return fmt.Errorf("error writing to config file: %v", err)
				}
			} else {
				fmt.Println("Asset server credentials verified successfully!")
			}
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
