package server

import (
	"colonies/internal/logging"
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/rpc"
	"colonies/pkg/security"
	"colonies/pkg/security/crypto"
	"colonies/pkg/security/validator"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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

func (server *ColoniesServer) handleWSRequest(c *gin.Context) {
	w := c.Writer
	r := c.Request
	var wsupgrader = websocket.Upgrader{}
	var err error
	var wsConn *websocket.Conn
	wsConn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		wsPayloadType, data, err := wsConn.ReadMessage()
		if err != nil {
			return
		}

		rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(data))
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}

		recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
		if server.handleError(c, err, http.StatusForbidden) {
			return
		}

		switch rpcMsg.PayloadType {
		case rpc.SubscribeProcessesPayloadType:
			msg, err := rpc.CreateSubscribeProcessesMsgFromJSON(rpcMsg.DecodePayload())
			if server.handleError(c, err, http.StatusBadRequest) {
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
				return
			}

			processSubcription := createProcessesSubscription(wsConn, wsPayloadType, msg.RuntimeType, msg.Timeout, msg.State)
			server.controller.subscribeProcesses(recoveredID, processSubcription)

		case rpc.SubscribeProcessPayloadType:
			msg, err := rpc.CreateSubscribeProcessMsgFromJSON(rpcMsg.DecodePayload())
			if server.handleError(c, err, http.StatusBadRequest) {
				return
			}
			if msg.MsgType != rpcMsg.PayloadType {
				server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
				return
			}

			processSubcription := createProcessSubscription(wsConn, wsPayloadType, msg.ProcessID, msg.Timeout, msg.State)
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
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	switch rpcMsg.PayloadType {
	// Colony operations
	case rpc.AddColonyPayloadType:
		server.handleAddColonyRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.DeleteColonyPayloadType:
		server.handleDeleteColonyRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColoniesPayloadType:
		server.handleGetColoniesRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetColonyPayloadType:
		server.handleGetColonyRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Runtime operations
	case rpc.AddRuntimePayloadType:
		server.handleAddRuntimeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimesPayloadType:
		server.handleGetRuntimesRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetRuntimePayloadType:
		server.handleGetRuntimeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.ApproveRuntimePayloadType:
		server.handleApproveRuntimeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.RejectRuntimePayloadType:
		server.handleRejectRuntimeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Process operations
	case rpc.SubmitProcessSpecPayloadType:
		server.handleSubmitProcessSpec(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.AssignProcessPayloadType:
		server.handleAssignProcessRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessesPayloadType:
		server.handleGetProcessesRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetProcessPayloadType:
		server.handleGetProcessRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.MarkSuccessfulPayloadType:
		server.handleMarkSuccessfulRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.MarkFailedPayloadType:
		server.handleMarkFailedRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())

	// Attribute operations
	case rpc.AddAttributePayloadType:
		server.handleAddAttributeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	case rpc.GetAttributePayloadType:
		server.handleGetAttributeRequest(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload())
	default:
		if server.handleError(c, errors.New("Invalid RPC message type"), http.StatusForbidden) {
			return
		}
	}
}

