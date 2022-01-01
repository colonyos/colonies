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
	controller        *ColoniesController
	rootPassword      string
	tlsPrivateKeyPath string
	tlsCertPath       string
	port              int
	httpServer        *http.Server
	crypto            security.Crypto
	validator         security.Validator
}

func CreateColoniesServer(db database.Database, port int, rootPassword string, tlsPrivateKeyPath string, tlsCertPath string, debug bool) *ColoniesServer {
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
	server.controller = CreateColoniesController(db)
	server.rootPassword = rootPassword
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
	server.ginHandler.GET("/events", server.handleWebsocketRequest)
}

func (server *ColoniesServer) handleWebsocketRequest(c *gin.Context) {
	wshandler(c.Writer, c.Request)
}

var wsupgrader = websocket.Upgrader{} // use default options

func wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("got " + string(msg))
		conn.WriteMessage(t, msg)
	}
}

func (server *ColoniesServer) handleEndpointRequest(c *gin.Context) {
	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	jsonString := string(jsonBytes)

	msgType := rpc.DetermineMsgType(jsonString)

	recoveredID := ""
	if msgType == rpc.GetColoniesMsgType || msgType == rpc.AddColonyMsgType {
		// AddColony and GetColonies requires root password instead of signatures
	} else {
		signature := c.GetHeader("Signature")
		recoveredID, err = server.crypto.RecoverID(jsonString, signature)
		if server.handleError(c, err, http.StatusForbidden) {
			return
		}
	}

	switch msgType {
	// Colony operations
	case rpc.AddColonyMsgType:
		server.handleAddColonyRequest(c, jsonString)
	case rpc.GetColoniesMsgType:
		server.handleGetColoniesRequest(c, jsonString)
	case rpc.GetColonyMsgType:
		server.handleGetColonyRequest(c, recoveredID, jsonString)

	// Runtime operations
	case rpc.AddRuntimeMsgType:
		server.handleAddRuntimeRequest(c, recoveredID, jsonString)
	case rpc.GetRuntimesMsgType:
		server.handleGetRuntimesRequest(c, recoveredID, jsonString)
	case rpc.GetRuntimeMsgType:
		server.handleGetRuntimeRequest(c, recoveredID, jsonString)
	case rpc.ApproveRuntimeMsgType:
		server.handleApproveRuntimeRequest(c, recoveredID, jsonString)
	case rpc.RejectRuntimeMsgType:
		server.handleRejectRuntimeRequest(c, recoveredID, jsonString)

	// Process operations
	case rpc.SubmitProcessSpecMsgType:
		server.handleSubmitProcessSpec(c, recoveredID, jsonString)
	case rpc.AssignProcessMsgType:
		server.handleAssignProcessRequest(c, recoveredID, jsonString)
	case rpc.GetProcessesMsgType:
		server.handleGetProcessesRequest(c, recoveredID, jsonString)
	case rpc.GetProcessMsgType:
		server.handleGetProcessRequest(c, recoveredID, jsonString)
	case rpc.MarkSuccessfulMsgType:
		server.handleMarkSuccessfulRequest(c, recoveredID, jsonString)
	case rpc.MarkFailedMsgType:
		server.handleMarkFailedRequest(c, recoveredID, jsonString)

	// Attribute operations
	case rpc.AddAttributeMsgType:
		server.handleAddAttributeRequest(c, recoveredID, jsonString)
	case rpc.GetAttributeMsgType:
		server.handleGetAttributeRequest(c, recoveredID, jsonString)
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
		failureString, err := failure.ToJSON()
		if err != nil {
			logging.Log().Error(err)
		}
		c.String(http.StatusBadRequest, failureString)

		return true
	}

	return false
}

