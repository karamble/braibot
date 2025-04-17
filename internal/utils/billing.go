package utils

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	kit "github.com/vctt94/bisonbotkit"
)

// BillingResult contains the result of a billing operation
type BillingResult struct {
	Success      bool
	ChargedDCR   float64
	ChargedUSD   float64
	NewBalance   float64
	ErrorMessage string
}

// CheckAndProcessBilling handles the complete billing process for a user
// Returns a BillingResult with the operation details
func CheckAndProcessBilling(ctx context.Context, bot *kit.Bot, dbManager *database.DBManager, pm types.ReceivedPM, costUSD float64, debug bool) (*BillingResult, error) {
	// Convert USD cost to DCR
	dcrAmount, err := USDToDCR(costUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert USD to DCR: %v", err)
	}

	// Convert DCR amount to atoms for comparison (1 DCR = 1e11 atoms)
	dcrAtoms := int64(dcrAmount * 1e11)

	// Get user balance in atoms
	userIDStr := GetUserIDString(pm.Uid)
	balance, err := dbManager.GetBalance(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	// Debug information
	if debug {
		fmt.Print(FormatDebugBalanceInfo(userIDStr, balance, costUSD, dcrAmount, dcrAtoms))
	}

	// Check if user has sufficient balance
	if balance < dcrAtoms {
		balanceDCR := float64(balance) / 1e11
		return &BillingResult{
			Success:      false,
			ErrorMessage: FormatInsufficientBalanceMessageWithUSD(dcrAmount, balanceDCR, costUSD),
		}, nil
	}

	// Deduct balance using CheckAndDeductBalance
	hasBalance, err := dbManager.CheckAndDeductBalance(pm.Uid, costUSD, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct balance: %v", err)
	}
	if !hasBalance {
		return &BillingResult{
			Success:      false,
			ErrorMessage: "Failed to deduct balance. Please try again.",
		}, nil
	}

	// Get updated balance for billing message
	newBalance, err := dbManager.GetUserBalance(pm.Uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated balance: %v", err)
	}

	// Debug information after deduction
	if debug {
		fmt.Print(FormatDebugAfterDeduction(int64(newBalance * 1e11)))
	}

	return &BillingResult{
		Success:    true,
		ChargedDCR: dcrAmount,
		ChargedUSD: costUSD,
		NewBalance: newBalance,
	}, nil
}

// SendBillingMessage sends a billing message to the user
func SendBillingMessage(ctx context.Context, bot *kit.Bot, pm types.ReceivedPM, result *BillingResult) error {
	if !result.Success {
		return bot.SendPM(ctx, pm.Nick, result.ErrorMessage)
	}

	billingMsg := FormatBillingMessage(result.ChargedDCR, result.ChargedUSD, result.NewBalance)
	return bot.SendPM(ctx, pm.Nick, billingMsg)
}
