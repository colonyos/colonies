package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
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
	libp2pHost     host.Host
	libp2pDHT      *dht.IpfsDHT
	libp2pCtx      context.Context
	libp2pCancel   context.CancelFunc
	libp2pTCPAddr  string // TCP multiaddress
	libp2pQUICAddr string // QUIC multiaddress
}

// GetBackendTypeFromEnv returns the backend type from environment variables
// Supports: COLONIES_BACKEND_TYPE environment variable
// Valid values: "gin", "libp2p"
// Default: "gin"
func GetBackendTypeFromEnv() BackendType {
	backendEnv := strings.ToLower(os.Getenv("COLONIES_BACKEND_TYPE"))

	var backendType BackendType
	switch backendEnv {
	case "gin", "":
		backendType = GinBackendType
	case "libp2p":
		backendType = LibP2PBackendType
	default:
		log.WithField("COLONIES_BACKEND_TYPE", backendEnv).Warn("Unknown backend type, defaulting to Gin")
		backendType = GinBackendType
	}

	log.WithFields(log.Fields{
		"COLONIES_BACKEND_TYPE": backendEnv,
		"SelectedBackend":       backendType,
	}).Info("Backend type determined from environment")

	return backendType
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
	log.WithFields(log.Fields{
		"BackendType": backendType,
		"HTTPPort":    port,
		"LibP2PPort":  os.Getenv("COLONIES_LIBP2P_PORT"),
	}).Info("=== INITIALIZING COLONIES SERVER BACKEND ===")

	switch backendType {
	case GinBackendType:
		log.WithField("BackendType", backendType).Info("✓ Creating server with Gin HTTP backend")
		server.backend = gin.NewCORSBackend()
		server.engine = server.backend.NewEngineWithDefaults()
		// Add CORS middleware
		server.engine.Use(server.backend.CORS())
		server.server = server.backend.NewServer(port, server.engine)
		server.libp2pEnabled = false
		log.Info("✓ Gin HTTP backend initialized successfully")
	case LibP2PBackendType:
		log.WithFields(log.Fields{
			"BackendType":      backendType,
			"HTTPPort":         port,
			"LibP2PPort":       os.Getenv("COLONIES_LIBP2P_PORT"),
			"LibP2PIdentitySet": os.Getenv("COLONIES_LIBP2P_IDENTITY") != "",
		}).Info("✓ Creating server with LibP2P backend (with HTTP fallback)")
		// LibP2P backend also initializes HTTP endpoints for compatibility
		// This allows existing clients to continue working while adding P2P capabilities
		server.backend = gin.NewCORSBackend()
		server.engine = server.backend.NewEngineWithDefaults()
		server.engine.Use(server.backend.CORS())
		server.server = server.backend.NewServer(port, server.engine)
		// Enable LibP2P networking in addition to HTTP
		server.libp2pEnabled = true
		log.Info("✓ LibP2P backend initialized successfully (HTTP + P2P)")
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
		if err := server.setupLibP2P(port, thisNode); err != nil {
			log.WithError(err).Fatal("Failed to setup LibP2P")
		}
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
	// Setup HTTP routes for both Gin and LibP2P backends
	if server.engine != nil {
		server.engine.POST("/api", server.handleAPIRequest)
		server.engine.GET("/health", server.handleHealthRequest)
		// Note: realtime handler now uses backend abstraction (but maintains /pubsub endpoint for compatibility)
		server.engine.GET("/pubsub", func(c backends.Context) {
			server.realtimeHandlers.HandleWSRequest(c)
		})

		if server.libp2pEnabled {
			log.Info("HTTP routes configured for LibP2P backend (with HTTP compatibility)")
		} else {
			log.Info("HTTP routes configured for Gin backend")
		}
	}
}

func (server *Server) setupLibP2P(port int, thisNode cluster.Node) error {
	log.WithField("Port", port).Info("Initializing LibP2P networking")

	// Get LibP2P port from environment
	libp2pPortStr := os.Getenv("COLONIES_LIBP2P_PORT")
	if libp2pPortStr == "" {
		err := errors.New("COLONIES_LIBP2P_PORT environment variable must be set for LibP2P backend")
		log.Error(err)
		return err
	}

	libp2pPort, err := strconv.Atoi(libp2pPortStr)
	if err != nil {
		log.WithError(err).Error("Failed to parse COLONIES_LIBP2P_PORT")
		return err
	}

	// Create context for LibP2P
	server.libp2pCtx, server.libp2pCancel = context.WithCancel(context.Background())

	// Build libp2p options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", libp2pPort),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", libp2pPort+1),
		),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	}

	log.WithFields(log.Fields{
		"HTTPPort":   port,
		"LibP2PPort": libp2pPort,
		"QUICPort":   libp2pPort + 1,
	}).Info("LibP2P port configuration")

	// Check for predefined identity from environment
	if identityKey := os.Getenv("COLONIES_LIBP2P_IDENTITY"); identityKey != "" {
		// Decode hex string to bytes
		keyBytes, err := hex.DecodeString(identityKey)
		if err != nil {
			log.WithError(err).Error("Failed to decode LibP2P identity hex string")
			return err
		}

		// Unmarshal the private key
		privKey, err := libp2pcrypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal LibP2P private key")
			return err
		}

		opts = append(opts, libp2p.Identity(privKey))
		log.Info("Using predefined LibP2P identity from COLONIES_LIBP2P_IDENTITY")
	}

	// Create libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		log.WithError(err).Error("Failed to create libp2p host")
		return err
	}

	server.libp2pHost = h
	server.libp2pTCPAddr = fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", libp2pPort)
	server.libp2pQUICAddr = fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", libp2pPort+1)

	// Set protocol handler for RPC requests
	h.SetStreamHandler(protocol.ID("/colonies/rpc/1.0.0"), server.handleLibP2PStream)

	// Initialize DHT for peer discovery
	kadDHT, err := dht.New(server.libp2pCtx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		h.Close()
		log.WithError(err).Error("Failed to create DHT")
		return err
	}
	server.libp2pDHT = kadDHT

	// Bootstrap the DHT
	if err = kadDHT.Bootstrap(server.libp2pCtx); err != nil {
		h.Close()
		log.WithError(err).Error("Failed to bootstrap DHT")
		return err
	}

	log.WithFields(log.Fields{
		"PeerID": h.ID().String(),
		"Addrs":  h.Addrs(),
	}).Info("LibP2P host created successfully with DHT")

	// Connect to bootstrap peers if specified
	if bootstrapPeers := os.Getenv("COLONIES_LIBP2P_BOOTSTRAP_PEERS"); bootstrapPeers != "" {
		go server.connectToBootstrapPeers(bootstrapPeers)
	} else {
		log.Info("No bootstrap peers configured (COLONIES_LIBP2P_BOOTSTRAP_PEERS not set)")
		log.Info("Server will be discoverable via DHT but won't proactively connect to other peers")
	}

	// Start DHT advertisement
	go server.advertiseSelfInDHT()

	return nil
}

