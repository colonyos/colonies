package websocket

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer interface {
	HandleHTTPError(c *gin.Context, err error, errorCode int) bool
	SendHTTPReply(c *gin.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c *gin.Context, payloadType string)
	Validator() security.Validator
	ProcessDB() database.ProcessDatabase
	WSController() WSController
	ParseSignature(payload string, signature string) (string, error)
	GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error)
}

type WSController interface {
	SubscribeProcesses(executorID string, subscription *Subscription) error
	SubscribeProcess(executorID string, subscription *Subscription) error
}

type Handlers struct {
	server ColoniesServer
}

func NewHandlers(server ColoniesServer) *Handlers {
	return &Handlers{server: server}
}

func (h *Handlers) sendWSErrorMsg(err error, errorCode int, wsConn *websocket.Conn, wsMsgType int) error {
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

func (h *Handlers) HandleWSRequest(c *gin.Context) {
	w := c.Writer
	r := c.Request
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

			// This test is strictly not needed, since the request does not specifiy a colony, but is rather derived from the database
			err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
			if err != nil {
				err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
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

		case rpc.SubscribeProcessPayloadType:
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

			err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
			if err != nil {
				log.Error(err)
				err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, not member of colony")
				}
				return
			}
			process, err := h.server.ProcessDB().GetProcessByID(msg.ProcessID)
			if err != nil {
				err := h.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			if process == nil {
				err := h.sendWSErrorMsg(errors.New("Failed to subscribe to process, process not found"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			if process.FunctionSpec.Conditions.ColonyName != msg.ColonyName {
				err := h.sendWSErrorMsg(errors.New("Failed to subscribe to process, process does not belong to colony"), http.StatusForbidden, wsConn, wsMsgType)
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
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process")
				}
				return
			}
		}
	}
}