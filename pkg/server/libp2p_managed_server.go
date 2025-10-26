package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	libp2pbackend "github.com/colonyos/colonies/pkg/backends/libp2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	log "github.com/sirupsen/logrus"
)

const (
	ColoniesProcotolID = protocol.ID("/colonies/rpc/1.0.0")
)

// LibP2PManagedServer implements a libp2p-based server for distributed Colony operations
type LibP2PManagedServer struct {
	config *ServerConfig
	server *Server // Shared Colonies server instance
	
	// LibP2P components
	host            host.Host
	pubsub          *pubsub.PubSub
	realtimeHandler *libp2pbackend.P2PRealtimeHandler
	
	// Stream management
	streams     map[string]network.Stream
	streamsLock sync.RWMutex
	
	// Lifecycle management
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewLibP2PManagedServer creates a new libp2p managed server
func NewLibP2PManagedServer(config *ServerConfig, sharedResources *SharedResources) (*LibP2PManagedServer, error) {
	if config.BackendType != LibP2PBackendType {
		return nil, fmt.Errorf("invalid backend type for libp2p server: %s", config.BackendType)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create shared server instance  
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
	
	// Build libp2p options
	// LibP2PPort must be explicitly configured
	if config.LibP2PPort == 0 {
		cancel()
		return nil, fmt.Errorf("LibP2PPort must be configured (set COLONIES_LIBP2P_PORT environment variable)")
	}

	libp2pPort := config.LibP2PPort
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
		"HTTPPort":    config.Port,
		"LibP2PPort":  libp2pPort,
		"QUICPort":    libp2pPort + 1,
	}).Info("LibP2P port configuration")

	// Check for predefined identity from environment
	// This allows for deterministic peer IDs
	if identityKey := getLibP2PIdentityFromEnv(); identityKey != "" {
		privKey, err := parseLibP2PPrivateKey(identityKey)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to parse libp2p identity: %w", err)
		}
		opts = append(opts, libp2p.Identity(privKey))
		log.Info("Using predefined LibP2P identity from COLONIES_LIBP2P_IDENTITY")
	}

	// Create libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Create pubsub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	lms := &LibP2PManagedServer{
		config:  config,
		server:  server,
		host:    h,
		pubsub:  ps,
		streams: make(map[string]network.Stream),
		ctx:     ctx,
		cancel:  cancel,
	}

	// Create realtime handler
	lms.realtimeHandler = libp2pbackend.NewP2PRealtimeHandler(ps)

	// Set protocol handler for RPC requests
	h.SetStreamHandler(ColoniesProcotolID, lms.handleRPCStream)

	log.WithFields(log.Fields{
		"BackendType": LibP2PBackendType,
		"Port":        config.Port,
		"PeerID":      h.ID().String(),
		"Addrs":       h.Addrs(),
	}).Info("LibP2P managed server created")

	return lms, nil
}

// Start starts the libp2p server
func (lms *LibP2PManagedServer) Start() error {
	lms.mu.Lock()
	defer lms.mu.Unlock()
	
	if lms.running {
		return fmt.Errorf("libp2p server is already running")
	}
	
	log.WithFields(log.Fields{
		"BackendType": LibP2PBackendType,
		"Port":        lms.config.Port,
		"PeerID":      lms.host.ID().String(),
	}).Info("Starting LibP2P server")
	
	// Start the gin server in a goroutine (for HTTP endpoints)
	go func() {
		log.WithField("BackendType", LibP2PBackendType).Info("Starting HTTP server for libp2p backend")
		if err := lms.server.ServeForever(); err != nil {
			log.WithError(err).Error("HTTP server stopped with error")
		}
	}()
	
	lms.running = true
	
	// Start background routines
	go lms.processUpdateBroadcaster()
	go lms.peerDiscovery()
	
	log.WithFields(log.Fields{
		"BackendType": LibP2PBackendType,
		"PeerID":      lms.host.ID().String(),
		"Addrs":       lms.host.Addrs(),
	}).Info("LibP2P server started")
	
	return nil
}

// Stop stops the libp2p server gracefully
func (lms *LibP2PManagedServer) Stop(ctx context.Context) error {
	lms.mu.Lock()
	defer lms.mu.Unlock()
	
	if !lms.running {
		return nil
	}
	
	log.WithField("BackendType", LibP2PBackendType).Info("Stopping LibP2P server")
	
	// Cancel context to stop background routines
	lms.cancel()
	
	// Close all streams
	lms.streamsLock.Lock()
	for _, stream := range lms.streams {
		stream.Close()
	}
	lms.streamsLock.Unlock()
	
	// Stop shared server
	if lms.server != nil {
		lms.server.Shutdown()
	}
	
	// Close libp2p host
	if lms.host != nil {
		lms.host.Close()
	}
	
	lms.running = false
	
	log.WithField("BackendType", LibP2PBackendType).Info("LibP2P server stopped")
	
	return nil
}

// GetBackendType returns the backend type
func (lms *LibP2PManagedServer) GetBackendType() BackendType {
	return LibP2PBackendType
}

// GetPort returns the server port
func (lms *LibP2PManagedServer) GetPort() int {
	return lms.config.Port
}

