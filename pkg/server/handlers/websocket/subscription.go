package websocket

import (
	"context"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	servercommunication "github.com/colonyos/colonies/pkg/server/websocket"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Subscription struct {
	WsConn       *websocket.Conn
	WsMsgType    int
	Timeout      int
	ExecutorType string
	State        int
	ProcessID    string
}

type WSSubscriptionController struct {
	eventHandler *servercommunication.EventHandler
}

// Used by ColoniesServer
func CreateSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, state int, timeout int) *Subscription {
	return &Subscription{WsConn: wsConn,
		WsMsgType:    wsMsgType,
		Timeout:      timeout,
		ProcessID:    processID,
		ExecutorType: executorType,
		State:        state}
}

// Used by ColoniesServer
func CreateProcessesSubscription(wsConn *websocket.Conn, wsMsgType int, executorType string, timeout int, state int) *Subscription {
	return &Subscription{WsConn: wsConn,
		WsMsgType:    wsMsgType,
		Timeout:      timeout,
		ExecutorType: executorType,
		State:        state}
}

func CreateProcessSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, timeout int, state int) *Subscription {
	return &Subscription{WsConn: wsConn,
		WsMsgType:    wsMsgType,
		Timeout:      timeout,
		ProcessID:    processID,
		ExecutorType: executorType,
		State:        state}
}

// Used by coloniesController
func CreateWSSubscriptionController(eventHandler *servercommunication.EventHandler) *WSSubscriptionController {
	wsSubCtrl := &WSSubscriptionController{}
	wsSubCtrl.eventHandler = eventHandler

	return wsSubCtrl
}

func (wsSubCtrl *WSSubscriptionController) sendProcessToWS(executorID string,
	process *core.Process,
	wsConn *websocket.Conn,
	wsMsgType int,
	cancel func()) {
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create Process JSON when subscribing to processes")
		cancel()
	}
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create RPCReplyMsg when subscribing to processes")
		cancel()
	}
	rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create RPCReplyMsg JSON when subscribing to processes")
		cancel()
	}
	if wsConn != nil {
		err = wsConn.WriteMessage(wsMsgType, []byte(rpcReplyJSONString))
	} else {
		err = errors.New("WebSocket connection is nil")
	}
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to write RPCReplyMsg JSON to WS when subscribing to processes")
		cancel()
	}
}

// Used by coloniesController
func (wsSubCtrl *WSSubscriptionController) Subscribe(executorID string, processID string, subscription *Subscription) {
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Duration(subscription.Timeout)*time.Second)
		defer cancelCtx()

		processChan, errChan := wsSubCtrl.eventHandler.Subscribe(subscription.ExecutorType, subscription.State, processID, ctx)
		for {
			select {
			case err := <-errChan:
				log.WithFields(log.Fields{
					"ExecutorID":   executorID,
					"ExecutorType": subscription.ExecutorType,
					"State":        subscription.State,
					"Error":        err}).
					Debug("Subscriber timed out")
				if subscription.WsConn != nil {
					subscription.WsConn.Close()
				}
				return // This will kill the go-routine, also note all cancelCtx will result in an err to errChan
			case process := <-processChan:
				wsSubCtrl.sendProcessToWS(executorID, process, subscription.WsConn, subscription.WsMsgType, func() { cancelCtx() })
			}
		}
	}()
}

func (wsSubCtrl *WSSubscriptionController) AddProcessesSubscriber(executorID string, subscription *Subscription) {
	wsSubCtrl.Subscribe(executorID, "", subscription)
}

func (wsSubCtrl *WSSubscriptionController) AddProcessSubscriber(executorID string, process *core.Process, subscription *Subscription) {
	wsSubCtrl.Subscribe(executorID, process.ID, subscription)

	// Send an event immediately if process already have the state the subscriber is looking for
	// See unittest TestSubscribeChangeStateProcess2 for more info
	if process.State == subscription.State {
		wsSubCtrl.sendProcessToWS(executorID, process, subscription.WsConn, subscription.WsMsgType, func() {})
	}
}