package channel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChannel(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "",
	}

	err := router.Create(channel)
	assert.Nil(t, err)

	// Verify channel exists
	retrieved, err := router.Get("ch-123")
	assert.Nil(t, err)
	assert.Equal(t, "ch-123", retrieved.ID)
	assert.Equal(t, "input", retrieved.Name)
	assert.Equal(t, "user-789", retrieved.SubmitterID)
}

func TestCreateDuplicateChannel(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
	}

	err := router.Create(channel)
	assert.Nil(t, err)

	// Try to create duplicate
	err = router.Create(channel)
	assert.Equal(t, ErrChannelExists, err)
}

func TestGetByProcessAndName(t *testing.T) {
	router := NewRouter()

	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}

	router.Create(channel1)
	router.Create(channel2)

	// Get by name
	retrieved, err := router.GetByProcessAndName("proc-1", "output")
	assert.Nil(t, err)
	assert.Equal(t, "ch-2", retrieved.ID)

	// Not found
	_, err = router.GetByProcessAndName("proc-1", "nonexistent")
	assert.Equal(t, ErrChannelNotFound, err)
}

func TestAppendAndReadAfter(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Submitter appends
	err := router.Append("ch-123", "user-789", 1, 0, []byte("message 1"))
	assert.Nil(t, err)

	err = router.Append("ch-123", "user-789", 2, 0, []byte("message 2"))
	assert.Nil(t, err)

	// Executor appends
	err = router.Append("ch-123", "exec-abc", 1, 0, []byte("message 3"))
	assert.Nil(t, err)

	// Read all
	entries, err := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 3)

	// Read after index 1 (skip first entry)
	entries, err = router.ReadAfter("ch-123", "user-789", 1, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, int64(2), entries[0].Sequence) // Second message from user
	assert.Equal(t, int64(1), entries[1].Sequence) // First message from executor

	// Read with limit
	entries, err = router.ReadAfter("ch-123", "user-789", 0, 2)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
}

func TestAuthorization(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Submitter can append
	err := router.Append("ch-123", "user-789", 1, 0, []byte("allowed"))
	assert.Nil(t, err)

	// Executor can append
	err = router.Append("ch-123", "exec-abc", 1, 0, []byte("allowed"))
	assert.Nil(t, err)

	// Unauthorized user cannot append
	err = router.Append("ch-123", "other-user", 1, 0, []byte("denied"))
	assert.Equal(t, ErrUnauthorized, err)

	// Unauthorized user cannot read
	_, err = router.ReadAfter("ch-123", "other-user", 0, 0)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestSetExecutorID(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
		ExecutorID:  "", // Not assigned yet
	}
	router.Create(channel)

	// Initially executor cannot access
	err := router.Append("ch-123", "exec-abc", 1, 0, []byte("denied"))
	assert.Equal(t, ErrUnauthorized, err)

	// Set executor ID
	err = router.SetExecutorID("ch-123", "exec-abc")
	assert.Nil(t, err)

	// Now executor can access
	err = router.Append("ch-123", "exec-abc", 1, 0, []byte("allowed"))
	assert.Nil(t, err)
}

func TestSetExecutorIDForProcess(t *testing.T) {
	router := NewRouter()

	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}
	router.Create(channel1)
	router.Create(channel2)

	// Set executor for all channels
	err := router.SetExecutorIDForProcess("proc-1", "exec-1")
	assert.Nil(t, err)

	// Both channels should have executor
	retrieved1, _ := router.Get("ch-1")
	retrieved2, _ := router.Get("ch-2")
	assert.Equal(t, "exec-1", retrieved1.ExecutorID)
	assert.Equal(t, "exec-1", retrieved2.ExecutorID)
}

func TestCleanupProcess(t *testing.T) {
	router := NewRouter()

	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}
	channel3 := &Channel{
		ID:          "ch-3",
		ProcessID:   "proc-2",
		Name:        "input",
		SubmitterID: "user-2",
	}
	router.Create(channel1)
	router.Create(channel2)
	router.Create(channel3)

	// Cleanup proc-1
	router.CleanupProcess("proc-1")

	// ch-1 and ch-2 should be gone
	_, err := router.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)
	_, err = router.Get("ch-2")
	assert.Equal(t, ErrChannelNotFound, err)

	// ch-3 should still exist
	_, err = router.Get("ch-3")
	assert.Nil(t, err)
}

