package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/logging"
	"colonies/pkg/security"
	"context"
	"errors"
	"fmt"
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
	server.ginHandler.GET("/colonies", server.handleGetColoniesRequest)
	server.ginHandler.GET("/colonies/:colonyid", server.handleGetColonyRequest)
	server.ginHandler.POST("/colonies", server.handleAddColonyRequest)
	server.ginHandler.POST("/colonies/:colonyid/computers", server.handleAddComputerRequest)
	server.ginHandler.GET("/colonies/:colonyid/computers", server.handleGetComputersRequest)
	server.ginHandler.GET("/colonies/:colonyid/computers/:computerid", server.handleGetComputerRequest)
	server.ginHandler.PUT("/colonies/:colonyid/computers/:computerid/approve", server.handleApproveComputerRequest)
	server.ginHandler.PUT("/colonies/:colonyid/computers/:computerid/reject", server.handleRejectComputerRequest)
	server.ginHandler.POST("/colonies/:colonyid/processes", server.handleAddProcessRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes", server.handleGetProcessesRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes/:processid", server.handleGetProcessRequest)
	server.ginHandler.PUT("/colonies/:colonyid/processes/:processid/finish", server.handleFinishProcessRequest)
	server.ginHandler.PUT("/colonies/:colonyid/processes/:processid/failed", server.handleFailedProcessRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes/assign", server.handleAssignProcessRequest)
	server.ginHandler.POST("/colonies/:colonyid/processes/:processid/attributes", server.handleAddAttributeRequest)
	server.ginHandler.GET("/colonies/:colonyid/processes/:processid/attributes/:attributeid", server.handleGetAttributeRequest)
}

func (server *ColoniesServer) handleGetColoniesRequest(c *gin.Context) {
	err := security.RequireRoot(c.GetHeader("RootPassword"), server.rootPassword)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	colonies, err := server.controller.GetColonies()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := core.ConvertColonyArrayToJSON(colonies)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetColonyRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	colony, err := server.controller.GetColonyByID(colonyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := colony.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAddColonyRequest(c *gin.Context) {
	err := security.RequireRoot(c.GetHeader("RootPassword"), server.rootPassword)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colony, err := core.ConvertJSONToColony(string(jsonBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addedColony, err := server.controller.AddColony(colony)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := addedColony.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleAddComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	err := security.RequireColonyOwner(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
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

	computer, err := core.ConvertJSONToComputer(string(jsonBytes))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addedComputer, err := server.controller.AddComputer(computer)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := addedComputer.ToJSON()
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetComputersRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	computers, err := server.controller.GetComputerByColonyID(colonyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := core.ConvertComputerArrayToJSON(computers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	computerID := c.Param("computerid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyComputerMembership(computerID, colonyID, server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	computer, err := server.controller.GetComputerByID(computerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := computer.ToJSON()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleApproveComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	computerID := c.Param("computerid")

	err := security.RequireColonyOwner(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyComputerMembership(computerID, colonyID, server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.ApproveComputer(computerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (server *ColoniesServer) handleRejectComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	computerID := c.Param("computerid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyComputerMembership(computerID, colonyID, server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.RejectComputer(computerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (server *ColoniesServer) handleAddProcessRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
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

	processSpec, err := core.ConvertJSONToProcessSpec(string(jsonBytes))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	process := core.CreateProcess(processSpec.TargetColonyID, processSpec.TargetComputerIDs, processSpec.ComputerType, processSpec.Timeout, processSpec.MaxRetries, processSpec.Mem, processSpec.Cores, processSpec.GPUs)

	var attributes []*core.Attribute
	for key, value := range processSpec.In {
		attributes = append(attributes, core.CreateAttribute(process.ID(), core.IN, key, value))
	}

	process.SetInAttributes(attributes)

	addedProcess, err := server.controller.AddProcess(process)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := addedProcess.ToJSON()
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

func (server *ColoniesServer) handleGetProcessesRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	stateStr := c.GetHeader("State")
	countStr := c.GetHeader("Count")
	computerID := c.GetHeader("ComputerID")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyComputerMembership(computerID, colonyID, server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if stateStr == "" {
		errorMsg := "State must be specified"
		logging.Log().Warning(errors.New(errorMsg))
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if countStr == "" {
		errorMsg := "Count must be specified"
		logging.Log().Warning(errors.New(errorMsg))
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch state {
	case core.WAITING:
		processes, err := server.controller.FindWaitingProcesses(computerID, colonyID, count)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, jsonString)
	case core.RUNNING:
		fmt.Println("running")
	case core.SUCCESS:
		fmt.Println("success")
	case core.FAILED:
		fmt.Println("failed")
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

func (server *ColoniesServer) handleAssignProcessRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	computerID := c.GetHeader("ComputerID")

	err := security.RequireColonyMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyComputerMembership(computerID, colonyID, server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	process, err := server.controller.AssignProcess(computerID, colonyID)
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
