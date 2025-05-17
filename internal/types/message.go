package braibottypes

import (
	"github.com/companyzero/bisonrelay/zkidentity"
)

// MessageContext represents a unified message context for both PM and group chat messages
type MessageContext struct {
	// Common fields
	Nick    string // User's nickname
	Uid     []byte // User's ID
	Message string // The message content

	// Context-specific fields
	IsPM   bool               // Whether this is a private message
	Sender zkidentity.ShortID // Sender's ID (for PM)
	GC     string             // Group chat ID (for GC)
}

// ReceivedPM represents a received private message
type ReceivedPM struct {
	Nick string
	Uid  []byte
	Msg  struct {
		Message string
	}
}

// GCReceivedMsg represents a received group chat message
type GCReceivedMsg struct {
	Nick    string
	Uid     []byte
	GcAlias string
	Msg     struct {
		Message string
	}
}
