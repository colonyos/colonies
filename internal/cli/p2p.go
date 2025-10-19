package cli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	p2pCmd.AddCommand(lookupCmd)
	p2pCmd.AddCommand(generateCmd)
	p2pCmd.AddCommand(relayCmd)
	rootCmd.AddCommand(p2pCmd)

	lookupCmd.Flags().StringVarP(&Text, "rendezvous", "r", "colonies-server", "Rendezvous point to search for (default: colonies-server)")
	lookupCmd.Flags().IntVarP(&Timeout, "timeout", "t", 30, "Timeout in seconds for DHT discovery")
}

var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "LibP2P relay, DHT, and identity management",
	Long:  "LibP2P relay server, DHT debugging, and P2P identity generation commands",
}

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Look up what DHT nodes have stored about Colonies servers",
	Long:  "Queries the DHT network to see which peers are advertising at the rendezvous point. Useful for debugging peer discovery issues.",
	Run: func(cmd *cobra.Command, args []string) {
		rendezvous := Text
		if rendezvous == "" {
			rendezvous = "colonies-server"
		}

		log.WithFields(log.Fields{
			"Rendezvous": rendezvous,
			"Timeout":    Timeout,
		}).Info("Starting DHT lookup")

		err := performDHTLookup(rendezvous, time.Duration(Timeout)*time.Second)
		CheckError(err)
	},
}

func performDHTLookup(rendezvous string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
	defer cancel()

	// Create a temporary libp2p host for DHT queries
	log.Info("Creating temporary libp2p host...")
	h, err := libp2p.New(
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	defer h.Close()

	log.WithField("PeerID", h.ID().String()).Info("Created temporary libp2p host")

	// Create DHT in client mode
	log.Info("Initializing DHT in client mode...")
	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeClient))
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	// Bootstrap the DHT
	log.Info("Bootstrapping DHT...")
	if err = kadDHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Get bootstrap peers from environment or use defaults
	bootstrapPeersStr := getBootstrapPeers()
	if bootstrapPeersStr == "" {
		log.Warn("No bootstrap peers configured, using default libp2p bootstrap nodes")
		bootstrapPeersStr = "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN,/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa,/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb,/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt"
	}

	// Connect to bootstrap peers
	log.Info("Connecting to bootstrap peers...")
	bootstrapAddrs := parseBootstrapPeers(bootstrapPeersStr)
	bootstrapConnected := 0
	for _, addr := range bootstrapAddrs {
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.WithError(err).WithField("Addr", addr.String()).Debug("Failed to parse bootstrap peer")
			continue
		}

		connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
		err = h.Connect(connectCtx, *peerInfo)
		connectCancel()

		if err != nil {
			log.WithError(err).WithField("PeerID", peerInfo.ID.String()).Debug("Failed to connect to bootstrap peer")
		} else {
			log.WithField("PeerID", peerInfo.ID.String()).Info("Connected to bootstrap peer")
			bootstrapConnected++
		}
	}

	if bootstrapConnected == 0 {
		return fmt.Errorf("failed to connect to any bootstrap peers")
	}

	log.WithField("Connected", bootstrapConnected).Info("Successfully connected to bootstrap peers")

	// Create routing discovery
	routingDiscovery := routing.NewRoutingDiscovery(kadDHT)

	// Search for peers at the rendezvous point
	log.WithFields(log.Fields{
		"Rendezvous": rendezvous,
		"Timeout":    timeout,
	}).Info("Searching DHT for peers at rendezvous point...")

	searchCtx, searchCancel := context.WithTimeout(ctx, timeout)
	defer searchCancel()

	peerChan, err := routingDiscovery.FindPeers(searchCtx, rendezvous)
	if err != nil {
		return fmt.Errorf("failed to start DHT peer search: %w", err)
	}

	// Collect discovered peers
	discoveredPeers := make(map[peer.ID]*DiscoveredPeerInfo)

	// Start receiving peers
	go func() {
		for peer := range peerChan {
			// Skip ourselves
			if peer.ID == h.ID() {
				continue
			}

			// Check connectivity
			connectivity := h.Network().Connectedness(peer.ID)
			connected := connectivity == network.Connected

			// Try to connect if not already connected
			if !connected {
				connectCtx, connectCancel := context.WithTimeout(ctx, 5*time.Second)
				err := h.Connect(connectCtx, peer)
				connectCancel()
				if err == nil {
					connected = true
				}
			}

			discoveredPeers[peer.ID] = &DiscoveredPeerInfo{
				PeerID:      peer.ID.String(),
				Addrs:       peer.Addrs,
				Connected:   connected,
				DiscoveredAt: time.Now(),
			}

			log.WithFields(log.Fields{
				"PeerID":    peer.ID.String(),
				"Addrs":     peer.Addrs,
				"Connected": connected,
			}).Info("Discovered peer via DHT")
		}
	}()

	// Wait for search to complete or timeout
	<-searchCtx.Done()

	// Give goroutine a moment to finish processing
	time.Sleep(500 * time.Millisecond)

	// Print results
	log.Info("═══════════════════════════════════════════════════════════════")
	log.WithField("Rendezvous", rendezvous).Info("DHT Lookup Results")
	log.Info("═══════════════════════════════════════════════════════════════")

	if len(discoveredPeers) == 0 {
		log.Warn("No peers found at rendezvous point")
		log.Info("")
		log.Info("Possible reasons:")
		log.Info("  1. Server not advertising on DHT (check COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS)")
		log.Info("  2. Server behind CGNAT without relay (check if relay is working)")
		log.Info("  3. DHT propagation delay (advertisements can take 5-10 minutes)")
		log.Info("  4. Wrong rendezvous point (use --rendezvous to specify)")
		return nil
	}

	log.WithField("TotalPeers", len(discoveredPeers)).Info("Found peers")
	log.Info("")

	for _, peerInfo := range discoveredPeers {
		log.Info("───────────────────────────────────────────────────────────────")
		log.WithField("PeerID", peerInfo.PeerID).Info("Peer")
		log.WithField("Connected", peerInfo.Connected).Info("Connection Status")
		log.Info("Advertised Addresses:")
		for _, addr := range peerInfo.Addrs {
			log.WithField("Addr", addr.String()).Info("  • ")
		}
		log.WithField("DiscoveredAt", peerInfo.DiscoveredAt.Format(time.RFC3339)).Info("First Seen")
	}

	log.Info("═══════════════════════════════════════════════════════════════")

	return nil
}

