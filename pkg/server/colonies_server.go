package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/logging"
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

// params := c.Request.URL.Query()
// randomData := params["dummydata"][0]

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

func CreateColoniesServer(db database.Database, port int, rootPassword string, tlsPrivateKeyPath string, tlsCertPath string) *ColoniesServer {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

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

	logrus.SetLevel(logrus.DebugLevel)
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
}

// Security condition: Only system admins can get info about all colonies.
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

	jsonString, err := core.ColonyArrayToJSON(colonies)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

// Security condition: Only the colony owner can get colony info.
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

// Security condition: Only system admins can add a colony.
func (server *ColoniesServer) handleAddColonyRequest(c *gin.Context) {
	err := security.RequireRoot(c.GetHeader("RootPassword"), server.rootPassword)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	colony, err := core.CreateColonyFromJSON(string(jsonData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.AddColony(colony)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

// Security condition: Only a colony owner can add a computer.
func (server *ColoniesServer) handleAddComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	err := security.RequireColonyOwner(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	computer, err := core.CreateComputerFromJSON(string(jsonData))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.AddComputer(computer)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

// Security condition: Colony owner or computer members can get computer info.
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

	jsonString, err := core.ComputerArrayToJSON(computers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

// Security condition: Colony owner or computer members can get computer info.
func (server *ColoniesServer) handleGetComputerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	computerID := c.Param("computerid")

	err := security.RequireColonyOwnerOrMember(c.GetHeader("Id"), colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	computer, err := server.controller.GetComputer(computerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := computer.ToJSON()
	if err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		logging.Log().Warning("Server forced to shutdown:", err)
	}
}
