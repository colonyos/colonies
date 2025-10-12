package libp2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/client/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

const (
	ColoniesProcotolID = protocol.ID("/colonies/rpc/1.0.0")
	ConnectTimeout     = 60 * time.Second  // Increased for relay circuit establishment
	StreamTimeout      = 90 * time.Second  // Increased for relay latency
)

// LibP2PClientBackend implements peer-to-peer client backend using libp2p
type LibP2PClientBackend struct {
	config *backends.ClientConfig

	// LibP2P components
	host   host.Host
	pubsub *pubsub.PubSub
	dht    *dht.IpfsDHT

	// Peer management
	serverPeers     map[peer.ID]*ServerPeer
	serverPeersLock sync.RWMutex

	// Connection management
	ctx    context.Context
	cancel context.CancelFunc

	// Discovery and routing
	bootstrapPeers    []multiaddr.Multiaddr
	useDHTDiscovery   bool
	dhtRendezvous     string
	routingDiscovery  *routing.RoutingDiscovery
}

// ServerPeer represents a known Colonies server peer
type ServerPeer struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Active   bool
}

// CachedPeer represents a cached peer entry for serialization
type CachedPeer struct {
	PeerID    string    `json:"peer_id"`
	Addrs     []string  `json:"addrs"`
	LastSeen  time.Time `json:"last_seen"`
	Rendezvous string   `json:"rendezvous"`
}

