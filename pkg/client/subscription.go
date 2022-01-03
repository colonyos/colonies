package client

import (
	"colonies/pkg/core"

	"github.com/gorilla/websocket"
)

type ProcessSubscription struct {
	ProcessChan chan *core.Process
	ErrChan     chan error
	wsConn      *websocket.Conn
}

func createProcessSubscription(wsConn *websocket.Conn) *ProcessSubscription {
	subscription := &ProcessSubscription{}
	subscription.ProcessChan = make(chan *core.Process)
	subscription.ErrChan = make(chan error)
	subscription.wsConn = wsConn

	return subscription
}

func (subscription *ProcessSubscription) Close() error {
	return subscription.wsConn.Close()
}
