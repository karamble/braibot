package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/braibot/internal/database"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// BalanceCommand returns the balance command
func BalanceCommand(dbManager *database.DBManager) Command {
	return Command{
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
	}
}