// GetAddr returns the server address (multiaddr format)
func (lms *LibP2PManagedServer) GetAddr() string {
	if lms.host != nil && len(lms.host.Addrs()) > 0 {
		return fmt.Sprintf("%s/p2p/%s", lms.host.Addrs()[0].String(), lms.host.ID().String())
	}
	return fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", lms.config.LibP2PPort)
}

// IsRunning returns whether the server is running
func (lms *LibP2PManagedServer) IsRunning() bool {
	lms.mu.RLock()
	defer lms.mu.RUnlock()
	return lms.running
}

// HealthCheck performs a health check on the server
func (lms *LibP2PManagedServer) HealthCheck() error {
	if !lms.IsRunning() {
		return fmt.Errorf("libp2p server is not running")
	}
	
	// Check if host is still active
	if lms.host == nil {
		return fmt.Errorf("libp2p host is nil")
	}
	
	// Check network connectivity
	if len(lms.host.Network().Peers()) == 0 {
		log.WithField("BackendType", LibP2PBackendType).Warn("No peers connected")
		// Don't return error - this is normal during startup
	}
	
	return nil
}

// handleRPCStream handles incoming RPC streams
func (lms *LibP2PManagedServer) handleRPCStream(stream network.Stream) {
	defer stream.Close()
	
	peerID := stream.Conn().RemotePeer().String()
	log.WithFields(log.Fields{
		"PeerID":      peerID,
		"Protocol":    ColoniesProcotolID,
		"BackendType": LibP2PBackendType,
	}).Debug("Handling RPC stream")

	// Store stream for potential reuse
	lms.streamsLock.Lock()
	lms.streams[peerID] = stream
	lms.streamsLock.Unlock()
	
	// Clean up stream reference when done
	defer func() {
		lms.streamsLock.Lock()
		delete(lms.streams, peerID)
		lms.streamsLock.Unlock()
	}()

	// Create context adapter for this stream
	streamCtx := libp2pbackend.NewStreamContext(stream, lms.pubsub)
	
	// Read the incoming message
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read from stream")
		return
	}
	
	message := string(buf[:n])
	log.WithFields(log.Fields{
		"PeerID":      peerID,
		"MessageSize": n,
		"Message":     message,
	}).Debug("Received RPC message")
	
	// Process the message with the shared server
	lms.server.handleAPIRequest(streamCtx)
}

// processUpdateBroadcaster broadcasts process updates via pubsub
func (lms *LibP2PManagedServer) processUpdateBroadcaster() {
	// Subscribe to process changes from the shared server
	// This is a simplified implementation - in practice you'd want
	// proper event subscription from the process controller
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lms.ctx.Done():
			return
		case <-ticker.C:
			// In a real implementation, this would be driven by actual process events
			// For now, this is a placeholder for the broadcaster routine
		}
	}
}

// peerDiscovery handles peer discovery and connection management
func (lms *LibP2PManagedServer) peerDiscovery() {
	// Enable mDNS discovery for local peers
	// In production, you'd also want DHT and bootstrap nodes
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-lms.ctx.Done():
			return
		case <-ticker.C:
			peers := lms.host.Network().Peers()
			log.WithFields(log.Fields{
				"BackendType": LibP2PBackendType,
				"PeerCount":   len(peers),
			}).Debug("Peer discovery check")
		}
	}
}

// GetHost returns the libp2p host (for external integration)
func (lms *LibP2PManagedServer) GetHost() host.Host {
	return lms.host
}

// GetPubSub returns the pubsub instance (for external integration)
func (lms *LibP2PManagedServer) GetPubSub() *pubsub.PubSub {
	return lms.pubsub
}

// PublishProcessUpdate publishes a process update via the realtime handler
func (lms *LibP2PManagedServer) PublishProcessUpdate(process *core.Process) error {
	if lms.realtimeHandler != nil {
		return lms.realtimeHandler.PublishProcessUpdate(process)
	}
	return fmt.Errorf("realtime handler not available")
}

// LibP2PBackendFactory creates libp2p managed servers
type LibP2PBackendFactory struct{}

// NewLibP2PBackendFactory creates a new libp2p backend factory
func NewLibP2PBackendFactory() *LibP2PBackendFactory {
	return &LibP2PBackendFactory{}
}

// CreateServer creates a new libp2p managed server
func (lbf *LibP2PBackendFactory) CreateServer(config *ServerConfig, sharedResources *SharedResources) (ManagedServer, error) {
	return NewLibP2PManagedServer(config, sharedResources)
}

// GetBackendType returns the backend type this factory creates
func (lbf *LibP2PBackendFactory) GetBackendType() BackendType {
	return LibP2PBackendType
}

// getLibP2PIdentityFromEnv retrieves the LibP2P identity from environment
func getLibP2PIdentityFromEnv() string {
	return os.Getenv("COLONIES_LIBP2P_IDENTITY")
}

// parseLibP2PPrivateKey parses a hex-encoded private key for LibP2P
func parseLibP2PPrivateKey(hexKey string) (crypto.PrivKey, error) {
	// Decode hex string to bytes
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex key: %w", err)
	}

	// Unmarshal the private key
	privKey, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return privKey, nil
}