package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/backends/gin"
	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/security/validator"
	"github.com/colonyos/colonies/pkg/server/controllers"
	attributehandlers "github.com/colonyos/colonies/pkg/server/handlers/attribute"
	blueprinthandlers "github.com/colonyos/colonies/pkg/server/handlers/blueprint"
	channelhandlers "github.com/colonyos/colonies/pkg/server/handlers/channel"
	"github.com/colonyos/colonies/pkg/server/handlers/colony"
	cronhandlers "github.com/colonyos/colonies/pkg/server/handlers/cron"
	"github.com/colonyos/colonies/pkg/server/handlers/executor"
	filehandlers "github.com/colonyos/colonies/pkg/server/handlers/file"
	functionhandlers "github.com/colonyos/colonies/pkg/server/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/server/handlers/generator"
	locationhandlers "github.com/colonyos/colonies/pkg/server/handlers/location"
	loghandlers "github.com/colonyos/colonies/pkg/server/handlers/log"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/handlers/processgraph"
	realtimehandlers "github.com/colonyos/colonies/pkg/server/handlers/realtime"
	securityhandlers "github.com/colonyos/colonies/pkg/server/handlers/security"
	serverhandlers "github.com/colonyos/colonies/pkg/server/handlers/server"
	snapshothandlers "github.com/colonyos/colonies/pkg/server/handlers/snapshot"
	"github.com/colonyos/colonies/pkg/server/handlers/user"
	"github.com/colonyos/colonies/pkg/server/registry"
	log "github.com/sirupsen/logrus"
)

// WSController interface for WebSocket subscription management
type WSController interface {
	SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error
	SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error
}

type Server struct {
	backend                 backends.CORSBackend
	engine                  backends.Engine
	server                  backends.Server
	controller              controllers.Controller
	serverID                string
	tls                     bool
	tlsPrivateKeyPath       string
	tlsCertPath             string
	port                    int
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
	resourceDB              database.BlueprintDatabase
	securityDB              database.SecurityDatabase
	locationDB              database.LocationDatabase
	exclusiveAssign         bool
	allowExecutorReregister bool
	retention               bool
	retentionPolicy         int64
	retentionPeriod         int

	// Handler composition
	serverAdapter          *ServerAdapter
	handlerRegistry        *registry.HandlerRegistry
	userHandlers           *user.Handlers
	colonyHandlers         *colony.Handlers
	executorHandlers       *executor.Handlers
	processHandlers        *process.Handlers
	processgraphHandlers   *processgraph.Handlers
	serverHandlers         *serverhandlers.Handlers
	logHandlers            *loghandlers.Handlers
	snapshotHandlers       *snapshothandlers.Handlers
	attributeHandlers      *attributehandlers.Handlers
	cronHandlers           *cronhandlers.Handlers
	functionHandlers       *functionhandlers.Handlers
	generatorHandlers      *generatorhandlers.Handlers
	blueprintHandlers      *blueprinthandlers.Handlers
	securityHandlers       *securityhandlers.Handlers
	fileHandlers           *filehandlers.Handlers
	realtimeHandlers       *realtimehandlers.Handlers
	channelHandlers        *channelhandlers.Handlers
	locationHandlers       *locationhandlers.Handlers
	backendRealtimeHandler realtimehandlers.RealtimeHandler
	channelRouter          *channel.Router
}

func CreateServer(db database.Database,
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
	retentionPeriod int) *Server {
	return createServerInternal(db, port, tls, tlsPrivateKeyPath, tlsCertPath, thisNode, clusterConfig, etcdDataPath, generatorPeriod, cronPeriod, exclusiveAssign, allowExecutorReregister, retention, retentionPolicy, retentionPeriod)
}

func createServerInternal(db database.Database,
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
	retentionPeriod int) *Server {
	server := &Server{}

	log.WithFields(log.Fields{
		"HTTPPort": port,
	}).Info("=== INITIALIZING COLONIES SERVER ===")

	// Initialize Gin HTTP backend
	server.backend = gin.NewCORSBackend()
	server.engine = server.backend.NewEngineWithDefaults()
	server.engine.Use(server.backend.CORS())
	server.server = server.backend.NewServer(port, server.engine)
	log.Info("Gin HTTP backend initialized successfully")

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
	server.resourceDB = db
	server.securityDB = db
	server.locationDB = db

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
	server.blueprintHandlers = blueprinthandlers.NewHandlers(server.serverAdapter)
	server.securityHandlers = securityhandlers.NewHandlers(server.serverAdapter)
	server.realtimeHandlers = realtimehandlers.NewHandlers(server.serverAdapter)
	server.channelRouter = server.controller.GetChannelRouter()
	server.channelHandlers = channelhandlers.NewHandlers(server.serverAdapter)
	server.locationHandlers = locationhandlers.NewHandlers(server.serverAdapter)

	// Create backend-specific realtime handler
	server.backendRealtimeHandler = gin.NewRealtimeHandler(server.serverAdapter)

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

func (server *Server) SetAllowExecutorReregister(allow bool) {
	server.allowExecutorReregister = allow
}

// registerHandlers registers all handlers that support self-registration
func (server *Server) registerHandlers() {
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

	// Register channel handlers
	if err := server.channelHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register channel handlers")
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

	// Register blueprint handlers
	if err := server.blueprintHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register blueprint handlers")
	}

	// Register security handlers
	if err := server.securityHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register security handlers")
	}

	// Register location handlers
	if err := server.locationHandlers.RegisterHandlers(server.handlerRegistry); err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to register location handlers")
	}

	log.WithFields(log.Fields{
		"RegisteredHandlers": len(server.handlerRegistry.GetRegisteredTypes()),
		"HandlerTypes":       server.handlerRegistry.GetRegisteredTypes(),
	}).Info("Handler registration completed")
}