// DiscoveredPeerInfo holds information about a discovered peer
type DiscoveredPeerInfo struct {
	PeerID       string
	Addrs        []multiaddr.Multiaddr
	Connected    bool
	DiscoveredAt time.Time
}

// getBootstrapPeers returns bootstrap peers from environment
func getBootstrapPeers() string {
	// Try client-specific env vars first
	if peers := os.Getenv("COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS"); peers != "" {
		return peers
	}
	// Fall back to server env vars
	if peers := os.Getenv("COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS"); peers != "" {
		return peers
	}
	// Legacy
	if peers := os.Getenv("COLONIES_LIBP2P_BOOTSTRAP_PEERS"); peers != "" {
		return peers
	}
	return ""
}

// parseBootstrapPeers parses a comma-separated list of bootstrap peer multiaddresses
func parseBootstrapPeers(peersStr string) []multiaddr.Multiaddr {
	var addrs []multiaddr.Multiaddr

	if peersStr == "" {
		return addrs
	}

	peerList := strings.Split(peersStr, ",")
	for _, peerStr := range peerList {
		peerStr = strings.TrimSpace(peerStr)
		if peerStr == "" {
			continue
		}

		addr, err := multiaddr.NewMultiaddr(peerStr)
		if err != nil {
			log.WithError(err).WithField("Addr", peerStr).Warn("Failed to parse bootstrap peer address")
			continue
		}
		addrs = append(addrs, addr)
	}

	return addrs
}

