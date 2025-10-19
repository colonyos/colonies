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
