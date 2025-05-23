// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/commands"
	braiconfig "github.com/karamble/braibot/internal/config"
	"github.com/karamble/braibot/internal/database"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	kit "github.com/vctt94/bisonbotkit"
	botkitconfig "github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	botkitutils "github.com/vctt94/bisonbotkit/utils"
)

var (
	flagAppRoot = flag.String("approot", "~/.braibot", "Path to application data directory")
	flagDebug   = flag.Bool("debug", false, "Enable debug mode")
	dbManager   *database.DBManager     // Database manager for user balances
	debug       bool                    // Debug mode flag
	welcomeSent = make(map[string]bool) // Track users who have received welcome message
)

func realMain() error {
	flag.Parse()

	// Set debug mode
	debug = *flagDebug

	// Expand and clean the app root path
	appRoot := botkitutils.CleanAndExpandPath(*flagAppRoot)

	// Initialize database manager
	var err error
	dbManager, err = database.NewDBManager(appRoot)
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
	cfg, err := botkitconfig.LoadBotConfig(appRoot, "braibot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Wait for braibot.conf to be created
	configPath := filepath.Join(appRoot, "braibot.conf")
	configFileFound := false
	for i := 0; i < 30; i++ { // Try for 3 seconds (30 * 100ms)
		if _, err := os.Stat(configPath); err == nil {
			// File exists, proceed with checkAndUpdateConfig for braibot specific configuration fields
			if err := braiconfig.CheckAndUpdateConfig(cfg, appRoot); err != nil {
				return fmt.Errorf("failed to check and update config: %v", err)
			}
			configFileFound = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !configFileFound {
		return fmt.Errorf("config file '%s' not found after waiting", configPath)
	}

	// Create a bidirectional channel for PMs and tips
	pmChan := make(chan types.ReceivedPM)
	tipChan := make(chan types.ReceivedTip)
	tipProgressChan := make(chan types.TipProgressEvent)
	gcChan := make(chan types.GCReceivedMsg) // Add group chat channel

	// Set up PM channels/log
	cfg.PMChan = pmChan
	cfg.PMLog = logBackend.Logger("PM")

	// Set up GC channels/log
	cfg.GCChan = gcChan
	cfg.GCLog = logBackend.Logger("GC")

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

	// Initialize command registry
	commandRegistry := commands.InitializeCommands(dbManager, cfg, bot, debug)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received shutdown signal: %v", sig)
		bot.Close()
		cancel()
	}()

	// Add a goroutine to handle PMs using our bidirectional channel
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg.Message)

			// Convert UID to string ID for tracking
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			userIDStr := userID.String()

			// Check if the message is a command
			if cmd, args, isCmd := commands.IsCommand(pm.Msg.Message); isCmd {
				// Mark welcome as sent when user sends any command
				welcomeSent[userIDStr] = true

				if command, exists := commandRegistry.Get(cmd); exists {
					// Construct MessageContext for PM
					var senderID zkidentity.ShortID
					senderID.FromBytes(pm.Uid)
					msgCtx := braibottypes.MessageContext{
						Nick:    pm.Nick,
						Uid:     pm.Uid,
						Message: pm.Msg.Message,
						IsPM:    true,
						Sender:  senderID,
					}
					msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))
					handleErr := command.Handler.Handle(ctx, msgCtx, args, msgSender, dbManager)
					if handleErr != nil {
						// Check if the error is specifically ErrInsufficientBalance
						var insufErr *utils.ErrInsufficientBalance
						if errors.Is(handleErr, insufErr) {
							// Send the specific error message as PM, don't log as warning
							if pmErr := bot.SendPM(ctx, pm.Nick, handleErr.Error()); pmErr != nil {
								log.Warnf("Failed to send insufficient balance PM to %s: %v", pm.Nick, pmErr)
							}
						} else {
							// Log other command execution errors as warnings
							log.Warnf("Error executing command %s for user %s: %v", cmd, pm.Nick, handleErr)
						}
					}
				} else {
					// Send error message for unknown command
					bot.SendPM(ctx, pm.Nick, fmt.Sprintf("👋 Hi %s!\n\nI don't recognize that command. Use **!help** to see available commands.", pm.Nick))
				}
			} else if utils.IsAudioNote(pm.Msg.Message) {
				// Handle audio note
				audioData, err := utils.ExtractAudioNoteData(pm.Msg.Message)
				if err != nil {
					log.Warnf("Failed to extract audio data from message: %v", err)
					bot.SendPM(ctx, pm.Nick, "Sorry, I couldn't process your audio note. Please try again.")
					continue
				}

				// Construct MessageContext for PM
				var senderID zkidentity.ShortID
				senderID.FromBytes(pm.Uid)
				msgCtx := braibottypes.MessageContext{
					Nick:    pm.Nick,
					Uid:     pm.Uid,
					Message: pm.Msg.Message,
					IsPM:    true,
					Sender:  senderID,
				}

				// Create a message sender
				msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

				// Get the AI command handler
				aiCommand, exists := commandRegistry.Get("ai")
				if !exists {
					log.Warnf("AI command not found in registry")
					bot.SendPM(ctx, pm.Nick, "Sorry, the AI processing feature is currently unavailable.")
					continue
				}

				// Execute the AI command with the audio data
				handleErr := aiCommand.Handler.Handle(ctx, msgCtx, []string{audioData}, msgSender, dbManager)
				if handleErr != nil {
					log.Warnf("Error processing audio note: %v", handleErr)
					bot.SendPM(ctx, pm.Nick, "Sorry, I couldn't process your audio note. Please try again.")
				}
			} else if !welcomeSent[userIDStr] {
				// Send welcome message for non-command messages if not sent before
				welcomeMsg := fmt.Sprintf("👋 Hi %s! I'm BraiBot, your AI assistant powered by Decred.\n\n"+
					"To get started, use **!help** to see available commands.\n"+
					"You can also send me a tip to use AI features or\ncheck your balance with **!balance**.",
					pm.Nick)

				if err := bot.SendPM(ctx, pm.Nick, welcomeMsg); err != nil {
					log.Warnf("Error sending welcome message: %v", err)
				} else {
					// Mark welcome as sent for this user
					welcomeSent[userIDStr] = true
				}
			}
		}
	}()

	// Add a goroutine to handle group chat messages
	go func() {
		for gc := range gcChan {
			log.Infof("Received GC message from %s in %s: %s", gc.Nick, gc.GcAlias, gc.Msg.Message)

			// Convert UID to string ID for tracking
			var userID zkidentity.ShortID
			userID.FromBytes(gc.Uid)

			// Check if the message is a command
			if cmd, args, isCmd := commands.IsCommand(gc.Msg.Message); isCmd {
				if command, exists := commandRegistry.Get(cmd); exists {
					var senderID zkidentity.ShortID
					senderID.FromBytes(gc.Uid)
					msgCtx := braibottypes.MessageContext{
						Nick:    gc.Nick,
						Uid:     gc.Uid,
						Message: gc.Msg.Message,
						IsPM:    false,
						Sender:  senderID,
						GC:      gc.GcAlias,
					}
					msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))
					handleErr := command.Handler.Handle(ctx, msgCtx, args, msgSender, dbManager)
					if handleErr != nil {
						// Check if the error is specifically ErrInsufficientBalance
						var insufErr *utils.ErrInsufficientBalance
						if errors.Is(handleErr, insufErr) {
							// Send the specific error message to the group chat
							if gcErr := bot.SendGC(ctx, gc.GcAlias, handleErr.Error()); gcErr != nil {
								log.Warnf("Failed to send insufficient balance message to GC %s: %v", gc.GcAlias, gcErr)
							}
						} else {
							// Log other command execution errors as warnings
							log.Warnf("Error executing command %s for user %s in GC %s: %v", cmd, gc.Nick, gc.GcAlias, handleErr)
						}
					}
				} else {
					// Send error message for unknown command to the group chat
					bot.SendGC(ctx, gc.GcAlias, fmt.Sprintf("👋 Hi %s!\n\nI don't recognize that command. Use **!help** to see available commands.", gc.Nick))
				}
			}
		}
	}()

	// Add input handling goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
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
			dcrAmount := float64(tip.AmountMatoms) / 1e11

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

	// Run the bot
	err = bot.Run(ctx)
	log.Infof("Bot exited: %v", err)
	return err
}

func main() {
	if err := realMain(); err != nil {
		// Don't treat context canceled as an error since it's expected during shutdown
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
