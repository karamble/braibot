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
// It also sets default values for optional settings like billing_enabled.
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

	// Check for billingenabled setting
	if _, exists := cfg.ExtraConfig["billingenabled"]; !exists {
		// Prompt user interactively if setting is missing
		reader := bufio.NewReader(os.Stdin)
		var billingEnabledStr string
		for {
			fmt.Print("Do you want to enable billing? (yes/no): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				// Handle potential read error (e.g., EOF)
				fmt.Printf("\nError reading input: %v. Defaulting billing to DISABLED.\n", err)
				billingEnabledStr = "false"
				break
			}
			input = strings.ToLower(strings.TrimSpace(input))

			if input == "yes" || input == "y" {
				billingEnabledStr = "true"
				break
			} else if input == "no" || input == "n" {
				billingEnabledStr = "false"
				break
			} else {
				fmt.Println("Invalid input. Please enter 'yes' or 'no'.")
			}
		}

		// Store the chosen setting in the config map
		cfg.ExtraConfig["billingenabled"] = billingEnabledStr

		// Append the setting to the config file
		configPath := filepath.Join(appRoot, "braibot.conf")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			// Log error but continue, config map has the value
			fmt.Printf("WARN: Failed to open config file to append billingenabled setting: %v\n", err)
		} else {
			if _, err := f.WriteString(fmt.Sprintf("billingenabled=%s\n", billingEnabledStr)); err != nil {
				fmt.Printf("WARN: Failed to write billingenabled setting to config file: %v\n", err)
			}
			f.Close() // Close file only if opened successfully
		}

	} else {
		// Validate existing setting
		val := strings.ToLower(strings.TrimSpace(cfg.ExtraConfig["billingenabled"]))
		if val != "true" && val != "false" {
			fmt.Printf("Invalid value '%s' for billingenabled found in config, defaulting to true.\n", cfg.ExtraConfig["billingenabled"])
			cfg.ExtraConfig["billingenabled"] = "true"
			// Optionally write the default back to the config file
			// writeDefaultBillingSetting(appRoot, "true") // Need a function that takes value
		} else {
			// Store the validated, lowercased value back
			cfg.ExtraConfig["billingenabled"] = val
		}
	}

	// Check for optional webhook settings
	// Only prompt for webhook settings if the user wants to enable them
	if _, exists := cfg.ExtraConfig["webhookenabled"]; !exists {
		reader := bufio.NewReader(os.Stdin)
		var webhookEnabledStr string
		for {
			fmt.Print("Do you want to enable webhook functionality? (yes/no): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("\nError reading input: %v. Defaulting webhook to DISABLED.\n", err)
				webhookEnabledStr = "false"
				break
			}
			input = strings.ToLower(strings.TrimSpace(input))

			if input == "yes" || input == "y" {
				webhookEnabledStr = "true"
				break
			} else if input == "no" || input == "n" {
				webhookEnabledStr = "false"
				break
			} else {
				fmt.Println("Invalid input. Please enter 'yes' or 'no'.")
			}
		}

		// Store the webhook enabled setting
		cfg.ExtraConfig["webhookenabled"] = webhookEnabledStr

		// Write to config file
		configPath := filepath.Join(appRoot, "braibot.conf")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("WARN: Failed to open config file to append webhookenabled setting: %v\n", err)
		} else {
			if _, err := f.WriteString(fmt.Sprintf("webhookenabled=%s\n", webhookEnabledStr)); err != nil {
				fmt.Printf("WARN: Failed to write webhookenabled setting to config file: %v\n", err)
			}
			f.Close()
		}

		// If webhook is enabled, prompt for URL and API key
		if webhookEnabledStr == "true" {
			// Prompt for webhook URL
			fmt.Print("Enter your webhook URL: ")
			webhookURL, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("\nError reading webhook URL: %v. Webhook functionality will be disabled.\n", err)
				cfg.ExtraConfig["webhookenabled"] = "false"
			} else {
				webhookURL = strings.TrimSpace(webhookURL)
				if webhookURL != "" {
					cfg.ExtraConfig["webhookurl"] = webhookURL

					// Write to config file
					f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Printf("WARN: Failed to open config file to append webhookurl: %v\n", err)
					} else {
						if _, err := f.WriteString(fmt.Sprintf("webhookurl=%s\n", webhookURL)); err != nil {
							fmt.Printf("WARN: Failed to write webhookurl to config file: %v\n", err)
						}
						f.Close()
					}

					// Prompt for webhook API key
					fmt.Print("Enter your webhook API key: ")
					webhookAPIKey, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("\nError reading webhook API key: %v. Webhook functionality will be disabled.\n", err)
						cfg.ExtraConfig["webhookenabled"] = "false"
					} else {
						webhookAPIKey = strings.TrimSpace(webhookAPIKey)
						if webhookAPIKey != "" {
							cfg.ExtraConfig["webhookapikey"] = webhookAPIKey

							// Write to config file
							f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
							if err != nil {
								fmt.Printf("WARN: Failed to open config file to append webhookapikey: %v\n", err)
							} else {
								if _, err := f.WriteString(fmt.Sprintf("webhookapikey=%s\n", webhookAPIKey)); err != nil {
									fmt.Printf("WARN: Failed to write webhookapikey to config file: %v\n", err)
								}
								f.Close()
							}
						} else {
							fmt.Println("Webhook API key cannot be empty. Webhook functionality will be disabled.")
							cfg.ExtraConfig["webhookenabled"] = "false"
						}
					}
				} else {
					fmt.Println("Webhook URL cannot be empty. Webhook functionality will be disabled.")
					cfg.ExtraConfig["webhookenabled"] = "false"
				}
			}
		}
	} else {
		// Validate existing webhook enabled setting
		val := strings.ToLower(strings.TrimSpace(cfg.ExtraConfig["webhookenabled"]))
		if val != "true" && val != "false" {
			fmt.Printf("Invalid value '%s' for webhookenabled found in config, defaulting to false.\n", cfg.ExtraConfig["webhookenabled"])
			cfg.ExtraConfig["webhookenabled"] = "false"
		} else {
			// Store the validated, lowercased value back
			cfg.ExtraConfig["webhookenabled"] = val
		}
	}

	return nil
}

// Optional: Helper function to write the default billing setting if needed
/*
func writeDefaultBillingSetting(appRoot string, value string) {
	configPath := filepath.Join(appRoot, "braibot.conf")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("WARN: Failed to open config file to write default billingenabled: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("billingenabled=%s\n", value)); err != nil {
		fmt.Printf("WARN: Failed to write default billingenabled to config file: %v\n", err)
	}
}
*/

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
