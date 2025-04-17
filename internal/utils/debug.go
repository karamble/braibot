package utils

import (
	"fmt"
)

// FormatDebugBalanceInfo formats debug information about a user's balance
func FormatDebugBalanceInfo(userID string, balanceAtoms int64, costUSD float64, costDCR float64, costAtoms int64) string {
	return fmt.Sprintf("DEBUG - Balance check:\n"+
		"  User ID: %s\n"+
		"  Current balance (atoms): %d\n"+
		"  Cost in USD: $%.2f\n"+
		"  Cost in DCR: %.8f\n"+
		"  Cost in atoms: %d\n"+
		"  Balance in DCR: %.8f\n",
		userID, balanceAtoms, costUSD, costDCR, costAtoms, float64(balanceAtoms)/1e11)
}

// FormatDebugAfterDeduction formats debug information after balance deduction
func FormatDebugAfterDeduction(newBalanceAtoms int64) string {
	return fmt.Sprintf("DEBUG - After deduction:\n"+
		"  New balance (atoms): %d\n"+
		"  New balance in DCR: %.8f\n",
		newBalanceAtoms, float64(newBalanceAtoms)/1e11)
}

// FormatDebugCommandInfo formats debug information for a command
func FormatDebugCommandInfo(commandName string, userID string, balanceAtoms int64, costUSD float64, costDCR float64, costAtoms int64) string {
	return fmt.Sprintf("DEBUG - %s command:\n"+
		"  User ID: %s\n"+
		"  Current balance (atoms): %d\n"+
		"  Cost in USD: $%.2f\n"+
		"  Cost in DCR: %.8f\n"+
		"  Cost in atoms: %d\n"+
		"  Balance in DCR: %.8f\n",
		commandName, userID, balanceAtoms, costUSD, costDCR, costAtoms, float64(balanceAtoms)/1e11)
}
