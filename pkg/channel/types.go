package channel

import (
	"time"
)

// ChannelSpec defines a channel in a FunctionSpec
type ChannelSpec struct {
	Name string `json:"name"`
}

// MsgEntry represents a single message in a channel
type MsgEntry struct {
	Sequence  int64     `json:"sequence"`
	InReplyTo int64     `json:"inreplyto,omitempty"` // References sequence from other sender
	Timestamp time.Time `json:"timestamp"`
	SenderID  string    `json:"senderid"`
	Payload   []byte    `json:"payload"`
}

// Channel represents an append-only message log
type Channel struct {
	ID          string      `json:"id"`
	ProcessID   string      `json:"processid"`
	Name        string      `json:"name"`
	SubmitterID string      `json:"submitterid"` // Process submitter
	ExecutorID  string      `json:"executorid"`  // Assigned executor
	Sequence    int64       `json:"sequence"`
	Log         []*MsgEntry `json:"log"`
}

// ProcessInfo contains the minimal process information needed for authorization
type ProcessInfo struct {
	ID          string
	SubmitterID string
	ExecutorID  string
}
