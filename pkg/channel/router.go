package channel

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
	ErrUnauthorized    = errors.New("unauthorized access to channel")
	ErrChannelExists   = errors.New("channel already exists")
)

// Subscriber represents a channel subscriber waiting for new entries
type Subscriber struct {
	ch        chan *MsgEntry
	channelID string
}

// Router manages channels in memory
type Router struct {
	mu          sync.RWMutex
	channels    map[string]*Channel
	byProcess   map[string][]string // processID → []channelID
	replicator  Replicator
	syncMode    bool // If true, replication is synchronous (for testing)
	subMu       sync.RWMutex
	subscribers map[string][]*Subscriber // channelID → subscribers
}

// NewRouter creates a new channel router
func NewRouter() *Router {
	return &Router{
		channels:    make(map[string]*Channel),
		byProcess:   make(map[string][]string),
		replicator:  &NoOpReplicator{},
		subscribers: make(map[string][]*Subscriber),
	}
}

// NewRouterWithReplicator creates a router with a custom replicator
func NewRouterWithReplicator(replicator Replicator) *Router {
	return &Router{
		channels:    make(map[string]*Channel),
		byProcess:   make(map[string][]string),
		replicator:  replicator,
		subscribers: make(map[string][]*Subscriber),
	}
}

// SetReplicator sets the replicator for this router
func (r *Router) SetReplicator(replicator Replicator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replicator = replicator
}

// SetSyncMode enables synchronous replication (useful for testing)
func (r *Router) SetSyncMode(sync bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.syncMode = sync
}

// Create creates a new channel
func (r *Router) Create(channel *Channel) error {
	r.mu.Lock()

	if _, exists := r.channels[channel.ID]; exists {
		r.mu.Unlock()
		return ErrChannelExists
	}

	// Initialize log if nil
	if channel.Log == nil {
		channel.Log = make([]*MsgEntry, 0)
	}

	r.channels[channel.ID] = channel

	// Index by process
	r.byProcess[channel.ProcessID] = append(r.byProcess[channel.ProcessID], channel.ID)

	// Copy channel metadata for replication (avoid data race)
	channelMeta := &Channel{
		ID:          channel.ID,
		ProcessID:   channel.ProcessID,
		Name:        channel.Name,
		SubmitterID: channel.SubmitterID,
		ExecutorID:  channel.ExecutorID,
		Sequence:    0,
		Log:         nil,
	}

	r.mu.Unlock()

	// Replicate to peers
	if r.syncMode {
		r.replicator.ReplicateChannel(channelMeta)
	} else {
		go r.replicator.ReplicateChannel(channelMeta)
	}

	return nil
}

// CreateIfNotExists creates a channel only if it doesn't already exist
// Returns nil on success or if channel already exists (idempotent)
func (r *Router) CreateIfNotExists(channel *Channel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.channels[channel.ID]; exists {
		return nil // Already exists, success
	}

	// Initialize log if nil
	if channel.Log == nil {
		channel.Log = make([]*MsgEntry, 0)
	}

	r.channels[channel.ID] = channel

	// Index by process
	r.byProcess[channel.ProcessID] = append(r.byProcess[channel.ProcessID], channel.ID)

	// Note: No replication here - this is called from replication handler
	return nil
}

