package gin

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// RealtimeServer interface for servers that can handle realtime connections
type RealtimeServer interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	ParseSignature(payload string, signature string) (string, error)
	GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error)
	WSController() WSController
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