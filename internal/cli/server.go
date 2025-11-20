package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
	grpcPkg "github.com/colonyos/colonies/pkg/backends/grpc"
	coapPkg "github.com/colonyos/colonies/pkg/backends/coap"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(chServerIDCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverStatisticsCmd)
	serverCmd.AddCommand(serverAliveCmd)
	rootCmd.AddCommand(serverCmd)

	chServerIDCmd.Flags().StringVarP(&TargetServerID, "serverid", "", "", "Server Id")
	chServerIDCmd.MarkFlagRequired("serverid")

	serverCmd.PersistentFlags().StringVarP(&DBType, "dbtype", "", "postgresql", "Database type (postgresql or velocitydb)")
	serverCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "", "Colonies database host")
	serverCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	serverCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	serverCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	serverCmd.PersistentFlags().StringVarP(&DataDir, "datadir", "", "", "Data directory")
	serverCmd.PersistentFlags().StringVarP(&TLSCert, "tlscert", "", "", "TLS certificate (can also use COLONIES_SERVER_HTTP_TLS_CERT)")
	serverCmd.PersistentFlags().StringVarP(&TLSKey, "tlskey", "", "", "TLS key (can also use COLONIES_SERVER_HTTP_TLS_KEY)")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port (can also use COLONIES_SERVER_HTTP_PORT)")
	serverCmd.PersistentFlags().StringVarP(&EtcdName, "etcdname", "", "etcd", "Etcd name")
	serverCmd.PersistentFlags().StringVarP(&EtcdHost, "etcdhost", "", "0.0.0.0", "Etcd host name")
	serverCmd.PersistentFlags().IntVarP(&EtcdClientPort, "etcdclientport", "", 2379, "Etcd port")
	serverCmd.PersistentFlags().IntVarP(&EtcdPeerPort, "etcdpeerport", "", 2380, "Etcd peer port")
	serverCmd.PersistentFlags().IntVarP(&RelayPort, "relayport", "", 2381, "Colonies server relay port")
	serverCmd.PersistentFlags().StringSliceVarP(&EtcdCluster, "initial-cluster", "", make([]string, 0), "Cluster config, e.g. --etcdcluster server1=localhost:peerport:relayport:apiport,server2=localhost:peerport:relayport:apiport")
	serverCmd.PersistentFlags().StringVarP(&EtcdDataDir, "etcddatadir", "", "", "Etcd data dir")
	serverCmd.PersistentFlags().BoolVarP(&InitDB, "initdb", "", false, "Initialize DB")
	serverCmd.PersistentFlags().BoolVarP(&Insecure, "insecure", "", false, "Disable TLS")

	serverStatusCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	serverStatusCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	serverStatisticsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

// getServerBackendTypeFromEnv reads backend configuration from environment variables
// and returns the appropriate BackendType for server initialization.
// This function encapsulates all environment variable reading for backend configuration.
//
// Supports:
//   - COLONIES_SERVER_BACKENDS (new): "http", "libp2p", "http,libp2p"
//   - COLONIES_BACKEND_TYPE (legacy): "gin", "libp2p"
//
// Default: GinBackendType (HTTP only)
func getServerBackendTypeFromEnv() server.BackendType {
	// Try new COLONIES_SERVER_BACKENDS first
	backendsEnv := os.Getenv("COLONIES_SERVER_BACKENDS")
	if backendsEnv != "" {
		backendType := server.ParseServerBackendsFromEnv(backendsEnv)
		log.WithFields(log.Fields{
			"COLONIES_SERVER_BACKENDS": backendsEnv,
			"SelectedBackend":          backendType,
		}).Info("Backend type determined from COLONIES_SERVER_BACKENDS")
		return backendType
	}

	// Fall back to legacy COLONIES_BACKEND_TYPE
	backendEnv := strings.ToLower(os.Getenv("COLONIES_BACKEND_TYPE"))
	if backendEnv != "" {
		var backendType server.BackendType
		switch backendEnv {
		case "gin":
			backendType = server.GinBackendType
		case "libp2p":
			backendType = server.LibP2PBackendType
		default:
			log.WithField("COLONIES_BACKEND_TYPE", backendEnv).Warn("Unknown backend type, defaulting to Gin")
			backendType = server.GinBackendType
		}

		log.WithFields(log.Fields{
			"COLONIES_BACKEND_TYPE": backendEnv,
			"SelectedBackend":       backendType,
		}).Info("Backend type determined from COLONIES_BACKEND_TYPE (legacy)")
		return backendType
	}

	// Default to HTTP if no environment variable is set
	log.Info("No backend environment variable set, defaulting to HTTP")
	return server.GinBackendType
}