// Get retrieves a channel by ID
func (r *Router) Get(channelID string) (*Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channel, exists := r.channels[channelID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	return channel, nil
}

// GetByProcessAndName retrieves a channel by process ID and name
func (r *Router) GetByProcessAndName(processID string, name string) (*Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channelIDs, exists := r.byProcess[processID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	for _, id := range channelIDs {
		channel := r.channels[id]
		if channel.Name == name {
			return channel, nil
		}
	}

	return nil, ErrChannelNotFound
}

// GetChannelsByProcess retrieves all channels for a process
func (r *Router) GetChannelsByProcess(processID string) []*Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channelIDs, exists := r.byProcess[processID]
	if !exists {
		return nil
	}

	channels := make([]*Channel, 0, len(channelIDs))
	for _, id := range channelIDs {
		if channel, exists := r.channels[id]; exists {
			channels = append(channels, channel)
		}
	}

	return channels
}

// Append adds a message to a channel with client-assigned sequence number
func (r *Router) Append(channelID string, senderID string, sequence int64, inReplyTo int64, payload []byte) error {
	r.mu.RLock()
	channel, exists := r.channels[channelID]
	r.mu.RUnlock()

	if !exists {
		return ErrChannelNotFound
	}

	// Check authorization
	if err := r.authorize(channel, senderID); err != nil {
		return err
	}

	// Lock channel for writing
	r.mu.Lock()

	entry := &MsgEntry{
		Sequence:  sequence, // Client-assigned
		InReplyTo: inReplyTo,
		Timestamp: time.Now(),
		SenderID:  senderID,
		Payload:   payload,
	}
	channel.Log = append(channel.Log, entry)

	// Keep sorted by (SenderID, Sequence) for causal ordering
	sort.Slice(channel.Log, func(i, j int) bool {
		if channel.Log[i].SenderID == channel.Log[j].SenderID {
			return channel.Log[i].Sequence < channel.Log[j].Sequence
		}
		// For different senders, sort by timestamp
		return channel.Log[i].Timestamp.Before(channel.Log[j].Timestamp)
	})

	// Copy channel metadata for replication (avoid data race)
	channelMeta := &Channel{
		ID:          channel.ID,
		ProcessID:   channel.ProcessID,
		Name:        channel.Name,
		SubmitterID: channel.SubmitterID,
		ExecutorID:  channel.ExecutorID,
		Sequence:    0,
		Log:         nil,
	}

	r.mu.Unlock()

	// Notify subscribers (push-based)
	r.notifySubscribers(channelID, entry)

	// Replicate to peers - include channel info to handle race conditions
	if r.syncMode {
		r.replicator.ReplicateEntry(channelMeta, entry)
	} else {
		go r.replicator.ReplicateEntry(channelMeta, entry)
	}

	return nil
}

// ReadAfter reads entries after a given index (position in log)
// Since sequences are per-sender, we use index-based reading
func (r *Router) ReadAfter(channelID string, callerID string, afterIndex int64, limit int) ([]*MsgEntry, error) {
	r.mu.RLock()
	channel, exists := r.channels[channelID]
	r.mu.RUnlock()

	if !exists {
		return nil, ErrChannelNotFound
	}

	// Check authorization
	if err := r.authorize(channel, callerID); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// afterIndex is the last index read, so we start from afterIndex+1
	startIdx := int(afterIndex)
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(channel.Log) {
		return []*MsgEntry{}, nil
	}

	endIdx := len(channel.Log)
	if limit > 0 && startIdx+limit < endIdx {
		endIdx = startIdx + limit
	}

	result := make([]*MsgEntry, endIdx-startIdx)
	copy(result, channel.Log[startIdx:endIdx])

	return result, nil
}

// authorize checks if caller has access to channel
func (r *Router) authorize(channel *Channel, callerID string) error {
	if callerID != channel.SubmitterID && callerID != channel.ExecutorID {
		return ErrUnauthorized
	}
	return nil
}

// SetExecutorID updates the executor ID for a channel (called when process is assigned)
func (r *Router) SetExecutorID(channelID string, executorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	channel, exists := r.channels[channelID]
	if !exists {
		return ErrChannelNotFound
	}

	channel.ExecutorID = executorID
	return nil
}

// SetExecutorIDForProcess updates executor ID for all channels of a process
func (r *Router) SetExecutorIDForProcess(processID string, executorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	channelIDs, exists := r.byProcess[processID]
	if !exists {
		return nil // No channels for this process
	}

	for _, id := range channelIDs {
		if channel, exists := r.channels[id]; exists {
			channel.ExecutorID = executorID
		}
	}

	// Replicate to peers
	if r.syncMode {
		r.replicator.ReplicateExecutorAssignment(processID, executorID)
	} else {
		go r.replicator.ReplicateExecutorAssignment(processID, executorID)
	}

	return nil
}

// CleanupProcess removes all channels for a process
func (r *Router) CleanupProcess(processID string) {
	r.mu.Lock()

	channelIDs, exists := r.byProcess[processID]
	if !exists {
		r.mu.Unlock()
		return
	}

	// Collect channel IDs to clean up subscribers
	idsToClean := make([]string, len(channelIDs))
	copy(idsToClean, channelIDs)

	for _, id := range channelIDs {
		delete(r.channels, id)
	}

	delete(r.byProcess, processID)

	r.mu.Unlock()

	// Clean up subscribers for deleted channels
	r.subMu.Lock()
	for _, id := range idsToClean {
		// Close all subscriber channels
		for _, sub := range r.subscribers[id] {
			close(sub.ch)
		}
		delete(r.subscribers, id)
	}
	r.subMu.Unlock()

	// Replicate to peers
	if r.syncMode {
		r.replicator.ReplicateCleanup(processID)
	} else {
		go r.replicator.ReplicateCleanup(processID)
	}
}

// GetSequence returns the current sequence number for a channel
func (r *Router) GetSequence(channelID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channel, exists := r.channels[channelID]
	if !exists {
		return 0, ErrChannelNotFound
	}

	return channel.Sequence, nil
}

// GetLogSize returns the number of entries in a channel
func (r *Router) GetLogSize(channelID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channel, exists := r.channels[channelID]
	if !exists {
		return 0, ErrChannelNotFound
	}

	return len(channel.Log), nil
}

// ReplicateEntry adds an entry from replication (used for distributed setup)
func (r *Router) ReplicateEntry(channelID string, entry *MsgEntry) error {
	r.mu.Lock()

	channel, exists := r.channels[channelID]
	if !exists {
		r.mu.Unlock()
		return ErrChannelNotFound
	}

	// Idempotent - check if already have this entry (same sender + sequence)
	for _, e := range channel.Log {
		if e.SenderID == entry.SenderID && e.Sequence == entry.Sequence {
			r.mu.Unlock()
			return nil // Already have it
		}
	}

	channel.Log = append(channel.Log, entry)

	// Keep sorted by (SenderID, Sequence) for causal ordering
	sort.Slice(channel.Log, func(i, j int) bool {
		if channel.Log[i].SenderID == channel.Log[j].SenderID {
			return channel.Log[i].Sequence < channel.Log[j].Sequence
		}
		// For different senders, sort by timestamp
		return channel.Log[i].Timestamp.Before(channel.Log[j].Timestamp)
	})

	r.mu.Unlock()

	// Notify subscribers (push replicated entries to local subscribers)
	r.notifySubscribers(channelID, entry)

	return nil
}

// Stats returns statistics about the router
func (r *Router) Stats() (channelCount int, processCount int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.channels), len(r.byProcess)
}

