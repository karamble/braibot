// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
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
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/braibot/internal/commands"
	braiconfig "github.com/karamble/braibot/internal/config"
	"github.com/karamble/braibot/internal/database"
	kit "github.com/vctt94/bisonbotkit"
	botkitconfig "github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

var (
	flagAppRoot = flag.String("approot", "~/.braibot", "Path to application data directory")
	debug       = false             // Set to true for debugging
	dbManager   *database.DBManager // Database manager for user balances
)

func realMain() error {
	flag.Parse()

	// Expand and clean the app root path
	appRoot := utils.CleanAndExpandPath(*flagAppRoot)

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
	for i := 0; i < 5; i++ { // Try for 5 seconds
		if _, err := os.Stat(configPath); err == nil {
			// File exists, proceed with checkAndUpdateConfig for braibot specific configuration fields
			if err := braiconfig.CheckAndUpdateConfig(cfg, appRoot); err != nil {
				return fmt.Errorf("failed to check and update config: %v", err)
			}
			break
		}
		time.Sleep(time.Second)
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

	// Initialize command registry
	commandRegistry := commands.InitializeCommands(dbManager, debug)

	// Add a goroutine to handle PMs using our bidirectional channel
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg.Message)

			// Check if the message is a command
			if cmd, args, isCmd := commands.IsCommand(pm.Msg.Message); isCmd {
				if command, exists := commandRegistry.Get(cmd); exists {
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