// ============================================================================
// Generate Command - Generate LibP2P identity
// ============================================================================

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a LibP2P identity (private key and peer ID)",
	Long: `Generate a LibP2P Ed25519 identity that can be used with:
  - COLONIES_P2P_RELAY_IDENTITY (hex-encoded, for relay nodes)
  - COLONIES_P2P_RELAY_IDENTITY_FILE (base64-encoded, for relay nodes)
  - COLONIES_SERVER_LIBP2P_IDENTITY (hex-encoded, for Colonies server)

The identity can be used for relay servers, DHT bootstrap nodes, or Colonies servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Generate Ed25519 key from libp2p
		privKey, _, err := libp2pcrypto.GenerateEd25519Key(rand.Reader)
		CheckError(err)

		// Marshal the private key to bytes
		privKeyBytes, err := libp2pcrypto.MarshalPrivateKey(privKey)
		CheckError(err)

		// Get the peer ID from the private key
		peerID, err := peer.IDFromPrivateKey(privKey)
		CheckError(err)

		// Log the results
		log.WithFields(log.Fields{
			"PeerID":  peerID.String(),
			"PrvKey":  hex.EncodeToString(privKeyBytes),
			"Example": "/ip4/127.0.0.1/tcp/4002/p2p/" + peerID.String(),
		}).Info("Generated new LibP2P identity")

		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("LibP2P Identity Generated")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println()
		fmt.Printf("Peer ID: %s\n", peerID.String())
		fmt.Println()
		fmt.Println("To use this identity:")
		fmt.Println()
		fmt.Println("Option 1: For relay node (hex-encoded):")
		fmt.Printf("  export COLONIES_P2P_RELAY_IDENTITY=\"%s\"\n", hex.EncodeToString(privKeyBytes))
		fmt.Println()
		fmt.Println("Option 2: For relay node (base64-encoded):")
		fmt.Printf("  export COLONIES_P2P_RELAY_IDENTITY_FILE=\"%s\"\n", base64.StdEncoding.EncodeToString(privKeyBytes))
		fmt.Println()
		fmt.Println("Option 3: For Colonies server (hex-encoded):")
		fmt.Printf("  export COLONIES_SERVER_LIBP2P_IDENTITY=\"%s\"\n", hex.EncodeToString(privKeyBytes))
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println()
	},
}

// ============================================================================
// Relay Command - Run LibP2P relay/bootstrap node
// ============================================================================

var relayCmd = &cobra.Command{
	Use:   "relay",
	Short: "Start a LibP2P relay and DHT bootstrap node",
	Long: `Start a LibP2P Circuit Relay v2 server and DHT bootstrap node.

This relay enables NAT traversal for Colonies clients behind firewalls or CGNAT.
The relay also acts as a DHT bootstrap node for peer discovery.

Environment variables:
  COLONIES_P2P_RELAY_PUBLIC_IP      - Public IP address to announce (required for remote access)
  COLONIES_P2P_RELAY_PORT           - Port to listen on (default: 4002, both TCP and UDP/QUIC)
  COLONIES_P2P_RELAY_IDENTITY       - Hex-encoded LibP2P identity
  COLONIES_P2P_RELAY_IDENTITY_FILE  - Base64-encoded LibP2P identity (alternative to hex)

Note: If no identity is provided, a temporary one will be generated (not persistent).