func TestCleanupProcessWithMessages(t *testing.T) {
	router := NewRouter()

	// Create channels with messages
	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-1",
		ExecutorID:  "exec-1",
	}
	router.Create(channel)

	// Add messages to channel
	router.Append("ch-123", "user-1", 1, 0, []byte("message 1"))
	router.Append("ch-123", "user-1", 2, 0, []byte("message 2"))
	router.Append("ch-123", "exec-1", 1, 1, []byte("response 1"))

	// Verify messages exist
	entries, err := router.ReadAfter("ch-123", "user-1", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 3)

	// Cleanup process
	router.CleanupProcess("proc-456")

	// Channel should be gone
	_, err = router.Get("ch-123")
	assert.Equal(t, ErrChannelNotFound, err)

	// Cannot read messages anymore
	_, err = router.ReadAfter("ch-123", "user-1", 0, 0)
	assert.Equal(t, ErrChannelNotFound, err)

	// Cannot append messages anymore
	err = router.Append("ch-123", "user-1", 3, 0, []byte("message 3"))
	assert.Equal(t, ErrChannelNotFound, err)
}

func TestCleanupProcessMultipleChannels(t *testing.T) {
	router := NewRouter()

	// Create multiple channels for same process
	processID := "proc-test"
	channels := []*Channel{
		{ID: "ch-1", ProcessID: processID, Name: "chat", SubmitterID: "user-1"},
		{ID: "ch-2", ProcessID: processID, Name: "logs", SubmitterID: "user-1"},
		{ID: "ch-3", ProcessID: processID, Name: "metrics", SubmitterID: "user-1"},
	}

	for _, ch := range channels {
		router.Create(ch)
	}

	// Verify all channels exist
	for _, ch := range channels {
		retrieved, err := router.Get(ch.ID)
		assert.Nil(t, err)
		assert.Equal(t, ch.ID, retrieved.ID)
	}

	// Get channels by process
	processChannels := router.GetChannelsByProcess(processID)
	assert.Len(t, processChannels, 3)

	// Cleanup process
	router.CleanupProcess(processID)

	// All channels should be gone
	for _, ch := range channels {
		_, err := router.Get(ch.ID)
		assert.Equal(t, ErrChannelNotFound, err)
	}

	// Process index should be empty
	processChannels = router.GetChannelsByProcess(processID)
	assert.Len(t, processChannels, 0)
}

func TestCleanupProcessDoesNotAffectOtherProcesses(t *testing.T) {
	router := NewRouter()

	// Create channels for multiple processes
	channel1 := &Channel{ID: "ch-1", ProcessID: "proc-1", Name: "chat", SubmitterID: "user-1"}
	channel2 := &Channel{ID: "ch-2", ProcessID: "proc-1", Name: "logs", SubmitterID: "user-1"}
	channel3 := &Channel{ID: "ch-3", ProcessID: "proc-2", Name: "chat", SubmitterID: "user-2"}
	channel4 := &Channel{ID: "ch-4", ProcessID: "proc-3", Name: "chat", SubmitterID: "user-3"}

	router.Create(channel1)
	router.Create(channel2)
	router.Create(channel3)
	router.Create(channel4)

	// Add messages to all channels
	router.Append("ch-1", "user-1", 1, 0, []byte("proc-1 msg"))
	router.Append("ch-2", "user-1", 1, 0, []byte("proc-1 logs"))
	router.Append("ch-3", "user-2", 1, 0, []byte("proc-2 msg"))
	router.Append("ch-4", "user-3", 1, 0, []byte("proc-3 msg"))

	// Cleanup proc-1
	router.CleanupProcess("proc-1")

	// proc-1 channels should be gone
	_, err := router.Get("ch-1")
	assert.Equal(t, ErrChannelNotFound, err)
	_, err = router.Get("ch-2")
	assert.Equal(t, ErrChannelNotFound, err)

	// proc-2 and proc-3 channels should still exist with their messages
	retrieved, err := router.Get("ch-3")
	assert.Nil(t, err)
	assert.Equal(t, "ch-3", retrieved.ID)
	entries, err := router.ReadAfter("ch-3", "user-2", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte("proc-2 msg"), entries[0].Payload)

	retrieved, err = router.Get("ch-4")
	assert.Nil(t, err)
	assert.Equal(t, "ch-4", retrieved.ID)
	entries, err = router.ReadAfter("ch-4", "user-3", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []byte("proc-3 msg"), entries[0].Payload)
}

