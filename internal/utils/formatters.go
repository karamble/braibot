package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
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
func HandleServiceResultOrError(ctx context.Context, bot *kit.Bot, pm types.ReceivedPM, commandName string, result interface{}, err error) error {
	// 1. Check direct error from the service call
	if err != nil {
		var insufficientBalanceErr *ErrInsufficientBalance // Use utils.ErrInsufficientBalance
		switch {
		case errors.As(err, &insufficientBalanceErr):
			pmMsg := fmt.Sprintf("%s generation failed: %s", commandName, insufficientBalanceErr.Error())
			_ = bot.SendPM(ctx, pm.Nick, pmMsg)
			return nil // Error handled (user notified)
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			fmt.Printf("INFO [%s] User %s: Context canceled/deadline exceeded: %v\n", commandName, pm.Nick, err)
			return nil // Error handled (clean termination)
		default:
			fmt.Printf("ERROR [%s direct] User %s: %v\n", commandName, pm.Nick, err)
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

				errMsg := fmt.Sprintf("ERROR [%s internal] User %s: %s generation failed internally", commandName, pm.Nick, commandName)
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
func FormatCommandHelpHeader(commandName string, model fal.Model, userID zkidentity.ShortID, dbManager *database.DBManager) string {
	// Get user balance
	userIDStr := userID.String()
	balanceAtoms, err := dbManager.GetBalance(userIDStr)
	balanceDCR := float64(balanceAtoms) / 1e11
	balanceStr := ""
	if err != nil {
		fmt.Printf("ERROR [FormatCommandHelpHeader] Failed to get balance for %s: %v\n", userIDStr, err)
		balanceStr = "(unavailable)"
	} else {
		balanceStr = fmt.Sprintf("%.8f DCR", balanceDCR)
	}

	// Convert model price to DCR
	modelDcrPrice, err := USDToDCR(model.PriceUSD)
	priceStr := ""
	if err != nil {
		fmt.Printf("ERROR [FormatCommandHelpHeader] Failed to convert USD to DCR: %v\n", err)
		priceStr = fmt.Sprintf("$%.2f USD (DCR price unavailable)", model.PriceUSD)
	} else {
		priceStr = fmt.Sprintf("$%.2f USD (%.8f DCR)", model.PriceUSD, modelDcrPrice)
	}

	// Determine command emoji
	commandEmoji := "✨" // Default
	switch commandName {
	case "text2image", "image2image":
		commandEmoji = "🎨"
	case "text2speech":
		commandEmoji = "🗣️"
	case "image2video", "text2video":
		commandEmoji = "🎬"
	}

	// Format the header
	header := fmt.Sprintf("%s Using Model: **%s**\n", commandEmoji, model.Name)
	header += fmt.Sprintf("📄 Description: %s\n", model.Description)
	header += fmt.Sprintf("💲 Price: %s\n", priceStr)
	header += fmt.Sprintf("💰 Your Balance: %s\n\n", balanceStr)

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