// connectToBootstrapPeers connects to the specified bootstrap peers for DHT bootstrapping
func (server *Server) connectToBootstrapPeers(bootstrapPeers string) {
	log.WithField("BootstrapPeers", bootstrapPeers).Info("Connecting to bootstrap peers...")

	// Parse comma-separated multiaddresses
	peers := strings.Split(bootstrapPeers, ",")
	successCount := 0
	failCount := 0

	for _, peerAddr := range peers {
		peerAddr = strings.TrimSpace(peerAddr)
		if peerAddr == "" {
			continue
		}

		// Parse the multiaddress
		maddr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			log.WithFields(log.Fields{
				"Address": peerAddr,
				"Error":   err,
			}).Warn("Failed to parse bootstrap peer address")
			failCount++
			continue
		}

		// Extract peer info
		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.WithFields(log.Fields{
				"Address": peerAddr,
				"Error":   err,
			}).Warn("Failed to extract peer info from address")
			failCount++
			continue
		}

		// Connect to the peer
		if err := server.libp2pHost.Connect(server.libp2pCtx, *peerInfo); err != nil {
			log.WithFields(log.Fields{
				"PeerID": peerInfo.ID.String(),
				"Addrs":  peerInfo.Addrs,
				"Error":  err,
			}).Warn("Failed to connect to bootstrap peer")
			failCount++
		} else {
			log.WithFields(log.Fields{
				"PeerID": peerInfo.ID.String(),
				"Addrs":  peerInfo.Addrs,
			}).Info("Successfully connected to bootstrap peer")
			successCount++
		}
	}

	log.WithFields(log.Fields{
		"TotalPeers":      len(peers),
		"SuccessfulConns": successCount,
		"FailedConns":     failCount,
	}).Info("Bootstrap peer connection completed")
}

