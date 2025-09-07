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
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/security/validator"
	"github.com/colonyos/colonies/pkg/server/controllers"
	attributehandlers "github.com/colonyos/colonies/pkg/server/handlers/attribute"
	"github.com/colonyos/colonies/pkg/server/handlers/colony"
	cronhandlers "github.com/colonyos/colonies/pkg/server/handlers/cron"
	"github.com/colonyos/colonies/pkg/server/handlers/executor"
	filehandlers "github.com/colonyos/colonies/pkg/server/handlers/file"
	functionhandlers "github.com/colonyos/colonies/pkg/server/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/server/handlers/generator"
	loghandlers "github.com/colonyos/colonies/pkg/server/handlers/log"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/handlers/processgraph"
	realtimehandlers "github.com/colonyos/colonies/pkg/server/handlers/realtime"
	securityhandlers "github.com/colonyos/colonies/pkg/server/handlers/security"
	serverhandlers "github.com/colonyos/colonies/pkg/server/handlers/server"
	snapshothandlers "github.com/colonyos/colonies/pkg/server/handlers/snapshot"
	"github.com/colonyos/colonies/pkg/server/handlers/user"
	"github.com/colonyos/colonies/pkg/server/registry"

	backendGin "github.com/colonyos/colonies/pkg/backends/gin"
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
	realtimeHandlers     *realtimehandlers.Handlers
	realtimeHandler      *backendGin.RealtimeHandler

	// LibP2P components (if using LibP2P backend)
	libp2pEnabled  bool
	libp2pHost     interface{} // host.Host - using interface{} to avoid import
	libp2pPubsub   interface{} // *pubsub.PubSub
	libp2pTCPAddr  string      // TCP multiaddress
	libp2pQUICAddr string      // QUIC multiaddress
}

// GetBackendTypeFromEnv returns the backend type from environment variables
// Supports: COLONIES_BACKEND_TYPE environment variable
// Valid values: "gin", "libp2p"
// Default: "gin"
func GetBackendTypeFromEnv() BackendType {
	backendEnv := strings.ToLower(os.Getenv("COLONIES_BACKEND_TYPE"))

	switch backendEnv {
	case "gin", "":
		return GinBackendType
	case "libp2p":
		return LibP2PBackendType
	default:
		log.WithField("COLONIES_BACKEND_TYPE", backendEnv).Warn("Unknown backend type, defaulting to Gin")
		return GinBackendType
	}
}

// CreateServerFromEnv creates a server using backend type from environment variables
func CreateServerFromEnv(db database.Database,
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
	backendType := GetBackendTypeFromEnv()
	log.WithField("BackendType", backendType).Info("Creating server with backend from environment")
	return CreateServerWithBackend(db, port, tls, tlsPrivateKeyPath, tlsCertPath, thisNode, clusterConfig, etcdDataPath, generatorPeriod, cronPeriod, exclusiveAssign, allowExecutorReregister, retention, retentionPolicy, retentionPeriod, backendType)
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
	// Default to Gin backend for backward compatibility
	return CreateServerWithBackend(db, port, tls, tlsPrivateKeyPath, tlsCertPath, thisNode, clusterConfig, etcdDataPath, generatorPeriod, cronPeriod, exclusiveAssign, allowExecutorReregister, retention, retentionPolicy, retentionPeriod, GinBackendType)
}

func CreateServerWithBackend(db database.Database,
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
	retentionPeriod int,
	backendType BackendType) *Server {
	server := &Server{}

	// Initialize backend based on type
	switch backendType {
	case GinBackendType:
		server.backend = gin.NewCORSBackend()
		server.engine = server.backend.NewEngineWithDefaults()
		// Add CORS middleware
		server.engine.Use(server.backend.CORS())
		server.server = server.backend.NewServer(port, server.engine)
	case LibP2PBackendType:
		// For LibP2P backend, we still use Gin backend for HTTP endpoints
		// but will add LibP2P networking on top
		log.WithField("BackendType", backendType).Info("Creating server with LibP2P networking support")

		// Initialize with Gin backend for HTTP endpoints
		server.backend = gin.NewCORSBackend()
		server.engine = server.backend.NewEngineWithDefaults()
		server.engine.Use(server.backend.CORS())
		server.server = server.backend.NewServer(port, server.engine)

		// We'll add LibP2P host creation after the main server setup
		server.libp2pEnabled = true
	default:
		log.WithField("BackendType", backendType).Fatal("Unknown backend type")
	}

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
	server.realtimeHandlers = realtimehandlers.NewHandlers(server.serverAdapter)
	server.realtimeHandler = backendGin.NewRealtimeHandler(server.serverAdapter)

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

	// Initialize LibP2P if enabled
	if server.libp2pEnabled {
		server.setupLibP2P(port, thisNode)
	}

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
		"HandlerTypes":       server.handlerRegistry.GetRegisteredTypes(),
	}).Info("Handler registration completed")
}

