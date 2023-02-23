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
	wsConn       *websocket.Conn
	wsMsgType    int
	timeout      int
	executorType string
	state        int
	processID    string
}

type wsSubscriptionController struct {
	eventHandler *eventHandler
}

// Used by ColoniesServer
func createSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, state int, timeout int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:    wsMsgType,
		timeout:      timeout,
		processID:    processID,
		executorType: executorType,
		state:        state}
}

// Used by ColoniesServer
func createProcessesSubscription(wsConn *websocket.Conn, wsMsgType int, executorType string, timeout int, state int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:    wsMsgType,
		timeout:      timeout,
		executorType: executorType,
		state:        state}
}

func createProcessSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, timeout int, state int) *subscription {
	return &subscription{wsConn: wsConn,
		wsMsgType:    wsMsgType,
		timeout:      timeout,
		processID:    processID,
		executorType: executorType,
		state:        state}
}

// Used by coloniesController
func createWSSubscriptionController(eventHandler *eventHandler) *wsSubscriptionController {
	wsSubCtrl := &wsSubscriptionController{}
	wsSubCtrl.eventHandler = eventHandler

	return wsSubCtrl
}

func (wsSubCtrl *wsSubscriptionController) sendProcessToWS(executorID string,
	process *core.Process,
	wsConn *websocket.Conn,
	wsMsgType int,
	cancel func()) {
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.ProcessSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create Process JSON when subscribing to processes")
		cancel()
	}
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.ProcessSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create RPCReplyMsg when subscribing to processes")
		cancel()
	}
	rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.ProcessSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create RPCReplyMsg JSON when subscribing to processes")
		cancel()
	}
	err = wsConn.WriteMessage(wsMsgType, []byte(rpcReplyJSONString))
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.ProcessSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to write RPCReplyMsg JSON to WS when subscribing to processes")
		cancel()
	}
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) subscribe(executorID string, processID string, subscription *subscription) {
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Duration(subscription.timeout)*time.Second)
		defer cancelCtx()

		processChan, errChan := wsSubCtrl.eventHandler.subscribe(subscription.executorType, subscription.state, processID, ctx)
		for {
			select {
			case err := <-errChan:
				log.WithFields(log.Fields{
					"ExecutorID":   executorID,
					"ExecutorType": subscription.executorType,
					"State":        subscription.state,
					"Error":        err}).
					Debug("Subscriber timed out")
				subscription.wsConn.Close()
				return // This will kill the go-routine, also note all cancelCtx will result in an err to errChan
			case process := <-processChan:
				wsSubCtrl.sendProcessToWS(executorID, process, subscription.wsConn, subscription.wsMsgType, func() { cancelCtx() })
			}
		}
	}()
}

func (wsSubCtrl *wsSubscriptionController) addProcessesSubscriber(executorID string, subscription *subscription) {
	wsSubCtrl.subscribe(executorID, "", subscription)
}

func (wsSubCtrl *wsSubscriptionController) addProcessSubscriber(executorID string, process *core.Process, subscription *subscription) {
	wsSubCtrl.subscribe(executorID, process.ID, subscription)

	// Send an event immediately if process already have the state the subscriber is looking for
	// See unittest TestSubscribeChangeStateProcess2 for more info
	if process.State == subscription.state {
		wsSubCtrl.sendProcessToWS(executorID, process, subscription.wsConn, subscription.wsMsgType, func() {})
	}
}
