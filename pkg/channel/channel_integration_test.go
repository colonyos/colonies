package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestChannelEndToEnd tests the complete channel workflow:
// 1. Process submission creates channels
// 2. Client sends messages
// 3. Executor reads and replies with InReplyTo
// 4. Client reads responses
func TestChannelEndToEnd(t *testing.T) {
	router := NewRouter()

	// Simulate process submission with channel "chat"
	processID := "proc-123"
	submitterID := "user-456"
	executorID := "exec-789"

	// Create channel (normally done by controller when process is submitted)
	channel := &Channel{
		ID:          "ch-abc",
		ProcessID:   processID,
		Name:        "chat",
		SubmitterID: submitterID,
		ExecutorID:  "", // Not assigned yet
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Simulate executor assignment (normally done by controller)
	err = router.SetExecutorID("ch-abc", executorID)
	assert.Nil(t, err)

	// Client sends first message with sequence 1
	err = router.Append("ch-abc", submitterID, 1, 0, []byte("What is 2+2?"))
	assert.Nil(t, err)

	// Executor reads messages from index 0
	entries, err := router.ReadAfter("ch-abc", executorID, 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, submitterID, entries[0].SenderID)
	assert.Equal(t, []byte("What is 2+2?"), entries[0].Payload)
	assert.Equal(t, int64(0), entries[0].InReplyTo) // No reply reference

	// Executor replies with sequence 1, referencing client's sequence 1
	err = router.Append("ch-abc", executorID, 1, 1, []byte("4"))
	assert.Nil(t, err)

	// Client reads new messages from index 1 (after their own message)
	entries, err = router.ReadAfter("ch-abc", submitterID, 1, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, executorID, entries[0].SenderID)
	assert.Equal(t, []byte("4"), entries[0].Payload)
	assert.Equal(t, int64(1), entries[0].InReplyTo) // References client's seq 1

	// Client sends second message with sequence 2
	err = router.Append("ch-abc", submitterID, 2, 0, []byte("What is 3+3?"))
	assert.Nil(t, err)

	// Executor reads from index 2 (after previous response)
	entries, err = router.ReadAfter("ch-abc", executorID, 2, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, int64(2), entries[0].Sequence)
	assert.Equal(t, submitterID, entries[0].SenderID)
	assert.Equal(t, []byte("What is 3+3?"), entries[0].Payload)

	// Executor replies with sequence 2, referencing client's sequence 2
	err = router.Append("ch-abc", executorID, 2, 2, []byte("6"))
	assert.Nil(t, err)

	// Client reads all messages from start
	entries, err = router.ReadAfter("ch-abc", submitterID, 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 4)

	// Verify ordering and correlation
	// Entry 0: Client seq 1 (question)
	assert.Equal(t, submitterID, entries[0].SenderID)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, []byte("What is 2+2?"), entries[0].Payload)

	// Entry 1: Executor seq 1 (response to client seq 1)
	assert.Equal(t, executorID, entries[1].SenderID)
	assert.Equal(t, int64(1), entries[1].Sequence)
	assert.Equal(t, int64(1), entries[1].InReplyTo)
	assert.Equal(t, []byte("4"), entries[1].Payload)

	// Entry 2: Client seq 2 (question)
	assert.Equal(t, submitterID, entries[2].SenderID)
	assert.Equal(t, int64(2), entries[2].Sequence)
	assert.Equal(t, []byte("What is 3+3?"), entries[2].Payload)

	// Entry 3: Executor seq 2 (response to client seq 2)
	assert.Equal(t, executorID, entries[3].SenderID)
	assert.Equal(t, int64(2), entries[3].Sequence)
	assert.Equal(t, int64(2), entries[3].InReplyTo)
	assert.Equal(t, []byte("6"), entries[3].Payload)
}

// TestChannelStreamingTokens tests streaming responses like an LLM
func TestChannelStreamingTokens(t *testing.T) {
	router := NewRouter()

	processID := "proc-llm"
	submitterID := "user-123"
	executorID := "ollama-exec"

	// Create channel
	channel := &Channel{
		ID:          "ch-stream",
		ProcessID:   processID,
		Name:        "chat",
		SubmitterID: submitterID,
		ExecutorID:  executorID,
	}
	router.Create(channel)

	// Client sends prompt
	router.Append("ch-stream", submitterID, 1, 0, []byte("Tell me a story"))

	// Executor streams tokens, all referencing client's sequence 1
	router.Append("ch-stream", executorID, 1, 1, []byte("Once"))
	router.Append("ch-stream", executorID, 2, 1, []byte(" upon"))
	router.Append("ch-stream", executorID, 3, 1, []byte(" a"))
	router.Append("ch-stream", executorID, 4, 1, []byte(" time"))

	// Client reads all tokens
	entries, err := router.ReadAfter("ch-stream", submitterID, 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 5) // 1 prompt + 4 tokens

	// Verify all executor responses reference the same client message
	for i := 1; i < 5; i++ {
		assert.Equal(t, executorID, entries[i].SenderID)
		assert.Equal(t, int64(1), entries[i].InReplyTo) // All reference client seq 1
	}

	// Reconstruct the streamed response
	var response string
	for i := 1; i < 5; i++ {
		response += string(entries[i].Payload)
	}
	assert.Equal(t, "Once upon a time", response)
}

// TestChannelMultipleWritersOrdering tests that messages from different senders
// maintain causal ordering
func TestChannelMultipleWritersOrdering(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-multi",
		ProcessID:   "proc-multi",
		Name:        "chat",
		SubmitterID: "user-1",
		ExecutorID:  "exec-1",
	}
	router.Create(channel)

	// Client sends messages
	router.Append("ch-multi", "user-1", 1, 0, []byte("Q1"))
	router.Append("ch-multi", "user-1", 2, 0, []byte("Q2"))

	// Executor sends responses (may arrive out of order from different servers)
	router.Append("ch-multi", "exec-1", 2, 2, []byte("A2"))
	router.Append("ch-multi", "exec-1", 1, 1, []byte("A1"))

	// Read all entries
	entries, err := router.ReadAfter("ch-multi", "user-1", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 4)

	// Verify ordering: client messages sorted by sequence, executor messages sorted by sequence
	// Within same sender: seq 1 < seq 2
	assert.Equal(t, "user-1", entries[0].SenderID)
	assert.Equal(t, int64(1), entries[0].Sequence)

	assert.Equal(t, "user-1", entries[1].SenderID)
	assert.Equal(t, int64(2), entries[1].Sequence)

	assert.Equal(t, "exec-1", entries[2].SenderID)
	assert.Equal(t, int64(1), entries[2].Sequence)

	assert.Equal(t, "exec-1", entries[3].SenderID)
	assert.Equal(t, int64(2), entries[3].Sequence)
}

// TestChannelCleanup verifies channels are removed when process completes
func TestChannelCleanup(t *testing.T) {
	router := NewRouter()

	processID := "proc-cleanup"
	channel1 := &Channel{ID: "ch-1", ProcessID: processID, Name: "chat", SubmitterID: "user-1"}
	channel2 := &Channel{ID: "ch-2", ProcessID: processID, Name: "logs", SubmitterID: "user-1"}

	router.Create(channel1)
	router.Create(channel2)

	// Add some messages
	router.Append("ch-1", "user-1", 1, 0, []byte("message"))

	// Simulate process completion
	router.CleanupProcess(processID)

	// Channels should be gone
	_, err := router.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)

	_, err = router.Get("ch-2")
	assert.Equal(t, ErrChannelNotFound, err)
}
