package gin

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/channel"
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

	// Get channel by process and name
	ch, err := h.server.ChannelRouter().GetByProcessAndName(msg.ProcessID, msg.Name)
	if err != nil {
		err := h.sendWSErrorMsg(errors.New("Channel not found"), http.StatusNotFound, wsConn, wsMsgType)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to channel, channel not found")
		}
		return
	}

	// Determine caller ID - either submitter or executor
	callerID := recoveredID
	if recoveredID == process.InitiatorID {
		callerID = process.InitiatorID
	} else if recoveredID == process.AssignedExecutorID {
		callerID = process.AssignedExecutorID
	}

	log.WithFields(log.Fields{"ProcessID": msg.ProcessID, "Channel": msg.Name, "CallerID": callerID}).Debug("WebSocket channel subscription started")

	// Long-poll loop: continuously check for new messages
	// AfterSeq is now used as index (position in log)
	lastIndex := msg.AfterSeq
	timeout := time.Duration(msg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 100 * time.Millisecond

	for time.Now().Before(deadline) {
		entries, err := h.server.ChannelRouter().ReadAfter(ch.ID, callerID, lastIndex, 0)
		if err != nil {
			if err == channel.ErrUnauthorized {
				h.sendWSErrorMsg(errors.New("Not authorized to read from channel"), http.StatusForbidden, wsConn, wsMsgType)
			} else {
				h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
			}
			return
		}

		if len(entries) > 0 {
			// Send entries to client
			jsonBytes, err := json.Marshal(entries)
			if err != nil {
				h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
				return
			}

			replyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeChannelPayloadType, string(jsonBytes))
			if err != nil {
				h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
				return
			}

			jsonString, err := replyMsg.ToJSON()
			if err != nil {
				h.sendWSErrorMsg(err, http.StatusInternalServerError, wsConn, wsMsgType)
				return
			}

			err = wsConn.WriteMessage(wsMsgType, []byte(jsonString))
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to send channel entries to WebSocket")
				return
			}

			// Update last index for next poll
			lastIndex += int64(len(entries))
		}

		time.Sleep(pollInterval)
	}

	// Send empty response on timeout
	replyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeChannelPayloadType, "[]")
	if err != nil {
		return
	}
	jsonString, err := replyMsg.ToJSON()
	if err != nil {
		return
	}
	wsConn.WriteMessage(wsMsgType, []byte(jsonString))
}