// advertiseSelfInDHT advertises this server in the DHT for peer discovery
func (server *Server) advertiseSelfInDHT() {
	if server.libp2pDHT == nil {
		log.Warn("DHT not initialized, skipping advertisement")
		return
	}

	// Create a routing discovery service
	routingDiscovery := routing.NewRoutingDiscovery(server.libp2pDHT)

	// Advertise with a rendezvous string
	rendezvous := "colonies-server"
	log.WithField("Rendezvous", rendezvous).Info("Starting DHT advertisement...")

	// Initial advertisement
	util.Advertise(server.libp2pCtx, routingDiscovery, rendezvous)
	log.WithFields(log.Fields{
		"Rendezvous": rendezvous,
		"PeerID":     server.libp2pHost.ID().String(),
	}).Info("Initial DHT advertisement completed")

	// Re-advertise periodically (every 30 minutes)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			util.Advertise(server.libp2pCtx, routingDiscovery, rendezvous)
			log.WithFields(log.Fields{
				"Rendezvous": rendezvous,
				"PeerID":     server.libp2pHost.ID().String(),
			}).Debug("Re-advertised in DHT")
		case <-server.libp2pCtx.Done():
			log.Info("DHT advertisement goroutine stopped")
			return
		}
	}
}

func (server *Server) startLibP2PNetworking() {
	if server.libp2pHost == nil {
		log.Warn("LibP2P host not initialized, skipping LibP2P networking start")
		return
	}

	log.WithFields(log.Fields{
		"PeerID": server.libp2pHost.ID().String(),
		"Addrs":  server.libp2pHost.Addrs(),
	}).Info("LibP2P server listening and ready for connections")

	// Keep the goroutine alive
	<-server.libp2pCtx.Done()
	log.Info("LibP2P networking goroutine stopped")
}

func (server *Server) handleLibP2PStream(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer().String()
	log.WithFields(log.Fields{
		"PeerID":   peerID,
		"Protocol": "/colonies/rpc/1.0.0",
	}).Debug("Handling LibP2P RPC stream")

	// Read the incoming message
	buf := make([]byte, 65536) // 64KB buffer
	n, err := stream.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read from LibP2P stream")
		return
	}

	jsonBytes := buf[:n]
	log.WithFields(log.Fields{
		"PeerID":      peerID,
		"MessageSize": n,
	}).Debug("Received LibP2P RPC message")

	// Parse and handle the RPC message (same logic as HTTP)
	rpcMsg, err := rpc.CreateRPCMsgFromJSON(string(jsonBytes))
	if err != nil {
		log.WithError(err).Error("Failed to parse RPC message from LibP2P stream")
		server.sendLibP2PError(stream, err, http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"PayloadType": rpcMsg.PayloadType,
		"PeerID":      peerID,
	}).Debug("Processing LibP2P RPC message")

	// Handle version requests without signature validation
	if rpcMsg.PayloadType == rpc.VersionPayloadType {
		server.processLibP2PVersionRequest(stream)
		return
	}

	// Validate signature
	recoveredID, err := server.parseSignature(rpcMsg.Payload, rpcMsg.Signature)
	if err != nil {
		log.WithError(err).Error("Failed to validate signature for LibP2P message")
		server.sendLibP2PError(stream, err, http.StatusForbidden)
		return
	}

	// Create a simple response writer that captures the RPC response
	responseCapture := &libp2pResponseWriter{data: make(map[string]interface{})}

	// Try to handle with registered handlers
	handled := server.handlerRegistry.HandleRequestWithRaw(
		responseCapture,
		recoveredID,
		rpcMsg.PayloadType,
		rpcMsg.DecodePayload(),
		string(jsonBytes),
	)

	if !handled {
		errMsg := "invalid rpcMsg.PayloadType: " + rpcMsg.PayloadType
		log.Error(errMsg)
		server.sendLibP2PError(stream, errors.New(errMsg), http.StatusForbidden)
		return
	}

	// Send the response back through the LibP2P stream
	if responseCapture.response != "" {
		_, err = stream.Write([]byte(responseCapture.response))
		if err != nil {
			log.WithError(err).Error("Failed to write response to LibP2P stream")
		}
	}
}

