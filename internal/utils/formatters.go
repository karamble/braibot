package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/companyzero/bisonrelay/zkidentity"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// ServiceResult defines an interface for common fields in service results.
// Assumes result structs have `Success bool` and `Error error` fields or methods.
type ServiceResult interface {
	IsSuccess() bool
	GetError() error
}

// HandleServiceResultOrError encapsulates common error handling for service calls.
// It checks for direct errors (like context cancellation, insufficient balance)
// and then checks the success status within the result.
// `result` is expected to be a pointer to a struct with `Success bool` and `Error error` fields.
// Returns nil if the error was handled (PM sent/logged appropriately), otherwise returns the error to propagate.
func HandleServiceResultOrError(ctx context.Context, bot *kit.Bot, msgCtx braibottypes.MessageContext, commandName string, result interface{}, err error) error {
	// Create message sender
	sender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

	// 1. Check direct error from the service call
	if err != nil {
		var insufficientBalanceErr *ErrInsufficientBalance // Use utils.ErrInsufficientBalance
		switch {
		case errors.As(err, &insufficientBalanceErr):
			pmMsg := fmt.Sprintf("%s generation failed: %s", commandName, insufficientBalanceErr.Error())
			_ = sender.SendMessage(ctx, msgCtx, pmMsg)
			return nil // Error handled (user notified)
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			fmt.Printf("INFO [%s] User %s: Context canceled/deadline exceeded: %v\n", commandName, msgCtx.Nick, err)
			return nil // Error handled (clean termination)
		default:
			fmt.Printf("ERROR [%s direct] User %s: %v\n", commandName, msgCtx.Nick, err)
			return err // Propagate error
		}
	}

	// 2. Check if the operation failed internally within the service
	// We need to use reflection to check common fields `Success` and `Error`
	if result != nil {
		val := reflect.ValueOf(result)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			successField := val.FieldByName("Success")
			errorField := val.FieldByName("Error")

			if successField.IsValid() && successField.Kind() == reflect.Bool && !successField.Bool() {
				var internalErr error
				if errorField.IsValid() && !errorField.IsNil() {
					if e, ok := errorField.Interface().(error); ok {
						internalErr = e
					}
				}

				errMsg := fmt.Sprintf("ERROR [%s internal] User %s: %s generation failed internally", commandName, msgCtx.Nick, commandName)
				if internalErr != nil {
					errMsg += fmt.Sprintf(": %v", internalErr)
					fmt.Println(errMsg)
					return fmt.Errorf("%s generation failed: %w", commandName, internalErr)
				} else {
					fmt.Println(errMsg)
					return fmt.Errorf("%s generation failed internally", commandName)
				}
			}
		}
	}

	// Success
	return nil
}

// FormatCommandHelpHeader generates the standard header for command help messages.
func FormatCommandHelpHeader(commandName string, model fal.Model, userID zkidentity.ShortID, dbManager braibottypes.DBManagerInterface) string {
	// Get user's balance
	userIDStr := userID.String()
	balance, err := dbManager.GetBalance(userIDStr)
	if err != nil {
		fmt.Printf("ERROR [FormatCommandHelpHeader] Failed to get balance for %s: %v\n", userIDStr, err)
		balance = 0
	}
	balanceDCR := float64(balance) / 1e11

	// Get current exchange rate for USD value
	dcrPrice, _, err := GetDCRPrice()
	if err != nil {
		fmt.Printf("ERROR [FormatCommandHelpHeader] Failed to convert USD to DCR: %v\n", err)
		dcrPrice = 0
	}
	usdValue := balanceDCR * dcrPrice

	// Format header
	header := fmt.Sprintf("🤖 **%s Model Help**\n\n", strings.Title(commandName))
	header += fmt.Sprintf("💰 **Your Balance:** %.8f DCR ($%.2f USD)\n\n", balanceDCR, usdValue)
	header += fmt.Sprintf("🎯 **Model:** %s\n", model.Name)
	header += fmt.Sprintf("💵 **Price:** $%.2f USD\n\n", model.PriceUSD)

	return header
}

// FormatBalanceMessage formats a balance message with DCR and USD values
func FormatBalanceMessage(balanceDCR float64, dcrPrice float64) string {
	usdValue := balanceDCR * dcrPrice
	return fmt.Sprintf("💰 Your Balance:\n• %.8f DCR\n• $%.2f USD",
		balanceDCR, usdValue)
}

// FormatBillingMessage formats a billing message with charged amount and remaining balance
func FormatBillingMessage(chargedDCR float64, chargedUSD float64, remainingBalance float64) string {
	return fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
		chargedDCR, chargedUSD, remainingBalance)
}

// FormatThousands formats a float64 with commas as thousands separators, rounded to the nearest integer.
func FormatThousands(n float64) string {
	// Format with 8 decimal places first
	str := fmt.Sprintf("%.8f", n)

	// Split into integer and decimal parts
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return str
	}

	intPart := parts[0]
	decPart := parts[1]

	// Handle negative numbers
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}

	// Add thousands separators
	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}

	// Recombine parts
	if negative {
		intPart = "-" + intPart
	}

	return intPart + "." + decPart
}

// FormatUSDThousands formats a float64 as USD with thousands separators and 2 decimal places.
func FormatUSDThousands(n float64) string {
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]
	negative := false
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = intPart[1:]
	}
	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}
	if negative {
		intPart = "-" + intPart
	}
	return intPart + "." + decPart
}

// IsAudioNote checks if a message contains an audio note embed
func IsAudioNote(message string) bool {
	return strings.Contains(message, "--embed[alt=Audio note,type=audio/ogg")
}

// ExtractAudioNoteData extracts the base64 audio data from an audio note message
func ExtractAudioNoteData(message string) (string, error) {
	// Find the data field in the embed
	dataStart := strings.Index(message, "data=")
	if dataStart == -1 {
		return "", fmt.Errorf("no data field found in audio note")
	}
	dataStart += 5 // Skip "data="

	// Find the end of the data (last ]--)
	dataEnd := strings.LastIndex(message, "]--")
	if dataEnd == -1 {
		return "", fmt.Errorf("invalid audio note format")
	}

	// Extract the base64 data
	data := message[dataStart:dataEnd]
	return data, nil
}
