package server

import (
	"errors"
	"net/http"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func (server *ColoniesServer) sendWSErrorMsg(err error, errorCode int, wsConn *websocket.Conn, wsMsgType int) error {
	rpcErrorReplyMSg, err := server.generateRPCErrorMsg(err, errorCode)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call server.generateRPCErrorMsg()")
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

func (server *ColoniesServer) handleWSRequest(c *gin.Context) {
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
			return
		}

		rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(data))
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}

		recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
		if server.handleHTTPError(c, err, http.StatusForbidden) {
			return
		}

		switch rpcMsg.PayloadType {
		case rpc.SubscribeProcessesPayloadType:
			msg, err := rpc.CreateSubscribeProcessesMsgFromJSON(rpcMsg.DecodePayload())
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
				}
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				errMsg := "Failed to subscribe to processes, msg.msgType does not match rpcMsg.PayloadType"
				err := server.sendWSErrorMsg(errors.New(errMsg), http.StatusForbidden, wsConn, wsMsgType)
				log.Error(errMsg)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			executor, err := server.controller.getExecutor(recoveredID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
				}
				return
			}
			if executor == nil {
				err := server.sendWSErrorMsg(errors.New("Failed to subscribe to processes, executor not found"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			// This test is strictly not needed, since the request does not specifiy a colony, but is rather derived from the database
			err = server.validator.RequireExecutorMembership(recoveredID, executor.ColonyID, true)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to processes, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			processSubcription := createProcessesSubscription(wsConn, wsMsgType, msg.ExecutorType, msg.Timeout, msg.State)
			server.controller.subscribeProcesses(recoveredID, processSubcription)

		case rpc.SubscribeProcessPayloadType:
			msg, err := rpc.CreateSubscribeProcessMsgFromJSON(rpcMsg.DecodePayload())
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				err := server.sendWSErrorMsg(errors.New("Failed to subscribe to process, msg.msgType does not match rpcMsg.PayloadType"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			executor, err := server.controller.getExecutor(recoveredID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, Failed to call server.sendWSErrorMsg()")
				}
				return
			}
			if executor == nil {
				err := server.sendWSErrorMsg(errors.New("Failed to subscribe to process, executor not found"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			// This test is strictly not needed, since the request does not specifiy a colony, but is rather
			// derived from the database
			err = server.validator.RequireExecutorMembership(recoveredID, executor.ColonyID, true)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to subscribe to process, failed to call server.sendWSErrorMsg()")
				}
				return
			}

			processSubcription := createProcessSubscription(wsConn, wsMsgType, msg.ProcessID, msg.ExecutorType, msg.Timeout, msg.State)
			server.controller.subscribeProcess(recoveredID, processSubcription)
		}
	}
}