func TestCleanupProcessIdempotent(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-1",
	}
	router.Create(channel)

	// Cleanup once
	router.CleanupProcess("proc-456")

	// Verify channel is gone
	_, err := router.Get("ch-123")
	assert.Equal(t, ErrChannelNotFound, err)

	// Cleanup again - should not panic or error
	router.CleanupProcess("proc-456")

	// Cleanup non-existent process - should not panic
	router.CleanupProcess("non-existent-process")
}

func TestCleanupProcessMemoryReclamation(t *testing.T) {
	router := NewRouter()

	// Create channel with large message log
	channel := &Channel{
		ID:          "ch-large",
		ProcessID:   "proc-large",
		Name:        "chat",
		SubmitterID: "user-1",
		ExecutorID:  "exec-1",
	}
	router.Create(channel)

	// Add 1000 messages
	for i := 1; i <= 1000; i++ {
		payload := make([]byte, 1024) // 1KB per message
		router.Append("ch-large", "user-1", int64(i), 0, payload)
	}

	// Verify log size
	size, err := router.GetLogSize("ch-large")
	assert.Nil(t, err)
	assert.Equal(t, 1000, size)

	// Cleanup process
	router.CleanupProcess("proc-large")

	// Channel should be gone
	_, err = router.Get("ch-large")
	assert.Equal(t, ErrChannelNotFound, err)

	// Verify we can't get log size anymore
	_, err = router.GetLogSize("ch-large")
	assert.Equal(t, ErrChannelNotFound, err)

	// Note: In Go, memory is reclaimed by GC when there are no more references
	// After cleanup, both the Channel object and its Log array should be eligible for GC
}

func TestGetChannelsByProcess(t *testing.T) {
	router := NewRouter()

	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "input",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-1",
		Name:        "output",
		SubmitterID: "user-1",
	}
	router.Create(channel1)
	router.Create(channel2)

	channels := router.GetChannelsByProcess("proc-1")
	assert.Len(t, channels, 2)
}

func TestReplicateEntry(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Replicate entries out of order
	entry3 := &MsgEntry{Sequence: 3, Payload: []byte("three")}
	entry1 := &MsgEntry{Sequence: 1, Payload: []byte("one")}
	entry2 := &MsgEntry{Sequence: 2, Payload: []byte("two")}

	router.ReplicateEntry("ch-123", entry3)
	router.ReplicateEntry("ch-123", entry1)
	router.ReplicateEntry("ch-123", entry2)

	// Should be sorted
	entries, _ := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Len(t, entries, 3)
	assert.Equal(t, int64(1), entries[0].Sequence)
	assert.Equal(t, int64(2), entries[1].Sequence)
	assert.Equal(t, int64(3), entries[2].Sequence)

	// Replicate duplicate - should be idempotent
	err := router.ReplicateEntry("ch-123", entry2)
	assert.Nil(t, err)
	entries, _ = router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Len(t, entries, 3)
}

func TestStats(t *testing.T) {
	router := NewRouter()

	channel1 := &Channel{ID: "ch-1", ProcessID: "proc-1", Name: "input", SubmitterID: "user-1"}
	channel2 := &Channel{ID: "ch-2", ProcessID: "proc-1", Name: "output", SubmitterID: "user-1"}
	channel3 := &Channel{ID: "ch-3", ProcessID: "proc-2", Name: "input", SubmitterID: "user-2"}
	router.Create(channel1)
	router.Create(channel2)
	router.Create(channel3)

	channelCount, processCount := router.Stats()
	assert.Equal(t, 3, channelCount)
	assert.Equal(t, 2, processCount)
}

func TestGetLogSize(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "input",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Initially empty
	size, _ := router.GetLogSize("ch-123")
	assert.Equal(t, 0, size)

	// After appending
	router.Append("ch-123", "user-789", 1, 0, []byte("msg1"))
	router.Append("ch-123", "user-789", 2, 0, []byte("msg2"))

	size, _ = router.GetLogSize("ch-123")
	assert.Equal(t, 2, size)
}

