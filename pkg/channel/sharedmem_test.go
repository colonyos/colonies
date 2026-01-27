package channel

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MemoryNetwork is an in-memory network implementation for testing
type MemoryNetwork struct {
	mu          sync.RWMutex
	subscribers []chan []byte
	closed      bool
}

func NewMemoryNetwork() *MemoryNetwork {
	return &MemoryNetwork{
		subscribers: make([]chan []byte, 0),
	}
}

func (n *MemoryNetwork) Broadcast(msg []byte) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.closed {
		return nil
	}

	for _, sub := range n.subscribers {
		select {
		case sub <- msg:
		default:
		}
	}
	return nil
}

func (n *MemoryNetwork) Subscribe() chan []byte {
	return n.SubscribeWithBuffer(100)
}

func (n *MemoryNetwork) SubscribeWithBuffer(bufferSize int) chan []byte {
	n.mu.Lock()
	defer n.mu.Unlock()

	ch := make(chan []byte, bufferSize)
	n.subscribers = append(n.subscribers, ch)
	return ch
}

func (n *MemoryNetwork) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.closed = true
	for _, sub := range n.subscribers {
		close(sub)
	}
}

func TestSharedMemBroadcastAndReceive(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	msgChan := network.Subscribe()
	go func() {
		for msg := range msgChan {
			sm.HandleIncoming(msg)
		}
	}()

	entry := &MsgEntry{
		Sequence:  1,
		Timestamp: time.Now(),
		SenderID:  "executor-1",
		Payload:   []byte("Hello, World!"),
		Type:      MsgTypeData,
	}

	err := sm.Broadcast("process-123", "output", entry)
	assert.Nil(t, err)

	select {
	case msg := <-sm.Receive():
		assert.Equal(t, "process-123", msg.ProcessID)
		assert.Equal(t, "output", msg.ChannelName)
		assert.Equal(t, int64(1), msg.Entry.Sequence)
		assert.Equal(t, []byte("Hello, World!"), msg.Entry.Payload)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestSharedMemReceiveWithContext(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	msgChan := network.Subscribe()
	go func() {
		for msg := range msgChan {
			sm.HandleIncoming(msg)
		}
	}()

	entry := &MsgEntry{
		Sequence:  1,
		Timestamp: time.Now(),
		SenderID:  "executor-1",
		Payload:   []byte("context test"),
	}

	sm.Broadcast("process-ctx", "output", entry)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	msg := sm.ReceiveWithContext(ctx)
	assert.NotNil(t, msg)
	assert.Equal(t, "process-ctx", msg.ProcessID)
}

func TestSharedMemReceiveWithContextTimeout(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	msg := sm.ReceiveWithContext(ctx)
	assert.Nil(t, msg, "Should return nil on context timeout")
}

func TestSharedMemMultipleNodes(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	var receivedCount int32
	var wg sync.WaitGroup

	nodes := make([]*SharedMem, 3)
	for i := 0; i < 3; i++ {
		nodes[i] = NewSharedMem(network, 100)

		msgChan := network.Subscribe()
		go func(sm *SharedMem) {
			for msg := range msgChan {
				sm.HandleIncoming(msg)
			}
		}(nodes[i])

		wg.Add(1)
		go func(sm *SharedMem) {
			for msg := range sm.Receive() {
				if msg != nil {
					atomic.AddInt32(&receivedCount, 1)
				}
			}
			wg.Done()
		}(nodes[i])
	}

	entry := &MsgEntry{
		Sequence:  1,
		Timestamp: time.Now(),
		SenderID:  "executor-1",
		Payload:   []byte("broadcast test"),
	}

	nodes[0].Broadcast("process-456", "stream", entry)

	time.Sleep(100 * time.Millisecond)

	for _, node := range nodes {
		node.Close()
	}
	wg.Wait()

	assert.Equal(t, int32(3), atomic.LoadInt32(&receivedCount))
}

func TestSharedMemOrderPreservation(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 200)
	defer sm.Close()

	msgChan := network.Subscribe()
	go func() {
		for msg := range msgChan {
			sm.HandleIncoming(msg)
		}
	}()

	for i := int64(1); i <= 100; i++ {
		entry := &MsgEntry{
			Sequence:  i,
			Timestamp: time.Now(),
			SenderID:  "executor-1",
			Payload:   []byte("message"),
		}
		sm.Broadcast("process-789", "output", entry)
	}

	receivedEntries := make([]*MsgEntry, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for len(receivedEntries) < 100 {
		msg := sm.ReceiveWithContext(ctx)
		if msg == nil {
			break
		}
		receivedEntries = append(receivedEntries, msg.Entry)
	}

	assert.Len(t, receivedEntries, 100)
	for i, entry := range receivedEntries {
		assert.Equal(t, int64(i+1), entry.Sequence)
	}
}

func TestSharedMemConcurrentBroadcast(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 1000)
	defer sm.Close()

	msgChan := network.SubscribeWithBuffer(1000)
	go func() {
		for msg := range msgChan {
			sm.HandleIncoming(msg)
		}
	}()

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 50

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < messagesPerGoroutine; i++ {
				entry := &MsgEntry{
					Sequence:  int64(i + 1),
					Timestamp: time.Now(),
					SenderID:  "executor-1",
					Payload:   []byte("concurrent message"),
				}
				sm.Broadcast("process-concurrent", "output", entry)
			}
		}(g)
	}

	wg.Wait()

	var receivedCount int32
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for {
		msg := sm.ReceiveWithContext(ctx)
		if msg == nil {
			break
		}
		atomic.AddInt32(&receivedCount, 1)
	}

	expectedTotal := int32(numGoroutines * messagesPerGoroutine)
	assert.Equal(t, expectedTotal, atomic.LoadInt32(&receivedCount))
}

