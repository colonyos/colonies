package channel

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/constants"
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
	router := NewRouterWithoutRateLimit() // Use no rate limit for this test

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

// TestDoubleCloseRace tests the race condition where CleanupProcess and Unsubscribe
// both try to close the same subscriber channel concurrently.
func TestDoubleCloseRace(t *testing.T) {
	// Run multiple iterations to increase chance of hitting the race
	for i := 0; i < 100; i++ {
		router := NewRouter()

		channel := &Channel{
			ID:          "ch-123",
			ProcessID:   "proc-456",
			Name:        "chat",
			SubmitterID: "user-789",
		}
		router.Create(channel)

		// Subscribe to the channel
		subChan, err := router.Subscribe("ch-123", "user-789")
		assert.Nil(t, err)
		assert.NotNil(t, subChan)

		// Use a WaitGroup to synchronize the goroutines
		var wg sync.WaitGroup
		wg.Add(2)

		// Use a channel to synchronize start
		start := make(chan struct{})

		// Goroutine 1: CleanupProcess
		go func() {
			defer wg.Done()
			<-start // Wait for signal
			router.CleanupProcess("proc-456")
		}()

		// Goroutine 2: Unsubscribe
		go func() {
			defer wg.Done()
			<-start // Wait for signal
			router.Unsubscribe("ch-123", subChan)
		}()

		// Start both goroutines at the same time
		close(start)

		// Wait for both to complete - should not panic
		wg.Wait()

		// Verify cleanup happened
		assert.Equal(t, 0, router.SubscriberCount("ch-123"))
		_, err = router.Get("ch-123")
		assert.Equal(t, ErrChannelNotFound, err)
	}
}

// TestDoubleCloseSequential tests the sequential case where CleanupProcess
// runs first and then Unsubscribe is called on the already-closed channel.
func TestDoubleCloseSequential(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Subscribe to get a channel
	subChan, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)
	assert.NotNil(t, subChan)

	// First: CleanupProcess closes the channel
	router.CleanupProcess("proc-456")

	// Verify channel is closed by trying to receive
	_, ok := <-subChan
	assert.False(t, ok, "Channel should be closed")

	// Second: Unsubscribe is called with the already-closed channel
	// This should NOT panic
	router.Unsubscribe("ch-123", subChan)

	// Verify state
	assert.Equal(t, 0, router.SubscriberCount("ch-123"))
}

// TestMultipleSubscribersCleanupRace tests race with multiple subscribers
func TestMultipleSubscribersCleanupRace(t *testing.T) {
	for i := 0; i < 50; i++ {
		router := NewRouter()

		channel := &Channel{
			ID:          "ch-123",
			ProcessID:   "proc-456",
			Name:        "chat",
			SubmitterID: "user-789",
			ExecutorID:  "exec-abc",
		}
		router.Create(channel)

		// Create multiple subscribers
		sub1, _ := router.Subscribe("ch-123", "user-789")
		sub2, _ := router.Subscribe("ch-123", "exec-abc")

		var wg sync.WaitGroup
		wg.Add(3)

		start := make(chan struct{})

		// Goroutine 1: CleanupProcess
		go func() {
			defer wg.Done()
			<-start
			router.CleanupProcess("proc-456")
		}()

		// Goroutine 2: Unsubscribe sub1
		go func() {
			defer wg.Done()
			<-start
			router.Unsubscribe("ch-123", sub1)
		}()

		// Goroutine 3: Unsubscribe sub2
		go func() {
			defer wg.Done()
			<-start
			router.Unsubscribe("ch-123", sub2)
		}()

		close(start)
		wg.Wait()

		// Should complete without panic
		assert.Equal(t, 0, router.SubscriberCount("ch-123"))
	}
}

// Rate Limiting Tests

func TestRateLimiterBasic(t *testing.T) {
	limiter := NewRateLimiter(10, 5) // 10 max tokens, 5 per second refill

	// Should allow first 10 requests (burst)
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow(), "Request %d should be allowed", i)
	}

	// 11th request should be denied
	assert.False(t, limiter.Allow(), "11th request should be denied")
}

