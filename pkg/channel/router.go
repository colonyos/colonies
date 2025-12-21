package channel

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/constants"
	log "github.com/sirupsen/logrus"
)

var (
	ErrChannelNotFound      = errors.New("channel not found")
	ErrUnauthorized         = errors.New("unauthorized access to channel")
	ErrChannelExists        = errors.New("channel already exists")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrMessageTooLarge      = errors.New("message payload exceeds maximum size")
	ErrChannelFull          = errors.New("channel log is full")
	ErrTooManyChannels      = errors.New("process has too many channels")
	ErrSubscriberTooSlow    = errors.New("subscriber disconnected: buffer full")
)

// Subscriber represents a channel subscriber waiting for new entries
type Subscriber struct {
	ch        chan *MsgEntry
	channelID string
	closed    bool // Set to true when subscriber is disconnected for being too slow
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64   // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new token bucket rate limiter
func NewRateLimiter(maxTokens float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens, // Start with full bucket
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// Tokens returns the current number of tokens (for testing)
func (rl *RateLimiter) Tokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens
}

// Router manages channels in memory
type Router struct {
	mu          sync.RWMutex
	channels    map[string]*Channel
	byProcess   map[string][]string      // processID -> []channelID
	subMu       sync.RWMutex
	subscribers map[string][]*Subscriber // channelID -> subscribers

	// Rate limiting
	rateLimitMu      sync.RWMutex
	rateLimiters     map[string]*RateLimiter // processID -> limiter
	rateLimitEnabled bool

	// Log size limiting
	maxLogEntries int

	// Channel count limiting
	maxChannelsPerProcess int

	// Subscriber buffer size
	subscriberBufferSize int
}

// NewRouter creates a new channel router with rate limiting enabled
func NewRouter() *Router {
	return &Router{
		channels:              make(map[string]*Channel),
		byProcess:             make(map[string][]string),
		subscribers:           make(map[string][]*Subscriber),
		rateLimiters:          make(map[string]*RateLimiter),
		rateLimitEnabled:      true,
		maxLogEntries:         constants.CHANNEL_MAX_LOG_ENTRIES,
		maxChannelsPerProcess: constants.CHANNEL_MAX_CHANNELS_PER_PROCESS,
		subscriberBufferSize:  constants.CHANNEL_SUBSCRIBER_BUFFER_SIZE,
	}
}

// NewRouterWithoutRateLimit creates a router without rate limiting (for testing)
func NewRouterWithoutRateLimit() *Router {
	return &Router{
		channels:              make(map[string]*Channel),
		byProcess:             make(map[string][]string),
		subscribers:           make(map[string][]*Subscriber),
		rateLimiters:          make(map[string]*RateLimiter),
		rateLimitEnabled:      false,
		maxLogEntries:         constants.CHANNEL_MAX_LOG_ENTRIES,
		maxChannelsPerProcess: constants.CHANNEL_MAX_CHANNELS_PER_PROCESS,
		subscriberBufferSize:  constants.CHANNEL_SUBSCRIBER_BUFFER_SIZE,
	}
}

// SetMaxLogEntries sets the maximum log entries per channel (for testing)
func (r *Router) SetMaxLogEntries(max int) {
	r.maxLogEntries = max
}

// SetSubscriberBufferSize sets the subscriber buffer size (for testing)
func (r *Router) SetSubscriberBufferSize(size int) {
	r.subscriberBufferSize = size
}

// SetMaxChannelsPerProcess sets the maximum channels per process (for testing)
func (r *Router) SetMaxChannelsPerProcess(max int) {
	r.maxChannelsPerProcess = max
}

// SetRateLimitEnabled enables or disables rate limiting
func (r *Router) SetRateLimitEnabled(enabled bool) {
	r.rateLimitMu.Lock()
	defer r.rateLimitMu.Unlock()
	r.rateLimitEnabled = enabled
}

// getRateLimiter returns the rate limiter for a process, creating one if needed
func (r *Router) getRateLimiter(processID string) *RateLimiter {
	r.rateLimitMu.RLock()
	limiter, exists := r.rateLimiters[processID]
	r.rateLimitMu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter
	r.rateLimitMu.Lock()
	defer r.rateLimitMu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := r.rateLimiters[processID]; exists {
		return limiter
	}

	limiter = NewRateLimiter(
		float64(constants.CHANNEL_RATE_LIMIT_BURST_SIZE),
		constants.CHANNEL_RATE_LIMIT_MESSAGES_PER_SECOND,
	)
	r.rateLimiters[processID] = limiter
	return limiter
}

// Create creates a new channel
func (r *Router) Create(channel *Channel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.channels[channel.ID]; exists {
		return ErrChannelExists
	}

	// Check channel count limit per process
	if len(r.byProcess[channel.ProcessID]) >= r.maxChannelsPerProcess {
		return ErrTooManyChannels
	}

	// Initialize log if nil
	if channel.Log == nil {
		channel.Log = make([]*MsgEntry, 0)
	}

	r.channels[channel.ID] = channel

	// Index by process
	r.byProcess[channel.ProcessID] = append(r.byProcess[channel.ProcessID], channel.ID)

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

	// Check channel count limit per process
	if len(r.byProcess[channel.ProcessID]) >= r.maxChannelsPerProcess {
		return ErrTooManyChannels
	}

	// Initialize log if nil
	if channel.Log == nil {
		channel.Log = make([]*MsgEntry, 0)
	}

	r.channels[channel.ID] = channel

	// Index by process
	r.byProcess[channel.ProcessID] = append(r.byProcess[channel.ProcessID], channel.ID)

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
	// Check message size before acquiring lock
	if len(payload) > constants.CHANNEL_MAX_MESSAGE_SIZE {
		return ErrMessageTooLarge
	}

	r.mu.Lock()
	channel, exists := r.channels[channelID]

	if !exists {
		r.mu.Unlock()
		return ErrChannelNotFound
	}

	// Check authorization (while holding lock to avoid race with SetExecutorIDForProcess)
	if err := r.authorize(channel, senderID, "append"); err != nil {
		r.mu.Unlock()
		return err
	}

	// Check rate limit (per process)
	r.rateLimitMu.RLock()
	rateLimitEnabled := r.rateLimitEnabled
	r.rateLimitMu.RUnlock()

	if rateLimitEnabled {
		limiter := r.getRateLimiter(channel.ProcessID)
		if !limiter.Allow() {
			r.mu.Unlock()
			return ErrRateLimitExceeded
		}
	}

	// Check channel log size limit
	if len(channel.Log) >= r.maxLogEntries {
		r.mu.Unlock()
		return ErrChannelFull
	}

	entry := &MsgEntry{
		Sequence:  sequence, // Client-assigned
		InReplyTo: inReplyTo,
		Timestamp: time.Now(),
		SenderID:  senderID,
		Payload:   payload,
		Type:      MsgTypeData,
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

	// Notify subscribers (push-based)
	r.notifySubscribers(channelID, entry)

	return nil
}

// AppendWithType adds a typed message to a channel (e.g., "end" for end-of-stream)
func (r *Router) AppendWithType(channelID string, senderID string, sequence int64, inReplyTo int64, payload []byte, msgType string) error {
	// Check message size before acquiring lock
	if len(payload) > constants.CHANNEL_MAX_MESSAGE_SIZE {
		return ErrMessageTooLarge
	}

	r.mu.Lock()
	channel, exists := r.channels[channelID]

	if !exists {
		r.mu.Unlock()
		return ErrChannelNotFound
	}

	// Check authorization (while holding lock to avoid race with SetExecutorIDForProcess)
	if err := r.authorize(channel, senderID, "append"); err != nil {
		r.mu.Unlock()
		return err
	}

	// Check rate limit (per process)
	r.rateLimitMu.RLock()
	rateLimitEnabled := r.rateLimitEnabled
	r.rateLimitMu.RUnlock()

	if rateLimitEnabled {
		limiter := r.getRateLimiter(channel.ProcessID)
		if !limiter.Allow() {
			r.mu.Unlock()
			return ErrRateLimitExceeded
		}
	}

	// Check channel log size limit
	if len(channel.Log) >= r.maxLogEntries {
		r.mu.Unlock()
		return ErrChannelFull
	}

	entry := &MsgEntry{
		Sequence:  sequence, // Client-assigned
		InReplyTo: inReplyTo,
		Timestamp: time.Now(),
		SenderID:  senderID,
		Payload:   payload,
		Type:      msgType,
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

	// Notify subscribers (push-based)
	r.notifySubscribers(channelID, entry)

	return nil
}

// ReadAfter reads entries after a given index (position in log)
// Since sequences are per-sender, we use index-based reading
// limit=0 means no limit
func (r *Router) ReadAfter(channelID string, callerID string, afterIndex int64, limit int) ([]*MsgEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channel, exists := r.channels[channelID]
	if !exists {
		return nil, ErrChannelNotFound
	}

	// Check authorization (while holding lock to avoid race with SetExecutorIDForProcess)
	if err := r.authorize(channel, callerID, "read"); err != nil {
		return nil, err
	}

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
func (r *Router) authorize(channel *Channel, callerID string, operation string) error {
	if callerID != channel.SubmitterID && callerID != channel.ExecutorID {
		log.WithFields(log.Fields{
			"channelID":   channel.ID,
			"channelName": channel.Name,
			"processID":   channel.ProcessID,
			"callerID":    callerID,
			"submitterID": channel.SubmitterID,
			"executorID":  channel.ExecutorID,
			"operation":   operation,
		}).Warn("Channel authorization failed")
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
			// Only close if not already closed (could be closed due to slow subscriber)
			if !sub.closed {
				close(sub.ch)
			}
		}
		delete(r.subscribers, id)
	}
	r.subMu.Unlock()

	// Clean up rate limiter for this process
	r.rateLimitMu.Lock()
	delete(r.rateLimiters, processID)
	r.rateLimitMu.Unlock()
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
	if !exists {
		r.mu.RUnlock()
		return nil, ErrChannelNotFound
	}

	// Verify authorization (while holding lock to avoid race with SetExecutorIDForProcess)
	if err := r.authorize(channel, callerID, "subscribe"); err != nil {
		r.mu.RUnlock()
		return nil, err
	}
	r.mu.RUnlock()

	ch := make(chan *MsgEntry, r.subscriberBufferSize)
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
			// Only close if not already closed (could be closed due to slow subscriber)
			if !sub.closed {
				close(ch)
			}
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
	var slowSubscribers []*Subscriber

	r.subMu.RLock()
	for _, sub := range r.subscribers[channelID] {
		if sub.closed {
			continue // Already disconnected
		}
		select {
		case sub.ch <- entry:
			// Successfully sent
		default:
			// Channel full, subscriber too slow - mark for disconnection
			log.WithFields(log.Fields{
				"channelID":  channelID,
				"bufferSize": r.subscriberBufferSize,
			}).Warn("Subscriber disconnected: buffer full, too slow to consume messages")
			sub.closed = true
			// Drain one message to make room for error message
			select {
			case <-sub.ch:
			default:
			}
			// Send error message so client knows why they were disconnected
			sub.ch <- &MsgEntry{Error: ErrSubscriberTooSlow.Error()}
			close(sub.ch)
			slowSubscribers = append(slowSubscribers, sub)
		}
	}
	r.subMu.RUnlock()

	// Clean up disconnected subscribers
	if len(slowSubscribers) > 0 {
		r.cleanupSlowSubscribers(channelID, slowSubscribers)
	}
}

// cleanupSlowSubscribers removes disconnected subscribers from the list
func (r *Router) cleanupSlowSubscribers(channelID string, toRemove []*Subscriber) {
	r.subMu.Lock()
	defer r.subMu.Unlock()

	subs := r.subscribers[channelID]
	remaining := make([]*Subscriber, 0, len(subs))
	for _, sub := range subs {
		found := false
		for _, remove := range toRemove {
			if sub == remove {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, sub)
		}
	}

	if len(remaining) == 0 {
		delete(r.subscribers, channelID)
	} else {
		r.subscribers[channelID] = remaining
	}
}

// SubscriberCount returns the number of subscribers for a channel (for testing)
func (r *Router) SubscriberCount(channelID string) int {
	r.subMu.RLock()
	defer r.subMu.RUnlock()
	return len(r.subscribers[channelID])
}
