package service

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/security/validator"
	"github.com/colonyos/colonies/pkg/service/controllers"
	"github.com/colonyos/colonies/pkg/service/registry"
	attributehandlers "github.com/colonyos/colonies/pkg/service/handlers/attribute"
	cronhandlers "github.com/colonyos/colonies/pkg/service/handlers/cron"
	filehandlers "github.com/colonyos/colonies/pkg/service/handlers/file"
	functionhandlers "github.com/colonyos/colonies/pkg/service/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/service/handlers/generator"
	securityhandlers "github.com/colonyos/colonies/pkg/service/handlers/security"
	"github.com/colonyos/colonies/pkg/service/handlers/user"
	"github.com/colonyos/colonies/pkg/service/handlers/colony"
	"github.com/colonyos/colonies/pkg/service/handlers/executor"
	loghandlers "github.com/colonyos/colonies/pkg/service/handlers/log"
	"github.com/colonyos/colonies/pkg/service/handlers/process"
	"github.com/colonyos/colonies/pkg/service/handlers/processgraph"
	serverhandlers "github.com/colonyos/colonies/pkg/service/handlers/server"
	snapshothandlers "github.com/colonyos/colonies/pkg/service/handlers/snapshot"
	websockethandlers "github.com/colonyos/colonies/pkg/service/handlers/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type ColoniesServer struct {
	ginHandler              *gin.Engine
	controller              controllers.Controller
	serverID                string
	tls                     bool
	tlsPrivateKeyPath       string
	tlsCertPath             string
	port                    int
	httpServer              *http.Server
	crypto                  security.Crypto
	validator               security.Validator
	userDB                  database.UserDatabase
	colonyDB                database.ColonyDatabase
	executorDB              database.ExecutorDatabase
	functionDB              database.FunctionDatabase
	processDB               database.ProcessDatabase
	attributeDB             database.AttributeDatabase
	processGraphDB          database.ProcessGraphDatabase
	generatorDB             database.GeneratorDatabase
	cronDB                  database.CronDatabase
	logDB                   database.LogDatabase
	fileDB                  database.FileDatabase
	snapshotDB              database.SnapshotDatabase
	securityDB              database.SecurityDatabase
	exclusiveAssign         bool
	allowExecutorReregister bool
	retention               bool
	retentionPolicy         int64
	retentionPeriod         int
	
	// Handler composition
	serverAdapter        *ServerAdapter
	handlerRegistry      *registry.HandlerRegistry
	userHandlers         *user.Handlers
	colonyHandlers       *colony.Handlers
	executorHandlers     *executor.Handlers
	processHandlers      *process.Handlers
	processgraphHandlers *processgraph.Handlers
	serverHandlers       *serverhandlers.Handlers
	logHandlers          *loghandlers.Handlers
	snapshotHandlers     *snapshothandlers.Handlers
	attributeHandlers    *attributehandlers.Handlers
	cronHandlers         *cronhandlers.Handlers
	functionHandlers     *functionhandlers.Handlers
	generatorHandlers    *generatorhandlers.Handlers
	securityHandlers     *securityhandlers.Handlers
	fileHandlers         *filehandlers.Handlers
	websocketHandlers    *websockethandlers.Handlers
}

func CreateColoniesServer(db database.Database,
	port int,
	tls bool,
	tlsPrivateKeyPath string,
	tlsCertPath string,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int,
	exclusiveAssign bool,
	allowExecutorReregister bool,
	retention bool,
	retentionPolicy int64,
	retentionPeriod int) *ColoniesServer {
	server := &ColoniesServer{}
	server.ginHandler = gin.Default()
	server.ginHandler.Use(cors.Default())

	// Set all the specific database interfaces
	server.userDB = db
	server.colonyDB = db
	server.executorDB = db
	server.functionDB = db
	server.processDB = db
	server.attributeDB = db
	server.processGraphDB = db
	server.generatorDB = db
	server.cronDB = db
	server.logDB = db
	server.fileDB = db
	server.snapshotDB = db
	server.securityDB = db

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.ginHandler,
	}

	server.httpServer = httpServer
	server.controller = controllers.CreateColoniesController(db, thisNode, clusterConfig, etcdDataPath, generatorPeriod, cronPeriod, retention, retentionPolicy, retentionPeriod)

	server.tls = tls
	server.port = port
	server.tlsPrivateKeyPath = tlsPrivateKeyPath
	server.tlsCertPath = tlsCertPath
	server.crypto = crypto.CreateCrypto()
	server.validator = validator.CreateValidator(db)
	server.exclusiveAssign = exclusiveAssign
	server.allowExecutorReregister = allowExecutorReregister
	server.retention = retention
	server.retentionPolicy = retentionPolicy

	// Initialize server adapter and handler structs
	server.serverAdapter = NewServerAdapter(server)
	server.handlerRegistry = registry.NewHandlerRegistry()
	server.userHandlers = user.NewHandlers(server.serverAdapter)
	server.colonyHandlers = colony.NewHandlers(server.serverAdapter)
	server.executorHandlers = executor.NewHandlers(server.serverAdapter)
	server.processHandlers = process.NewHandlers(server.serverAdapter)
	server.processgraphHandlers = processgraph.NewHandlers(server.serverAdapter.ProcessgraphServer())
	server.serverHandlers = serverhandlers.NewHandlers(server.serverAdapter.ServerServer())
	server.logHandlers = loghandlers.NewHandlers(server.serverAdapter)
	server.snapshotHandlers = snapshothandlers.NewHandlers(server.serverAdapter)
	server.attributeHandlers = attributehandlers.NewHandlers(server.serverAdapter)
	server.cronHandlers = cronhandlers.NewHandlers(server.serverAdapter)
	server.fileHandlers = filehandlers.NewHandlers(server.serverAdapter)
	server.functionHandlers = functionhandlers.NewHandlers(server.serverAdapter)
	server.generatorHandlers = generatorhandlers.NewHandlers(server.serverAdapter)
	server.securityHandlers = securityhandlers.NewHandlers(server.serverAdapter)
	server.websocketHandlers = websockethandlers.NewHandlers(server.serverAdapter)
	
	// Register all handlers that implement self-registration
	server.registerHandlers()

	log.WithFields(log.Fields{"Port": port,
		"TLS":                     tls,
		"TLSPrivateKeyPath":       tlsPrivateKeyPath,
		"TLSCertPath":             tlsCertPath,
		"APIPort":                 thisNode.APIPort,
		"EtcdClientPort":          thisNode.EtcdClientPort,
		"EtcdPeerPort":            thisNode.EtcdPeerPort,
		"EtcdDataPath":            etcdDataPath,
		"Host":                    thisNode.Host,
		"RelayPort":               thisNode.RelayPort,
		"Name":                    thisNode.Name,
		"GeneratorPeriod":         generatorPeriod,
		"CronPeriod":              cronPeriod,
		"AllowExecutorReregister": allowExecutorReregister,
		"ExclusiveAssign":         exclusiveAssign,
		"Retention":               retention,
		"RetentionPolicy":         retentionPolicy}).
		Info("Starting Colonies server")

	server.setupRoutes()

	return server
}