func TestRateLimiterRefill(t *testing.T) {
	limiter := NewRateLimiter(5, 100) // 5 max tokens, 100 per second refill

	// Exhaust all tokens
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow())
	}
	assert.False(t, limiter.Allow())

	// Wait for refill (50ms should give ~5 tokens at 100/sec)
	time.Sleep(60 * time.Millisecond)

	// Should allow more requests now
	assert.True(t, limiter.Allow(), "Should allow after refill")
}

func TestRateLimitEnabledByDefault(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Should allow up to burst size (500 with current constants)
	// Just test that a few messages work
	for i := 1; i <= 10; i++ {
		err := router.Append("ch-123", "user-789", int64(i), 0, []byte("test"))
		assert.Nil(t, err, "Message %d should be allowed", i)
	}
}

func TestRateLimitExceeded(t *testing.T) {
	router := NewRouterForTesting()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	burstSize := constants.CHANNEL_RATE_LIMIT_BURST_SIZE

	// Send messages up to burst size - all should succeed
	for i := 1; i <= burstSize; i++ {
		err := router.Append("ch-123", "user-789", int64(i), 0, []byte("test"))
		assert.Nil(t, err, "Message %d should be allowed (within burst)", i)
	}

	// Next message should fail with rate limit exceeded
	err := router.Append("ch-123", "user-789", int64(burstSize+1), 0, []byte("test"))
	assert.Equal(t, ErrRateLimitExceeded, err)
}

func TestRateLimitDisabled(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Should allow unlimited messages when rate limiting is disabled
	for i := 1; i <= 1000; i++ {
		err := router.Append("ch-123", "user-789", int64(i), 0, []byte("test"))
		assert.Nil(t, err, "Message %d should be allowed without rate limit", i)
	}
}

func TestRateLimitPerProcess(t *testing.T) {
	router := NewRouterForTesting()

	// Create channels for two different processes
	channel1 := &Channel{
		ID:          "ch-1",
		ProcessID:   "proc-1",
		Name:        "chat",
		SubmitterID: "user-1",
	}
	channel2 := &Channel{
		ID:          "ch-2",
		ProcessID:   "proc-2",
		Name:        "chat",
		SubmitterID: "user-2",
	}
	router.Create(channel1)
	router.Create(channel2)

	burstSize := constants.CHANNEL_RATE_LIMIT_BURST_SIZE

	// Exhaust proc-1's limit
	for i := 1; i <= burstSize; i++ {
		err := router.Append("ch-1", "user-1", int64(i), 0, []byte("test"))
		assert.Nil(t, err)
	}
	err := router.Append("ch-1", "user-1", int64(burstSize+1), 0, []byte("test"))
	assert.Equal(t, ErrRateLimitExceeded, err)

	// proc-2 should still have its own independent limit
	err = router.Append("ch-2", "user-2", 1, 0, []byte("test"))
	assert.Nil(t, err, "Different process should have independent rate limit")
}

func TestRateLimitCleanupOnProcessClose(t *testing.T) {
	router := NewRouter()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Trigger limiter creation
	router.Append("ch-123", "user-789", 1, 0, []byte("test"))

	// Verify limiter exists
	router.rateLimitMu.RLock()
	_, exists := router.rateLimiters["proc-456"]
	router.rateLimitMu.RUnlock()
	assert.True(t, exists, "Rate limiter should exist")

	// Cleanup process
	router.CleanupProcess("proc-456")

	// Verify limiter is cleaned up
	router.rateLimitMu.RLock()
	_, exists = router.rateLimiters["proc-456"]
	router.rateLimitMu.RUnlock()
	assert.False(t, exists, "Rate limiter should be cleaned up")
}

// Message Size Tests

func TestMessageSizeWithinLimit(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Message at max size should succeed
	maxPayload := make([]byte, constants.CHANNEL_MAX_MESSAGE_SIZE)
	err := router.Append("ch-123", "user-789", 1, 0, maxPayload)
	assert.Nil(t, err, "Message at max size should be allowed")

	// Verify message was stored
	entries, err := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Len(t, entries[0].Payload, constants.CHANNEL_MAX_MESSAGE_SIZE)
}

func TestMessageSizeExceedsLimit(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Message exceeding max size should fail
	oversizedPayload := make([]byte, constants.CHANNEL_MAX_MESSAGE_SIZE+1)
	err := router.Append("ch-123", "user-789", 1, 0, oversizedPayload)
	assert.Equal(t, ErrMessageTooLarge, err)

	// Verify no message was stored
	entries, err := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 0)
}

