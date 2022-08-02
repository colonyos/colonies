package server

import (
	"context"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type subscription struct {
	wsConn      *websocket.Conn
	wsMsgType   int
	timeout     int
	runtimeType string
	state       int
	processID   string
}

type wsSubscriptionController struct {
	eventHandler *eventHandler
}

// Used by ColoniesServer
func createSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, runtimeType string, state int, timeout int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:   wsMsgType,
		timeout:     timeout,
		processID:   processID,
		runtimeType: runtimeType,
		state:       state}
}

// Used by ColoniesServer
func createProcessesSubscription(wsConn *websocket.Conn, wsMsgType int, runtimeType string, timeout int, state int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:   wsMsgType,
		timeout:     timeout,
		runtimeType: runtimeType,
		state:       state}
}

func createProcessSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, runtimeType string, timeout int, state int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:   wsMsgType,
		timeout:     timeout,
		processID:   processID,
		runtimeType: runtimeType,
		state:       state}
}

// Used by coloniesController
func createWSSubscriptionController(eventHandler *eventHandler) *wsSubscriptionController {
	wsSubCtrl := &wsSubscriptionController{}
	wsSubCtrl.eventHandler = eventHandler

	return wsSubCtrl
}

func (wsSubCtrl *wsSubscriptionController) sendProcessToWS(runtimeID string,
	process *core.Process,
	wsConn *websocket.Conn,
	wsMsgType int,
	cancel func()) {
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"RuntimeID":   runtimeID,
			"RuntimeType": process.ProcessSpec.Conditions.RuntimeType,
			"State":       process.State,
			"Err":         err}).
			Error("Failed to create Process JSON when subscribing to processes")
		cancel()
	}
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
	if err != nil {
		log.WithFields(log.Fields{
			"RuntimeID":   runtimeID,
			"RuntimeType": process.ProcessSpec.Conditions.RuntimeType,
			"State":       process.State,
			"Err":         err}).
			Error("Failed to create RPCReplyMsg when subscribing to processes")
		cancel()
	}
	rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"RuntimeID":   runtimeID,
			"RuntimeType": process.ProcessSpec.Conditions.RuntimeType,
			"State":       process.State,
			"Err":         err}).
			Error("Failed to create RPCReplyMsg JSON when subscribing to processes")
		cancel()
	}
	err = wsConn.WriteMessage(wsMsgType, []byte(rpcReplyJSONString))
	if err != nil {
		log.WithFields(log.Fields{
			"RuntimeID":   runtimeID,
			"RuntimeType": process.ProcessSpec.Conditions.RuntimeType,
			"State":       process.State,
			"Err":         err}).
			Error("Failed to write RPCReplyMsg JSON to WS when subscribing to processes")
		cancel()
	}
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) subscribe(runtimeID string, processID string, subscription *subscription) {
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Duration(subscription.timeout)*time.Second)
		defer cancelCtx()

		processChan, errChan := wsSubCtrl.eventHandler.subscribe(subscription.runtimeType, subscription.state, processID, ctx)
		for {
			select {
			case err := <-errChan:
				log.WithFields(log.Fields{
					"RuntimeID":   runtimeID,
					"RuntimeType": subscription.runtimeType,
					"State":       subscription.state,
					"Err":         err}).
					Error("Failed to subscribe to processes")
				return // This will kill the go-routine, also note all cancelCtx will result in an err to errChan
			case process := <-processChan:
				wsSubCtrl.sendProcessToWS(runtimeID, process, subscription.wsConn, subscription.wsMsgType, func() { cancelCtx() })
			}
		}
	}()
}

func (wsSubCtrl *wsSubscriptionController) addProcessesSubscriber(runtimeID string, subscription *subscription) {
	wsSubCtrl.subscribe(runtimeID, "", subscription)
}

func (wsSubCtrl *wsSubscriptionController) addProcessSubscriber(runtimeID string, process *core.Process, subscription *subscription) {
	wsSubCtrl.subscribe(runtimeID, process.ID, subscription)

	// Send an event immediately if process already have the state the subscriber is looking for
	// See unittest TestSubscribeChangeStateProcess2 for more info
	if process.State == subscription.state {
		wsSubCtrl.sendProcessToWS(runtimeID, process, subscription.wsConn, subscription.wsMsgType, func() {})
	}
}