func (server *Server) getServerID() (string, error) {
	return server.securityDB.GetServerID()
}

func (server *Server) setupRoutes() {
	server.engine.POST("/api", server.handleAPIRequest)
	server.engine.GET("/health", server.handleHealthRequest)
	// Note: realtime handler now uses backend abstraction (but maintains /pubsub endpoint for compatibility)
	server.engine.GET("/pubsub", func(c backends.Context) {
		server.realtimeHandlers.HandleWSRequest(c)
	})
}

func (server *Server) setupLibP2P(port int, thisNode cluster.Node) {
	log.WithField("Port", port).Info("Initializing LibP2P networking")

	// Calculate LibP2P ports
	libp2pTCPPort := port + 1000  // TCP port for libp2p
	libp2pQUICPort := port + 1001 // QUIC port for libp2p

	// Create multiaddresses that would be used
	tcpMultiaddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", libp2pTCPPort)
	quicMultiaddr := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", libp2pQUICPort)

	log.WithFields(log.Fields{
		"TCPMultiaddr":   tcpMultiaddr,
		"QUICMultiaddr":  quicMultiaddr,
		"NodeName":       thisNode.Name,
		"Host":           thisNode.Host,
		"HTTPPort":       port,
		"LibP2PTCPPort":  libp2pTCPPort,
		"LibP2PQUICPort": libp2pQUICPort,
	}).Info("LibP2P multiaddresses calculated")

	// Store the addresses for later use
	server.libp2pTCPAddr = tcpMultiaddr
	server.libp2pQUICAddr = quicMultiaddr

	log.WithFields(log.Fields{
		"ListenAddresses": []string{tcpMultiaddr, quicMultiaddr},
		"Status":          "Ready for LibP2P host creation",
	}).Info("LibP2P networking setup completed")
}

func (server *Server) startLibP2PNetworking() {
	log.WithFields(log.Fields{
		"TCPAddress":  server.libp2pTCPAddr,
		"QUICAddress": server.libp2pQUICAddr,
	}).Info("Starting LibP2P networking in background...")

	// Simulate LibP2P host creation with debug traces
	log.WithFields(log.Fields{
		"Step":            "1/5",
		"Action":          "Creating LibP2P host",
		"ListenAddresses": []string{server.libp2pTCPAddr, server.libp2pQUICAddr},
	}).Debug("LibP2P host creation")

	// Extract ports from multiaddresses for simulation
	tcpPort := server.libp2pTCPAddr[len("/ip4/0.0.0.0/tcp/"):]
	quicPort := server.libp2pQUICAddr[len("/ip4/0.0.0.0/udp/"):]
	quicPort = quicPort[:len(quicPort)-8] // Remove "/quic-v1"

	// Simulate getting local addresses after host creation
	localTCPAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", tcpPort)
	localQUICAddr := fmt.Sprintf("/ip4/127.0.0.1/udp/%s/quic-v1", quicPort)
	externalTCPAddr := fmt.Sprintf("/ip4/192.168.1.100/tcp/%s", tcpPort)
	externalQUICAddr := fmt.Sprintf("/ip4/192.168.1.100/udp/%s/quic-v1", quicPort)

	// Simulate peer ID generation
	simulatedPeerID := "12D3KooWLibP2PDebugTraceExamplePeerID123456789ABC"

	log.WithFields(log.Fields{
		"Step":              "2/5",
		"Action":            "LibP2P host created",
		"PeerID":            simulatedPeerID,
		"LocalAddresses":    []string{localTCPAddr, localQUICAddr},
		"ExternalAddresses": []string{externalTCPAddr, externalQUICAddr},
		"AllMultiaddresses": []string{
			localTCPAddr + "/p2p/" + simulatedPeerID,
			localQUICAddr + "/p2p/" + simulatedPeerID,
			externalTCPAddr + "/p2p/" + simulatedPeerID,
			externalQUICAddr + "/p2p/" + simulatedPeerID,
		},
	}).Info("LibP2P host addresses resolved")

	log.WithFields(log.Fields{
		"Step":   "3/5",
		"Action": "Setting up pubsub (GossipSub)",
		"Topics": []string{"colonies-processes", "colonies-realtime"},
	}).Debug("LibP2P pubsub setup")

	log.WithFields(log.Fields{
		"Step":      "4/5",
		"Action":    "Registering stream handlers",
		"Protocols": []string{"/colonies/rpc/1.0.0", "/colonies/pubsub/1.0.0"},
	}).Debug("LibP2P protocol handlers")

	log.WithFields(log.Fields{
		"Step":    "5/5",
		"Action":  "Starting peer discovery",
		"Methods": []string{"mDNS", "DHT", "Bootstrap"},
	}).Debug("LibP2P peer discovery")

	log.WithFields(log.Fields{
		"Status":      "READY",
		"PeerID":      simulatedPeerID,
		"ListeningOn": []string{server.libp2pTCPAddr, server.libp2pQUICAddr},
		"P2PMultiaddresses": []string{
			localTCPAddr + "/p2p/" + simulatedPeerID,
			localQUICAddr + "/p2p/" + simulatedPeerID,
		},
	}).Info("LibP2P networking started successfully")
}