// getLibP2PConfigFromEnv reads LibP2P-specific configuration from environment variables.
// Returns nil if LibP2P is not configured or configuration is incomplete.
//
// Supports both new and legacy environment variable names for backward compatibility:
//   - COLONIES_SERVER_LIBP2P_PORT (new) or COLONIES_LIBP2P_PORT (legacy) - REQUIRED
//   - COLONIES_SERVER_LIBP2P_IDENTITY (new) or COLONIES_LIBP2P_IDENTITY (legacy) - optional
//   - COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS (new) or COLONIES_LIBP2P_BOOTSTRAP_PEERS (legacy) - optional
//   - COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS (new) or COLONIES_LIBP2P_ANNOUNCE_ADDRS (legacy) - optional
//
// Note: Announce addresses are optional. LibP2P will automatically detect external addresses
// using AutoNAT, UPnP/NAT-PMP, and NAT hole punching (DCUtR).
func getLibP2PConfigFromEnv() *server.LibP2PConfig {
	// Try new variable names first, fall back to legacy
	portStr := os.Getenv("COLONIES_SERVER_LIBP2P_PORT")
	if portStr == "" {
		portStr = os.Getenv("COLONIES_LIBP2P_PORT") // Backward compatibility
	}

	if portStr == "" {
		log.Debug("No LibP2P port configured (COLONIES_SERVER_LIBP2P_PORT not set)")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.WithError(err).Error("Failed to parse LibP2P port, LibP2P backend will not be available")
		return nil
	}

	identity := os.Getenv("COLONIES_SERVER_LIBP2P_IDENTITY")
	if identity == "" {
		identity = os.Getenv("COLONIES_LIBP2P_IDENTITY") // Backward compatibility
	}

	bootstrapPeers := os.Getenv("COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS")
	if bootstrapPeers == "" {
		bootstrapPeers = os.Getenv("COLONIES_LIBP2P_BOOTSTRAP_PEERS") // Backward compatibility
	}

	announceAddrs := os.Getenv("COLONIES_SERVER_LIBP2P_ANNOUNCE_ADDRS")
	if announceAddrs == "" {
		announceAddrs = os.Getenv("COLONIES_LIBP2P_ANNOUNCE_ADDRS") // Backward compatibility
	}

	config := &server.LibP2PConfig{
		Port:            port,
		Identity:        identity,
		BootstrapPeers:  bootstrapPeers,
		AnnounceAddrs:   announceAddrs,
	}

	log.WithFields(log.Fields{
		"Port":                port,
		"IdentityConfigured":  identity != "",
		"BootstrapPeersCount": len(strings.Split(bootstrapPeers, ",")),
		"AnnounceAddrs":       announceAddrs,
	}).Info("LibP2P configuration loaded from environment")

	return config
}

// getCoAPConfigFromEnv reads CoAP-specific configuration from environment variables.
// Returns nil if CoAP is not configured or configuration is incomplete.
//
// Supports:
//   - COLONIES_SERVER_COAP_PORT: CoAP server port (required)
func getCoAPConfigFromEnv() *server.CoAPConfig {
	portStr := os.Getenv("COLONIES_SERVER_COAP_PORT")
	if portStr == "" {
		log.Debug("No CoAP port configured (COLONIES_SERVER_COAP_PORT not set)")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.WithError(err).Error("Failed to parse CoAP port, CoAP backend will not be available")
		return nil
	}

	config := &server.CoAPConfig{
		Port: port,
	}

	log.WithFields(log.Fields{
		"Port": port,
	}).Info("CoAP configuration loaded from environment")

	return config
}

