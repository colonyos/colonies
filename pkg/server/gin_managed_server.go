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

// GinManagedServer wraps the existing Server to make it compatible with ServerManager
type GinManagedServer struct {
	server  *Server
	config  *ServerConfig
	mu      sync.RWMutex
	running bool
}

// NewGinManagedServer creates a new gin managed server
func NewGinManagedServer(config *ServerConfig, sharedResources *SharedResources) (*GinManagedServer, error) {
	if config.BackendType != GinBackendType {
		return nil, fmt.Errorf("invalid backend type for gin server: %s", config.BackendType)
	}
	
	// Create the underlying server using the existing CreateServer function
	server := CreateServer(
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
	)
	
	return &GinManagedServer{
		server: server,
		config: config,
	}, nil
}

// Start starts the gin server
func (gms *GinManagedServer) Start() error {
	gms.mu.Lock()
	defer gms.mu.Unlock()
	
	if gms.running {
		return errors.New("gin server is already running")
	}
	
	// Start the server in a goroutine
	go func() {
		log.WithFields(log.Fields{
			"BackendType": GinBackendType,
			"Port":        gms.config.Port,
			"TLS":         gms.config.TLS,
		}).Info("Starting Gin server")
		
		gms.mu.Lock()
		gms.running = true
		gms.mu.Unlock()
		
		var err error
		if gms.config.TLS {
			err = gms.server.ServeForever()
		} else {
			err = gms.server.ServeForever()
		}
		
		gms.mu.Lock()
		gms.running = false
		gms.mu.Unlock()
		
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithFields(log.Fields{
				"BackendType": GinBackendType,
				"Error":       err,
			}).Error("Gin server stopped with error")
		} else {
			log.WithField("BackendType", GinBackendType).Info("Gin server stopped")
		}
	}()
	
	// Wait a moment to ensure server started
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// Stop stops the gin server gracefully
func (gms *GinManagedServer) Stop(ctx context.Context) error {
	gms.mu.RLock()
	if !gms.running {
		gms.mu.RUnlock()
		return nil
	}
	gms.mu.RUnlock()
	
	log.WithField("BackendType", GinBackendType).Info("Stopping Gin server")
	
	// Create a channel to signal completion
	done := make(chan error, 1)
	
	go func() {
		gms.server.Shutdown()
		done <- nil
	}()
	
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to shutdown gin server: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("gin server shutdown timed out: %w", ctx.Err())
	}
}

// GetBackendType returns the backend type
func (gms *GinManagedServer) GetBackendType() BackendType {
	return GinBackendType
}

// GetPort returns the server port
func (gms *GinManagedServer) GetPort() int {
	return gms.config.Port
}

// GetAddr returns the server address
func (gms *GinManagedServer) GetAddr() string {
	return fmt.Sprintf(":%d", gms.config.Port)
}

// IsRunning returns whether the server is running
func (gms *GinManagedServer) IsRunning() bool {
	gms.mu.RLock()
	defer gms.mu.RUnlock()
	return gms.running
}

// HealthCheck performs a health check on the server
func (gms *GinManagedServer) HealthCheck() error {
	if !gms.IsRunning() {
		return errors.New("gin server is not running")
	}
	
	// Perform a simple HTTP health check
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	protocol := "http"
	if gms.config.TLS {
		protocol = "https"
	}
	
	url := fmt.Sprintf("%s://localhost:%d/health", protocol, gms.config.Port)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return nil
}

// GetServer returns the underlying server (for compatibility)
func (gms *GinManagedServer) GetServer() *Server {
	return gms.server
}

// GinBackendFactory creates gin managed servers
type GinBackendFactory struct{}

// NewGinBackendFactory creates a new gin backend factory
func NewGinBackendFactory() *GinBackendFactory {
	return &GinBackendFactory{}
}

// CreateServer creates a new gin managed server
func (gbf *GinBackendFactory) CreateServer(config *ServerConfig, sharedResources *SharedResources) (ManagedServer, error) {
	return NewGinManagedServer(config, sharedResources)
}

// GetBackendType returns the backend type this factory creates
func (gbf *GinBackendFactory) GetBackendType() BackendType {
	return GinBackendType
}