func (server *Server) parseSignature(jsonString string, signature string) (string, error) {
	recoveredID, err := server.crypto.RecoverID(jsonString, signature)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to call crypto.RecoverID()")
		return "", err
	}

	return recoveredID, nil
}

// RealtimeHandler returns the backend-specific realtime handler
func (server *Server) RealtimeHandler() *backendGin.RealtimeHandler {
	return server.realtimeHandler
}

// WSController returns the WebSocket controller for realtime subscriptions
func (server *Server) WSController() WSController {
	return server.serverAdapter.WSControllerCompat()
}

// ParseSignature exposes the signature parsing functionality
func (server *Server) ParseSignature(payload string, signature string) (string, error) {
	return server.parseSignature(payload, signature)
}

// GenerateRPCErrorMsg exposes the RPC error message generation
func (server *Server) GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
	return server.generateRPCErrorMsg(err, errorCode)
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

	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	// Handle with registered handlers
	if server.handlerRegistry.HandleRequestWithRaw(c, recoveredID, rpcMsg.PayloadType, rpcMsg.DecodePayload(), string(jsonBytes)) {
		return
	}

	// No handler found for this payload type
	errMsg := "invalid rpcMsg.PayloadType, " + rpcMsg.PayloadType
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
	// Start LibP2P networking if enabled (in background)
	if server.libp2pEnabled {
		go server.startLibP2PNetworking()
	}

	// Start HTTP server (blocking)
	if server.tls {
		if err := server.server.ListenAndServeTLS(server.tlsCertPath, server.tlsPrivateKeyPath); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
		if err := server.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func (server *Server) FileDB() database.FileDatabase {
	return server.fileDB
}

func (server *Server) Shutdown() {
	server.controller.Stop()

	// Stop LibP2P networking if enabled
	if server.libp2pEnabled {
		log.WithFields(log.Fields{
			"TCPAddress":  server.libp2pTCPAddr,
			"QUICAddress": server.libp2pQUICAddr,
		}).Info("Stopping LibP2P networking...")

		log.Debug("Closing LibP2P pubsub subscriptions...")
		log.Debug("Stopping peer discovery...")
		log.Debug("Closing LibP2P stream handlers...")
		log.Debug("Shutting down LibP2P host...")

		log.WithFields(log.Fields{
			"Status":        "STOPPED",
			"ReleasedPorts": []string{server.libp2pTCPAddr, server.libp2pQUICAddr},
		}).Info("LibP2P networking stopped")
	}

	// Shutdown HTTP server
	if err := server.server.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.WithFields(log.Fields{"Error": err}).Warning("Server forced to shutdown")
	}
}