// Subscribe registers for push notifications on a channel
// Returns a channel that receives entries as they're appended
func (r *Router) Subscribe(channelID string, callerID string) (chan *MsgEntry, error) {
	r.mu.RLock()
	channel, exists := r.channels[channelID]
	r.mu.RUnlock()

	if !exists {
		return nil, ErrChannelNotFound
	}

	// Verify authorization
	if err := r.authorize(channel, callerID); err != nil {
		return nil, err
	}

	ch := make(chan *MsgEntry, 100)
	sub := &Subscriber{ch: ch, channelID: channelID}

	r.subMu.Lock()
	r.subscribers[channelID] = append(r.subscribers[channelID], sub)
	r.subMu.Unlock()

	return ch, nil
}

// Unsubscribe removes a subscriber from a channel
func (r *Router) Unsubscribe(channelID string, ch chan *MsgEntry) {
	r.subMu.Lock()
	defer r.subMu.Unlock()

	subs := r.subscribers[channelID]
	for i, sub := range subs {
		if sub.ch == ch {
			// Remove subscriber by replacing with last element and truncating
			r.subscribers[channelID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}

	// Clean up empty subscriber lists
	if len(r.subscribers[channelID]) == 0 {
		delete(r.subscribers, channelID)
	}
}

// notifySubscribers sends an entry to all subscribers of a channel
func (r *Router) notifySubscribers(channelID string, entry *MsgEntry) {
	r.subMu.RLock()
	defer r.subMu.RUnlock()

	for _, sub := range r.subscribers[channelID] {
		select {
		case sub.ch <- entry:
			// Successfully sent
		default:
			// Channel full, subscriber too slow - skip to avoid blocking
		}
	}
}

// SubscriberCount returns the number of subscribers for a channel (for testing)
func (r *Router) SubscriberCount(channelID string) int {
	r.subMu.RLock()
	defer r.subMu.RUnlock()
	return len(r.subscribers[channelID])
}
