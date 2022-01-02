package server

import (
	"time"

	"github.com/gorilla/websocket"
)

type processesSubscription struct {
	wsConn              *websocket.Conn
	wsMsgType           int
	subscriptionTimeout time.Time
	runtimeType         string
	state               int
}

func createProcessesSubscription(wsConn *websocket.Conn,
	wsMsgType int,
	runtimeType string,
	timeout int,
	state int) *processesSubscription {
	return &processesSubscription{wsConn: wsConn,
		wsMsgType:           wsMsgType,
		subscriptionTimeout: time.Now(),
		runtimeType:         runtimeType,
		state:               state}
}
