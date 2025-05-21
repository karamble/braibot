package utils

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	kit "github.com/vctt94/bisonbotkit"
)

// ErrInsufficientBalance is a custom error type for insufficient funds.
type ErrInsufficientBalance struct {
	Message string
}

// Error implements the error interface.
func (e *ErrInsufficientBalance) Error() string {
	return e.Message
}

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

	// Only send billing information in PMs
	billingMsg := FormatBillingMessage(result.ChargedDCR, result.ChargedUSD, result.NewBalance)
	return bot.SendPM(ctx, pm.Nick, billingMsg)
}

// CheckBalance checks if a user has sufficient balance for a given cost in USD, without deducting.
// It returns the required DCR amount, the current balance in DCR,
// and potentially an ErrInsufficientBalance or other critical error.
// If billingEnabled is false, it returns success (nil error).
func CheckBalance(ctx context.Context, dbManager *database.DBManager, userID []byte, costUSD float64, debug bool, billingEnabled bool) (requiredDCR float64, currentBalanceDCR float64, err error) {
	// Get current balance regardless of billing status for reporting
	userIDStr := GetUserIDString(userID)
	balanceAtoms, balanceErr := dbManager.GetBalance(userIDStr)
	if balanceErr != nil {
		// Return this error even if billing is disabled, as it prevents knowing the balance
		err = fmt.Errorf("failed to get balance: %v", balanceErr)
		return
	}
	currentBalanceDCR = float64(balanceAtoms) / 1e11

	// If billing is disabled, return success (nil error)
	if !billingEnabled {
		requiredDCR = 0 // No cost applied
		return          // err is nil
	}

	// --- Billing is enabled, perform normal checks ---

	// Convert USD cost to DCR
	requiredDCR, err = USDToDCR(costUSD)
	if err != nil {
		err = fmt.Errorf("failed to convert USD to DCR: %v", err)
		return
	}

	// Convert DCR amount to atoms for comparison (1 DCR = 1e11 atoms)
	dcrAtoms := int64(requiredDCR * 1e11)

	// Debug information
	if debug {
		fmt.Print(FormatDebugBalanceInfo(userIDStr, balanceAtoms, costUSD, requiredDCR, dcrAtoms))
	}

	// Check if user has sufficient balance
	if balanceAtoms < dcrAtoms {
		// Return the specific error type with the formatted message
		err = &ErrInsufficientBalance{
			Message: FormatInsufficientBalanceMessageWithUSD(requiredDCR, currentBalanceDCR, costUSD),
		}
		return // Return the insufficient balance error
	}

	// Sufficient balance, return success (nil error)
	return
}

// DeductBalance deducts the specified cost in USD from the user's balance.
// It assumes the balance check has already passed IF billing is enabled.
// Returns the amount charged in DCR, the new balance in DCR, and any error encountered.
// If billingEnabled is false, it returns zero charged and the current balance without hitting the DB.
func DeductBalance(ctx context.Context, dbManager *database.DBManager, userID []byte, costUSD float64, debug bool, billingEnabled bool) (chargedDCR float64, newBalanceDCR float64, err error) {
	// Get current balance first
	currentBalanceDCR, balanceErr := dbManager.GetUserBalance(userID) // Assuming GetUserBalance returns DCR
	if balanceErr != nil {
		err = fmt.Errorf("failed to get current balance before deduction attempt: %v", balanceErr)
		return
	}

	// If billing is disabled, do nothing and return current balance
	if !billingEnabled {
		chargedDCR = 0
		newBalanceDCR = currentBalanceDCR
		return // Success (no-op)
	}

	// --- Billing is enabled, perform deduction ---

	// Deduct balance using CheckAndDeductBalance
	// Note: CheckAndDeductBalance itself likely converts USD internally based on its signature
	hasBalanceAfterDeduct, err := dbManager.CheckAndDeductBalance(userID, costUSD, debug)
	if err != nil {
		err = fmt.Errorf("failed to deduct balance: %v", err)
		newBalanceDCR = currentBalanceDCR // Return pre-deduction balance on error
		return
	}
	// This check might be redundant if CheckBalance was called first,
	// but CheckAndDeductBalance performs an atomic check-and-deduct.
	if !hasBalanceAfterDeduct {
		// This indicates a potential race condition or logic error if CheckBalance passed moments before.
		err = fmt.Errorf("deduction failed despite prior check (insufficient funds or race condition)")
		newBalanceDCR = currentBalanceDCR // Return pre-deduction balance on error
		return
	}

	// Convert charged amount back to DCR for reporting (if needed, depends on CheckAndDeductBalance return)
	// Assuming costUSD was the amount successfully deducted in USD terms.
	chargedDCR, convertErr := USDToDCR(costUSD)
	if convertErr != nil {
		// Log this error, but the deduction likely succeeded, so proceed with getting new balance.
		fmt.Printf("WARN: Failed to convert charged USD to DCR for reporting: %v\n", convertErr)
		chargedDCR = 0 // Assign a placeholder
	}

	// Get updated balance for result
	finalBalanceDCR, finalBalanceErr := dbManager.GetUserBalance(userID) // Assuming GetUserBalance returns DCR
	if finalBalanceErr != nil {
		// The deduction likely succeeded, but we failed to get the final balance.
		err = fmt.Errorf("deduction likely succeeded, but failed to get updated balance: %v", finalBalanceErr)
		newBalanceDCR = currentBalanceDCR // Return pre-deduction balance as best guess
		return
	}
	newBalanceDCR = finalBalanceDCR

	// Debug information after deduction
	if debug {
		fmt.Print(FormatDebugAfterDeduction(int64(newBalanceDCR * 1e11)))
	}

	return // Success
}
