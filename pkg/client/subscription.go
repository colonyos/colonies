package client

import (
	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
)

type ProcessSubscription struct {
	ProcessChan chan *core.Process
	ErrChan     chan error
	conn        backends.RealtimeConnection
}

func createProcessSubscription(conn backends.RealtimeConnection) *ProcessSubscription {
	subscription := &ProcessSubscription{}
	subscription.ProcessChan = make(chan *core.Process)
	subscription.ErrChan = make(chan error)
	subscription.conn = conn

	return subscription
}

func (subscription *ProcessSubscription) Close() error {
	return subscription.conn.Close()
}

// FileSubscription is the client-side handle to a SubscribeFiles
// websocket. Drain EventChan for normal events; ErrChan carries
// connection errors and overflow signals (backends.ErrSubscriberOverflowed
// is delivered as an error with that exact message). Close releases the
// underlying websocket and any goroutines reading from it.
type FileSubscription struct {
	EventChan chan *core.FileEvent
	ErrChan   chan error
	conn      backends.RealtimeConnection
}

func createFileSubscription(conn backends.RealtimeConnection) *FileSubscription {
	return &FileSubscription{
		EventChan: make(chan *core.FileEvent),
		ErrChan:   make(chan error),
		conn:      conn,
	}
}

func (subscription *FileSubscription) Close() error {
	return subscription.conn.Close()
}
