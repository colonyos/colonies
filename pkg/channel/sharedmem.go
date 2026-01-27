package channel

import (
	"context"
	"encoding/json"
	"sync"
)

// Network interface for broadcasting messages to other nodes
type Network interface {
	Broadcast(msg []byte) error
}

// ChannelMessage is the wire format for broadcasting channel entries
type ChannelMessage struct {
	ProcessID   string    `json:"processid"`
	ChannelName string    `json:"channelname"`
	Entry       *MsgEntry `json:"entry"`
}

// SharedMem handles broadcasting and receiving channel messages across nodes
type SharedMem struct {
	network         Network
	receiveChan     chan *ChannelMessage
	activeProcesses map[string]struct{} // set of active process IDs
	mu              sync.RWMutex
	closed          bool
}

// NewSharedMem creates a new SharedMem instance
func NewSharedMem(network Network, bufferSize int) *SharedMem {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	return &SharedMem{
		network:         network,
		receiveChan:     make(chan *ChannelMessage, bufferSize),
		activeProcesses: make(map[string]struct{}),
	}
}

// Broadcast sends a channel entry to all nodes in the cluster
func (sm *SharedMem) Broadcast(processID, channelName string, entry *MsgEntry) error {
	sm.mu.Lock()
	sm.activeProcesses[processID] = struct{}{}
	sm.mu.Unlock()

	msg := ChannelMessage{
		ProcessID:   processID,
		ChannelName: channelName,
		Entry:       entry,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return sm.network.Broadcast(data)
}

// Receive returns a channel that receives messages from the network
func (sm *SharedMem) Receive() <-chan *ChannelMessage {
	return sm.receiveChan
}

// HandleIncoming processes an incoming message from the network
func (sm *SharedMem) HandleIncoming(data []byte) error {
	var msg ChannelMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.closed {
		return nil
	}

	sm.activeProcesses[msg.ProcessID] = struct{}{}

	select {
	case sm.receiveChan <- &msg:
	default:
	}

	return nil
}

// CloseProcess removes a process from tracking
func (sm *SharedMem) CloseProcess(processID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.activeProcesses, processID)
}

// GetActiveProcesses returns a snapshot of all active process IDs
func (sm *SharedMem) GetActiveProcesses() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	processes := make([]string, 0, len(sm.activeProcesses))
	for processID := range sm.activeProcesses {
		processes = append(processes, processID)
	}
	return processes
}

// ActiveProcessCount returns the number of active processes
func (sm *SharedMem) ActiveProcessCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.activeProcesses)
}

// Close closes the receive channel
func (sm *SharedMem) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.closed {
		sm.closed = true
		close(sm.receiveChan)
	}
}

// ReceiveWithContext receives a message with context for timeout/cancellation
func (sm *SharedMem) ReceiveWithContext(ctx context.Context) *ChannelMessage {
	select {
	case msg, ok := <-sm.receiveChan:
		if !ok {
			return nil
		}
		return msg
	case <-ctx.Done():
		return nil
	}
}