Example:
  COLONIES_P2P_RELAY_PUBLIC_IP=46.62.173.145 \
  COLONIES_P2P_RELAY_IDENTITY="080112..." \
  colonies p2p relay`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting Colonies LibP2P Relay/Bootstrap Node...")

		// Load or generate identity
		privKey, err := loadOrGenerateRelayIdentity()
		CheckError(err)

		// Get relay configuration from environment
		publicIP := os.Getenv("COLONIES_P2P_RELAY_PUBLIC_IP")
		port := os.Getenv("COLONIES_P2P_RELAY_PORT")
		if port == "" {
			port = "4002" // Default port
		}

		// Configure relay service with infinite limits for testing
		relayOpts := []relayv2.Option{
			relayv2.WithInfiniteLimits(), // Use infinite limits for testing
		}

		// Create resource manager with infinite limits
		rmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
		CheckError(err)

		// Use non-standard ports to avoid random public libp2p nodes
		listenTCP := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)
		listenQUIC := fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic-v1", port)

		opts := []libp2p.Option{
			libp2p.Identity(privKey),
			libp2p.ListenAddrStrings(
				listenTCP,  // TCP on all interfaces
				listenQUIC, // QUIC on all interfaces
			),
			libp2p.DefaultTransports,
			libp2p.EnableRelay(),
			libp2p.EnableRelayService(relayOpts...),
			libp2p.ResourceManager(rmgr),
			libp2p.EnableNATService(),
			libp2p.NATPortMap(),
			libp2p.DefaultMuxers,
			libp2p.DefaultSecurity,
		}

		log.WithFields(log.Fields{
			"Port":     port,
			"PublicIP": publicIP,
		}).Info("Relay configuration")

		// Configure address announcement
		opts = append(opts, libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			if publicIP != "" {
				publicAddrs := []multiaddr.Multiaddr{}
				publicTCP, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", publicIP, port))
				publicQUIC, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/%s/quic-v1", publicIP, port))
				if publicTCP != nil {
					publicAddrs = append(publicAddrs, publicTCP)
				}
				if publicQUIC != nil {
					publicAddrs = append(publicAddrs, publicQUIC)
				}
				return publicAddrs
			}

			// Filter out private IPs to avoid poisoning DHT
			var filtered []multiaddr.Multiaddr
			for _, addr := range addrs {
				s := addr.String()
				// Skip private IPv4 ranges
				if strings.Contains(s, "/ip4/10.") ||
					strings.Contains(s, "/ip4/192.168.") ||
					strings.Contains(s, "/ip4/172.16.") ||
					strings.Contains(s, "/ip4/172.17.") ||
					strings.Contains(s, "/ip4/172.18.") ||
					strings.Contains(s, "/ip4/172.19.") ||
					strings.Contains(s, "/ip4/172.20.") ||
					strings.Contains(s, "/ip4/172.21.") ||
					strings.Contains(s, "/ip4/172.22.") ||
					strings.Contains(s, "/ip4/172.23.") ||
					strings.Contains(s, "/ip4/172.24.") ||
					strings.Contains(s, "/ip4/172.25.") ||
					strings.Contains(s, "/ip4/172.26.") ||
					strings.Contains(s, "/ip4/172.27.") ||
					strings.Contains(s, "/ip4/172.28.") ||
					strings.Contains(s, "/ip4/172.29.") ||
					strings.Contains(s, "/ip4/172.30.") ||
					strings.Contains(s, "/ip4/172.31.") ||
					strings.Contains(s, "/ip4/127.") {
					continue
				}
				filtered = append(filtered, addr)
			}
			if len(filtered) == 0 {
				log.Warn("No public addresses detected - set PUBLIC_IP environment variable")
			}
			return filtered
		}))

		// Create libp2p host
		log.Info("Creating libp2p host with relay service...")
		h, err := libp2p.New(opts...)
		CheckError(err)

		log.WithField("PeerID", h.ID().String()).Info("Relay host created successfully")

		// Set up network event notifications
		setupRelayNetworkNotifications(h)

		// Verify relay protocols
		verifyRelayProtocols(h)

		// Log listening addresses
		log.WithField("Count", len(h.Addrs())).Info("Listening addresses:")
		for _, addr := range h.Addrs() {
			fullAddr := fmt.Sprintf("%s/p2p/%s", addr, h.ID().String())
			log.WithField("Addr", fullAddr).Info("  Listening on")
		}

		// Start DHT in server mode
		ctx := context.Background()
		log.Info("Initializing DHT in server mode...")
		kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
		CheckError(err)

		log.Info("Bootstrapping DHT...")
		err = kadDHT.Bootstrap(ctx)
		CheckError(err)

		// Connect to public bootstrap nodes
		connectToPublicBootstrap(ctx, h)

		log.Info("DHT bootstrap node started successfully")
		log.WithField("RoutingTableSize", kadDHT.RoutingTable().Size()).Info("DHT ready")

		// Print configuration instructions
		printRelayConfigurationInstructions(h)

		// Monitor relay statistics
		go monitorRelayStats(h, kadDHT)

		// Wait for interrupt
		log.Info("Relay/Bootstrap node is running. Press Ctrl+C to stop.")
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Info("Shutting down...")
		h.Close()
	},
}

// ============================================================================
// Relay Helper Functions
// ============================================================================

func loadOrGenerateRelayIdentity() (libp2pcrypto.PrivKey, error) {
	// Try COLONIES_P2P_RELAY_IDENTITY env var first (hex-encoded)
	if identityHex := os.Getenv("COLONIES_P2P_RELAY_IDENTITY"); identityHex != "" {
		log.Info("Using identity from COLONIES_P2P_RELAY_IDENTITY (hex)")
		privKeyBytes, err := hex.DecodeString(identityHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode COLONIES_P2P_RELAY_IDENTITY: %w", err)
		}
		privKey, err := libp2pcrypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
		}
		pid, _ := peer.IDFromPrivateKey(privKey)
		log.WithField("PeerID", pid.String()).Info("Identity loaded from environment")
		return privKey, nil
	}

	// Try COLONIES_P2P_RELAY_IDENTITY_FILE env var (base64-encoded content)
	if identityBase64 := os.Getenv("COLONIES_P2P_RELAY_IDENTITY_FILE"); identityBase64 != "" {
		log.Info("Using identity from COLONIES_P2P_RELAY_IDENTITY_FILE (base64)")
		privKeyBytes, err := base64.StdEncoding.DecodeString(identityBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode COLONIES_P2P_RELAY_IDENTITY_FILE (expected base64): %w", err)
		}

		privKey, err := libp2pcrypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
		}

		pid, _ := peer.IDFromPrivateKey(privKey)
		log.WithField("PeerID", pid.String()).Info("Identity loaded from environment")
		return privKey, nil
	}

	// No identity provided - generate a new one and display it
	log.Warn("No identity configured - generating temporary identity")
	log.Warn("This identity will NOT persist across restarts!")
	log.Warn("Set COLONIES_P2P_RELAY_IDENTITY or COLONIES_P2P_RELAY_IDENTITY_FILE to use a persistent identity")

	privKey, _, err := libp2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	pid, _ := peer.IDFromPrivateKey(privKey)
	privKeyBytes, _ := libp2pcrypto.MarshalPrivateKey(privKey)

	log.WithField("PeerID", pid.String()).Warn("Generated temporary identity")
	log.Warn("═══════════════════════════════════════════════════════════════")
	log.Warn("To make this identity persistent, set one of:")
	log.Warnf("  export COLONIES_P2P_RELAY_IDENTITY=\"%s\"", hex.EncodeToString(privKeyBytes))
	log.Warnf("  export COLONIES_P2P_RELAY_IDENTITY_FILE=\"%s\"", base64.StdEncoding.EncodeToString(privKeyBytes))
	log.Warn("═══════════════════════════════════════════════════════════════")

	return privKey, nil
}

func verifyRelayProtocols(h host.Host) {
	streamProtos := h.Mux().Protocols()
	log.WithField("Count", len(streamProtos)).Info("Registered protocols")

	relayHopProto := "/libp2p/circuit/relay/0.2.0/hop"
	relayStopProto := "/libp2p/circuit/relay/0.2.0/stop"

	supportsHop := false
	supportsStop := false
	for _, proto := range streamProtos {
		if string(proto) == relayHopProto {
			supportsHop = true
		}
		if string(proto) == relayStopProto {
			supportsStop = true
		}
	}

	if supportsHop {
		log.Info("✓ Relay HOP protocol registered (relay server)")
	} else {
		log.Warn("⚠️  Relay HOP protocol NOT FOUND")
	}

	if supportsStop {
		log.Info("✓ Relay STOP protocol registered (relay client)")
	} else {
		log.Warn("⚠️  Relay STOP protocol NOT FOUND")
	}
}

func setupRelayNetworkNotifications(h host.Host) {
	notifee := &network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			streams := c.GetStreams()
			hasRelayStream := false
			for _, s := range streams {
				proto := s.Protocol()
				if strings.Contains(string(proto), "circuit") || strings.Contains(string(proto), "relay") {
					hasRelayStream = true
					break
				}
			}

			if hasRelayStream || len(streams) > 1 {
				log.WithFields(log.Fields{
					"Peer":      c.RemotePeer().ShortString(),
					"Addr":      c.RemoteMultiaddr(),
					"Transport": c.ConnState().Transport,
					"Streams":   len(streams),
				}).Info("🔗 Client connected")
			}
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			streams := c.GetStreams()
			hadRelayStream := false
			for _, s := range streams {
				proto := s.Protocol()
				if strings.Contains(string(proto), "circuit") || strings.Contains(string(proto), "relay") {
					hadRelayStream = true
					break
				}
			}

			if hadRelayStream {
				log.WithFields(log.Fields{
					"Peer":     c.RemotePeer().ShortString(),
					"Duration": time.Since(c.Stat().Opened),
				}).Info("❌ Client disconnected")
			}
		},
	}

	h.Network().Notify(notifee)
	log.Info("Network event notifications enabled")
}

func connectToPublicBootstrap(ctx context.Context, h host.Host) {
	publicBootstrapPeers := []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	}

	log.Info("Connecting to public DHT bootstrap nodes...")
	connectedCount := 0
	for _, addrStr := range publicBootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.WithError(err).WithField("Addr", addrStr).Warn("Failed to parse bootstrap address")
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.WithError(err).Warn("Failed to get peer info")
			continue
		}

		connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := h.Connect(connectCtx, *peerInfo); err != nil {
			log.WithError(err).WithField("Peer", peerInfo.ID.ShortString()).Debug("Failed to connect to bootstrap peer")
			cancel()
			continue
		}
		cancel()

		connectedCount++
		log.WithField("Peer", peerInfo.ID.ShortString()).Info("✓ Connected to public bootstrap peer")
	}

	log.WithField("Connected", fmt.Sprintf("%d/%d", connectedCount, len(publicBootstrapPeers))).Info("Bootstrap connection complete")
}

func printRelayConfigurationInstructions(h host.Host) {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("CONFIGURATION INSTRUCTIONS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Add these lines to your docker-compose.env file:")
	fmt.Println()

	if len(h.Addrs()) > 0 {
		var tcpAddr, quicAddr string

		for _, addr := range h.Addrs() {
			addrStr := addr.String()
			if strings.Contains(addrStr, "127.0.0.1") {
				continue
			}

			if strings.Contains(addrStr, "/tcp/") {
				tcpAddr = fmt.Sprintf("%s/p2p/%s", addr, h.ID())
			} else if strings.Contains(addrStr, "/quic") {
				quicAddr = fmt.Sprintf("%s/p2p/%s", addr, h.ID())
			}
		}

		if tcpAddr != "" {
			fmt.Println("# Server configuration (TCP works on home networks)")
			fmt.Printf("export COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS=\"%s\"\n", tcpAddr)
			fmt.Println()
		}
		if quicAddr != "" {
			fmt.Println("# Client configuration (QUIC recommended for 5G/mobile)")
			fmt.Printf("export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS=\"%s\"\n", quicAddr)
		} else if tcpAddr != "" {
			fmt.Println("# Client configuration (fallback to TCP)")
			fmt.Printf("export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS=\"%s\"\n", tcpAddr)
		}
	}

	fmt.Println()
	fmt.Println("NOTE: QUIC (UDP) is recommended for clients behind CGNAT (5G/mobile)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}

func monitorRelayStats(h host.Host, kadDHT *dht.IpfsDHT) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		peers := h.Network().Peers()
		conns := h.Network().Conns()

		relayConns := 0
		tcpConns := 0
		quicConns := 0

		for _, conn := range conns {
			remoteAddr := conn.RemoteMultiaddr().String()
			if strings.Contains(remoteAddr, "/p2p-circuit") {
				relayConns++
			}
			if strings.Contains(remoteAddr, "/tcp/") {
				tcpConns++
			}
			if strings.Contains(remoteAddr, "/quic") {
				quicConns++
			}
		}

		log.Info("═══════════════════════════════════════════════════════")
		log.WithFields(log.Fields{
			"Peers":       len(peers),
			"Connections": len(conns),
			"TCP":         tcpConns,
			"QUIC":        quicConns,
			"Relay":       relayConns,
			"DHT":         kadDHT.RoutingTable().Size(),
		}).Info("📊 Relay Statistics")

		if len(peers) > 0 && len(peers) <= 5 {
			log.Info("Connected peers:")
			for _, p := range peers {
				peerConns := h.Network().ConnsToPeer(p)
				log.WithFields(log.Fields{
					"Peer":  p.ShortString(),
					"Conns": len(peerConns),
				}).Info("  •")
			}
		} else if len(peers) > 5 {
			log.WithField("Total", len(peers)).Info("Connected peers (showing first 5)")
			for i, p := range peers[:5] {
				peerConns := h.Network().ConnsToPeer(p)
				log.WithFields(log.Fields{
					"Index": i + 1,
					"Peer":  p.ShortString(),
					"Conns": len(peerConns),
				}).Info("  •")
			}
		}
		log.Info("═══════════════════════════════════════════════════════")
	}
}
