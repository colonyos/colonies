package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/colonyos/colonies/internal/logging"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/security/validator"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ColoniesServer struct {
	ginHandler        *gin.Engine
	controller        *coloniesController
	serverID          string
	tlsPrivateKeyPath string
	tlsCertPath       string
	port              int
	httpServer        *http.Server
	crypto            security.Crypto
	validator         security.Validator
}

func CreateColoniesServer(db database.Database, port int, serverID string, tlsPrivateKeyPath string, tlsCertPath string, debug bool) *ColoniesServer {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = ioutil.Discard
		logging.DisableDebug()
	}

	server := &ColoniesServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer
	server.controller = createColoniesController(db)
	server.serverID = serverID
	server.port = port
	server.tlsPrivateKeyPath = tlsPrivateKeyPath
	server.tlsCertPath = tlsCertPath
	server.crypto = crypto.CreateCrypto()
	server.validator = validator.CreateValidator(db)

	server.setupRoutes()

	logging.Log().Info("Starting Colonies API server at port: " + strconv.Itoa(port))

	return server
}

func (server *ColoniesServer) setupRoutes() {
	server.ginHandler.POST("/api", server.handleEndpointRequest)
	server.ginHandler.GET("/pubsub", server.handleWSRequest)
}

func (server *ColoniesServer) sendWSErrorMsg(err error, errorCode int, wsConn *websocket.Conn, wsMsgType int) error {
	rpcErrorReplyMSg, err := server.generateRPCErrorMsg(err, errorCode)
	if err != nil {
		return err
	}

	jsonString, err := rpcErrorReplyMSg.ToJSON()
	if err != nil {
		return err
	}

	err = wsConn.WriteMessage(wsMsgType, []byte(jsonString))
	if err != nil {
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
		fmt.Println(err)
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
					logging.Log().Error(err)
				}
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				err := server.sendWSErrorMsg(errors.New("msg.msgType does not match rpcMsg.PayloadType"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			runtime, err := server.controller.getRuntimeByID(recoveredID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}
			if runtime == nil {
				err := server.sendWSErrorMsg(errors.New("runtime not found"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			// This test is strictly not needed, since the request does not specifiy a colony, but is rather
			// derived from the database
			err = server.validator.RequireRuntimeMembership(recoveredID, runtime.ColonyID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			processSubcription := createProcessesSubscription(wsConn, wsMsgType, msg.RuntimeType, msg.Timeout, msg.State)
			server.controller.subscribeProcesses(recoveredID, processSubcription)

		case rpc.SubscribeProcessPayloadType:
			msg, err := rpc.CreateSubscribeProcessMsgFromJSON(rpcMsg.DecodePayload())
			if server.handleHTTPError(c, err, http.StatusBadRequest) {
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				err := server.sendWSErrorMsg(errors.New("msg.msgType does not match rpcMsg.PayloadType"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			runtime, err := server.controller.getRuntimeByID(recoveredID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}
			if runtime == nil {
				err := server.sendWSErrorMsg(errors.New("runtime not found"), http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			// This test is strictly not needed, since the request does not specifiy a colony, but is rather
			// derived from the database
			err = server.validator.RequireRuntimeMembership(recoveredID, runtime.ColonyID)
			if err != nil {
				err := server.sendWSErrorMsg(err, http.StatusForbidden, wsConn, wsMsgType)
				if err != nil {
					logging.Log().Error(err)
				}
				return
			}

			processSubcription := createProcessSubscription(wsConn, wsMsgType, msg.ProcessID, msg.Timeout, msg.State)
			server.controller.subscribeProcess(recoveredID, processSubcription)
		}
	}
}

func (server *ColoniesServer) parseSignature(jsonString string, signature string) (string, error) {
	recoveredID, err := server.crypto.RecoverID(jsonString, signature)
	if err != nil {
		return "", err
	}

	return recoveredID, nil
}

func (server *ColoniesServer) handleEndpointRequest(c *gin.Context) {
	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	switch rpcMsg.PayloadType {
	// Colony operations
	case rpc.AddColonyPayloadType:
		server.handleAddColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteColonyPayloadType:
		server.handleDeleteColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColoniesPayloadType:
		server.handleGetColoniesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyPayloadType:
		server.handleGetColonyHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Runtime operations
	case rpc.AddRuntimePayloadType:
		server.handleAddRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimesPayloadType:
		server.handleGetRuntimesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimePayloadType:
		server.handleGetRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ApproveRuntimePayloadType:
		server.handleApproveRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RejectRuntimePayloadType:
		server.handleRejectRuntimeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Process operations
	case rpc.SubmitProcessSpecPayloadType:
		server.handleSubmitProcessSpecHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.AssignProcessPayloadType:
		server.handleAssignProcessHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessesPayloadType:
		server.handleGetProcessesHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessPayloadType:
		server.handleGetProcessHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseSuccessfulPayloadType:
		server.handleCloseSuccessfulHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.CloseFailedPayloadType:
		server.handleCloseFailedHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Attribute operations
	case rpc.AddAttributePayloadType:
		server.handleAddAttributeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetAttributePayloadType:
		server.handleGetAttributeHTTPRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	default:
		if server.handleHTTPError(c, errors.New("invalid rpcMsg.PayloadType"), http.StatusForbidden) {
			return
		}
	}
}

func (server *ColoniesServer) generateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
	failure := core.CreateFailure(errorCode, err.Error())
	jsonString, err := failure.ToJSON()
	if err != nil {
		return nil, err
	}
	rpcReplyMsg, err := rpc.CreateRPCErrorReplyMsg(rpc.ErrorPayloadType, jsonString)
	if err != nil {
		return nil, err
	}

	return rpcReplyMsg, nil
}

func (server *ColoniesServer) handleHTTPError(c *gin.Context, err error, errorCode int) bool {
	if err != nil {
		logging.Log().Warning(err)
		rpcReplyMsg, err := server.generateRPCErrorMsg(err, errorCode)
		if err != nil {
			logging.Log().Error(err)
		}
		rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
		if err != nil {
			logging.Log().Error(err)
		}

		c.String(errorCode, rpcReplyMsgJSONString)
		return true
	}

	return false
}

func (server *ColoniesServer) sendHTTPReply(c *gin.Context, payloadType string, jsonString string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) sendEmptyHTTPReply(c *gin.Context, payloadType string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, "{}")
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) handleAddColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if msg.Colony == nil {
		server.handleHTTPError(c, errors.New("colony is nil"), http.StatusBadRequest)
		return
	}

	if len(msg.Colony.ID) != 64 {
		server.handleHTTPError(c, errors.New("invalid colony id length"), http.StatusBadRequest)
		return
	}

	addedColony, err := server.controller.addColony(msg.Colony)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedColony == nil {
		server.handleHTTPError(c, errors.New("addedColony is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteColonyHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteColonyMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteColony(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleGetColoniesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := server.controller.getColonies()
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertColonyArrayToJSON(colonies)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetColonyHTTPRequest(c *gin.Context, recoveredID, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.getColonyByID(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if colony == nil {
		server.handleHTTPError(c, errors.New("colony is nil"), http.StatusInternalServerError)
	}

	jsonString, err = colony.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAddRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Runtime == nil {
		server.handleHTTPError(c, errors.New("runtime is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.Runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	addedRuntime, err := server.controller.addRuntime(msg.Runtime)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedRuntime == nil {
		server.handleHTTPError(c, errors.New("addedRuntime is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedRuntime.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	runtimes, err := server.controller.getRuntimeByColonyID(msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertRuntimeArrayToJSON(runtimes)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleApproveRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateApproveRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.approveRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleRejectRuntimeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRejectRuntimeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleHTTPError(c, errors.New("runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.rejectRuntime(msg.RuntimeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleSubmitProcessSpecHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitProcessSpecMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.ProcessSpec == nil {
		server.handleHTTPError(c, errors.New("msg.ProcessSpec is nil"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.ProcessSpec)
	addedProcess, err := server.controller.addProcess(process)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedProcess == nil {
		server.handleHTTPError(c, errors.New("addedProcess is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedProcess.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAssignProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.msgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	process, err := server.controller.assignProcess(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessesHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.findWaitingProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := server.controller.findRunningProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := server.controller.findSuccessfulProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := server.controller.findFailedProcesses(msg.ColonyID, msg.Count)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleHTTPError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendHTTPReply(c, payloadType, jsonString)
	default:
		err := errors.New("invalid msg.State")
		server.handleHTTPError(c, err, http.StatusBadRequest)
		return
	}
}

func (server *ColoniesServer) handleGetProcessHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleCloseSuccessfulHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseSuccessfulMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("not allowed to close process as successful")
		server.handleHTTPError(c, err, http.StatusForbidden)
	}

	err = server.controller.closeSuccessful(process.ID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleCloseFailedHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateCloseFailedMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("not allowed to close process as failed")
		server.handleHTTPError(c, err, http.StatusForbidden)
	}

	err = server.controller.closeFailed(process.ID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyHTTPReply(c, payloadType)
}

func (server *ColoniesServer) handleAddAttributeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddAttributeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}
	if msg.Attribute == nil {
		server.handleHTTPError(c, errors.New("msg.Attribute is nil"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.Attribute.TargetID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("only runtime with id <" + process.AssignedRuntimeID + "> is allowed to set attributes")
		server.handleHTTPError(c, err, http.StatusForbidden)
		return
	}

	msg.Attribute.GenerateID()

	addedAttribute, err := server.controller.addAttribute(msg.Attribute)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if addedAttribute == nil {
		server.handleHTTPError(c, errors.New("addedAttribute is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedAttribute.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetAttributeHTTPRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetAttributeMsgFromJSON(jsonString)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleHTTPError(c, errors.New("failed to parse JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleHTTPError(c, errors.New("msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	attribute, err := server.controller.getAttribute(msg.AttributeID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if attribute == nil {
		server.handleHTTPError(c, errors.New("attribute is nil"), http.StatusInternalServerError)
	}

	process, err := server.controller.getProcessByID(attribute.TargetID)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleHTTPError(c, errors.New("process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = attribute.ToJSON()
	if server.handleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendHTTPReply(c, payloadType, jsonString)
}

func (server *ColoniesServer) numberOfProcessesSubscribers() int {
	return server.controller.numberOfProcessesSubscribers()
}

func (server *ColoniesServer) numberOfProcessSubscribers() int {
	return server.controller.numberOfProcessSubscribers()
}

func (server *ColoniesServer) ServeForever() error {
	if err := server.httpServer.ListenAndServeTLS(server.tlsCertPath, server.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *ColoniesServer) Shutdown() {
	server.controller.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		logging.Log().Warning("Server forced to shutdown:", err)
	}
}