// DHTCache represents the DHT peer cache structure
type DHTCache struct {
	Version int          `json:"version"`
	Updated time.Time    `json:"updated"`
	Peers   []CachedPeer `json:"peers"`
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
		config:        config,
		host:          h,
		pubsub:        ps,
		serverPeers:   make(map[peer.ID]*ServerPeer),
		ctx:           ctx,
		cancel:        cancel,
		dhtRendezvous: "colonies-server", // Default rendezvous point
	}

	// Parse config.Host - it can be:
	// 1. A full multiaddress: /dns/localhost/tcp/5000/p2p/12D3Koo... (direct connect)
	// 2. Just "dht" or "dht:rendezvous-name" (DHT-based discovery)
	// 3. Plain hostname like "localhost" (will use DHT discovery)
	// 4. Empty (will use DHT with default rendezvous)
	if config.Host != "" {
		if config.Host == "dht" || (len(config.Host) >= 4 && config.Host[:4] == "dht:") {
			// DHT-based discovery
			backend.useDHTDiscovery = true
			if len(config.Host) > 4 && config.Host[3] == ':' {
				backend.dhtRendezvous = config.Host[4:]
			}
			logrus.WithField("rendezvous", backend.dhtRendezvous).Info("Using DHT-based peer discovery")
		} else if addr, err := multiaddr.NewMultiaddr(config.Host); err == nil {
			// Successfully parsed as multiaddress - this is a direct connection target
			backend.bootstrapPeers = []multiaddr.Multiaddr{addr}
			logrus.WithField("addr", addr.String()).Info("Using direct multiaddress connection")
		} else {
			// Not a valid multiaddress - treat as plain hostname and use DHT discovery
			// This handles cases like "localhost" or "example.com" gracefully
			backend.useDHTDiscovery = true
			logrus.WithFields(logrus.Fields{
				"host":       config.Host,
				"rendezvous": backend.dhtRendezvous,
			}).Info("Plain hostname provided, using DHT-based peer discovery")
		}
	} else {
		// No host specified, default to DHT discovery
		backend.useDHTDiscovery = true
		logrus.WithField("rendezvous", backend.dhtRendezvous).Info("No server host specified, using DHT-based discovery")
	}

	// Parse additional bootstrap peers from config if specified
	if config.BootstrapPeers != "" {
		peerAddrs := strings.Split(config.BootstrapPeers, ",")
		for _, peerAddr := range peerAddrs {
			peerAddr = strings.TrimSpace(peerAddr)
			if peerAddr == "" {
				continue
			}
			if addr, err := multiaddr.NewMultiaddr(peerAddr); err == nil {
				backend.bootstrapPeers = append(backend.bootstrapPeers, addr)
			} else {
				logrus.WithError(err).WithField("addr", peerAddr).Warn("Failed to parse bootstrap peer address")
			}
		}
	}

	// Initialize DHT if needed
	if backend.useDHTDiscovery || len(backend.bootstrapPeers) == 0 {
		kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeClient))
		if err != nil {
			h.Close()
			cancel()
			return nil, fmt.Errorf("failed to create DHT: %w", err)
		}
		backend.dht = kadDHT

		// Bootstrap the DHT
		if err = kadDHT.Bootstrap(ctx); err != nil {
			h.Close()
			cancel()
			return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
		}

		backend.routingDiscovery = routing.NewRoutingDiscovery(kadDHT)
		logrus.Info("DHT initialized for client")
	}

	// Load DHT cache if using DHT discovery
	if backend.useDHTDiscovery {
		if err := backend.loadDHTCache(); err != nil {
			logrus.WithError(err).Warn("Failed to load DHT cache, will perform fresh discovery")
		}
	}

	// Start peer discovery
	go backend.discoverPeers()

	// If using DHT discovery and no cached peers found, start searching
	backend.serverPeersLock.RLock()
	hasCachedPeers := len(backend.serverPeers) > 0
	backend.serverPeersLock.RUnlock()

	if backend.useDHTDiscovery && backend.routingDiscovery != nil && !hasCachedPeers {
		// Start initial DHT search in background
		// If no peers found initially, getActivePeer() will trigger more searches
		go backend.discoverViaDHT()
	}

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
	// If using DHT discovery, retry for up to 30 seconds
	maxRetries := 30
	retryDelay := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		l.serverPeersLock.RLock()
		// Find the most recently seen active peer
		var bestPeer *ServerPeer
		var latestTime time.Time

		for _, peer := range l.serverPeers {
			if peer.Active && peer.LastSeen.After(latestTime) {
				bestPeer = peer
				latestTime = peer.LastSeen
			}
		}
		l.serverPeersLock.RUnlock()

		if bestPeer != nil {
			return bestPeer, nil
		}

		// If using DHT discovery and no peers found, trigger discovery and wait
		if l.useDHTDiscovery && attempt < maxRetries-1 {
			if attempt == 0 {
				logrus.Info("No active peers found, waiting for DHT discovery...")
			}

			// Trigger DHT search every 5 attempts to increase chances of finding peers
			// DHT advertisements take time to propagate through the network
			if attempt%5 == 0 && l.routingDiscovery != nil {
				go l.discoverViaDHT()
			}

			time.Sleep(retryDelay)
			continue
		}

		break
	}

	return nil, fmt.Errorf("no active server peers available")
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
	// If using DHT, we'll test if they're Colonies servers
	// If not using DHT, we'll add them directly as servers
	for _, addr := range l.bootstrapPeers {
		// When using DHT, bootstrap peers need to be tested first
		// When not using DHT, they're known servers
		addAsServer := !l.useDHTDiscovery
		go l.connectToBootstrapPeer(addr, addAsServer)
	}

	// If using DHT, also try to probe bootstrap peers for Colonies protocol
	if l.useDHTDiscovery {
		time.Sleep(2 * time.Second) // Give connections time to establish
		go l.probeBootstrapPeersForColoniesProtocol()
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
func (l *LibP2PClientBackend) connectToBootstrapPeer(addr multiaddr.Multiaddr, addAsServer bool) {
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

	// Only add as server peer if requested (e.g., when it's a known Colonies server)
	if addAsServer {
		l.serverPeersLock.Lock()
		l.serverPeers[peerInfo.ID] = &ServerPeer{
			ID:       peerInfo.ID,
			Addrs:    peerInfo.Addrs,
			LastSeen: time.Now(),
			Active:   true,
		}
		l.serverPeersLock.Unlock()
	}

	logrus.WithFields(logrus.Fields{
		"peer_id":     peerInfo.ID.String(),
		"addr":        addr.String(),
		"add_as_server": addAsServer,
	}).Info("Connected to bootstrap peer")
}

// probeBootstrapPeersForColoniesProtocol checks if connected bootstrap peers support Colonies protocol
func (l *LibP2PClientBackend) probeBootstrapPeersForColoniesProtocol() {
	// Get all connected peers
	connectedPeers := l.host.Network().Peers()

	for _, peerID := range connectedPeers {
		// Check if this peer supports our protocol
		protocols, err := l.host.Peerstore().GetProtocols(peerID)
		if err != nil {
			continue
		}

		// Check if peer supports Colonies RPC protocol
		supportsColonies := false
		for _, proto := range protocols {
			if proto == ColoniesProcotolID {
				supportsColonies = true
				break
			}
		}

		if supportsColonies {
			// This bootstrap peer is a Colonies server!
			addrs := l.host.Peerstore().Addrs(peerID)
			l.serverPeersLock.Lock()
			if _, exists := l.serverPeers[peerID]; !exists {
				l.serverPeers[peerID] = &ServerPeer{
					ID:       peerID,
					Addrs:    addrs,
					LastSeen: time.Now(),
					Active:   true,
				}
				logrus.WithFields(logrus.Fields{
					"peer_id": peerID.String(),
					"addrs":   addrs,
				}).Info("Bootstrap peer supports Colonies protocol, added as server")

				// Save cache since we found a working server peer
				l.serverPeersLock.Unlock()
				if err := l.saveDHTCache(); err != nil {
					logrus.WithError(err).Warn("Failed to save DHT cache after discovering bootstrap peer")
				}
				l.serverPeersLock.Lock()
			}
			l.serverPeersLock.Unlock()
		}
	}
}

// performPeerDiscovery performs periodic peer discovery
func (l *LibP2PClientBackend) performPeerDiscovery() {
	// Check connectivity to known peers and try to reconnect to inactive ones
	l.serverPeersLock.Lock()
	for peerID, serverPeer := range l.serverPeers {
		// Check if peer is still connected
		if l.host.Network().Connectedness(peerID) != network.Connected {
			// Peer disconnected - try to reconnect
			peerInfo := peer.AddrInfo{
				ID:    peerID,
				Addrs: serverPeer.Addrs,
			}

			ctx, cancel := context.WithTimeout(l.ctx, ConnectTimeout)
			err := l.host.Connect(ctx, peerInfo)
			cancel()

			if err != nil {
				serverPeer.Active = false
				logrus.WithError(err).WithField("peer_id", peerID.String()).Debug("Failed to reconnect to peer")
			} else {
				serverPeer.Active = true
				serverPeer.LastSeen = time.Now()
				logrus.WithField("peer_id", peerID.String()).Info("Successfully reconnected to peer")
			}
		} else {
			serverPeer.Active = true
			serverPeer.LastSeen = time.Now()
		}
	}
	l.serverPeersLock.Unlock()

	// If using DHT discovery and no active peers, search for more
	l.serverPeersLock.RLock()
	activePeers := 0
	for _, peer := range l.serverPeers {
		if peer.Active {
			activePeers++
		}
	}
	l.serverPeersLock.RUnlock()

	if l.useDHTDiscovery && activePeers == 0 && l.routingDiscovery != nil {
		go l.discoverViaDHT()
	}

	logrus.WithFields(logrus.Fields{
		"active_peers": activePeers,
		"total_peers":  len(l.serverPeers),
	}).Debug("Peer discovery status")
}

// discoverViaDHT discovers peers via DHT using the rendezvous point
func (l *LibP2PClientBackend) discoverViaDHT() {
	logrus.WithField("rendezvous", l.dhtRendezvous).Info("Searching for peers via DHT...")

	ctx, cancel := context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	// Find peers advertising at the rendezvous point
	peerChan, err := l.routingDiscovery.FindPeers(ctx, l.dhtRendezvous)
	if err != nil {
		logrus.WithError(err).Warn("Failed to initiate DHT peer discovery")
		return
	}

	found := 0
	connected := 0
	for peer := range peerChan {
		// Skip ourselves
		if peer.ID == l.host.ID() {
			continue
		}

		found++
		isConnected := false

		// Try to connect to the peer
		if l.host.Network().Connectedness(peer.ID) != network.Connected {
			ctx, cancel := context.WithTimeout(l.ctx, ConnectTimeout)
			err := l.host.Connect(ctx, peer)
			cancel()

			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"peer_id": peer.ID.String(),
					"addrs":   peer.Addrs,
				}).Warn("Failed initial connection to discovered peer (will retry later)")
				// Don't skip! Add peer with Active=false so we can retry later
			} else {
				isConnected = true
				connected++
				logrus.WithFields(logrus.Fields{
					"peer_id": peer.ID.String(),
					"addrs":   peer.Addrs,
				}).Info("Discovered and connected to peer via DHT")
			}
		} else {
			isConnected = true
			connected++
			logrus.WithFields(logrus.Fields{
				"peer_id": peer.ID.String(),
				"addrs":   peer.Addrs,
			}).Info("Discovered peer via DHT (already connected)")
		}

		// Add to known server peers (even if initial connection failed)
		// performPeerDiscovery() will retry connections to inactive peers
		l.serverPeersLock.Lock()
		l.serverPeers[peer.ID] = &ServerPeer{
			ID:       peer.ID,
			Addrs:    peer.Addrs,
			LastSeen: time.Now(),
			Active:   isConnected,
		}
		l.serverPeersLock.Unlock()
	}

	logrus.WithFields(logrus.Fields{
		"rendezvous":      l.dhtRendezvous,
		"peers_found":     found,
		"peers_connected": connected,
	}).Info("DHT peer discovery completed")

	// Save cache if we found any peers
	if found > 0 {
		if err := l.saveDHTCache(); err != nil {
			logrus.WithError(err).Warn("Failed to save DHT cache")
		}
	}
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

