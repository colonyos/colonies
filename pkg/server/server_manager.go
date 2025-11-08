package server

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/server/controllers"
	log "github.com/sirupsen/logrus"
)

// BackendType represents different server backend types
type BackendType string

const (
	GinBackendType                   BackendType = "gin"
	LibP2PBackendType                BackendType = "libp2p"
	GRPCBackendType                  BackendType = "grpc"
	CoAPBackendType                  BackendType = "coap"
	HTTPGRPCBackendType              BackendType = "http+grpc"
	HTTPGRPCLibP2PBackendType        BackendType = "http+grpc+libp2p"
	HTTPGRPCLibP2PCoAPBackendType    BackendType = "http+grpc+libp2p+coap"
)

// ManagedServer represents a server instance managed by ServerManager
type ManagedServer interface {
	// Lifecycle management
	Start() error
	Stop(ctx context.Context) error
	
	// Server info
	GetBackendType() BackendType
	GetPort() int
	GetAddr() string
	IsRunning() bool
	
	// Health checks
	HealthCheck() error
}

// ServerConfig holds configuration for a managed server
type ServerConfig struct {
	BackendType             BackendType
	Port                   int
	LibP2PPort             int  // Port for LibP2P transport (required for LibP2P backend)
	TLS                    bool
	TLSPrivateKeyPath      string
	TLSCertPath            string
	GRPCConfig             *GRPCConfig  // gRPC-specific configuration (required for gRPC backend)
	ExclusiveAssign        bool
	AllowExecutorReregister bool
	Retention              bool
	RetentionPolicy        int64
	RetentionPeriod        int
	Enabled                bool
}

// ServerManager manages multiple server backends
type ServerManager struct {
	// Shared services
	db              database.Database
	thisNode        cluster.Node
	clusterConfig   cluster.Config
	etcdDataPath    string
	generatorPeriod int
	cronPeriod      int
	
	// Managed servers
	servers    map[BackendType]ManagedServer
	configs    map[BackendType]*ServerConfig
	mu         sync.RWMutex
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	
	// Backend factories
	backendFactories map[BackendType]BackendFactory
}

// BackendFactory creates backend-specific servers
type BackendFactory interface {
	CreateServer(config *ServerConfig, sharedResources *SharedResources) (ManagedServer, error)
	GetBackendType() BackendType
}

// SharedResources contains services shared between all server backends
type SharedResources struct {
	DB              database.Database
	ThisNode        cluster.Node
	ClusterConfig   cluster.Config
	EtcdDataPath    string
	GeneratorPeriod int
	CronPeriod      int
	Controller      controllers.Controller  // Shared controller to avoid etcd port conflicts
	BaseServer      *Server                  // Shared base server for handler registration
}

// NewServerManager creates a new ServerManager
func NewServerManager(
	db database.Database,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int,
) *ServerManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ServerManager{
		db:              db,
		thisNode:        thisNode,
		clusterConfig:   clusterConfig,
		etcdDataPath:    etcdDataPath,
		generatorPeriod: generatorPeriod,
		cronPeriod:      cronPeriod,
		servers:         make(map[BackendType]ManagedServer),
		configs:         make(map[BackendType]*ServerConfig),
		ctx:             ctx,
		cancel:          cancel,
		backendFactories: make(map[BackendType]BackendFactory),
	}
}

// RegisterBackendFactory registers a factory for creating backend servers
func (sm *ServerManager) RegisterBackendFactory(factory BackendFactory) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return errors.New("cannot register backend factory while server manager is running")
	}
	
	backendType := factory.GetBackendType()
	if _, exists := sm.backendFactories[backendType]; exists {
		return fmt.Errorf("backend factory for type %s already registered", backendType)
	}
	
	sm.backendFactories[backendType] = factory
	log.WithField("BackendType", backendType).Info("Backend factory registered")
	return nil
}

// AddServerConfig adds configuration for a server backend
func (sm *ServerManager) AddServerConfig(config *ServerConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return errors.New("cannot add server config while server manager is running")
	}
	
	if !config.Enabled {
		log.WithField("BackendType", config.BackendType).Info("Backend disabled in config")
		return nil
	}
	
	if _, exists := sm.configs[config.BackendType]; exists {
		return fmt.Errorf("server config for backend %s already exists", config.BackendType)
	}
	
	sm.configs[config.BackendType] = config
	log.WithFields(log.Fields{
		"BackendType": config.BackendType,
		"Port":        config.Port,
		"TLS":         config.TLS,
	}).Info("Server config added")
	
	return nil
}

