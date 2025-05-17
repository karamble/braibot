package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
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

// FormatThousands formats a float64 with dots as thousands separators, rounded to the nearest integer.
func FormatThousands(n float64) string {
	// Round to the nearest integer
	rounded := int64(math.Round(n)) // Use int64 for potentially large numbers
	s := strconv.FormatInt(rounded, 10)

	nDigits := len(s)
	if rounded < 0 {
		nDigits-- // Sign doesn't count
	}
	if nDigits <= 3 {
		return s // No separator needed
	}

	// Calculate the number of separators needed
	numSeparators := (nDigits - 1) / 3
	resultLen := len(s) + numSeparators
	result := make([]byte, resultLen)

	sepIdx := resultLen - 1
	srcIdx := len(s) - 1
	digitsSinceSep := 0

	for srcIdx >= 0 {
		if s[srcIdx] == '-' {
			result[0] = '-'
			break
		}

		if digitsSinceSep == 3 {
			result[sepIdx] = '.'
			sepIdx--
			digitsSinceSep = 0
		}

		result[sepIdx] = s[srcIdx]
		sepIdx--
		srcIdx--
		digitsSinceSep++
	}

	return string(result)
}
