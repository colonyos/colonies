package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// GRPCManagedServer wraps the existing Server to make it compatible with ServerManager
type GRPCManagedServer struct {
	server  *Server
	config  *ServerConfig
	mu      sync.RWMutex
	running bool
}

// NewGRPCManagedServer creates a new gRPC managed server
func NewGRPCManagedServer(config *ServerConfig, sharedResources *SharedResources) (*GRPCManagedServer, error) {
	if config.BackendType != GRPCBackendType {
		return nil, fmt.Errorf("invalid backend type for gRPC server: %s", config.BackendType)
	}

	if config.GRPCConfig == nil {
		return nil, errors.New("gRPC config is required for gRPC backend")
	}

	// Create the underlying server using the existing CreateServerWithBackendType function
	server := CreateServerWithBackendType(
		sharedResources.DB,
		config.Port,
		config.TLS,
		config.TLSPrivateKeyPath,
		config.TLSCertPath,
		sharedResources.ThisNode,
		sharedResources.ClusterConfig,
		sharedResources.EtcdDataPath,
		sharedResources.GeneratorPeriod,
		sharedResources.CronPeriod,
		config.ExclusiveAssign,
		config.AllowExecutorReregister,
		config.Retention,
		config.RetentionPolicy,
		config.RetentionPeriod,
		GRPCBackendType,
		nil, // libp2pConfig
		config.GRPCConfig,
		nil, // coapConfig
	)

	return &GRPCManagedServer{
		server: server,
		config: config,
	}, nil
}

// Start starts the gRPC server
func (gms *GRPCManagedServer) Start() error {
	gms.mu.Lock()
	defer gms.mu.Unlock()

	if gms.running {
		return errors.New("gRPC server is already running")
	}

	// Start the server in a goroutine
	go func() {
		log.WithFields(log.Fields{
			"BackendType": GRPCBackendType,
			"Port":        gms.config.GRPCConfig.Port,
			"TLS":         !gms.config.GRPCConfig.Insecure,
		}).Info("Starting gRPC server")

		gms.mu.Lock()
		gms.running = true
		gms.mu.Unlock()

		err := gms.server.ServeForever()

		gms.mu.Lock()
		gms.running = false
		gms.mu.Unlock()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithFields(log.Fields{
				"BackendType": GRPCBackendType,
				"Error":       err,
			}).Error("gRPC server stopped with error")
		} else {
			log.WithField("BackendType", GRPCBackendType).Info("gRPC server stopped")
		}
	}()

	// Wait a moment to ensure server started
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop stops the gRPC server gracefully
func (gms *GRPCManagedServer) Stop(ctx context.Context) error {
	gms.mu.RLock()
	if !gms.running {
		gms.mu.RUnlock()
		return nil
	}
	gms.mu.RUnlock()

	log.WithField("BackendType", GRPCBackendType).Info("Stopping gRPC server")

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		gms.server.Shutdown()
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to shutdown gRPC server: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("gRPC server shutdown timed out: %w", ctx.Err())
	}
}

// GetBackendType returns the backend type
func (gms *GRPCManagedServer) GetBackendType() BackendType {
	return GRPCBackendType
}

// GetPort returns the server port
func (gms *GRPCManagedServer) GetPort() int {
	if gms.config.GRPCConfig != nil {
		return gms.config.GRPCConfig.Port
	}
	return gms.config.Port
}

// GetAddr returns the server address
func (gms *GRPCManagedServer) GetAddr() string {
	return fmt.Sprintf(":%d", gms.GetPort())
}

// IsRunning returns whether the server is running
func (gms *GRPCManagedServer) IsRunning() bool {
	gms.mu.RLock()
	defer gms.mu.RUnlock()
	return gms.running
}

// HealthCheck performs a health check on the server
func (gms *GRPCManagedServer) HealthCheck() error {
	if !gms.IsRunning() {
		return errors.New("gRPC server is not running")
	}

	// For gRPC, we could implement a proper health check using gRPC health protocol
	// For now, just check if the server is running
	return nil
}

// GetServer returns the underlying server (for compatibility)
func (gms *GRPCManagedServer) GetServer() *Server {
	return gms.server
}

// GRPCBackendFactory creates gRPC managed servers
type GRPCBackendFactory struct{}

// NewGRPCBackendFactory creates a new gRPC backend factory
func NewGRPCBackendFactory() *GRPCBackendFactory {
	return &GRPCBackendFactory{}
}

// CreateServer creates a new gRPC managed server
func (gbf *GRPCBackendFactory) CreateServer(config *ServerConfig, sharedResources *SharedResources) (ManagedServer, error) {
	return NewGRPCManagedServer(config, sharedResources)
}

// GetBackendType returns the backend type this factory creates
func (gbf *GRPCBackendFactory) GetBackendType() BackendType {
	return GRPCBackendType
}
