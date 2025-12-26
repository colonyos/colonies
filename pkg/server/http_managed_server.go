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

// HTTPManagedServer wraps the existing Server to make it compatible with ServerManager
type HTTPManagedServer struct {
	server  *Server
	config  *ServerConfig
	mu      sync.RWMutex
	running bool
}

// NewHTTPManagedServer creates a new HTTP managed server
func NewHTTPManagedServer(config *ServerConfig, sharedResources *SharedResources) (*HTTPManagedServer, error) {
	if config.BackendType != GinBackendType {
		return nil, fmt.Errorf("invalid backend type for HTTP server: %s", config.BackendType)
	}

	// Create the underlying server
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

	return &HTTPManagedServer{
		server: server,
		config: config,
	}, nil
}

// Start starts the HTTP server
func (hms *HTTPManagedServer) Start() error {
	hms.mu.Lock()
	defer hms.mu.Unlock()

	if hms.running {
		return errors.New("HTTP server is already running")
	}

	// Start the server in a goroutine
	go func() {
		log.WithFields(log.Fields{
			"BackendType": GinBackendType,
			"Port":        hms.config.Port,
			"TLS":         hms.config.TLS,
		}).Info("Starting HTTP server")

		hms.mu.Lock()
		hms.running = true
		hms.mu.Unlock()

		err := hms.server.ServeForever()

		hms.mu.Lock()
		hms.running = false
		hms.mu.Unlock()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithFields(log.Fields{
				"BackendType": GinBackendType,
				"Error":       err,
			}).Error("HTTP server stopped with error")
		} else {
			log.WithField("BackendType", GinBackendType).Info("HTTP server stopped")
		}
	}()

	// Wait a moment to ensure server started
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop stops the HTTP server gracefully
func (hms *HTTPManagedServer) Stop(ctx context.Context) error {
	hms.mu.RLock()
	if !hms.running {
		hms.mu.RUnlock()
		return nil
	}
	hms.mu.RUnlock()

	log.WithField("BackendType", GinBackendType).Info("Stopping HTTP server")

	// Create a channel to signal completion
	done := make(chan error, 1)

	go func() {
		hms.server.Shutdown()
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("HTTP server shutdown timed out: %w", ctx.Err())
	}
}

// GetBackendType returns the backend type
func (hms *HTTPManagedServer) GetBackendType() BackendType {
	return GinBackendType
}

// GetPort returns the server port
func (hms *HTTPManagedServer) GetPort() int {
	return hms.config.Port
}

// GetAddr returns the server address
func (hms *HTTPManagedServer) GetAddr() string {
	return fmt.Sprintf(":%d", hms.GetPort())
}

// IsRunning returns whether the server is running
func (hms *HTTPManagedServer) IsRunning() bool {
	hms.mu.RLock()
	defer hms.mu.RUnlock()
	return hms.running
}

// HealthCheck performs a health check on the server
func (hms *HTTPManagedServer) HealthCheck() error {
	if !hms.IsRunning() {
		return errors.New("HTTP server is not running")
	}

	// Could add more sophisticated health check here (e.g., make HTTP request to /health endpoint)
	return nil
}

// GetServer returns the underlying server (for compatibility)
func (hms *HTTPManagedServer) GetServer() *Server {
	return hms.server
}

// HTTPBackendFactory creates HTTP managed servers
type HTTPBackendFactory struct{}

// NewHTTPBackendFactory creates a new HTTP backend factory
func NewHTTPBackendFactory() *HTTPBackendFactory {
	return &HTTPBackendFactory{}
}

// CreateServer creates a new HTTP managed server
func (hbf *HTTPBackendFactory) CreateServer(config *ServerConfig, sharedResources *SharedResources) (ManagedServer, error) {
	return NewHTTPManagedServer(config, sharedResources)
}

// GetBackendType returns the backend type this factory creates
func (hbf *HTTPBackendFactory) GetBackendType() BackendType {
	return GinBackendType
}