func (server *ColoniesServer) handleAddColonyRequest(c *gin.Context, jsonString string) {
	msg, err := rpc.CreateAddColonyMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddColonyMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRoot(msg.RootPassword, server.rootPassword)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedColony, err := server.controller.AddColony(msg.Colony)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetColoniesRequest(c *gin.Context, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetColoniesMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRoot(msg.RootPassword, server.rootPassword)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := server.controller.GetColonies()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if colonies == nil {
		server.handleError(c, errors.New("handleGetColoniesRequest: colonies is nil"), http.StatusInternalServerError)
	}

	jsonString, err = core.ConvertColonyArrayToJSON(colonies)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetColonyRequest(c *gin.Context, recoveredID, jsonString string) {
	msg, err := rpc.CreateGetColonyMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetColonyMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.GetColonyByID(msg.ColonyID)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAddRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateAddRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddRuntimeMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireColonyOwner(recoveredID, msg.Runtime.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedRuntime, err := server.controller.AddRuntime(msg.Runtime)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetRuntimesRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetRuntimesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetRuntimesMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	runtimes, err := server.controller.GetRuntimeByColonyID(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtimes == nil {
		server.handleError(c, errors.New("handleGetRuntimesRequest: runtimes is nil"), http.StatusInternalServerError)
	}

	jsonString, err = core.ConvertRuntimeArrayToJSON(runtimes)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetRuntimeMsg JSON"), http.StatusBadRequest)
	}

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleApproveRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateApproveRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse ApproveRuntimeMsg JSON"), http.StatusBadRequest)
	}

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
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

	err = server.controller.ApproveRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleRejectRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateRejectRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse RejectRuntimeMsg JSON"), http.StatusBadRequest)
	}

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
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

	err = server.controller.RejectRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleSubmitProcessSpec(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateSubmitProcessSpecMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse SubmitRuntimeMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ProcessSpec.Conditions.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.ProcessSpec)
	addedProcess, err := server.controller.AddProcess(process)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAssignProcessRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AssignRuntimeMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process, err := server.controller.AssignProcess(recoveredID, msg.ColonyID)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetProcessesRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetProcessesMsg JSON"), http.StatusBadRequest)
	}

	err = server.validator.RequireRuntimeMembership(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.FindWaitingProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		if processes == nil {
			server.handleError(c, errors.New("handleGetProcessesRequest (WAITING): processes is nil"), http.StatusInternalServerError)
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		c.String(http.StatusOK, jsonString)
	case core.RUNNING:
		processes, err := server.controller.FindRunningProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		if processes == nil {
			server.handleError(c, errors.New("handleGetProcessesRequest (RUNNING): processes is nil"), http.StatusInternalServerError)
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		c.String(http.StatusOK, jsonString)
	case core.SUCCESS:
		processes, err := server.controller.FindSuccessfulProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		if processes == nil {
			server.handleError(c, errors.New("handleGetProcessesRequest (SUCCESS): processes is nil"), http.StatusInternalServerError)
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		c.String(http.StatusOK, jsonString)
	case core.FAILED:
		processes, err := server.controller.FindFailedProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		if processes == nil {
			server.handleError(c, errors.New("handleGetProcessesRequest (FAILED): processes is nil"), http.StatusInternalServerError)
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
		}
		c.String(http.StatusOK, jsonString)
	default:
		err := errors.New("Invalid state")
		server.handleError(c, err, http.StatusBadRequest)
		return
	}
}

func (server *ColoniesServer) handleGetProcessRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetProcessMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetProcessMsg JSON"), http.StatusBadRequest)
	}

	process, err := server.controller.GetProcessByID(msg.ProcessID)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleMarkSuccessfulRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateMarkSuccessfulMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse MarkSuccessfulMsg JSON"), http.StatusBadRequest)
	}

	process, err := server.controller.GetProcessByID(msg.ProcessID)
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

	err = server.controller.MarkSuccessful(process.ID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, "")
}

func (server *ColoniesServer) handleMarkFailedRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateMarkFailedMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse MarkedFailedMsg JSON"), http.StatusBadRequest)
	}

	process, err := server.controller.GetProcessByID(msg.ProcessID)
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

	err = server.controller.MarkFailed(process.ID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, "")
}

func (server *ColoniesServer) handleAddAttributeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateAddAttributeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse AddAttributeMsg JSON"), http.StatusBadRequest)
	}

	process, err := server.controller.GetProcessByID(msg.Attribute.TargetID)
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

	addedAttribute, err := server.controller.AddAttribute(msg.Attribute)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetAttributeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetAttributeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if msg == nil {
		server.handleError(c, errors.New("Failed to parse GetAttributeMsg JSON"), http.StatusBadRequest)
	}

	attribute, err := server.controller.GetAttribute(msg.AttributeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if attribute == nil {
		server.handleError(c, errors.New("handleGetAttributeRequest: attribute is nil"), http.StatusInternalServerError)
	}

	process, err := server.controller.GetProcessByID(attribute.TargetID)
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

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) ServeForever() error {
	if err := server.httpServer.ListenAndServeTLS(server.tlsCertPath, server.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (server *ColoniesServer) Shutdown() {
	server.controller.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		logging.Log().Warning("Server forced to shutdown:", err)
	}
}
