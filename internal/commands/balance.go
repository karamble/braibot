package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/utils"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// BalanceCommand returns the balance command
func BalanceCommand(dbManager *database.DBManager, debug bool, billingEnabled bool) Command {
	return Command{
		Name:        "balance",
		Description: "💰 Check your current DCR balance available for AI tasks.",
		Category:    "🎯 Basic",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// If billing is disabled, inform the user and exit
			if !billingEnabled {
				return bot.SendPM(ctx, pm.Nick, "💰 Billing is currently disabled for this bot.")
			}

			// --- Billing is enabled, proceed with balance check ---

			// Convert UID to string ID for database
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			userIDStr := userID.String()

			// Get current balance in atoms
			balance, err := dbManager.GetBalance(userIDStr)
			if err != nil {
				return fmt.Errorf("failed to get balance: %v", err)
			}

			// Convert balance to DCR for display (1 DCR = 1e11 atoms)
			balanceDCR := float64(balance) / 1e11

			// Debug information
			if debug {
				fmt.Printf("DEBUG - Balance command:\n")
				fmt.Printf("  User ID: %s\n", userIDStr)
				fmt.Printf("  Balance (atoms): %d\n", balance)
				fmt.Printf("  Balance in DCR: %.8f\n", balanceDCR)
			}

			// Get current exchange rate using utils
			dcrPrice, _, err := utils.GetDCRPrice()
			if err != nil {
				// Inform user about the rate error but still show balance
				// Use only DCR balance if rate fails
				msg := fmt.Sprintf("💰 Your Balance:\n• %.8f DCR\n(Could not fetch current DCR/USD rate: %v)", balanceDCR, err)
				return bot.SendPM(ctx, pm.Nick, msg)
			}

			// Send balance information using the formatter
			message := utils.FormatBalanceMessage(balanceDCR, dcrPrice)
			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
