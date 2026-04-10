package utils

import (
	"context"
	"fmt"

	"github.com/karamble/braibot/internal/database"
)

// ErrInsufficientBalance is a custom error type for insufficient funds.
type ErrInsufficientBalance struct {
	Message string
}

// Error implements the error interface.
func (e *ErrInsufficientBalance) Error() string {
	return e.Message
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

	// Convert USD cost to DCR and then to atoms for the database layer
	chargedDCR, convertErr := USDToDCR(costUSD)
	if convertErr != nil {
		err = fmt.Errorf("failed to convert USD to DCR: %v", convertErr)
		newBalanceDCR = currentBalanceDCR
		return
	}
	costAtoms := int64(chargedDCR * 1e11)

	// Deduct balance using CheckAndDeductBalance (atomic check-and-deduct)
	hasBalanceAfterDeduct, err := dbManager.CheckAndDeductBalance(userID, costAtoms, debug)
	if err != nil {
		err = fmt.Errorf("failed to deduct balance: %v", err)
		newBalanceDCR = currentBalanceDCR // Return pre-deduction balance on error
		return
	}
	if !hasBalanceAfterDeduct {
		err = fmt.Errorf("deduction failed despite prior check (insufficient funds or race condition)")
		newBalanceDCR = currentBalanceDCR // Return pre-deduction balance on error
		return
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
