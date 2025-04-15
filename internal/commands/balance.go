package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// BalanceCommand returns the balance command
func BalanceCommand(dbManager *database.DBManager, debug bool) Command {
	return Command{
		Name:        "balance",
		Description: "Check your current balance",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
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

			// Get current exchange rate
			dcrPrice, _, err := GetDCRPrice()
			if err != nil {
				return fmt.Errorf("failed to get DCR price: %v", err)
			}

			// Calculate USD value
			usdValue := balanceDCR * dcrPrice

			// Send balance information
			message := fmt.Sprintf("💰 Your Balance:\n• %.8f DCR\n• $%.2f USD",
				balanceDCR, usdValue)
			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
