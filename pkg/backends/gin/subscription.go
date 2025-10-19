package gin

import (
	"context"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// SubscriptionController implements the backends.RealtimeSubscriptionController interface
type SubscriptionController struct {
	eventHandler backends.RealtimeEventHandler
}

// NewSubscriptionController creates a new WebSocket subscription controller
func NewSubscriptionController(eventHandler backends.RealtimeEventHandler) backends.RealtimeSubscriptionController {
	return &SubscriptionController{eventHandler: eventHandler}
}

func (ctrl *SubscriptionController) sendProcessToConnection(executorID string,
	process *core.Process,
	conn backends.RealtimeConnection,
	msgType int,
	cancel func()) error {
	
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to create Process JSON when subscribing to processes")
		cancel()
		return err
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
		return err
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
		return err
	}
	
	if conn == nil || !conn.IsOpen() {
		err = errors.New("connection is closed")
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Connection is closed when subscribing to processes")
		cancel()
		return err
	}
	
	err = conn.WriteMessage(msgType, []byte(rpcReplyJSONString))
	if err != nil {
		log.WithFields(log.Fields{
			"ExecutorID":   executorID,
			"ExecutorType": process.FunctionSpec.Conditions.ExecutorType,
			"State":        process.State,
			"Error":        err}).
			Error("Failed to write RPCReplyMsg JSON to connection when subscribing to processes")
		cancel()
		return err
	}
	
	return nil
}

func (ctrl *SubscriptionController) subscribe(executorID string, processID string, subscription *backends.RealtimeSubscription) {
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Duration(subscription.Timeout)*time.Second)
		defer cancelCtx()

		processChan, errChan := ctrl.eventHandler.Subscribe(subscription.ExecutorType, subscription.State, processID, ctx)
		for {
			select {
			case err := <-errChan:
				log.WithFields(log.Fields{
					"ExecutorID":   executorID,
					"ExecutorType": subscription.ExecutorType,
					"State":        subscription.State,
					"Error":        err}).
					Debug("Subscriber timed out")
				if subscription.Connection != nil && subscription.Connection.IsOpen() {
					subscription.Connection.Close()
				}
				return // This will kill the go-routine, also note all cancelCtx will result in an err to errChan
			case process := <-processChan:
				ctrl.sendProcessToConnection(executorID, process, subscription.Connection, subscription.MsgType, func() { cancelCtx() })
			}
		}
	}()
}

// AddProcessesSubscriber implements backends.RealtimeSubscriptionController
func (ctrl *SubscriptionController) AddProcessesSubscriber(executorID string, subscription *backends.RealtimeSubscription) error {
	ctrl.subscribe(executorID, "", subscription)
	return nil
}

// AddProcessSubscriber implements backends.RealtimeSubscriptionController
func (ctrl *SubscriptionController) AddProcessSubscriber(executorID string, process *core.Process, subscription *backends.RealtimeSubscription) error {
	ctrl.subscribe(executorID, process.ID, subscription)

	// Send an event immediately if process already have the state the subscriber is looking for
	// See unittest TestSubscribeChangeStateProcess2 for more info
	if process.State == subscription.State {
		ctrl.sendProcessToConnection(executorID, process, subscription.Connection, subscription.MsgType, func() {})
	}
	
	return nil
}

// Utility functions for backward compatibility with existing websocket code
func CreateSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, state int, timeout int) *backends.RealtimeSubscription {
	return &backends.RealtimeSubscription{
		Connection:   NewWebSocketConnection(wsConn),
		MsgType:      wsMsgType,
		Timeout:      timeout,
		ProcessID:    processID,
		ExecutorType: executorType,
		State:        state,
	}
}

func CreateProcessesSubscription(wsConn *websocket.Conn, wsMsgType int, executorType string, timeout int, state int) *backends.RealtimeSubscription {
	return &backends.RealtimeSubscription{
		Connection:   NewWebSocketConnection(wsConn),
		MsgType:      wsMsgType,
		Timeout:      timeout,
		ExecutorType: executorType,
		State:        state,
	}
}

func CreateProcessSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, executorType string, timeout int, state int) *backends.RealtimeSubscription {
	return &backends.RealtimeSubscription{
		Connection:   NewWebSocketConnection(wsConn),
		MsgType:      wsMsgType,
		Timeout:      timeout,
		ProcessID:    processID,
		ExecutorType: executorType,
		State:        state,
	}
}