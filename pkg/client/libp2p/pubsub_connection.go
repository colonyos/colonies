package libp2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"
)

// PubSubRealtimeConnection implements RealtimeConnection using libp2p pubsub
type PubSubRealtimeConnection struct {
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
	ctx          context.Context
	localPeerID  string // Our own peer ID to filter out our messages
	
	// Message handling
	messageQueue chan *pubsub.Message
	closed       bool
	closedLock   sync.RWMutex
}

// NewPubSubRealtimeConnection creates a new pubsub-based realtime connection
func NewPubSubRealtimeConnection(topic *pubsub.Topic, subscription *pubsub.Subscription, ctx context.Context, localPeerID string) *PubSubRealtimeConnection {
	conn := &PubSubRealtimeConnection{
		topic:        topic,
		subscription: subscription,
		ctx:          ctx,
		localPeerID:  localPeerID,
		messageQueue: make(chan *pubsub.Message, 100), // Buffer for messages
	}
	
	// Start message reader goroutine
	go conn.messageReader()
	
	return conn
}

// WriteMessage writes a message to the pubsub topic
func (p *PubSubRealtimeConnection) WriteMessage(messageType int, data []byte) error {
	p.closedLock.RLock()
	defer p.closedLock.RUnlock()
	
	if p.closed {
		return fmt.Errorf("connection is closed")
	}
	
	// For pubsub, messageType is ignored - we just publish the data
	err := p.topic.Publish(p.ctx, data)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish message to pubsub topic")
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	return nil
}

// ReadMessage reads a message from the pubsub subscription
func (p *PubSubRealtimeConnection) ReadMessage() (messageType int, data []byte, err error) {
	select {
	case msg := <-p.messageQueue:
		if msg == nil {
			return 0, nil, fmt.Errorf("connection closed")
		}
		// Return TextMessage type for compatibility with WebSocket interface
		return backends.TextMessage, msg.Data, nil
		
	case <-p.ctx.Done():
		return 0, nil, fmt.Errorf("context cancelled")
	}
}

// Close closes the pubsub realtime connection
func (p *PubSubRealtimeConnection) Close() error {
	p.closedLock.Lock()
	defer p.closedLock.Unlock()
	
	if p.closed {
		return nil
	}
	
	p.closed = true
	
	// Close subscription
	if p.subscription != nil {
		p.subscription.Cancel()
	}
	
	// Close topic
	if p.topic != nil {
		p.topic.Close()
	}
	
	// Close message queue
	close(p.messageQueue)
	
	logrus.Debug("PubSub realtime connection closed")
	return nil
}

// SetReadLimit sets the maximum size for incoming messages
func (p *PubSubRealtimeConnection) SetReadLimit(limit int64) {
	// In libp2p pubsub, message size limits are typically configured
	// at the pubsub level, not per-subscription
	logrus.WithField("limit", limit).Debug("SetReadLimit called (not implemented for pubsub)")
}

// messageReader reads messages from pubsub and forwards them to the queue
func (p *PubSubRealtimeConnection) messageReader() {
	defer func() {
		logrus.Debug("PubSub message reader stopped")
	}()
	
	for {
		msg, err := p.subscription.Next(p.ctx)
		if err != nil {
			if p.ctx.Err() != nil {
				// Context cancelled, normal shutdown
				return
			}
			logrus.WithError(err).Error("Failed to read message from pubsub subscription")
			continue
		}
		
		// Skip messages from ourselves
		if msg.ReceivedFrom.String() == p.localPeerID {
			continue
		}
		
		select {
		case p.messageQueue <- msg:
			// Message queued successfully
		case <-p.ctx.Done():
			return
		default:
			// Queue is full, drop oldest message
			select {
			case <-p.messageQueue:
				p.messageQueue <- msg
			default:
			}
		}
	}
}

// Compile-time check that PubSubRealtimeConnection implements RealtimeConnection
var _ backends.RealtimeConnection = (*PubSubRealtimeConnection)(nil)