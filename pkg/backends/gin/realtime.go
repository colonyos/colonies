package gin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// RealtimeServer interface for servers that can handle realtime connections
type RealtimeServer interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	ParseSignature(payload string, signature string) (string, error)
	GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error)
	WSController() WSController
	ChannelRouter() *channel.Router
	ProcessDB() database.ProcessDatabase
	Validator() security.Validator
}

// WSController interface for WebSocket handlers
type WSController interface {
	SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error
	SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error
}

// RealtimeHandler handles WebSocket connections for gin backend
type RealtimeHandler struct {
	server RealtimeServer
}

// NewRealtimeHandler creates a new realtime handler for gin backend
func NewRealtimeHandler(server RealtimeServer) *RealtimeHandler {
	return &RealtimeHandler{server: server}
}

func (h *RealtimeHandler) sendWSErrorMsg(err error, errorCode int, wsConn *websocket.Conn, wsMsgType int) error {
	rpcErrorReplyMSg, err := h.server.GenerateRPCErrorMsg(err, errorCode)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call server.GenerateRPCErrorMsg()")
		return err
	}

	jsonString, err := rpcErrorReplyMSg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call rpcErrorReplyMSg.ToJSON()")
		return err
	}

	err = wsConn.WriteMessage(wsMsgType, []byte(jsonString))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call wsConn.WriteMessage()")
		return err
	}

	return nil
}

// HandleWSRequest handles WebSocket upgrade and message processing for Gin
func (h *RealtimeHandler) HandleWSRequest(c backends.Context) {
	// For WebSocket handling, we need to access the underlying gin context
	// Cast to ContextAdapter to get the underlying gin.Context
	ginAdapter, ok := c.(*ContextAdapter)
	if !ok {
		log.Error("WebSocket handler requires gin context adapter")
		return
	}
	ginCtx := ginAdapter.GinContext()
	w := ginCtx.Writer
	r := ginCtx.Request
	
	var wsupgrader = websocket.Upgrader{}
	wsupgrader.CheckOrigin = func(r *http.Request) bool { return true } // TODO: Insecure
	var err error
	var wsConn *websocket.Conn
	wsConn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call wsupgrader.Upgrade()")
		return
	}

	for {
		wsMsgType, data, err := wsConn.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}

		rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(data))
		if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
			return
		}

		recoveredID, err := h.server.ParseSignature(rpcMsg.Payload, rpcMsg.Signature)
		if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
			return
		}

		switch rpcMsg.PayloadType {
		case rpc.SubscribeProcessesPayloadType:
			h.handleSubscribeProcesses(c, rpcMsg, recoveredID, wsConn, wsMsgType)
		case rpc.SubscribeProcessPayloadType:
			h.handleSubscribeProcess(c, rpcMsg, recoveredID, wsConn, wsMsgType)
		case rpc.SubscribeChannelPayloadType:
			h.handleSubscribeChannel(c, rpcMsg, recoveredID, wsConn, wsMsgType)
		}
	}
}

func (h *RealtimeHandler) handleSubscribeProcesses(c backends.Context, rpcMsg *rpc.RPCMsg, recoveredID string, wsConn *websocket.Conn, wsMsgType int) {
	msg, err := rpc.CreateSubscribeProcessesMsgFromJSON(rpcMsg.DecodePayload())
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
		}
		return
	}
	if msg.MsgType != rpcMsg.PayloadType {
		errMsg := "Failed to subscribe to processes, msg.msgType does not match rpcMsg.PayloadType"
		err := h.sendWSErrorMsg(errors.New(errMsg), http.StatusForbidden, wsConn, wsMsgType)
		log.Error(errMsg)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
		}
		return
	}

	processSubcription := CreateProcessesSubscription(wsConn, wsMsgType, msg.ExecutorType, msg.Timeout, msg.State)
	err = h.server.WSController().SubscribeProcesses(recoveredID, processSubcription)
	if err != nil {
		err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes")
		}
		return
	}
}

func (h *RealtimeHandler) handleSubscribeProcess(c backends.Context, rpcMsg *rpc.RPCMsg, recoveredID string, wsConn *websocket.Conn, wsMsgType int) {
	msg, err := rpc.CreateSubscribeProcessMsgFromJSON(rpcMsg.DecodePayload())
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg.MsgType != rpcMsg.PayloadType {
		err := h.sendWSErrorMsg(errors.New("Failed to subscribe to process, msg.msgType does not match rpcMsg.PayloadType"), http.StatusForbidden, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
		}
		return
	}

	processSubcription := CreateProcessSubscription(wsConn, wsMsgType, msg.ProcessID, msg.ExecutorType, msg.Timeout, msg.State)
	err = h.server.WSController().SubscribeProcess(recoveredID, processSubcription)
	if err != nil {
		err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
		}
		return
	}
}

