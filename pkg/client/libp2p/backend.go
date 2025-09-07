package libp2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

const (
	ColoniesProcotolID = protocol.ID("/colonies/rpc/1.0.0")
	ConnectTimeout     = 30 * time.Second
	StreamTimeout      = 60 * time.Second
)

// LibP2PClientBackend implements peer-to-peer client backend using libp2p
type LibP2PClientBackend struct {
	config *backends.ClientConfig
	
	// LibP2P components
	host   host.Host
	pubsub *pubsub.PubSub
	
	// Peer management
	serverPeers     map[peer.ID]*ServerPeer
	serverPeersLock sync.RWMutex
	
	// Connection management
	ctx    context.Context
	cancel context.CancelFunc
	
	// Discovery and routing
	bootstrapPeers []multiaddr.Multiaddr
}

// ServerPeer represents a known Colonies server peer
type ServerPeer struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Active   bool
}

// NewLibP2PClientBackend creates a new libp2p client backend
func NewLibP2PClientBackend(config *backends.ClientConfig) (*LibP2PClientBackend, error) {
	if config.BackendType != backends.LibP2PClientBackendType {
		return nil, fmt.Errorf("invalid backend type for libp2p client")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create libp2p host with minimal configuration for client
	h, err := libp2p.New(
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Create pubsub for realtime communication
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	backend := &LibP2PClientBackend{
		config:      config,
		host:        h,
		pubsub:      ps,
		serverPeers: make(map[peer.ID]*ServerPeer),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Parse bootstrap peers from config.Host (assuming it contains multiaddr)
	if config.Host != "" {
		if addr, err := multiaddr.NewMultiaddr(config.Host); err == nil {
			backend.bootstrapPeers = []multiaddr.Multiaddr{addr}
		}
	}

	// Start peer discovery
	go backend.discoverPeers()

	logrus.WithFields(logrus.Fields{
		"peer_id":         h.ID().String(),
		"bootstrap_peers": len(backend.bootstrapPeers),
	}).Info("LibP2P client backend initialized")

	return backend, nil
}

// SendRawMessage sends a raw JSON message via libp2p stream
func (l *LibP2PClientBackend) SendRawMessage(jsonString string, insecure bool) (string, error) {
	// Get an active server peer
	serverPeer, err := l.getActivePeer()
	if err != nil {
		return "", fmt.Errorf("no active server peer: %w", err)
	}

	// Open stream to server
	stream, err := l.host.NewStream(l.ctx, serverPeer.ID, ColoniesProcotolID)
	if err != nil {
		l.markPeerInactive(serverPeer.ID)
		return "", fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Set stream timeout
	stream.SetDeadline(time.Now().Add(StreamTimeout))

	// Send message
	_, err = stream.Write([]byte(jsonString))
	if err != nil {
		return "", fmt.Errorf("failed to write to stream: %w", err)
	}

	// Read response
	response, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(response), nil
}

// SendMessage sends an RPC message with authentication via libp2p stream
func (l *LibP2PClientBackend) SendMessage(method string, jsonString string, prvKey string, insecure bool, ctx context.Context) (string, error) {
	var rpcMsg *rpc.RPCMsg
	var err error
	
	if insecure {
		rpcMsg, err = rpc.CreateInsecureRPCMsg(method, jsonString)
		if err != nil {
			return "", err
		}
	} else {
		rpcMsg, err = rpc.CreateRPCMsg(method, jsonString, prvKey)
		if err != nil {
			return "", err
		}
	}
	
	rpcJSONString, err := rpcMsg.ToJSON()
	if err != nil {
		return "", err
	}

	// Get an active server peer
	serverPeer, err := l.getActivePeer()
	if err != nil {
		return "", fmt.Errorf("no active server peer: %w", err)
	}

	// Open stream to server with context
	stream, err := l.host.NewStream(ctx, serverPeer.ID, ColoniesProcotolID)
	if err != nil {
		l.markPeerInactive(serverPeer.ID)
		return "", fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Set stream timeout
	stream.SetDeadline(time.Now().Add(StreamTimeout))

	// Send RPC message
	_, err = stream.Write([]byte(rpcJSONString))
	if err != nil {
		return "", fmt.Errorf("failed to write to stream: %w", err)
	}

	// Read response
	response, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	respBodyString := string(response)

	// Parse RPC reply
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(respBodyString)
	if err != nil {
		return "", fmt.Errorf("expected a valid Colonies RPC message, but got: %s", respBodyString)
	}

	if rpcReplyMsg.Error {
		failure, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
		if err != nil {
			return "", err
		}
		return "", &core.ColoniesError{Status: failure.Status, Message: failure.Message}
	}

	return rpcReplyMsg.DecodePayload(), nil
}

// EstablishRealtimeConn establishes a real-time connection via libp2p pubsub
func (l *LibP2PClientBackend) EstablishRealtimeConn(jsonString string) (backends.RealtimeConnection, error) {
	// Parse the subscription request to determine topic
	var rpcMsg rpc.RPCMsg
	err := json.Unmarshal([]byte(jsonString), &rpcMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscription request: %w", err)
	}

	topic, err := l.getTopicForSubscription(&rpcMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to determine topic: %w", err)
	}

	logrus.WithField("topic", topic).Info("Establishing libp2p pubsub realtime connection")

	// Join the topic
	topicHandle, err := l.pubsub.Join(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic %s: %w", topic, err)
	}

	// Subscribe to the topic
	subscription, err := topicHandle.Subscribe()
	if err != nil {
		topicHandle.Close()
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	// Create and return the pubsub connection
	return NewPubSubRealtimeConnection(topicHandle, subscription, l.ctx, l.host.ID().String()), nil
}

// CheckHealth checks the health of libp2p connections
func (l *LibP2PClientBackend) CheckHealth() error {
	// Check if host is still active
	if l.host == nil {
		return fmt.Errorf("libp2p host is nil")
	}

	// Check if we have any active peers
	l.serverPeersLock.RLock()
	activePeers := 0
	for _, peer := range l.serverPeers {
		if peer.Active {
			activePeers++
		}
	}
	l.serverPeersLock.RUnlock()

	if activePeers == 0 {
		return fmt.Errorf("no active server peers")
	}

	return nil
}

// Close closes the libp2p backend and cleans up resources
func (l *LibP2PClientBackend) Close() error {
	l.cancel()
	
	if l.host != nil {
		return l.host.Close()
	}
	
	return nil
}

// getActivePeer returns an active server peer
func (l *LibP2PClientBackend) getActivePeer() (*ServerPeer, error) {
	l.serverPeersLock.RLock()
	defer l.serverPeersLock.RUnlock()

	// Find the most recently seen active peer
	var bestPeer *ServerPeer
	var latestTime time.Time

	for _, peer := range l.serverPeers {
		if peer.Active && peer.LastSeen.After(latestTime) {
			bestPeer = peer
			latestTime = peer.LastSeen
		}
	}

	if bestPeer == nil {
		return nil, fmt.Errorf("no active server peers available")
	}

	return bestPeer, nil
}

// markPeerInactive marks a peer as inactive
func (l *LibP2PClientBackend) markPeerInactive(peerID peer.ID) {
	l.serverPeersLock.Lock()
	defer l.serverPeersLock.Unlock()

	if peer, exists := l.serverPeers[peerID]; exists {
		peer.Active = false
		logrus.WithField("peer_id", peerID.String()).Warn("Marked peer as inactive")
	}
}

// discoverPeers handles peer discovery
func (l *LibP2PClientBackend) discoverPeers() {
	// Connect to bootstrap peers
	for _, addr := range l.bootstrapPeers {
		go l.connectToBootstrapPeer(addr)
	}

	// Periodic peer discovery
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.performPeerDiscovery()
		}
	}
}

// connectToBootstrapPeer connects to a bootstrap peer
func (l *LibP2PClientBackend) connectToBootstrapPeer(addr multiaddr.Multiaddr) {
	ctx, cancel := context.WithTimeout(l.ctx, ConnectTimeout)
	defer cancel()

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		logrus.WithError(err).WithField("addr", addr.String()).Error("Failed to parse peer address")
		return
	}

	err = l.host.Connect(ctx, *peerInfo)
	if err != nil {
		logrus.WithError(err).WithField("peer_id", peerInfo.ID.String()).Warn("Failed to connect to bootstrap peer")
		return
	}

	// Add to known server peers
	l.serverPeersLock.Lock()
	l.serverPeers[peerInfo.ID] = &ServerPeer{
		ID:       peerInfo.ID,
		Addrs:    peerInfo.Addrs,
		LastSeen: time.Now(),
		Active:   true,
	}
	l.serverPeersLock.Unlock()

	logrus.WithFields(logrus.Fields{
		"peer_id": peerInfo.ID.String(),
		"addr":    addr.String(),
	}).Info("Connected to bootstrap peer")
}

// performPeerDiscovery performs periodic peer discovery
func (l *LibP2PClientBackend) performPeerDiscovery() {
	// Check connectivity to known peers
	l.serverPeersLock.Lock()
	for peerID, serverPeer := range l.serverPeers {
		// Check if peer is still connected
		if l.host.Network().Connectedness(peerID) != network.Connected {
			serverPeer.Active = false
		} else {
			serverPeer.Active = true
			serverPeer.LastSeen = time.Now()
		}
	}
	l.serverPeersLock.Unlock()

	// Log current peer status
	l.serverPeersLock.RLock()
	activePeers := 0
	for _, peer := range l.serverPeers {
		if peer.Active {
			activePeers++
		}
	}
	l.serverPeersLock.RUnlock()

	logrus.WithFields(logrus.Fields{
		"active_peers": activePeers,
		"total_peers":  len(l.serverPeers),
	}).Debug("Peer discovery status")
}

// getTopicForSubscription determines the appropriate pubsub topic based on the subscription request
func (l *LibP2PClientBackend) getTopicForSubscription(rpcMsg *rpc.RPCMsg) (string, error) {
	switch rpcMsg.PayloadType {
	case rpc.SubscribeProcessesPayloadType:
		var msg rpc.SubscribeProcessesMsg
		err := json.Unmarshal([]byte(rpcMsg.DecodePayload()), &msg)
		if err != nil {
			return "", fmt.Errorf("failed to parse subscribe processes message: %w", err)
		}
		return fmt.Sprintf("/colonies/%s/processes", msg.ColonyName), nil
		
	case rpc.SubscribeProcessPayloadType:
		var msg rpc.SubscribeProcessMsg
		err := json.Unmarshal([]byte(rpcMsg.DecodePayload()), &msg)
		if err != nil {
			return "", fmt.Errorf("failed to parse subscribe process message: %w", err)
		}
		return fmt.Sprintf("/colonies/%s/process/%s", msg.ColonyName, msg.ProcessID), nil
		
	default:
		return "", fmt.Errorf("unsupported subscription type: %s", rpcMsg.PayloadType)
	}
}

// LibP2PClientBackendFactory creates libp2p client backends
type LibP2PClientBackendFactory struct{}

// NewLibP2PClientBackendFactory creates a new libp2p client backend factory
func NewLibP2PClientBackendFactory() *LibP2PClientBackendFactory {
	return &LibP2PClientBackendFactory{}
}

// CreateBackend creates a new libp2p client backend
func (f *LibP2PClientBackendFactory) CreateBackend(config *backends.ClientConfig) (backends.ClientBackend, error) {
	return NewLibP2PClientBackend(config)
}

// GetBackendType returns the backend type this factory creates
func (f *LibP2PClientBackendFactory) GetBackendType() backends.ClientBackendType {
	return backends.LibP2PClientBackendType
}

// Compile-time checks that LibP2PClientBackend implements the required interfaces
var _ backends.ClientBackend = (*LibP2PClientBackend)(nil)
var _ backends.RealtimeBackend = (*LibP2PClientBackend)(nil)
var _ backends.ClientBackendWithRealtime = (*LibP2PClientBackend)(nil)