// getGRPCConfigFromEnv reads gRPC-specific configuration from environment variables.
// Returns nil if gRPC is not configured or configuration is incomplete.
//
// Supports:
//   - COLONIES_SERVER_GRPC_PORT: gRPC server port (required)
//   - COLONIES_SERVER_GRPC_TLS_CERT: Path to TLS certificate file (optional)
//   - COLONIES_SERVER_GRPC_TLS_KEY: Path to TLS private key file (optional)
//   - COLONIES_SERVER_GRPC_INSECURE: Run without TLS (default: false)
func getGRPCConfigFromEnv() *server.GRPCConfig {
	portStr := os.Getenv("COLONIES_SERVER_GRPC_PORT")
	if portStr == "" {
		log.Debug("No gRPC port configured (COLONIES_SERVER_GRPC_PORT not set)")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.WithError(err).Error("Failed to parse gRPC port, gRPC backend will not be available")
		return nil
	}

	tlsCert := os.Getenv("COLONIES_SERVER_GRPC_TLS_CERT")
	tlsKey := os.Getenv("COLONIES_SERVER_GRPC_TLS_KEY")

	// Use COLONIES_TLS for consistency: "true" = TLS enabled, "false" = no TLS
	tlsStr := os.Getenv("COLONIES_TLS")
	insecure := false
	if tlsStr == "false" {
		insecure = true
	} else if tlsStr == "true" {
		insecure = false
	}

	config := &server.GRPCConfig{
		Port:     port,
		TLSCert:  tlsCert,
		TLSKey:   tlsKey,
		Insecure: insecure,
	}

	log.WithFields(log.Fields{
		"Port":              port,
		"TLSCertConfigured": tlsCert != "",
		"TLSKeyConfigured":  tlsKey != "",
		"Insecure":          insecure,
	}).Info("gRPC configuration loaded from environment")

	return config
}

// startSingleBackendServer starts a server with a single backend (original behavior)
func startSingleBackendServer(
	db database.Database,
	node cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	backendType server.BackendType,
	libp2pConfig *server.LibP2PConfig,
	grpcConfig *server.GRPCConfig,
	coapConfig *server.CoAPConfig,
) {
	retentionPeriod := 60000 // Run retention worker once a minute

	srv := server.CreateServerWithBackendType(
		db,
		ServerPort,
		UseTLS,
		TLSKey,
		TLSCert,
		node,
		clusterConfig,
		etcdDataPath,
		GeneratorCheckerPeriod,
		CronCheckerPeriod,
		ExclusiveAssign,
		AllowExecutorReregister,
		Retention,
		RetentionPolicy,
		retentionPeriod,
		backendType,
		libp2pConfig,
		grpcConfig,
		coapConfig,
	)

	for {
		err := srv.ServeForever()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to start Colonies Server")
			time.Sleep(1 * time.Second)
		}
	}
}