func TestChannelNotFound(t *testing.T) {
	router := NewRouter()

	_, err := router.Get("nonexistent")
	assert.Equal(t, ErrChannelNotFound, err)

	err = router.Append("nonexistent", "user", 1, 0, []byte("data"))
	assert.Equal(t, ErrChannelNotFound, err)

	_, err = router.ReadAfter("nonexistent", "user", 0, 0)
	assert.Equal(t, ErrChannelNotFound, err)
}

func TestSubscribe(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Submitter can subscribe
	ch, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)
	assert.NotNil(t, ch)

	// Verify subscriber count
	assert.Equal(t, 1, router.SubscriberCount("ch-123"))

	// Unsubscribe
	router.Unsubscribe("ch-123", ch)
	assert.Equal(t, 0, router.SubscriberCount("ch-123"))
}

func TestSubscribeUnauthorized(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Unauthorized user cannot subscribe
	_, err := router.Subscribe("ch-123", "other-user")
	assert.Equal(t, ErrUnauthorized, err)
}

func TestSubscribeNotFound(t *testing.T) {
	router := NewRouter()

	_, err := router.Subscribe("nonexistent", "user")
	assert.Equal(t, ErrChannelNotFound, err)
}

func TestSubscribePushNotification(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Subscribe
	subChan, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)

	// Append message - should trigger push notification
	err = router.Append("ch-123", "exec-abc", 1, 0, []byte("hello"))
	assert.Nil(t, err)

	// Should receive the entry on the subscription channel
	select {
	case entry := <-subChan:
		assert.Equal(t, int64(1), entry.Sequence)
		assert.Equal(t, []byte("hello"), entry.Payload)
	default:
		t.Fatal("Expected to receive entry on subscription channel")
	}

	router.Unsubscribe("ch-123", subChan)
}

func TestSubscribeMultipleSubscribers(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-abc",
	}
	router.Create(channel)

	// Multiple subscribers
	sub1, _ := router.Subscribe("ch-123", "user-789")
	sub2, _ := router.Subscribe("ch-123", "exec-abc")

	assert.Equal(t, 2, router.SubscriberCount("ch-123"))

	// Append message - both should receive
	router.Append("ch-123", "user-789", 1, 0, []byte("broadcast"))

	// Both subscribers should receive
	select {
	case entry := <-sub1:
		assert.Equal(t, []byte("broadcast"), entry.Payload)
	default:
		t.Fatal("Subscriber 1 did not receive entry")
	}

	select {
	case entry := <-sub2:
		assert.Equal(t, []byte("broadcast"), entry.Payload)
	default:
		t.Fatal("Subscriber 2 did not receive entry")
	}

	// Unsubscribe one
	router.Unsubscribe("ch-123", sub1)
	assert.Equal(t, 1, router.SubscriberCount("ch-123"))

	// Remaining subscriber should still receive
	router.Append("ch-123", "user-789", 2, 0, []byte("second"))
	select {
	case entry := <-sub2:
		assert.Equal(t, []byte("second"), entry.Payload)
	default:
		t.Fatal("Subscriber 2 did not receive second entry")
	}

	router.Unsubscribe("ch-123", sub2)
}

func TestSubscribeReplicatedEntry(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Subscribe
	subChan, _ := router.Subscribe("ch-123", "user-789")

	// Replicate entry - should also trigger notification
	entry := &MsgEntry{
		Sequence:  1,
		SenderID:  "remote-exec",
		Payload:   []byte("replicated"),
	}
	router.ReplicateEntry("ch-123", entry)

	// Should receive the replicated entry
	select {
	case received := <-subChan:
		assert.Equal(t, []byte("replicated"), received.Payload)
	default:
		t.Fatal("Expected to receive replicated entry")
	}

	router.Unsubscribe("ch-123", subChan)
}

func TestSubscribeCleanup(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Subscribe
	subChan, _ := router.Subscribe("ch-123", "user-789")
	assert.Equal(t, 1, router.SubscriberCount("ch-123"))

	// Cleanup process - should remove channel and subscribers
	router.CleanupProcess("proc-456")

	// Subscriber count should be 0 (channel gone)
	assert.Equal(t, 0, router.SubscriberCount("ch-123"))

	// Unsubscribe should not panic on cleaned up channel
	router.Unsubscribe("ch-123", subChan)
}