func TestMessageSizeEmptyPayload(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Empty payload should succeed
	err := router.Append("ch-123", "user-789", 1, 0, []byte{})
	assert.Nil(t, err, "Empty message should be allowed")

	// Nil payload should also succeed
	err = router.Append("ch-123", "user-789", 2, 0, nil)
	assert.Nil(t, err, "Nil payload should be allowed")
}

// Channel Log Size Tests

func TestChannelLogFull(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetMaxLogEntries(5) // Small limit for testing

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Fill up to limit
	for i := 1; i <= 5; i++ {
		err := router.Append("ch-123", "user-789", int64(i), 0, []byte("test"))
		assert.Nil(t, err, "Message %d should be allowed", i)
	}

	// Verify log is full
	size, err := router.GetLogSize("ch-123")
	assert.Nil(t, err)
	assert.Equal(t, 5, size)

	// Next message should fail
	err = router.Append("ch-123", "user-789", 6, 0, []byte("test"))
	assert.Equal(t, ErrChannelFull, err)

	// Log size should still be 5
	size, err = router.GetLogSize("ch-123")
	assert.Nil(t, err)
	assert.Equal(t, 5, size)
}

func TestChannelLogFullWithDefaultLimit(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
	}
	router.Create(channel)

	// Should allow messages up to the default limit
	// Just verify a reasonable number works
	for i := 1; i <= 100; i++ {
		err := router.Append("ch-123", "user-789", int64(i), 0, []byte("test"))
		assert.Nil(t, err, "Message %d should be allowed with default limit", i)
	}
}

// Channel Count Limit Tests

func TestTooManyChannelsPerProcess(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetMaxChannelsPerProcess(3) // Small limit for testing

	processID := "proc-456"

	// Create channels up to limit
	for i := 1; i <= 3; i++ {
		channel := &Channel{
			ID:          fmt.Sprintf("ch-%d", i),
			ProcessID:   processID,
			Name:        fmt.Sprintf("channel%d", i),
			SubmitterID: "user-789",
		}
		err := router.Create(channel)
		assert.Nil(t, err, "Channel %d should be allowed", i)
	}

	// Verify 3 channels exist
	channels := router.GetChannelsByProcess(processID)
	assert.Len(t, channels, 3)

	// 4th channel should fail
	channel4 := &Channel{
		ID:          "ch-4",
		ProcessID:   processID,
		Name:        "channel4",
		SubmitterID: "user-789",
	}
	err := router.Create(channel4)
	assert.Equal(t, ErrTooManyChannels, err)

	// Still only 3 channels
	channels = router.GetChannelsByProcess(processID)
	assert.Len(t, channels, 3)
}

func TestTooManyChannelsCreateIfNotExists(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetMaxChannelsPerProcess(2)

	processID := "proc-456"

	// Create 2 channels
	channel1 := &Channel{ID: "ch-1", ProcessID: processID, Name: "c1", SubmitterID: "user"}
	channel2 := &Channel{ID: "ch-2", ProcessID: processID, Name: "c2", SubmitterID: "user"}

	err := router.CreateIfNotExists(channel1)
	assert.Nil(t, err)
	err = router.CreateIfNotExists(channel2)
	assert.Nil(t, err)

	// 3rd channel should fail
	channel3 := &Channel{ID: "ch-3", ProcessID: processID, Name: "c3", SubmitterID: "user"}
	err = router.CreateIfNotExists(channel3)
	assert.Equal(t, ErrTooManyChannels, err)

	// Calling CreateIfNotExists on existing channel should still succeed (idempotent)
	err = router.CreateIfNotExists(channel1)
	assert.Nil(t, err, "CreateIfNotExists on existing channel should succeed")
}

