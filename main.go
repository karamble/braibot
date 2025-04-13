// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/braibot/internal/audio"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

var (
	flagAppRoot = flag.String("approot", "~/.braibot", "Path to application data directory")
	debug       = true     // Set to true for debugging
	dbManager   *DBManager // Database manager for user balances
)

// Command represents a bot command
type Command struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error
}

// Available commands
var commands map[string]Command

func init() {
	commands = map[string]Command{
		"help": {
			Name:        "help",
			Description: "Shows this help message",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				helpMsg := "| Command | Description |\n| -------- | ----------- |\n"
				for _, cmd := range commands {
					helpMsg += fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description)
				}
				return bot.SendPM(ctx, pm.Nick, helpMsg)
			},
		},
		"listmodels": {
			Name:        "listmodels",
			Description: "Lists available models for a specific command. Usage: !listmodels [command]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) == 0 {
					return bot.SendPM(ctx, pm.Nick, "Please specify a command. Usage: !listmodels [command]")
				}

				commandName := strings.ToLower(args[0])

				var modelList string
				var models map[string]falapi.Model

				switch commandName {
				case "text2image":
					models = falapi.Text2ImageModels
					modelList = "Available models for text2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
				case "text2speech":
					models = falapi.Text2SpeechModels
					modelList = "Available models for text2speech:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
				default:
					return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image' or 'text2speech'.")
				}

				for _, model := range models {
					modelList += fmt.Sprintf("| %s | %s | $%.2f |\n", model.Name, model.Description, model.Price)
				}

				return bot.SendPM(ctx, pm.Nick, modelList)
			},
		},
		"setmodel": {
			Name:        "setmodel",
			Description: "Sets the model to use for specified commands. Usage: !setmodel [command] [modelname]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) < 2 {
					return bot.SendPM(ctx, pm.Nick, "Please specify a command and a model name. Usage: !setmodel [command] [modelname]")
				}
				commandName := args[0]
				modelName := args[1]

				// Check if the command is valid
				if _, exists := commands[commandName]; !exists {
					return bot.SendPM(ctx, pm.Nick, "Invalid command name. Use !listmodels to see available commands.")
				}

				// Check if the model is valid for the specific command
				var models map[string]falapi.Model
				switch commandName {
				case "text2image":
					models = falapi.Text2ImageModels
				case "text2speech":
					models = falapi.Text2SpeechModels
				default:
					return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image' or 'text2speech'.")
				}

				if _, exists := models[modelName]; exists {
					falapi.SetDefaultModel(commandName, modelName)
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Model for %s set to: %s", commandName, modelName))
				}

				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Invalid model name for %s. Use !listmodels %s to see available models.", commandName, commandName))
			},
		},
		"text2image": {
			Name:        "text2image",
			Description: "Generates an image from text prompt. Usage: !text2image [prompt]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) == 0 {
					return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2image [prompt]")
				}

				prompt := strings.Join(args, " ")

				// Create Fal.ai client
				client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

				// Get model configuration
				modelName, exists := falapi.GetDefaultModel("text2image")
				if !exists {
					return fmt.Errorf("no default model found for text2image")
				}
				model, exists := falapi.GetModel(modelName, "text2image")
				if !exists {
					return fmt.Errorf("model not found: %s", modelName)
				}

				// Generate image
				imageResp, err := client.GenerateImage(ctx, prompt, model.Name, bot, pm.Nick)
				if err != nil {
					return err
				}

				// Assuming the first image is the one we want to send
				if len(imageResp.Images) > 0 {
					imageURL := imageResp.Images[0].URL
					// Fetch the image data
					imgResp, err := http.Get(imageURL)
					if err != nil {
						return err
					}
					defer imgResp.Body.Close()

					imgData, err := io.ReadAll(imgResp.Body)
					if err != nil {
						return err
					}

					// Encode the image data to base64
					encodedImage := base64.StdEncoding.EncodeToString(imgData)

					// Determine the image type from ContentType
					var imageType string
					switch imageResp.Images[0].ContentType {
					case "image/jpeg":
						imageType = "image/jpeg"
					case "image/png":
						imageType = "image/png"
					case "image/webp":
						imageType = "image/webp"
					default:
						imageType = "image/jpeg" // Fallback to jpeg if unknown
					}

					// Create the message with embedded image, using the user's prompt as the alt text
					message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--", url.QueryEscape(prompt), imageType, encodedImage)
					return bot.SendPM(ctx, pm.Nick, message)
				} else {
					return bot.SendPM(ctx, pm.Nick, "No images were generated.")
				}
			},
		},
		"balance": {
			Name:        "balance",
			Description: "Shows your current balance",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				// Convert UID to string ID for database, just like in tip handler
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)
				userIDStr := userID.String()

				// Get balance from database using the proper ID
				balance, err := dbManager.GetBalance(userIDStr)
				if err != nil {
					return fmt.Errorf("failed to get balance: %v", err)
				}

				// Convert to DCR using dcrutil, same as in tip handler
				dcrBalance := dcrutil.Amount(balance / 1e3).ToCoin()

				// Send balance message
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Your current balance is %.8f DCR", dcrBalance))
			},
		},
		"rate": {
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
		},
		"text2speech": {
			Name:        "text2speech",
			Description: "Converts text to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) < 2 {
					voiceList := "Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl"
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Please provide a voice ID and text. Usage: !text2speech [voice_id] [text]\n\n%s", voiceList))
				}

				voiceID := args[0]
				text := strings.Join(args[1:], " ")

				// Create Fal.ai client
				client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

				// Get model configuration
				modelName, exists := falapi.GetDefaultModel("text2speech")
				if !exists {
					return fmt.Errorf("no default model found for text2speech")
				}
				_, exists = falapi.GetModel(modelName, "text2speech")
				if !exists {
					return fmt.Errorf("model not found: %s", modelName)
				}

				// Generate speech
				audioResp, err := client.GenerateSpeech(ctx, text, voiceID, bot, pm.Nick)
				if err != nil {
					return err
				}

				// Fetch the audio data
				resp, err := http.Get(audioResp.Audio.URL)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				audioData, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				// Convert PCM to Opus using the audio package
				oggData, err := audio.ConvertPCMToOpus(audioData)
				if err != nil {
					return err
				}

				// Encode the audio data to base64
				encodedAudio := base64.StdEncoding.EncodeToString(oggData)

				// Create the message with embedded audio
				message := fmt.Sprintf("--embed[alt=%s,type=audio/ogg,data=%s]--", url.QueryEscape(text), encodedAudio)
				return bot.SendPM(ctx, pm.Nick, message)
			},
		},
	}
}