// libp2pResponseWriter is a minimal implementation of backends.Context for LibP2P streams
type libp2pResponseWriter struct {
	statusCode int
	response   string
	data       map[string]interface{}
	aborted    bool
}

func (w *libp2pResponseWriter) String(code int, format string, values ...interface{}) {
	w.statusCode = code
	w.response = fmt.Sprintf(format, values...)
}

func (w *libp2pResponseWriter) JSON(code int, obj interface{}) {
	w.statusCode = code
	// This is simplified - in production you'd marshal to JSON
	w.response = fmt.Sprintf("%v", obj)
}

func (w *libp2pResponseWriter) XML(code int, obj interface{}) {
	w.statusCode = code
	w.response = fmt.Sprintf("%v", obj)
}

func (w *libp2pResponseWriter) Data(code int, contentType string, data []byte) {
	w.statusCode = code
	w.response = string(data)
}

func (w *libp2pResponseWriter) Status(code int) {
	w.statusCode = code
}

func (w *libp2pResponseWriter) Request() *http.Request {
	return nil
}

func (w *libp2pResponseWriter) ReadBody() ([]byte, error) {
	return nil, errors.New("ReadBody not supported in LibP2P context")
}

func (w *libp2pResponseWriter) GetHeader(key string) string {
	return ""
}

func (w *libp2pResponseWriter) Header(key, value string) {
	// No-op
}

func (w *libp2pResponseWriter) Param(key string) string {
	return ""
}

func (w *libp2pResponseWriter) Query(key string) string {
	return ""
}

func (w *libp2pResponseWriter) DefaultQuery(key, defaultValue string) string {
	return defaultValue
}

func (w *libp2pResponseWriter) PostForm(key string) string {
	return ""
}

func (w *libp2pResponseWriter) DefaultPostForm(key, defaultValue string) string {
	return defaultValue
}

func (w *libp2pResponseWriter) Bind(obj interface{}) error {
	return errors.New("Bind not supported in LibP2P context")
}

func (w *libp2pResponseWriter) ShouldBind(obj interface{}) error {
	return errors.New("ShouldBind not supported in LibP2P context")
}

func (w *libp2pResponseWriter) BindJSON(obj interface{}) error {
	return errors.New("BindJSON not supported in LibP2P context")
}

func (w *libp2pResponseWriter) ShouldBindJSON(obj interface{}) error {
	return errors.New("ShouldBindJSON not supported in LibP2P context")
}

func (w *libp2pResponseWriter) Set(key string, value interface{}) {
	w.data[key] = value
}

func (w *libp2pResponseWriter) Get(key string) (interface{}, bool) {
	val, exists := w.data[key]
	return val, exists
}