// startHTTPAndGRPCServer starts both HTTP and gRPC servers sharing the same base Server
func startHTTPAndGRPCServer(
	db database.Database,
	node cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	grpcConfig *server.GRPCConfig,
) {
	retentionPeriod := 60000 // Run retention worker once a minute

	// Create ONE HTTP server with all the controller/database logic
	httpServer := server.CreateServerWithBackendType(
		db,
		ServerPort,
		UseTLS,
		TLSKey,
		TLSCert,
		node,
		clusterConfig,
		etcdDataPath,
		GeneratorCheckerPeriod,
		CronCheckerPeriod,
		ExclusiveAssign,
		AllowExecutorReregister,
		Retention,
		RetentionPolicy,
		retentionPeriod,
		server.GinBackendType, // Create as HTTP server
		nil,                   // no libp2p
		nil,                   // no grpc in Server struct
		nil,                   // no coap in Server struct
	)

	// Create a gRPC backend and server that shares the HTTP server's handlers
	grpcBackend := grpcPkg.NewBackend().(*grpcPkg.Backend)
	grpcServer := grpcBackend.NewServerWithAddrAndHandler(
		fmt.Sprintf(":%d", grpcConfig.Port),
		httpServer, // httpServer implements RPCHandler interface via HandleRPC method
	)

	log.WithFields(log.Fields{
		"HTTPPort": ServerPort,
		"GRPCPort": grpcConfig.Port,
	}).Info("Starting HTTP and gRPC servers with shared controller")

	// Start gRPC server in background
	go func() {
		log.WithField("Port", grpcConfig.Port).Info("Starting gRPC server")
		if err := grpcServer.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("gRPC server failed")
		}
	}()

	// Start HTTP server (blocking)
	for {
		err := httpServer.ServeForever()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to start HTTP server")
			time.Sleep(1 * time.Second)
		}
	}
}

// startHTTPGRPCAndLibP2PServer starts HTTP (with LibP2P) and gRPC servers sharing the same base Server
func startHTTPGRPCAndLibP2PServer(
	db database.Database,
	node cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	libp2pConfig *server.LibP2PConfig,
	grpcConfig *server.GRPCConfig,
) {
	retentionPeriod := 60000 // Run retention worker once a minute

	// Create ONE HTTP server with LibP2P enabled (combines HTTP + LibP2P functionality)
	httpServer := server.CreateServerWithBackendType(
		db,
		ServerPort,
		UseTLS,
		TLSKey,
		TLSCert,
		node,
		clusterConfig,
		etcdDataPath,
		GeneratorCheckerPeriod,
		CronCheckerPeriod,
		ExclusiveAssign,
		AllowExecutorReregister,
		Retention,
		RetentionPolicy,
		retentionPeriod,
		server.LibP2PBackendType, // This creates HTTP + LibP2P
		libp2pConfig,
		nil, // no grpc in Server struct
		nil, // no coap in Server struct
	)

	// Create a gRPC backend and server that shares the HTTP server's handlers
	grpcBackend := grpcPkg.NewBackend().(*grpcPkg.Backend)
	grpcServer := grpcBackend.NewServerWithAddrAndHandler(
		fmt.Sprintf(":%d", grpcConfig.Port),
		httpServer, // httpServer implements RPCHandler interface via HandleRPC method
	)

	log.WithFields(log.Fields{
		"HTTPPort":   ServerPort,
		"GRPCPort":   grpcConfig.Port,
		"LibP2PPort": libp2pConfig.Port,
	}).Info("Starting HTTP, gRPC, and LibP2P servers with shared controller")

	// Start gRPC server in background
	go func() {
		log.WithField("Port", grpcConfig.Port).Info("Starting gRPC server")
		if err := grpcServer.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("gRPC server failed")
		}
	}()

	// Start HTTP server with LibP2P (blocking)
	for {
		err := httpServer.ServeForever()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to start HTTP/LibP2P server")
			time.Sleep(1 * time.Second)
		}
	}
}