func (h *RealtimeHandler) handleSubscribeChannel(c backends.Context, rpcMsg *rpc.RPCMsg, recoveredID string, wsConn *websocket.Conn, wsMsgType int) {
	msg, err := rpc.CreateSubscribeChannelMsgFromJSON(rpcMsg.DecodePayload())
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		err := h.sendWSErrorMsg(err, http.StatusBadRequest, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, failed to send error message")
		}
		return
	}

	if msg.MsgType != rpcMsg.PayloadType {
		err := h.sendWSErrorMsg(errors.New("Failed to subscribe to channel, msg.MsgType does not match rpcMsg.PayloadType"), http.StatusBadRequest, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, failed to send error message")
		}
		return
	}

	// Get the process to verify membership
	process, err := h.server.ProcessDB().GetProcessByID(msg.ProcessID)
	if err != nil || process == nil {
		err := h.sendWSErrorMsg(errors.New("Process not found"), http.StatusNotFound, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, process not found")
		}
		return
	}

	// Verify colony membership
	err = h.server.Validator().RequireMembership(recoveredID, process.FunctionSpec.Conditions.ColonyName, true)
	if err != nil {
		err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, membership check failed")
		}
		return
	}

	// Get channel by process and name, creating on demand if necessary for cluster scenarios
	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name}).Info("Looking up channel for WebSocket subscription")
	ch, err := h.server.ChannelRouter().GetByProcessAndName(msg.ProcessID, msg.Name)
	if err != nil {
		log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "Error": err, "ErrorType": fmt.Sprintf("%T", err)}).Info("Channel lookup failed, attempting lazy creation")
		if errors.Is(err, channel.ErrChannelNotFound) {
			// Try lazy creation - channel may not have replicated to this server yet
			ch, err = h.ensureChannelExists(process, msg.Name)
			if err != nil {
				err := h.sendWSErrorMsg(errors.New("Channel not found"), http.StatusNotFound, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, channel not found")
				}
				return
			}
			log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name}).Info("Created channel on demand for WebSocket subscription (cluster lazy creation)")
		} else {
			err := h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel")
			}
			return
		}
	}

	// Determine caller ID - either submitter or executor
	callerID := recoveredID
	if recoveredID == process.InitiatorID {
		callerID = process.InitiatorID
	} else if recoveredID == process.AssignedExecutorID {
		callerID = process.AssignedExecutorID
	}

	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "CallerID": callerID, "Timeout": msg.Timeout}).Info("WebSocket channel subscription started (push-based)")

	// Subscribe to push notifications
	entryChan, err := h.server.ChannelRouter().Subscribe(ch.ID, callerID)
	if err != nil {
		if err == channel.ErrUnauthorized {
			h.sendWSErrorMsg(errors.New("Not authorized to subscribe to channel"), http.StatusForbidden, wsConn, wsMsgType)
		} else {
			h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
		}
		return
	}
	defer h.server.ChannelRouter().Unsubscribe(ch.ID, entryChan)

	// First, send any existing entries after the requested index
	lastIndex := msg.AfterSeq
	existingEntries, err := h.server.ChannelRouter().ReadAfter(ch.ID, callerID, lastIndex, 0)
	if err == nil && len(existingEntries) > 0 {
		if err := h.sendChannelEntries(existingEntries, wsConn, wsMsgType); err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to send existing channel entries")
			return
		}
		lastIndex += int64(len(existingEntries))
	}

	// Set up timeout
	timeout := time.Duration(msg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Push-based streaming loop
	for {
		select {
		case entry, ok := <-entryChan:
			if !ok {
				// Channel closed (unsubscribed)
				return
			}

			// Send entry immediately to WebSocket
			if err := h.sendChannelEntries([]*channel.MsgEntry{entry}, wsConn, wsMsgType); err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to send channel entry to WebSocket")
				return
			}

		case <-timer.C:
			// Timeout - send empty response and close
			log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name}).Debug("WebSocket channel subscription timeout")
			replyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeChannelPayloadType, "[]")
			if err != nil {
				return
			}
			jsonString, err := replyMsg.ToJSON()
			if err != nil {
				return
			}
			wsConn.WriteMessage(wsMsgType, []byte(jsonString))
			return
		}
	}
}

// ensureChannelExists creates a channel on demand if it's defined in the process spec
// but doesn't exist locally. This handles cluster scenarios where a client connects
// to a different server than where the process was originally submitted.
func (h *RealtimeHandler) ensureChannelExists(process *core.Process, channelName string) (*channel.Channel, error) {
	// Don't create channels for closed processes (SUCCESS or FAILED)
	// Channels are cleaned up when processes close
	if process.State == core.SUCCESS || process.State == core.FAILED {
		return nil, channel.ErrChannelNotFound
	}

	// Check if this channel is defined in the process spec
	channelDefined := false
	for _, ch := range process.FunctionSpec.Channels {
		if ch == channelName {
			channelDefined = true
			break
		}
	}

	if !channelDefined {
		return nil, channel.ErrChannelNotFound
	}

	// Create the channel on demand
	ch := &channel.Channel{
		ID:          process.ID + "_" + channelName, // Deterministic ID
		ProcessID:   process.ID,
		Name:        channelName,
		SubmitterID: process.InitiatorID,
		ExecutorID:  process.AssignedExecutorID,
	}

	// Use CreateIfNotExists to handle concurrent creation
	if err := h.server.ChannelRouter().CreateIfNotExists(ch); err != nil {
		return nil, err
	}

	// Return the channel (might have been created by another goroutine)
	return h.server.ChannelRouter().GetByProcessAndName(process.ID, channelName)
}

// sendChannelEntries sends channel entries to a WebSocket connection
func (h *RealtimeHandler) sendChannelEntries(entries []*channel.MsgEntry, wsConn *websocket.Conn, wsMsgType int) error {
	jsonBytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	replyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeChannelPayloadType, string(jsonBytes))
	if err != nil {
		return err
	}

	jsonString, err := replyMsg.ToJSON()
	if err != nil {
		return err
	}

	return wsConn.WriteMessage(wsMsgType, []byte(jsonString))
}