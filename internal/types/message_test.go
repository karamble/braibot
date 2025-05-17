package braibottypes

import (
	"context"
	"fmt"
	"testing"

	"github.com/companyzero/bisonrelay/zkidentity"
)

// MockBot is a mock implementation of BotInterface for testing
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

// TestMessageContext tests the MessageContext struct
func TestMessageContext(t *testing.T) {
	// Create a PM context
	pmCtx := MessageContext{
		Nick:    "user1",
		Uid:     []byte("123"),
		Message: "Hello",
		IsPM:    true,
		Sender:  zkidentity.ShortID{},
	}

	// Create a GC context
	gcCtx := MessageContext{
		Nick:    "user2",
		Uid:     []byte("456"),
		Message: "Hi",
		IsPM:    false,
		GC:      "general",
	}

	// Test PM context
	if pmCtx.Nick != "user1" {
		t.Errorf("Expected PM context Nick to be 'user1', got '%s'", pmCtx.Nick)
	}
	if pmCtx.IsPM != true {
		t.Errorf("Expected PM context IsPM to be true, got false")
	}

	// Test GC context
	if gcCtx.Nick != "user2" {
		t.Errorf("Expected GC context Nick to be 'user2', got '%s'", gcCtx.Nick)
	}
	if gcCtx.IsPM != false {
		t.Errorf("Expected GC context IsPM to be false, got true")
	}
	if gcCtx.GC != "general" {
		t.Errorf("Expected GC context GC to be 'general', got '%s'", gcCtx.GC)
	}
}

// TestMessageSender tests the MessageSender struct
func TestMessageSender(t *testing.T) {
	mockBot := &MockBot{}
	sender := NewMessageSender(mockBot)

	// Test sending a PM
	pmCtx := MessageContext{
		Nick:    "user1",
		Uid:     []byte("123"),
		Message: "Hello",
		IsPM:    true,
		Sender:  zkidentity.ShortID{},
	}
	err := sender.SendMessage(context.Background(), pmCtx, "Hi")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test sending a GC message
	gcCtx := MessageContext{
		Nick:    "user2",
		Uid:     []byte("456"),
		Message: "Hi",
		IsPM:    false,
		GC:      "general",
	}
	err = sender.SendMessage(context.Background(), gcCtx, "Hello")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test error handling
	err = sender.SendErrorMessage(context.Background(), pmCtx, fmt.Errorf("test error"))
	if err != nil {
		t.Errorf("SendErrorMessage failed: %v", err)
	}
	if mockBot.lastPM != "❌ Error: test error" {
		t.Errorf("Expected lastPM to be '❌ Error: test error', got '%s'", mockBot.lastPM)
	}

	// Test success message
	err = sender.SendSuccessMessage(context.Background(), pmCtx, "Operation completed")
	if err != nil {
		t.Errorf("SendSuccessMessage failed: %v", err)
	}
	if mockBot.lastPM != "✅ Operation completed" {
		t.Errorf("Expected lastPM to be '✅ Operation completed', got '%s'", mockBot.lastPM)
	}

	// Test error propagation
	mockBot.lastError = fmt.Errorf("mock error")
	err = sender.SendMessage(context.Background(), pmCtx, "Should fail")
	if err != mockBot.lastError {
		t.Errorf("Expected error to be propagated, got %v", err)
	}
}
