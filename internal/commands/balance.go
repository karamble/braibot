package commands

import (
	"context"
	"fmt"

	braibottypes "github.com/karamble/braibot/internal/types"
)

// BalanceCommand returns the balance command
func BalanceCommand() braibottypes.Command {
	return braibottypes.Command{
		Name:        "balance",
		Description: "💰 Show your current balance",
		Category:    "�� Basic",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Only respond in private messages
			if !msgCtx.IsPM {
				return nil
			}

			userID := string(msgCtx.Uid)
			balance, err := db.GetBalance(userID)
			if err != nil {
				return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to get balance: %v", err))
			}
			balanceMsg := fmt.Sprintf("💰 Your Balance: %d atoms", balance)
			return sender.SendMessage(ctx, msgCtx, balanceMsg)
		}),
	}
}