func (server *ColoniesServer) SetAllowExecutorReregister(allow bool) {
	server.allowExecutorReregister = allow
}

// registerHandlers registers all handlers that support self-registration
func (server *ColoniesServer) registerHandlers() {
	// Register attribute handlers
	if err := server.attributeHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register attribute handlers")
	}

	// Register user handlers
	if err := server.userHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register user handlers")
	}

	// Register colony handlers
	if err := server.colonyHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register colony handlers")
	}

	// Register executor handlers
	if err := server.executorHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register executor handlers")
	}

	// Register function handlers
	if err := server.functionHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register function handlers")
	}

	// Register cron handlers
	if err := server.cronHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register cron handlers")
	}

	// Register generator handlers
	if err := server.generatorHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register generator handlers")
	}

	// Register server handlers
	if err := server.serverHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register server handlers")
	}

	// Register process graph handlers
	if err := server.processgraphHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register process graph handlers")
	}

	// Register log handlers
	if err := server.logHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register log handlers")
	}

	// Register process handlers
	if err := server.processHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register process handlers")
	}

	// Register file handlers
	if err := server.fileHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register file handlers")
	}

	// Register snapshot handlers
	if err := server.snapshotHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register snapshot handlers")
	}

	// Register security handlers
	if err := server.securityHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register security handlers")
	}

	log.WithFields(log.Fields{
		"RegisteredHandlers": len(server.handlerRegistry.GetRegisteredTypes()),
		"HandlerTypes":      server.handlerRegistry.GetRegisteredTypes(),
	}).Info("Handler registration completed")
}

func (server *ColoniesServer) getServerID() (string, error) {
	return server.securityDB.GetServerID()
}

func (server *ColoniesServer) setupRoutes() {
	server.ginHandler.POST("/api", server.handleAPIRequest)
	server.ginHandler.GET("/health", server.handleHealthRequest)
	server.ginHandler.GET("/pubsub", server.websocketHandlers.HandleWSRequest)
}

func (server *ColoniesServer) parseSignature(jsonString string, signature string) (string, error) {
	recoveredID, err := server.crypto.RecoverID(jsonString, signature)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call crypto.RecoverID()")
		return "", err
	}

	return recoveredID, nil
}

func (server *ColoniesServer) handleHealthRequest(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func (server *ColoniesServer) handleAPIRequest(c *gin.Context) {
	jsonBytes, err := ioutil.ReadAll(c.Request.Body)
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Error("Bad request")
		return
	}

	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if server.handleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	// Version does not require a valid private key
	if rpcMsg.PayloadType == rpc.VersionPayloadType {
		server.serverHandlers.HandleVersion(c, rpcMsg.PayloadType, rpcMsg.DecodePayload())
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.handleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Handle with registered handlers
	if server.handlerRegistry.HandleRequestWithRaw(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload(), string(jsonBytes)) {
		return
	}

	// No handler found for this payload type
	errMsg := "invalid rpcMsg.PayloadType, " + rpcMsg.PayloadType
	if server.handleHTTPError(c, errors.New(errMsg), http.StatusForbidden) {
		log.Error(errMsg)
		return
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
		if !strings.HasPrefix(err.Error(), "No processes can be selected for executor with Id") {
			log.Debug(err)
		}

		rpcReplyMsg, err := server.generateRPCErrorMsg(err, errorCode)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to call server.generateRPCErrorMsg()")
		}
		rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to call pcReplyMsg.ToJSON()")
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

func (server *ColoniesServer) ServeForever() error {
	if server.tls {
		if err := server.httpServer.ListenAndServeTLS(server.tlsCertPath, server.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
		if err := server.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func (server *ColoniesServer) FileDB() database.FileDatabase {
	return server.fileDB
}

func (server *ColoniesServer) Shutdown() {
	server.controller.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("ColoniesServer forced to shutdown")
	}
}