// startAllBackendsServer starts all four backends: HTTP (with LibP2P), gRPC, and CoAP - all sharing the same base Server
func startAllBackendsServer(
	db database.Database,
	node cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	libp2pConfig *server.LibP2PConfig,
	grpcConfig *server.GRPCConfig,
	coapConfig *server.CoAPConfig,
) {
	retentionPeriod := 60000 // Run retention worker once a minute

	// Create ONE HTTP server with LibP2P enabled (combines HTTP + LibP2P functionality)
	httpServer := server.CreateServerWithBackendType(
		db,
		ServerPort,
		UseTLS,
		TLSKey,
		TLSCert,
		node,
		clusterConfig,
		etcdDataPath,
		GeneratorCheckerPeriod,
		CronCheckerPeriod,
		ExclusiveAssign,
		AllowExecutorReregister,
		Retention,
		RetentionPolicy,
		retentionPeriod,
		server.LibP2PBackendType, // This creates HTTP + LibP2P
		libp2pConfig,
		nil, // no grpc in Server struct
		nil, // no coap in Server struct
	)

	// Create a gRPC backend and server that shares the HTTP server's handlers
	grpcBackend := grpcPkg.NewBackend().(*grpcPkg.Backend)
	grpcServer := grpcBackend.NewServerWithAddrAndHandler(
		fmt.Sprintf(":%d", grpcConfig.Port),
		httpServer, // httpServer implements RPCHandler interface via HandleRPC method
	)

	// Create a CoAP backend and server that shares the HTTP server's handlers
	coapBackend := coapPkg.NewBackend().(*coapPkg.Backend)
	coapServer := coapBackend.NewServerWithAddrAndHandler(
		fmt.Sprintf(":%d", coapConfig.Port),
		httpServer, // httpServer implements RPCHandler interface via HandleRPC method
	)

	log.WithFields(log.Fields{
		"HTTPPort":   ServerPort,
		"GRPCPort":   grpcConfig.Port,
		"CoAPPort":   coapConfig.Port,
		"LibP2PPort": libp2pConfig.Port,
	}).Info("Starting HTTP, gRPC, CoAP, and LibP2P servers with shared controller")

	// Start gRPC server in background
	go func() {
		log.WithField("Port", grpcConfig.Port).Info("Starting gRPC server")
		if err := grpcServer.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("gRPC server failed")
		}
	}()

	// Start CoAP server in background
	go func() {
		log.WithField("Port", coapConfig.Port).Info("Starting CoAP server")
		if err := coapServer.ListenAndServe(); err != nil {
			log.WithError(err).Fatal("CoAP server failed")
		}
	}()

	// Start HTTP server with LibP2P (blocking)
	for {
		err := httpServer.ServeForever()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to start HTTP/LibP2P server")
			time.Sleep(1 * time.Second)
		}
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage production server",
	Long:  "Manage production server",
}

