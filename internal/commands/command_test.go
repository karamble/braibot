package commands

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/types"
)

// MockBot implements BotInterface for testing
type MockBot struct {
	lastPM        string
	lastGC        string
	lastGCChannel string
	lastError     error
}

func (m *MockBot) SendPM(ctx context.Context, uid zkidentity.ShortID, msg string) error {
	m.lastPM = msg
	return m.lastError
}

func (m *MockBot) SendGC(ctx context.Context, gc string, msg string) error {
	m.lastGC = msg
	return m.lastError
}

func (m *MockBot) SendGCMessage(ctx context.Context, gc string, channel string, msg string) error {
	m.lastGC = msg
	m.lastGCChannel = channel
	return m.lastError
}

// MockDBManager implements DBManagerInterface for testing
type MockDBManager struct {
	balance int64
	err     error
}

func (m *MockDBManager) GetBalance(userID string) (int64, error) {
	return m.balance, m.err
}

func (m *MockDBManager) UpdateBalance(userID string, amount int64) error {
	return m.err
}

func (m *MockDBManager) Close() error {
	return nil
}

// Custom error type for testing
type testError string

func (e testError) Error() string {
	return string(e)
}

func TestCommandHandlers(t *testing.T) {
	// Create test dependencies
	mockDB := &MockDBManager{
		balance: 100000000, // 1 DCR in atoms
	}
	registry := NewRegistry()

	// Create a test ShortID
	var testID zkidentity.ShortID
	copy(testID[:], []byte{1, 2, 3, 4})

	// Test cases for each command
	testCases := []struct {
		name          string
		command       types.Command
		args          []string
		ctx           types.MessageContext
		expectedPM    string
		expectedGC    string
		expectedError bool
		errorMessage  testError
		setDBError    error
		setBotError   error
	}{
		{
			name:    "Help Command - No Args",
			command: HelpCommand(registry, mockDB),
			args:    []string{},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!help",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "🤖 **Welcome to BraiBot Help!**",
			expectedGC: "🤖 **Welcome to BraiBot Help!**",
		},
		{
			name:    "Help Command - With Command",
			command: HelpCommand(registry, mockDB),
			args:    []string{"text2image"},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!help text2image",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Command: !text2image",
			expectedGC: "Command: !text2image",
		},
		{
			name:    "Balance Command - Success",
			command: BalanceCommand(),
			args:    []string{},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!balance",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "💰 Your Balance:",
			expectedGC: "💰 Your Balance:",
		},
		{
			name:    "Balance Command - DB Error",
			command: BalanceCommand(),
			args:    []string{},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!balance",
				IsPM:    true,
				Sender:  testID,
			},
			setDBError:    errors.New("database error"),
			expectedError: true,
			errorMessage:  "failed to get balance",
		},
		{
			name:    "Rate Command - Success",
			command: RateCommand(),
			args:    []string{},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!rate",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Current Exchange Rates",
			expectedGC: "Current Exchange Rates",
		},
		{
			name:    "List Models Command - Success",
			command: ListModelsCommand(),
			args:    []string{"text2image"},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!listmodels text2image",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Available models for text2image:",
			expectedGC: "Available models for text2image:",
		},
		{
			name:    "List Models Command - Invalid Command",
			command: ListModelsCommand(),
			args:    []string{"invalid"},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!listmodels invalid",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Invalid command",
			expectedGC: "Invalid command",
		},
		{
			name:    "Set Model Command - Success",
			command: SetModelCommand(registry),
			args:    []string{"text2image", "stable-diffusion-xl"},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!setmodel text2image stable-diffusion-xl",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Model for text2image set to: stable-diffusion-xl",
			expectedGC: "Model for text2image set to: stable-diffusion-xl",
		},
		{
			name:    "Set Model Command - Invalid Model",
			command: SetModelCommand(registry),
			args:    []string{"text2image", "invalid-model"},
			ctx: types.MessageContext{
				Nick:    "testuser",
				Uid:     []byte{1, 2, 3, 4},
				Message: "!setmodel text2image invalid-model",
				IsPM:    true,
				Sender:  testID,
			},
			expectedPM: "Invalid model name",
			expectedGC: "Invalid model name",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test conditions
			mockDB.err = tc.setDBError
			mockBot := &MockBot{lastError: tc.setBotError}

			// Create message sender
			sender := types.NewMessageSender(mockBot)

			// Execute command
			err := tc.command.Handler.Handle(context.Background(), tc.ctx, tc.args, sender, mockDB)

			// Check error conditions
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), string(tc.errorMessage)) {
					t.Errorf("Expected error containing '%s', got '%v'", tc.errorMessage, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check PM message
			if tc.ctx.IsPM {
				if !strings.Contains(mockBot.lastPM, tc.expectedPM) {
					t.Errorf("Expected PM message containing '%s', got '%s'", tc.expectedPM, mockBot.lastPM)
				}
			}

			// Check GC message
			if !tc.ctx.IsPM {
				if !strings.Contains(mockBot.lastGC, tc.expectedGC) {
					t.Errorf("Expected GC message containing '%s', got '%s'", tc.expectedGC, mockBot.lastGC)
				}
			}
		})
	}
}
