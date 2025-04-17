package utils

import (
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
)

// GetUserIDString converts a user's UID bytes to a string ID
func GetUserIDString(uid []byte) string {
	var userID zkidentity.ShortID
	userID.FromBytes(uid)
	return userID.String()
}

// FormatInsufficientBalanceMessage formats a message for insufficient balance
func FormatInsufficientBalanceMessage(requiredDCR float64, currentDCR float64) string {
	return fmt.Sprintf("Insufficient balance. Required: %.8f DCR, Current: %.8f DCR", requiredDCR, currentDCR)
}

// FormatInsufficientBalanceMessageWithUSD formats a message for insufficient balance with USD value
func FormatInsufficientBalanceMessageWithUSD(requiredDCR float64, currentDCR float64, usdAmount float64) string {
	return fmt.Sprintf("Insufficient balance. You have %.8f DCR, but this operation requires %.8f DCR (%.2f USD). Please send a tip to use this feature.",
		currentDCR, requiredDCR, usdAmount)
}