func (server *ColoniesServer) handleError(c *gin.Context, err error, errorCode int) bool {
	if err != nil {
		logging.Log().Warning(err)
		failure := core.CreateFailure(errorCode, err.Error())
		jsonString, err := failure.ToJSON()
		if err != nil {
			logging.Log().Error(err)
		}
		rpcReplyMsg, err := rpc.CreateRPCErrorReplyMsg(rpc.ErrorPayloadType, jsonString)
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

func (server *ColoniesServer) sendReplyToClient(c *gin.Context, payloadType string, jsonString string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) sendEmptyReplyToClient(c *gin.Context, payloadType string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, "{}")
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *ColoniesServer) handleAddColonyRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddColonyMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedColony, err := server.controller.addColony(msg.Colony)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if addedColony == nil {
		server.handleError(c, errors.New("handleAddColonyRequest: addedColony is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleDeleteColonyRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateDeleteColonyMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse DeleteColonyMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.deleteColony(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyReplyToClient(c, payloadType)
}

func (server *ColoniesServer) handleGetColoniesRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetColoniesMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireServerOwner(recoveredID, server.serverID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := server.controller.getColonies()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertColonyArrayToJSON(colonies)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetColonyRequest(c *gin.Context, recoveredID, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetColonyMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.getColonyByID(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if colony == nil {
		server.handleError(c, errors.New("handleGetColonyRequest: colony is nil"), http.StatusInternalServerError)
	}

	jsonString, err = colony.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAddRuntimeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.Runtime.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedRuntime, err := server.controller.addRuntime(msg.Runtime)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if addedRuntime == nil {
		server.handleError(c, errors.New("handleAddRuntimeRequest: addedRuntime is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedRuntime.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimesRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetRuntimesMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	runtimes, err := server.controller.getRuntimeByColonyID(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = core.ConvertRuntimeArrayToJSON(runtimes)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetRuntimeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleError(c, errors.New("handleGetRuntimeRequest: runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, runtime.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleApproveRuntimeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateApproveRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse ApproveRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleError(c, errors.New("handleApproveColonyRequest: runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.approveRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyReplyToClient(c, payloadType)
}

func (server *ColoniesServer) handleRejectRuntimeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRejectRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse RejectRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	runtime, err := server.controller.getRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleError(c, errors.New("handleRejectRuntimeRequest: runtime is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireColonyOwner(recoveredID, runtime.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.rejectRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyReplyToClient(c, payloadType)
}

func (server *ColoniesServer) handleSubmitProcessSpec(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSubmitProcessSpecMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse SubmitRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.ProcessSpec)
	addedProcess, err := server.controller.addProcess(process)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if addedProcess == nil {
		server.handleError(c, errors.New("handleSubmitProcessSpecRequest: addedProcess is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedProcess.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleAssignProcessRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AssignRuntimeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process, err := server.controller.assignProcess(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleAssignRequest: process is nil"), http.StatusInternalServerError)
	}

	jsonString, err = process.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetProcessesRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetProcessesMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.findWaitingProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendReplyToClient(c, payloadType, jsonString)
	case core.RUNNING:
		processes, err := server.controller.findRunningProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendReplyToClient(c, payloadType, jsonString)
	case core.SUCCESS:
		processes, err := server.controller.findSuccessfulProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendReplyToClient(c, payloadType, jsonString)
	case core.FAILED:
		processes, err := server.controller.findFailedProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		server.sendReplyToClient(c, payloadType, jsonString)
	default:
		err := errors.New("Invalid state")
		server.handleError(c, err, http.StatusBadRequest)
		return
	}
}

func (server *ColoniesServer) handleGetProcessRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetProcessMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetProcessMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleGetProcessRequest: process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleMarkSuccessfulRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateMarkSuccessfulMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse MarkSuccessfulMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleMarkSuccessfulRequest: process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Only Runtime with Id <" + process.AssignedRuntimeID + "> is allowed to mark process as failed")
		server.handleError(c, err, http.StatusForbidden)
	}

	err = server.controller.markSuccessful(process.ID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyReplyToClient(c, payloadType)
}

func (server *ColoniesServer) handleMarkFailedRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateMarkFailedMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse MarkedFailedMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.ProcessID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleMarkFailedRequest: process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Only Runtime with Id <" + process.AssignedRuntimeID + "> is allowed to mark process as failed")
		server.handleError(c, err, http.StatusForbidden)
	}

	err = server.controller.markFailed(process.ID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	server.sendEmptyReplyToClient(c, payloadType)
}

func (server *ColoniesServer) handleAddAttributeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateAddAttributeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddAttributeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	process, err := server.controller.getProcessByID(msg.Attribute.TargetID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleAddAttributeRequest: process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Only Runtime with Id <" + process.AssignedRuntimeID + "> is allowed to set attributes")
		server.handleError(c, err, http.StatusForbidden)
		return
	}

	addedAttribute, err := server.controller.addAttribute(msg.Attribute)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if addedAttribute == nil {
		server.handleError(c, errors.New("handleAddAttributeRequest: addedAttribute is nil"), http.StatusInternalServerError)
	}

	jsonString, err = addedAttribute.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
}

func (server *ColoniesServer) handleGetAttributeRequest(c *gin.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetAttributeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetAttributeMsg JSON"), http.StatusBadRequest)
	}
	if msg.MsgType != payloadType {
		server.handleError(c, errors.New("MsgType does not match PayloadType"), http.StatusBadRequest)
		return
	}

	attribute, err := server.controller.getAttribute(msg.AttributeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if attribute == nil {
		server.handleError(c, errors.New("handleGetAttributeRequest: attribute is nil"), http.StatusInternalServerError)
	}

	process, err := server.controller.getProcessByID(attribute.TargetID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if process == nil {
		server.handleError(c, errors.New("handleGetAttributeRequest: process is nil"), http.StatusInternalServerError)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, process.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	if process.AssignedRuntimeID != recoveredID {
		err := errors.New("Only Runtime with Id <" + process.AssignedRuntimeID + "> is allowed to set attributes")
		server.handleError(c, err, http.StatusForbidden)
	}

	jsonString, err = attribute.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	server.sendReplyToClient(c, payloadType, jsonString)
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
