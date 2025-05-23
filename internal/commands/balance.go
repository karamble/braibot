package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
)

// BalanceCommand returns the balance command
func BalanceCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "balance",
		Description: "💰 Show your current balance",
		Category:    "Basic",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Only respond in private messages
			if !msgCtx.IsPM {
				return nil
			}

			// Convert UID to string ID for database
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			userIDStr := userID.String()

			balance, err := db.GetBalance(userIDStr)
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to get balance: %v", err))
			}

			// Convert atoms to DCR
			balanceDCR := float64(balance) / 1e11

			// Get current exchange rate for USD value
			dcrPrice, _, err := utils.GetDCRPrice()
			if err != nil {
				// Log the error but continue, showing balance without USD value
				fmt.Printf("ERROR [balance] Failed to get DCR price: %v\n", err)
				balanceMsg := fmt.Sprintf("💰 Your Balance: %s DCR", utils.FormatThousands(balanceDCR))
				return sender.SendMessage(ctx, msgCtx, balanceMsg)
			}

			// Calculate USD value
			usdValue := balanceDCR * dcrPrice

			// Format balance message with both DCR and USD values
			balanceMsg := fmt.Sprintf("💰 Your Balance:\n• DCR: %s DCR\n• USD: $%s USD",
				utils.FormatThousands(balanceDCR),
				utils.FormatThousands(usdValue))
			return sender.SendMessage(ctx, msgCtx, balanceMsg)
		}),
	}
}
