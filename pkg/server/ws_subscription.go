package server

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type processesSubscription struct {
	wsConn              *websocket.Conn
	wsMsgType           int
	subscriptionTimeout time.Time
	runtimeType         string
	state               int
}

type processSubscription struct {
	wsConn              *websocket.Conn
	wsMsgType           int
	subscriptionTimeout time.Time
	processID           string
	state               int
}

type wsSubscriptionController struct {
	processesSubscribers map[string]*processesSubscription
	processSubscribers   map[string]*processSubscription
}

// Used by ColoniesServer
func createProcessSubscription(wsConn *websocket.Conn, wsMsgType int, processID string, timeout int, state int) *processSubscription {
	return &processSubscription{wsConn: wsConn,
		wsMsgType:           wsMsgType,
		subscriptionTimeout: time.Now(),
		processID:           processID,
		state:               state}
}

// Used by ColoniesServer
func createProcessesSubscription(wsConn *websocket.Conn, wsMsgType int, runtimeType string, timeout int, state int) *processesSubscription {
	return &processesSubscription{wsConn: wsConn,
		wsMsgType:           wsMsgType,
		subscriptionTimeout: time.Now(),
		runtimeType:         runtimeType,
		state:               state}
}

// Used by coloniesController
func createWSSubscriptionController() *wsSubscriptionController {
	wsSubCtrl := &wsSubscriptionController{}
	wsSubCtrl.processesSubscribers = make(map[string]*processesSubscription)
	wsSubCtrl.processSubscribers = make(map[string]*processSubscription)

	return wsSubCtrl
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) addProcessesSubscriber(runtimeID string, subscription *processesSubscription) {
	wsSubCtrl.processesSubscribers[runtimeID] = subscription
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) addProcessSubscriber(runtimeID string, subscription *processSubscription, process *core.Process) {
	wsSubCtrl.processSubscribers[runtimeID] = subscription

	// Send an event immediately if process already have the state the subscriber is looking for
	// See unittest TestSubscribeChangeStateProcess2 for more info
	if process.State == subscription.state {
		wsSubCtrl.wsWriteProcessChangeEvent(process, runtimeID, subscription)
	}
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) sendProcessesEvent(process *core.Process) {
	for runtimeID, subscription := range wsSubCtrl.processesSubscribers {
		if subscription.runtimeType == process.ProcessSpec.Conditions.RuntimeType && subscription.state == process.State {
			jsonString, err := process.ToJSON()
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to parse JSON when removing processes event subscription")
				delete(wsSubCtrl.processesSubscribers, runtimeID)
			}
			rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to create RPC reply message when removing processes event subscription")
				delete(wsSubCtrl.processSubscribers, runtimeID)
			}

			rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to generate JSON when removing processes event subcription")
				delete(wsSubCtrl.processSubscribers, runtimeID)
			}
			err = subscription.wsConn.WriteMessage(subscription.wsMsgType, []byte(rpcReplyJSONString))
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Removing processes event subcription")
				delete(wsSubCtrl.processesSubscribers, runtimeID)
			}
		}
	}
}

// Used by coloniesController
func (wsSubCtrl *wsSubscriptionController) sendProcessChangeStateEvent(process *core.Process) {
	for runtimeID, subscription := range wsSubCtrl.processSubscribers {
		if subscription.processID == process.ID && subscription.state == process.State {
			wsSubCtrl.wsWriteProcessChangeEvent(process, runtimeID, subscription)
		}
	}
}

// Used by subscriptionController internally
func (wsSubCtrl *wsSubscriptionController) wsWriteProcessChangeEvent(process *core.Process, runtimeID string, subscription *processSubscription) {
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to parse JSON when removing process event subscription")
		delete(wsSubCtrl.processSubscribers, runtimeID)
	}

	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to create RPC reply message when removing process event subscription")
		delete(wsSubCtrl.processSubscribers, runtimeID)
	}

	rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to generate JSON when removing process event subcription")
		delete(wsSubCtrl.processSubscribers, runtimeID)
	}

	err = subscription.wsConn.WriteMessage(subscription.wsMsgType, []byte(rpcReplyJSONString))
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Removing process event subcription")
		delete(wsSubCtrl.processesSubscribers, runtimeID)
	}
}