// isCommand checks if a message is a command (starts with !)
func isCommand(msg string) (string, []string, bool) {
	if !strings.HasPrefix(msg, "!") {
		return "", nil, false
	}

	parts := strings.Fields(msg[1:]) // Remove ! and split
	if len(parts) == 0 {
		return "", nil, false
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	return cmd, args, true
}

func realMain() error {
	flag.Parse()

	// Expand and clean the app root path
	appRoot := utils.CleanAndExpandPath(*flagAppRoot)

	// Initialize database manager
	var err error
	dbManager, err = NewDBManager(appRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Initialize logging
	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appRoot, "logs", "braibot.log"),
		DebugLevel:     "info",
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logging: %v", err)
	}
	defer logBackend.Close()

	// Get a logger for the application
	log := logBackend.Logger("BraiBot")

	// Load bot configuration
	cfg, err := config.LoadBotConfig(appRoot, "braibot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create a bidirectional channel for PMs and tips
	pmChan := make(chan types.ReceivedPM)
	tipChan := make(chan types.ReceivedTip)
	tipProgressChan := make(chan types.TipProgressEvent)

	// Set up PM channels/log
	cfg.PMChan = pmChan
	cfg.PMLog = logBackend.Logger("PM")

	// Set up tip channels/logs
	cfg.TipLog = logBackend.Logger("TIP")
	cfg.TipProgressChan = tipProgressChan
	cfg.TipReceivedLog = logBackend.Logger("TIP_RECEIVED")
	cfg.TipReceivedChan = tipChan

	// Create new bot instance
	bot, err := kit.NewBot(cfg, logBackend)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	// Add a goroutine to handle PMs using our bidirectional channel
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg.Message)

			// Check if the message is a command
			if cmd, args, isCmd := isCommand(pm.Msg.Message); isCmd {
				if command, exists := commands[cmd]; exists {
					if err := command.Handler(context.Background(), bot, cfg, pm, args); err != nil {
						log.Warnf("Error executing command %s: %v", cmd, err)
					}
				} else {
					// Send error message for unknown command
					bot.SendPM(context.Background(), pm.Nick, "Unknown command. Use !help to see available commands.")
				}
			}
		}
	}()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add input handling goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			tokens := strings.SplitN(line, " ", 2)
			if len(tokens) != 2 {
				log.Warn("Invalid format. Use: <nick> <message>")
				continue
			}

			nick, msg := tokens[0], tokens[1]
			if err := bot.SendPM(ctx, nick, msg); err != nil {
				log.Warnf("Failed to send PM: %v", err)
				continue
			}
			log.Infof("-> %s: %s", nick, msg)
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("Error reading input: %v", err)
		}
	}()

	// Handle received tips
	go func() {
		for tip := range tipChan {
			// Convert UID to string ID for database
			var userID zkidentity.ShortID
			userID.FromBytes(tip.Uid)
			userIDStr := userID.String()

			// Update user's balance in the database
			err = dbManager.UpdateBalance(userIDStr, tip.AmountMatoms)
			if err != nil {
				log.Errorf("Failed to update balance: %v", err)
				continue
			}

			// Convert to DCR for display
			dcrAmount := dcrutil.Amount(tip.AmountMatoms / 1e3).ToCoin()

			log.Infof("Tip received: %.8f DCR from %s",
				dcrAmount,
				userIDStr)

			// Send thank you message
			bot.SendPM(ctx, userIDStr,
				fmt.Sprintf("Thank you for the tip of %.8f DCR!", dcrAmount))

			// Acknowledge the tip
			bot.AckTipReceived(ctx, tip.SequenceId)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received shutdown signal: %v", sig)
		bot.Close()
		cancel()
	}()

	// Run the bot with the cancellable context
	if err := bot.Run(ctx); err != nil {
		return fmt.Errorf("bot error: %v", err)
	}

	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
