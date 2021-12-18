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
	apiKey            string
	tlsPrivateKeyPath string
	tlsCertPath       string
	port              int
	httpServer        *http.Server
	ownership         security.Ownership
}

func CreateColoniesServer(db database.Database, port int, apiKey string, tlsPrivateKeyPath string, tlsCertPath string) *ColoniesServer {
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
	server.apiKey = apiKey
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
	server.ginHandler.POST("/colonies/:colonyid/workers", server.handleAddWorkerRequest)
	server.ginHandler.GET("/colonies/:colonyid/workers", server.handleGetWorkersRequest)
	server.ginHandler.GET("/colonies/:colonyid/workers/:workerid", server.handleGetWorkerRequest)
}

// Security condition: Only system admins can get info about all colonies.
func (server *ColoniesServer) handleGetColoniesRequest(c *gin.Context) {
	err := security.VerifyAPIKey(c.GetHeader("Api-Key"), server.apiKey)
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

// Security condition 1: Only the colony owner can get colony info.
// Security condition 2: Dummy data has to be valid.
func (server *ColoniesServer) handleGetColonyRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	err := security.VerifyColonyOwnership(colonyID, colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	colony, err := server.controller.GetColony(colonyID)
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
	err := security.VerifyAPIKey(c.GetHeader("Api-Key"), server.apiKey)
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

// Security condition: Only a colony owner can add a worker.
func (server *ColoniesServer) handleAddWorkerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	worker, err := core.CreateWorkerFromJSON(string(jsonData))
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = security.VerifyColonyOwnership(colonyID, colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	err = server.controller.AddWorker(worker)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "")
}

// Security condition: Colony owner or worker members can get worker info.
func (server *ColoniesServer) handleGetWorkersRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")

	workers, err := server.controller.GetWorkerByColonyID(colonyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := core.WorkerArrayToJSON(workers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jsonString)
}

// Security condition: Colony owner or worker members can get worker info.
func (server *ColoniesServer) handleGetWorkerRequest(c *gin.Context) {
	colonyID := c.Param("colonyid")
	workerID := c.Param("workerid")
	id := colonyID

	err := security.VerifyAccessRights(id, colonyID, c.GetHeader("Digest"), c.GetHeader("Signature"), server.ownership)
	if err != nil {
		logging.Log().Warning(err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	worker, err := server.controller.GetWorker(workerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonString, err := worker.ToJSON()
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