func TestSharedMemClose(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)

	sm.Close()

	_, ok := <-sm.Receive()
	assert.False(t, ok, "Receive channel should be closed")

	// Double close should not panic
	sm.Close()
}

func TestSharedMemInvalidJSON(t *testing.T) {
	sm := NewSharedMem(nil, 100)
	defer sm.Close()

	err := sm.HandleIncoming([]byte("not valid json"))
	assert.NotNil(t, err)
}

// Process tracking tests

func TestSharedMemTracksProcessOnBroadcast(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	assert.Equal(t, 0, sm.ActiveProcessCount())

	entry := &MsgEntry{Sequence: 1, Timestamp: time.Now(), SenderID: "executor-1", Payload: []byte("test")}
	sm.Broadcast("process-1", "output", entry)
	sm.Broadcast("process-2", "output", entry)
	sm.Broadcast("process-1", "output", entry) // duplicate

	assert.Equal(t, 2, sm.ActiveProcessCount())

	processes := sm.GetActiveProcesses()
	sort.Strings(processes)
	assert.Equal(t, []string{"process-1", "process-2"}, processes)
}

func TestSharedMemTracksProcessOnHandleIncoming(t *testing.T) {
	sm := NewSharedMem(nil, 100)
	defer sm.Close()

	assert.Equal(t, 0, sm.ActiveProcessCount())

	msg1 := `{"processid":"process-a","channelname":"out","entry":{"sequence":1}}`
	msg2 := `{"processid":"process-b","channelname":"out","entry":{"sequence":1}}`

	sm.HandleIncoming([]byte(msg1))
	sm.HandleIncoming([]byte(msg2))

	assert.Equal(t, 2, sm.ActiveProcessCount())

	processes := sm.GetActiveProcesses()
	sort.Strings(processes)
	assert.Equal(t, []string{"process-a", "process-b"}, processes)
}

func TestSharedMemCloseProcess(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	entry := &MsgEntry{Sequence: 1, Timestamp: time.Now(), SenderID: "executor-1", Payload: []byte("test")}
	sm.Broadcast("process-1", "output", entry)
	sm.Broadcast("process-2", "output", entry)
	sm.Broadcast("process-3", "output", entry)

	assert.Equal(t, 3, sm.ActiveProcessCount())

	sm.CloseProcess("process-2")

	assert.Equal(t, 2, sm.ActiveProcessCount())

	processes := sm.GetActiveProcesses()
	sort.Strings(processes)
	assert.Equal(t, []string{"process-1", "process-3"}, processes)
}

func TestSharedMemCloseProcessIdempotent(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	entry := &MsgEntry{Sequence: 1, Timestamp: time.Now(), SenderID: "executor-1", Payload: []byte("test")}
	sm.Broadcast("process-1", "output", entry)

	assert.Equal(t, 1, sm.ActiveProcessCount())

	sm.CloseProcess("process-1")
	sm.CloseProcess("process-1") // double close
	sm.CloseProcess("non-existent") // close non-existent

	assert.Equal(t, 0, sm.ActiveProcessCount())
}

func TestSharedMemGetActiveProcessesThreadSafe(t *testing.T) {
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	var wg sync.WaitGroup

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			entry := &MsgEntry{Sequence: int64(i), Timestamp: time.Now(), SenderID: "executor-1", Payload: []byte("test")}
			sm.Broadcast("process-"+string(rune('A'+i%26)), "output", entry)
		}
	}()

	// Reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = sm.GetActiveProcesses()
			_ = sm.ActiveProcessCount()
		}
	}()

	// Closer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			sm.CloseProcess("process-" + string(rune('A'+i%26)))
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()
	// Should complete without race conditions
}

func TestSharedMemExternalGCPattern(t *testing.T) {
	// This test demonstrates how external GC would work
	network := NewMemoryNetwork()
	defer network.Close()

	sm := NewSharedMem(network, 100)
	defer sm.Close()

	// Simulate some processes
	entry := &MsgEntry{Sequence: 1, Timestamp: time.Now(), SenderID: "executor-1", Payload: []byte("test")}
	sm.Broadcast("process-active-1", "output", entry)
	sm.Broadcast("process-active-2", "output", entry)
	sm.Broadcast("process-closed-1", "output", entry)
	sm.Broadcast("process-closed-2", "output", entry)

	assert.Equal(t, 4, sm.ActiveProcessCount())

	// External GC: check each process against "database"
	activeInDB := map[string]bool{
		"process-active-1": true,
		"process-active-2": true,
		"process-closed-1": false,
		"process-closed-2": false,
	}

	// GC loop (would run periodically in production)
	processes := sm.GetActiveProcesses()
	for _, processID := range processes {
		if !activeInDB[processID] {
			sm.CloseProcess(processID)
		}
	}

	assert.Equal(t, 2, sm.ActiveProcessCount())

	remaining := sm.GetActiveProcesses()
	sort.Strings(remaining)
	assert.Equal(t, []string{"process-active-1", "process-active-2"}, remaining)
}
