package database

import (
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
)

// CheckAndDeductBalance checks if a user has sufficient balance and deducts the cost if they do.
// costAtoms is the cost in atoms (1 DCR = 1e11 atoms). The caller is responsible for
// converting from USD/DCR to atoms before calling this function.
// Returns true if the operation was successful, false otherwise.
func (db *DBManager) CheckAndDeductBalance(uid []byte, costAtoms int64, debug bool) (bool, error) {
	// Convert UID to string ID for database
	var userID zkidentity.ShortID
	userID.FromBytes(uid)
	userIDStr := userID.String()

	// Get current balance
	balance, err := db.GetBalance(userIDStr)
	if err != nil {
		return false, fmt.Errorf("failed to get balance: %v", err)
	}

	// Debug information
	if debug {
		fmt.Printf("DEBUG - Balance check:\n")
		fmt.Printf("  User ID: %s\n", userIDStr)
		fmt.Printf("  Current balance (atoms): %d\n", balance)
		fmt.Printf("  Cost in atoms: %d\n", costAtoms)
		fmt.Printf("  Cost in DCR: %.8f\n", float64(costAtoms)/1e11)
		fmt.Printf("  Balance in DCR: %.8f\n", float64(balance)/1e11)
	}

	// Check if user has sufficient balance
	if balance < costAtoms {
		balanceDCR := float64(balance) / 1e11
		costDCR := float64(costAtoms) / 1e11
		return false, fmt.Errorf("insufficient balance. Required: %.8f DCR, Current: %.8f DCR", costDCR, balanceDCR)
	}

	// Deduct the cost from the user's balance (negative amount)
	err = db.UpdateBalance(userIDStr, -costAtoms)
	if err != nil {
		return false, fmt.Errorf("failed to deduct balance: %v", err)
	}

	// Debug information after deduction
	if debug {
		fmt.Printf("DEBUG - After deduction:\n")
		fmt.Printf("  New balance (atoms): %d\n", balance-costAtoms)
		fmt.Printf("  New balance in DCR: %.8f\n", float64(balance-costAtoms)/1e11)
	}

	return true, nil
}

// GetUserBalance gets the current balance of a user in DCR
func (db *DBManager) GetUserBalance(uid []byte) (float64, error) {
	// Convert UID to string ID for database
	var userID zkidentity.ShortID
	userID.FromBytes(uid)
	userIDStr := userID.String()

	// Get current balance in atoms
	balanceAtoms, err := db.GetBalance(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	// Convert atoms to DCR
	return float64(balanceAtoms) / 1e11, nil
}