var chServerIDCmd = &cobra.Command{
	Use:   "chid",
	Short: "Change server Id",
	Long:  "Change server Id",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(TargetServerID) != 64 {
			CheckError(errors.New("Invalid server Id length"))
		}

		CheckError(client.ChangeServerID(TargetServerID, ServerPrvKey))

		log.WithFields(log.Fields{
			"ServerId": ServerID}).
			Info("Changed server Id")
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status about a production server",
	Long:  "Show status about a production server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		client := setup()

		// Get server info from the running server via RPC
		serverInfo, err := client.GetServerInfo()
		CheckError(err)

		printServerStatusTable(serverInfo)
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a production server",
	Long:  "Start a production server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseEnv()

		// Read HTTP server configuration from environment if not set via flags
		if ServerPort == -1 {
			if portStr := os.Getenv("COLONIES_SERVER_HTTP_PORT"); portStr != "" {
				port, err := strconv.Atoi(portStr)
				if err == nil {
					ServerPort = port
					log.WithField("Port", ServerPort).Info("Using COLONIES_SERVER_HTTP_PORT from environment")
				}
			}
		}

		if TLSCert == "" {
			if cert := os.Getenv("COLONIES_SERVER_HTTP_TLS_CERT"); cert != "" {
				TLSCert = cert
				log.WithField("Cert", TLSCert).Info("Using COLONIES_SERVER_HTTP_TLS_CERT from environment")
			}
		}

		if TLSKey == "" {
			if key := os.Getenv("COLONIES_SERVER_HTTP_TLS_KEY"); key != "" {
				TLSKey = key
				log.WithField("Key", TLSKey).Info("Using COLONIES_SERVER_HTTP_TLS_KEY from environment")
			}
		}

		// Check TLS setting from environment - use single COLONIES_TLS variable
		if tlsStr := os.Getenv("COLONIES_TLS"); tlsStr != "" {
			if tlsStr == "true" {
				Insecure = false
			} else if tlsStr == "false" {
				Insecure = true
			}
		}

		log.WithFields(log.Fields{
			"Insecure": Insecure,
			"TLSKey":   TLSKey,
			"TLSCert":  TLSCert,
		}).Debug("TLS configuration check")

		if !Insecure {
			log.Info("TLS mode enabled, checking certificates...")
			_, err := os.Stat(TLSKey)
			if err != nil {
				CheckError(errors.New("Failed to load TLS Key: " + TLSKey))
				os.Exit(-1)
			}

			_, err = os.Stat(TLSCert)
			if err != nil {
				CheckError(errors.New("Failed to load TLS Cert: " + TLSCert))
				os.Exit(-1)
			}
		} else {
			log.Info("Insecure mode enabled, skipping TLS certificate checks")
		}

		log.WithFields(log.Fields{
			"DBType":      DBType,
			"DBHost":      DBHost,
			"DBPort":      DBPort,
			"DBName":      DBName,
			"DBPrefix":    DBPrefix,
			"TimescaleDB": TimescaleDB,
			"DataDir":     DataDir,
		}).Info("Server starting with database configuration")

		dbConfig := database.DatabaseConfig{
			Type:        database.DatabaseType(DBType),
			Host:        DBHost,
			Port:        DBPort,
			User:        DBUser,
			Password:    DBPassword,
			Name:        DBName,
			Prefix:      DBPrefix,
			TimescaleDB: TimescaleDB,
			DataDir:     DataDir,
		}

		log.WithField("DatabaseType", DBType).Info("Creating database using factory")
		db, err := database.CreateDatabase(dbConfig)
		CheckError(err)

		// Connect to database (only needed for PostgreSQL)
		if DBType == "postgresql" {
			log.Info("Connecting to PostgreSQL database...")
			for {
				if postgresDB, ok := db.(*postgresql.PQDatabase); ok {
					err := postgresDB.Connect()
					if err != nil {
						log.WithFields(log.Fields{"Error": err}).Error("Failed to connect to PostgreSQL database, retrying...")
						time.Sleep(1 * time.Second)
					} else {
						log.Info("Successfully connected to PostgreSQL database")
						break
					}
				}
			}
		}

		log.WithFields(log.Fields{
			"DatabaseType": DBType,
			"Host":         DBHost,
			"Port":         DBPort,
			"User":         DBUser,
			"Password":     "*******************",
			"Name":         DBName,
			"UseTLS":       UseTLS,
			"TimescaleDB":  TimescaleDB,
		}).Info("Database connection established")

		node := cluster.Node{Name: EtcdName, Host: EtcdHost, APIPort: ServerPort, EtcdClientPort: EtcdClientPort, EtcdPeerPort: EtcdPeerPort, RelayPort: RelayPort}
		clusterConfig := cluster.Config{}

		if len(EtcdCluster) > 0 {
			// Parse EtcdCluster flag
			errMsg := "Invalid cluster, try e.g. --etcdcluster server1=localhost:23100:25100:26100,server2=localhost:23101:25101:26101"
			for _, s := range EtcdCluster {
				split1 := strings.Split(s, "=")
				if len(split1) != 2 {
					CheckError(errors.New(errMsg))
				}
				name := split1[0]
				split2 := strings.Split(split1[1], ":")
				if len(split2) != 4 {
					CheckError(errors.New(errMsg))
				}
				host := split2[0]
				portStr1 := split2[1]
				etcPeerPort, err := strconv.Atoi(portStr1)
				CheckError(err)
				portStr2 := split2[2]
				relayPort, err := strconv.Atoi(portStr2)
				CheckError(err)
				portStr3 := split2[3]
				apiPort, err := strconv.Atoi(portStr3)
				CheckError(err)
				node := cluster.Node{Name: name, Host: host, EtcdClientPort: EtcdClientPort, EtcdPeerPort: etcPeerPort, RelayPort: relayPort, APIPort: apiPort}
				clusterConfig.AddNode(node)
			}
		} else {
			clusterConfig.AddNode(node)
		}

		if EtcdDataDir == "" {
			EtcdDataDir = "/tmp/colonies/prod/etcd"
			log.Warning("EtcdDataDir not specified, setting it to " + EtcdDataDir)
		}

		setupProfiler()

		// Determine backend type from environment
		backendType := getServerBackendTypeFromEnv()
		log.WithField("BackendType", backendType).Info("Server backend type determined from environment")

		// Get LibP2P configuration if LibP2P backend is selected
		var libp2pConfig *server.LibP2PConfig
		if backendType == server.LibP2PBackendType || backendType == server.HTTPGRPCLibP2PBackendType || backendType == server.HTTPGRPCLibP2PCoAPBackendType {
			libp2pConfig = getLibP2PConfigFromEnv()
			if libp2pConfig == nil {
				CheckError(errors.New("LibP2P backend selected but LibP2P configuration is incomplete (COLONIES_SERVER_LIBP2P_PORT must be set)"))
			}
		}

		// Get gRPC configuration if needed
		var grpcConfig *server.GRPCConfig
		if backendType == server.GRPCBackendType || backendType == server.HTTPGRPCBackendType || backendType == server.HTTPGRPCLibP2PBackendType || backendType == server.HTTPGRPCLibP2PCoAPBackendType {
			grpcConfig = getGRPCConfigFromEnv()
			if grpcConfig == nil {
				CheckError(errors.New("gRPC backend selected but gRPC configuration is incomplete (COLONIES_SERVER_GRPC_PORT must be set)"))
			}
		}

		// Get CoAP configuration if needed
		var coapConfig *server.CoAPConfig
		if backendType == server.CoAPBackendType || backendType == server.HTTPGRPCLibP2PCoAPBackendType {
			coapConfig = getCoAPConfigFromEnv()
			if coapConfig == nil {
				CheckError(errors.New("CoAP backend selected but CoAP configuration is incomplete (COLONIES_SERVER_COAP_PORT must be set)"))
			}
		}

		if InitDB {
			err := db.Initialize()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to call db.Initialize()")
			} else {
				log.WithFields(log.Fields{"ServerID": ServerID}).Info("Setting server ID")
				CheckError(db.SetServerID("", ServerID))

				log.Info("Colonies database initialized")
			}
		}

		// Route to appropriate startup function based on backend type
		if backendType == server.HTTPGRPCLibP2PCoAPBackendType {
			log.Info("Starting server with all four backends (HTTP + gRPC + LibP2P + CoAP)")
			startAllBackendsServer(db, node, clusterConfig, EtcdDataDir, libp2pConfig, grpcConfig, coapConfig)
		} else if backendType == server.HTTPGRPCLibP2PBackendType {
			log.Info("Starting server with all three backends (HTTP + gRPC + LibP2P)")
			startHTTPGRPCAndLibP2PServer(db, node, clusterConfig, EtcdDataDir, libp2pConfig, grpcConfig)
		} else if backendType == server.HTTPGRPCBackendType {
			log.Info("Starting server with multiple backends (HTTP + gRPC)")
			startHTTPAndGRPCServer(db, node, clusterConfig, EtcdDataDir, grpcConfig)
		} else {
			// Single backend - use existing approach
			log.WithField("BackendType", backendType).Info("Starting server with single backend")
			startSingleBackendServer(db, node, clusterConfig, EtcdDataDir, backendType, libp2pConfig, grpcConfig, coapConfig)
		}
	},
}

var serverStatisticsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show server statistics",
	Long:  "Show server statistics",
	Run: func(cmd *cobra.Command, args []string) {
		parseEnv()

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		stat, err := client.Statistics(ServerPrvKey)
		CheckError(err)

		printServerStatTable(stat)
	},
}

var serverAliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Check if a server is alive",
	Long:  "Check if a server is alive",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(0)
	},
}