// StartAll starts all configured and enabled servers
func (sm *ServerManager) StartAll() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return errors.New("server manager is already running")
	}
	
	sharedResources := &SharedResources{
		DB:              sm.db,
		ThisNode:        sm.thisNode,
		ClusterConfig:   sm.clusterConfig,
		EtcdDataPath:    sm.etcdDataPath,
		GeneratorPeriod: sm.generatorPeriod,
		CronPeriod:      sm.cronPeriod,
	}
	
	// Create and start servers for each enabled backend
	var errors []error
	for backendType, config := range sm.configs {
		if !config.Enabled {
			continue
		}
		
		factory, exists := sm.backendFactories[backendType]
		if !exists {
			err := fmt.Errorf("no factory registered for backend type %s", backendType)
			errors = append(errors, err)
			log.Error(err)
			continue
		}
		
		server, err := factory.CreateServer(config, sharedResources)
		if err != nil {
			err = fmt.Errorf("failed to create server for backend %s: %w", backendType, err)
			errors = append(errors, err)
			log.Error(err)
			continue
		}
		
		sm.servers[backendType] = server
		
		// Start server in goroutine
		go func(bt BackendType, srv ManagedServer) {
			log.WithField("BackendType", bt).Info("Starting server")
			if err := srv.Start(); err != nil {
				log.WithFields(log.Fields{
					"BackendType": bt,
					"Error":       err,
				}).Error("Failed to start server")
			}
		}(backendType, server)
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to start some servers: %v", errors)
	}
	
	sm.running = true
	log.WithField("ServerCount", len(sm.servers)).Info("ServerManager started")
	
	// Start health check routine
	go sm.healthCheckRoutine()
	
	return nil
}

// StopAll stops all running servers gracefully
func (sm *ServerManager) StopAll(timeout time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.running {
		return nil
	}
	
	// Cancel context to stop health checks
	sm.cancel()
	
	// Stop servers with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	var wg sync.WaitGroup
	var errors []error
	var errorsMu sync.Mutex
	
	for backendType, server := range sm.servers {
		wg.Add(1)
		go func(bt BackendType, srv ManagedServer) {
			defer wg.Done()
			
			log.WithField("BackendType", bt).Info("Stopping server")
			if err := srv.Stop(ctx); err != nil {
				errorsMu.Lock()
				errors = append(errors, fmt.Errorf("failed to stop %s server: %w", bt, err))
				errorsMu.Unlock()
				log.WithFields(log.Fields{
					"BackendType": bt,
					"Error":       err,
				}).Error("Failed to stop server")
			} else {
				log.WithField("BackendType", bt).Info("Server stopped")
			}
		}(backendType, server)
	}
	
	wg.Wait()
	
	sm.servers = make(map[BackendType]ManagedServer)
	sm.running = false
	
	log.Info("ServerManager stopped")
	
	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while stopping servers: %v", errors)
	}
	
	return nil
}

// GetServer returns a managed server by backend type
func (sm *ServerManager) GetServer(backendType BackendType) (ManagedServer, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	server, exists := sm.servers[backendType]
	return server, exists
}

// GetRunningServers returns all currently running servers
func (sm *ServerManager) GetRunningServers() map[BackendType]ManagedServer {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	running := make(map[BackendType]ManagedServer)
	for backendType, server := range sm.servers {
		if server.IsRunning() {
			running[backendType] = server
		}
	}
	
	return running
}

// IsRunning returns whether the server manager is running
func (sm *ServerManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// GetStatus returns status information about all servers
func (sm *ServerManager) GetStatus() map[BackendType]ServerStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	status := make(map[BackendType]ServerStatus)
	
	for backendType, server := range sm.servers {
		var healthError error
		if server.IsRunning() {
			healthError = server.HealthCheck()
		}
		
		status[backendType] = ServerStatus{
			BackendType:  backendType,
			Running:      server.IsRunning(),
			Port:         server.GetPort(),
			Addr:         server.GetAddr(),
			HealthError:  healthError,
		}
	}
	
	return status
}

// ServerStatus represents the status of a managed server
type ServerStatus struct {
	BackendType BackendType
	Running     bool
	Port        int
	Addr        string
	HealthError error
}

// healthCheckRoutine periodically checks server health
func (sm *ServerManager) healthCheckRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.performHealthChecks()
		}
	}
}

// performHealthChecks checks health of all running servers
func (sm *ServerManager) performHealthChecks() {
	sm.mu.RLock()
	servers := make(map[BackendType]ManagedServer)
	for k, v := range sm.servers {
		servers[k] = v
	}
	sm.mu.RUnlock()
	
	for backendType, server := range servers {
		if !server.IsRunning() {
			continue
		}
		
		if err := server.HealthCheck(); err != nil {
			log.WithFields(log.Fields{
				"BackendType": backendType,
				"Error":       err,
			}).Warn("Server health check failed")
		}
	}
}