func (server *Server) getServerID() (string, error) {
	return server.securityDB.GetServerID()
}

func (server *Server) setupRoutes() {
	// Setup HTTP routes for both Gin and LibP2P backends
	if server.engine != nil {
		server.engine.POST("/api", server.handleAPIRequest)
		server.engine.GET("/health", server.handleHealthRequest)
		// Note: realtime handler now uses backend abstraction (but maintains /pubsub endpoint for compatibility)
		server.engine.GET("/pubsub", func(c backends.Context) {
			server.realtimeHandlers.HandleWSRequest(c)
		})

		log.Info("HTTP routes configured")
	}
}

func (server *Server) handleHealthRequest(c backends.Context) {
	c.String(http.StatusOK, "")
}

func (server *Server) handleAPIRequest(c backends.Context) {
	jsonBytes, err := c.ReadBody()
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		log.WithFields(log.Fields{"Error": err}).Error("Bad request")
		return
	}

	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	// Version does not require a valid private key
	if rpcMsg.PayloadType == rpc.VersionPayloadType {
		server.serverHandlers.HandleVersion(c, rpcMsg.PayloadType, rpcMsg.DecodePayload())
		return
	}

	// GetServerInfo does not require a valid private key
	if rpcMsg.PayloadType == rpc.GetServerInfoPayloadType {
		server.handleGetServerInfoHTTP(c)
		return
	}

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Handle with registered handlers
	if server.handlerRegistry.HandleRequestWithRaw(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload(), string(jsonBytes)) {
		return
	}

	// No handler found for this payload type
	errMsg := "invalid rpcMsg.PayloadType: " + rpcMsg.PayloadType
	if rpcMsg.PayloadType == "" {
		errMsg = "invalid rpcMsg.PayloadType: empty (this usually means the client is misconfigured - check that COLONIES_PRVKEY and other environment variables are set correctly)"
	}
	if server.HandleHTTPError(c, errors.New(errMsg), http.StatusForbidden) {
		log.Error(errMsg)
		return
	}
}

func (server *Server) generateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
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

func (server *Server) parseSignature(jsonString string, signature string) (string, error) {
	recoveredID, err := server.crypto.RecoverID(jsonString, signature)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call crypto.RecoverID()")
		return "", err
	}

	return recoveredID, nil
}

// WSController returns the WebSocket controller for realtime subscriptions
func (server *Server) WSController() WSController {
	return server.serverAdapter.WSControllerCompat()
}

func (server *Server) buildServerInfoFromEnv() *core.ServerInfo {
	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		buildVersion = "dev"
	}
	buildTime := os.Getenv("BUILD_TIME")
	if buildTime == "" {
		buildTime = time.Now().Format(time.RFC3339)
	}

	serverInfo := core.CreateServerInfo(buildVersion, buildTime)

	// Default to HTTP backend
	host := os.Getenv("COLONIES_SERVER_HTTP_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	portStr := os.Getenv("COLONIES_SERVER_HTTP_PORT")
	port := 50080
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}
	tlsStr := os.Getenv("COLONIES_SERVER_HTTP_TLS")
	tls := tlsStr == "true"
	serverInfo.AddBackend("http", port, host, tls, false)

	return serverInfo
}

// handleGetServerInfoHTTP handles GetServerInfo requests for HTTP backend
func (server *Server) handleGetServerInfoHTTP(c backends.Context) {
	serverInfo := server.buildServerInfoFromEnv()

	jsonString, err := serverInfo.ToJSON()
	if server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.GetServerInfoPayloadType, jsonString)
	if server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *Server) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
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

func (server *Server) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, jsonString)
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *Server) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(payloadType, "{}")
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}
	rpcReplyMsgJSONString, err := rpcReplyMsg.ToJSON()
	if server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	c.String(http.StatusOK, rpcReplyMsgJSONString)
}

func (server *Server) ServeForever() error {
	// Start the backend server (blocking)
	if server.server != nil {
		return server.server.ListenAndServe()
	}

	return errors.New("no server configured")
}

func (server *Server) FileDB() database.FileDatabase {
	return server.fileDB
}

func (server *Server) Shutdown() {
	server.controller.Stop()

	// Shutdown HTTP server
	if server.server != nil {
		if err := server.server.ShutdownWithTimeout(5 * time.Second); err != nil {
			log.WithFields(log.Fields{"Error": err}).Warning("Server forced to shutdown")
		}
	}
}
