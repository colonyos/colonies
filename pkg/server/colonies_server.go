package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/logging"
	"colonies/pkg/rpc"
	"colonies/pkg/security"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	ownership         security.Ownership
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
	server.ownership = security.CreateOwnership(db)
	server.rootPassword = rootPassword
	server.port = port
	server.tlsPrivateKeyPath = tlsPrivateKeyPath
	server.tlsCertPath = tlsCertPath

	server.setupRoutes()

	logging.Log().Info("Starting Colonies API server at port: " + strconv.Itoa(port))

	return server
}

func (server *ColoniesServer) setupRoutes() {
	server.ginHandler.POST("/endpoint", server.handleEndpointRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes/:processid", server.handleGetProcessRequest)
	server.ginHandler.PUT("/colonies/:colonyid/processes/:processid/finish", server.handleFinishProcessRequest)
	server.ginHandler.PUT("/colonies/:colonyid/processes/:processid/failed", server.handleFailedProcessRequest)
	server.ginHandler.POST("/colonies/:colonyid/processes/:processid/attributes", server.handleAddAttributeRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes/:processid/attributes/:attributeid", server.handleGetAttributeRequest)
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
		recoveredID, err = security.RecoverID(jsonString, signature)
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

	// SECURITY: Check that the root password matches the server root password
	err = security.VerifyRoot(msg.RootPassword, server.rootPassword)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedColony, err := server.controller.AddColony(msg.Colony)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = addedColony.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetColoniesRequest(c *gin.Context, jsonString string) {
	msg, err := rpc.CreateGetColoniesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: Check that the root password matches the server root password
	err = security.VerifyRoot(msg.RootPassword, server.rootPassword)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colonies, err := server.controller.GetColonies()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
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

	// SECURITY: We need to check if the recoveredID is a member of the targeted colony
	err = security.VerifyRuntimeMembership(recoveredID, msg.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	colony, err := server.controller.GetColonyByID(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = colony.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAddRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateAddRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: We need to check if the recoveredID is the owner of the colony
	err = security.VerifyColonyOwner(recoveredID, msg.Runtime.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	addedRuntime, err := server.controller.AddRuntime(msg.Runtime)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = addedRuntime.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetRuntimesRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetRuntimesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: We need to check if the recoveredID is a member of the targeted colony
	err = security.VerifyRuntimeMembership(recoveredID, msg.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	runtimes, err := server.controller.GetRuntimeByColonyID(msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
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

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleError(c, errors.New("Runtime with Id <"+msg.RuntimeID+"> not found"), http.StatusBadRequest)
		return
	}

	// SECURITY: We need to check if the recoveredID is a member of the targeted colony
	err = security.VerifyRuntimeMembership(recoveredID, runtime.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleApproveRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateApproveRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}
	if runtime == nil {
		server.handleError(c, errors.New("Runtime with Id <"+msg.RuntimeID+"> not found"), http.StatusBadRequest)
		return
	}

	// SECURITY: We need to check if the recoveredID is the owner of the colony
	err = security.VerifyColonyOwner(recoveredID, runtime.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.ApproveRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleRejectRuntimeRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateRejectRuntimeMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	runtime, err := server.controller.GetRuntimeByID(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: We need to check if the recoveredID is the owner of the colony
	err = security.VerifyColonyOwner(recoveredID, runtime.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	err = server.controller.RejectRuntime(msg.RuntimeID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = runtime.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleSubmitProcessSpec(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateSubmitProcessSpecMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: We need to check if the recoveredID is a member of the targeted colony
	err = security.VerifyRuntimeMembership(recoveredID, msg.ProcessSpec.Conditions.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process := core.CreateProcess(msg.ProcessSpec)
	addedProcess, err := server.controller.AddProcess(process)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = addedProcess.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAssignProcessRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateAssignProcessMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	// SECURITY: We need to check if the recoveredID is a member of the targeted colony
	err = security.VerifyRuntimeMembership(recoveredID, msg.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	process, err := server.controller.AssignProcess(recoveredID, msg.ColonyID)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = process.ToJSON()
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetProcessesRequest(c *gin.Context, recoveredID string, jsonString string) {
	msg, err := rpc.CreateGetProcessesMsgFromJSON(jsonString)
	if server.handleError(c, err, http.StatusBadRequest) {
		return
	}

	err = security.VerifyRuntimeMembership(recoveredID, msg.ColonyID, server.ownership)
	if server.handleError(c, err, http.StatusForbidden) {
		return
	}

	switch msg.State {
	case core.WAITING:
		processes, err := server.controller.FindWaitingProcesses(msg.ColonyID, msg.Count)
		if server.handleError(c, err, http.StatusBadRequest) {
			return
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

func (server *ColoniesServer) handleGetProcessRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	processID := c.Param("processid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	process, err := server.controller.GetProcessByID(colonyID, processID)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := process.ToJSON()
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleFinishProcessRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	processID := c.Param("processid")

	err := security.RequireColonyMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.MarkSuccessful(c.GetHeader("Id"), processID)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (server *ColoniesServer) handleFailedProcessRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	processID := c.Param("processid")

	err := security.RequireColonyMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.MarkFailed(c.GetHeader("Id"), processID)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (server *ColoniesServer) handleAddAttributeRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	err := security.RequireColonyMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attribute, err := core.ConvertJSONToAttribute(string(jsonBytes))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addedAttribute, err := server.controller.AddAttribute(attribute)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := addedAttribute.ToJSON()
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetAttributeRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	//processID := c.Param("processid")
	attributeID := c.Param("attributeid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	attribute, err := server.controller.GetAttribute(attributeID)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := attribute.ToJSON()
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
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