func TestChannelLimitPerProcessIndependent(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetMaxChannelsPerProcess(2)

	// Create 2 channels for process 1
	for i := 1; i <= 2; i++ {
		channel := &Channel{
			ID:          fmt.Sprintf("proc1-ch-%d", i),
			ProcessID:   "proc-1",
			Name:        fmt.Sprintf("channel%d", i),
			SubmitterID: "user-1",
		}
		err := router.Create(channel)
		assert.Nil(t, err)
	}

	// Process 1 is at limit
	channel3 := &Channel{ID: "proc1-ch-3", ProcessID: "proc-1", Name: "c3", SubmitterID: "user-1"}
	err := router.Create(channel3)
	assert.Equal(t, ErrTooManyChannels, err)

	// Process 2 should have its own independent limit
	for i := 1; i <= 2; i++ {
		channel := &Channel{
			ID:          fmt.Sprintf("proc2-ch-%d", i),
			ProcessID:   "proc-2",
			Name:        fmt.Sprintf("channel%d", i),
			SubmitterID: "user-2",
		}
		err := router.Create(channel)
		assert.Nil(t, err, "Process 2 should have independent limit")
	}
}

func TestChannelLimitWithDefaultValue(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	// Default limit is 100, just verify we can create a reasonable number
	processID := "proc-456"
	for i := 1; i <= 50; i++ {
		channel := &Channel{
			ID:          fmt.Sprintf("ch-%d", i),
			ProcessID:   processID,
			Name:        fmt.Sprintf("channel%d", i),
			SubmitterID: "user-789",
		}
		err := router.Create(channel)
		assert.Nil(t, err, "Channel %d should be allowed with default limit", i)
	}
}

func TestSlowSubscriberDisconnected(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetSubscriberBufferSize(3) // Small buffer for testing

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "data",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Subscribe but don't consume messages
	subCh, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)
	assert.NotNil(t, subCh)

	// Fill the buffer (3 messages)
	for i := 0; i < 3; i++ {
		err := router.Append("ch-123", "exec-123", int64(i), 0, []byte(fmt.Sprintf("msg%d", i)))
		assert.Nil(t, err)
	}

	// Subscriber should still be connected (buffer is full but not overflowed)
	assert.Equal(t, 1, router.SubscriberCount("ch-123"))

	// Send one more message - this should trigger disconnection
	err = router.Append("ch-123", "exec-123", 3, 0, []byte("overflow"))
	assert.Nil(t, err)

	// Give cleanup time to complete
	time.Sleep(10 * time.Millisecond)

	// Subscriber should be disconnected
	assert.Equal(t, 0, router.SubscriberCount("ch-123"))

	// The subscriber channel should be closed
	_, ok := <-subCh
	// Keep draining until closed
	for ok {
		_, ok = <-subCh
	}
	assert.False(t, ok, "Subscriber channel should be closed")
}

func TestSlowSubscriberCanDetectDisconnection(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetSubscriberBufferSize(2)

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "data",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	subCh, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)

	// Fill buffer and overflow
	for i := 0; i < 5; i++ {
		router.Append("ch-123", "exec-123", int64(i), 0, []byte(fmt.Sprintf("msg%d", i)))
	}

	// Subscriber should receive error message before channel closes
	var lastMsg *MsgEntry
	for {
		select {
		case msg, ok := <-subCh:
			if !ok {
				// Channel is closed - verify we got error message
				assert.NotNil(t, lastMsg, "Should have received at least one message")
				assert.Equal(t, ErrSubscriberTooSlow.Error(), lastMsg.Error, "Last message should contain error")
				return
			}
			lastMsg = msg
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected subscriber channel to be closed")
		}
	}
}

func TestMultipleSubscribersOneSlowOneFast(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetSubscriberBufferSize(3)

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "data",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Subscribe two subscribers
	slowCh, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)
	fastCh, err := router.Subscribe("ch-123", "exec-123")
	assert.Nil(t, err)

	assert.Equal(t, 2, router.SubscriberCount("ch-123"))

	// Fast subscriber consumes in a goroutine
	fastReceived := make(chan int, 10)
	go func() {
		count := 0
		for range fastCh {
			count++
			fastReceived <- count
		}
		close(fastReceived)
	}()

	// Send messages - slow subscriber doesn't consume
	for i := 0; i < 10; i++ {
		err := router.Append("ch-123", "exec-123", int64(i), 0, []byte(fmt.Sprintf("msg%d", i)))
		assert.Nil(t, err)
		time.Sleep(5 * time.Millisecond) // Give fast subscriber time to consume
	}

	// Wait for fast subscriber to receive all messages
	time.Sleep(50 * time.Millisecond)

	// Slow subscriber should be disconnected, fast should remain
	assert.Equal(t, 1, router.SubscriberCount("ch-123"))

	// Verify slow subscriber channel is closed
	select {
	case _, ok := <-slowCh:
		if ok {
			// Drain remaining
			for range slowCh {
			}
		}
	default:
	}

	// Cleanup
	router.CleanupProcess("proc-456")
}