func (w *libp2pResponseWriter) GetString(key string) string {
	if val, exists := w.data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (w *libp2pResponseWriter) GetBool(key string) bool {
	if val, exists := w.data[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (w *libp2pResponseWriter) GetInt(key string) int {
	if val, exists := w.data[key]; exists {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

func (w *libp2pResponseWriter) GetInt64(key string) int64 {
	if val, exists := w.data[key]; exists {
		if i, ok := val.(int64); ok {
			return i
		}
	}
	return 0
}

func (w *libp2pResponseWriter) GetFloat64(key string) float64 {
	if val, exists := w.data[key]; exists {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0.0
}

func (w *libp2pResponseWriter) Abort() {
	w.aborted = true
}

func (w *libp2pResponseWriter) AbortWithStatus(code int) {
	w.statusCode = code
	w.aborted = true
}

func (w *libp2pResponseWriter) AbortWithStatusJSON(code int, jsonObj interface{}) {
	w.statusCode = code
	w.response = fmt.Sprintf("%v", jsonObj)
	w.aborted = true
}

func (w *libp2pResponseWriter) IsAborted() bool {
	return w.aborted
}

func (w *libp2pResponseWriter) Next() {
	// No-op for LibP2P context
}

func (server *Server) sendLibP2PError(stream network.Stream, err error, errorCode int) {
	rpcReplyMsg, genErr := server.generateRPCErrorMsg(err, errorCode)
	if genErr != nil {
		log.WithError(genErr).Error("Failed to generate RPC error message")
		return
	}

	rpcReplyMsgJSONString, genErr := rpcReplyMsg.ToJSON()
	if genErr != nil {
		log.WithError(genErr).Error("Failed to serialize RPC error message")
		return
	}

	_, writeErr := stream.Write([]byte(rpcReplyMsgJSONString))
	if writeErr != nil {
		log.WithError(writeErr).Error("Failed to write error to LibP2P stream")
	}
}

func (server *Server) processLibP2PVersionRequest(stream network.Stream) {
	// Get version info
	buildVersion := os.Getenv("BUILD_VERSION")
	if buildVersion == "" {
		buildVersion = "dev"
	}
	buildTime := os.Getenv("BUILD_TIME")
	if buildTime == "" {
		buildTime = time.Now().Format(time.RFC3339)
	}

	// Convert to JSON (simplified)
	response := fmt.Sprintf(`{"BuildVersion":"%s","BuildTime":"%s"}`, buildVersion, buildTime)

	// Wrap in RPC reply
	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.VersionPayloadType, response)
	if err != nil {
		log.WithError(err).Error("Failed to create version RPC reply")
		server.sendLibP2PError(stream, err, http.StatusInternalServerError)
		return
	}

	rpcReplyMsgJSON, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithError(err).Error("Failed to serialize version RPC reply")
		server.sendLibP2PError(stream, err, http.StatusInternalServerError)
		return
	}

	_, err = stream.Write([]byte(rpcReplyMsgJSON))
	if err != nil {
		log.WithError(err).Error("Failed to write version response to LibP2P stream")
	}

	log.WithFields(log.Fields{
		"BuildVersion": buildVersion,
		"BuildTime":    buildTime,
	}).Debug("Sent version response via LibP2P")
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
	// Start LibP2P networking in background if enabled
	if server.libp2pEnabled {
		log.Info("Starting LibP2P networking in background")
		go server.startLibP2PNetworking()
	}

	// Start HTTP server (blocking) - runs for both Gin and LibP2P backends
	log.Info("Starting HTTP server")
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
	if server.libp2pEnabled && server.libp2pHost != nil {
		log.WithFields(log.Fields{
			"TCPAddress":  server.libp2pTCPAddr,
			"QUICAddress": server.libp2pQUICAddr,
		}).Info("Stopping LibP2P networking...")

		// Cancel context to stop background goroutines
		if server.libp2pCancel != nil {
			server.libp2pCancel()
		}

		// Close LibP2P host
		if err := server.libp2pHost.Close(); err != nil {
			log.WithError(err).Error("Error closing LibP2P host")
		}

		log.WithFields(log.Fields{
			"Status":        "STOPPED",
			"ReleasedPorts": []string{server.libp2pTCPAddr, server.libp2pQUICAddr},
		}).Info("LibP2P networking stopped")
	}

	// Shutdown HTTP server (runs for both backends)
	if server.server != nil {
		if err := server.server.ShutdownWithTimeout(5 * time.Second); err != nil {
			log.WithFields(log.Fields{"Error": err}).Warning("Server forced to shutdown")
		}
	}
}