// getDHTCachePath returns the path to the DHT cache file
func getDHTCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create .colonies directory if it doesn't exist
	coloniesDir := filepath.Join(home, ".colonies")
	if err := os.MkdirAll(coloniesDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create .colonies directory: %w", err)
	}

	return filepath.Join(coloniesDir, "dht_cache"), nil
}

// loadDHTCache loads cached DHT peers from disk
func (l *LibP2PClientBackend) loadDHTCache() error {
	cachePath, err := getDHTCachePath()
	if err != nil {
		return err
	}

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		logrus.Debug("No DHT cache file found")
		return nil
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache
	var cache DHTCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("failed to parse cache file: %w", err)
	}

	// Cache is valid for 24 hours
	cacheExpiry := 24 * time.Hour
	if time.Since(cache.Updated) > cacheExpiry {
		logrus.WithField("age", time.Since(cache.Updated)).Info("DHT cache expired, will perform fresh discovery")
		return nil
	}

	// Try cached peers that match our rendezvous point
	loaded := 0
	for _, cachedPeer := range cache.Peers {
		// Only use peers from the same rendezvous point
		if cachedPeer.Rendezvous != l.dhtRendezvous {
			continue
		}

		// Parse peer ID
		peerID, err := peer.Decode(cachedPeer.PeerID)
		if err != nil {
			logrus.WithError(err).WithField("peer_id", cachedPeer.PeerID).Debug("Failed to decode cached peer ID")
			continue
		}

		// Parse multiaddrs
		var addrs []multiaddr.Multiaddr
		for _, addrStr := range cachedPeer.Addrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				logrus.WithError(err).WithField("addr", addrStr).Debug("Failed to parse cached address")
				continue
			}
			addrs = append(addrs, addr)
		}

		if len(addrs) == 0 {
			continue
		}

		// Try to connect to cached peer
		peerInfo := peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}

		ctx, cancel := context.WithTimeout(l.ctx, ConnectTimeout)
		err = l.host.Connect(ctx, peerInfo)
		cancel()

		if err != nil {
			logrus.WithError(err).WithField("peer_id", peerID.String()).Debug("Failed to connect to cached peer")
			continue
		}

		// Successfully connected - add to server peers
		l.serverPeersLock.Lock()
		l.serverPeers[peerID] = &ServerPeer{
			ID:       peerID,
			Addrs:    addrs,
			LastSeen: time.Now(),
			Active:   true,
		}
		l.serverPeersLock.Unlock()

		loaded++
		logrus.WithFields(logrus.Fields{
			"peer_id": peerID.String(),
			"addrs":   addrs,
		}).Info("Successfully connected to cached peer")
	}

	if loaded > 0 {
		logrus.WithField("peers_loaded", loaded).Info("Loaded peers from DHT cache")
	}

	return nil
}

// saveDHTCache saves current DHT peers to disk
func (l *LibP2PClientBackend) saveDHTCache() error {
	cachePath, err := getDHTCachePath()
	if err != nil {
		return err
	}

	l.serverPeersLock.RLock()
	defer l.serverPeersLock.RUnlock()

	// Build cache from active server peers
	var cachedPeers []CachedPeer
	for _, serverPeer := range l.serverPeers {
		if !serverPeer.Active {
			continue
		}

		// Convert multiaddrs to strings
		var addrStrs []string
		for _, addr := range serverPeer.Addrs {
			addrStrs = append(addrStrs, addr.String())
		}

		cachedPeers = append(cachedPeers, CachedPeer{
			PeerID:     serverPeer.ID.String(),
			Addrs:      addrStrs,
			LastSeen:   serverPeer.LastSeen,
			Rendezvous: l.dhtRendezvous,
		})
	}

	// Create cache structure
	cache := DHTCache{
		Version: 1,
		Updated: time.Now(),
		Peers:   cachedPeers,
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"path":  cachePath,
		"peers": len(cachedPeers),
	}).Debug("Saved DHT cache")

	return nil
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