func TestSubscriberBufferSizeDefault(t *testing.T) {
	router := NewRouter()

	// Verify default buffer size is from constants
	assert.Equal(t, constants.CHANNEL_SUBSCRIBER_BUFFER_SIZE, router.subscriberBufferSize)
}

func TestUnsubscribeAlreadyClosedSubscriber(t *testing.T) {
	router := NewRouterWithoutRateLimit()
	router.SetSubscriberBufferSize(2)

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "data",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	subCh, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)

	// Overflow to trigger disconnection
	for i := 0; i < 5; i++ {
		router.Append("ch-123", "exec-123", int64(i), 0, []byte("msg"))
	}

	time.Sleep(10 * time.Millisecond)

	// Subscriber should already be removed
	assert.Equal(t, 0, router.SubscriberCount("ch-123"))

	// Unsubscribe should not panic even though subscriber is already disconnected
	router.Unsubscribe("ch-123", subCh) // Should not panic
}

func TestAppendWithType(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Append regular message
	err = router.Append("ch-123", "exec-123", 1, 0, []byte("hello"))
	assert.Nil(t, err)

	// Append typed message (end-of-stream)
	err = router.AppendWithType("ch-123", "exec-123", 2, 0, nil, MsgTypeEnd)
	assert.Nil(t, err)

	// Read messages and verify types
	entries, err := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)

	// First message should be data type
	assert.Equal(t, MsgTypeData, entries[0].Type)
	assert.Equal(t, "hello", string(entries[0].Payload))

	// Second message should be end type
	assert.Equal(t, MsgTypeEnd, entries[1].Type)
}

func TestAppendWithTypeEndOfStream(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Subscribe to channel
	subCh, err := router.Subscribe("ch-123", "user-789")
	assert.Nil(t, err)

	// Append some messages then end marker
	err = router.Append("ch-123", "exec-123", 1, 0, []byte("token1"))
	assert.Nil(t, err)
	err = router.Append("ch-123", "exec-123", 2, 0, []byte("token2"))
	assert.Nil(t, err)
	err = router.AppendWithType("ch-123", "exec-123", 3, 0, nil, MsgTypeEnd)
	assert.Nil(t, err)

	// Receive messages
	var received []*MsgEntry
	for i := 0; i < 3; i++ {
		select {
		case msg := <-subCh:
			received = append(received, msg)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Timeout waiting for message")
		}
	}

	// Verify we got all messages in order
	assert.Len(t, received, 3)
	assert.Equal(t, "token1", string(received[0].Payload))
	assert.Equal(t, MsgTypeData, received[0].Type)
	assert.Equal(t, "token2", string(received[1].Payload))
	assert.Equal(t, MsgTypeData, received[1].Type)
	assert.Equal(t, MsgTypeEnd, received[2].Type)
}

func TestAppendWithTypeError(t *testing.T) {
	router := NewRouterWithoutRateLimit()

	channel := &Channel{
		ID:          "ch-123",
		ProcessID:   "proc-456",
		Name:        "chat",
		SubmitterID: "user-789",
		ExecutorID:  "exec-123",
	}
	err := router.Create(channel)
	assert.Nil(t, err)

	// Append error message
	err = router.AppendWithType("ch-123", "exec-123", 1, 0, []byte("something went wrong"), MsgTypeError)
	assert.Nil(t, err)

	// Read and verify
	entries, err := router.ReadAfter("ch-123", "user-789", 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, MsgTypeError, entries[0].Type)
	assert.Equal(t, "something went wrong", string(entries[0].Payload))
}

func TestMsgTypeConstants(t *testing.T) {
	// Verify constant values
	assert.Equal(t, "data", MsgTypeData)
	assert.Equal(t, "end", MsgTypeEnd)
	assert.Equal(t, "error", MsgTypeError)
}
