package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	log.Println("Starting Colonies Relay/Bootstrap Node...")

	// Load or generate identity
	privKey, err := loadOrGenerateIdentity()
	if err != nil {
		log.Fatal(err)
	}

	// Check for public IP from environment
	publicIP := os.Getenv("PUBLIC_IP")

	opts := []libp2p.Option{
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",        // TCP on all interfaces
			"/ip4/0.0.0.0/udp/4001/quic-v1", // QUIC on all interfaces
		),
		libp2p.EnableRelay(),              // Enable relay transport
		libp2p.EnableRelayService(         // Provide relay service
			relay.WithLimit(nil), // No limits (adjust for production)
		),
		libp2p.EnableNATService(),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	// If public IP is provided, announce it
	if publicIP != "" {
		log.Printf("Using public IP from environment: %s", publicIP)
		opts = append(opts, libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			// Replace private IPs with public IP
			publicAddrs := []multiaddr.Multiaddr{}

			// Add public IP announcements
			publicTCP, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/4001", publicIP))
			publicQUIC, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/4001/quic-v1", publicIP))

			if publicTCP != nil {
				publicAddrs = append(publicAddrs, publicTCP)
			}
			if publicQUIC != nil {
				publicAddrs = append(publicAddrs, publicQUIC)
			}

			return publicAddrs
		}))
	} else {
		log.Println("WARNING: PUBLIC_IP not set - relay will announce private IP addresses")
		log.Println("Set PUBLIC_IP environment variable: export PUBLIC_IP=your.public.ip.address")
	}

	// Create libp2p host with relay support
	log.Println("Creating libp2p host with options:")
	log.Printf("  - Identity: %v", privKey != nil)
	log.Printf("  - Listen: /ip4/0.0.0.0/tcp/4001, /ip4/0.0.0.0/udp/4001/quic-v1")
	log.Printf("  - Relay: enabled (client)")
	log.Printf("  - Relay Service: enabled (server)")
	log.Printf("  - NAT Service: enabled")
	log.Printf("  - Public IP: %s", publicIP)

	h, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	log.Printf("✓ Relay host created successfully")
	log.Printf("  Peer ID: %s", h.ID().String())

	// Set up network event notifications
	setupNetworkNotifications(h)

	// Verify relay protocols are registered
	streamProtos := h.Mux().Protocols()
	log.Printf("  Registered mux protocols (%d total):", len(streamProtos))
	for _, proto := range streamProtos {
		log.Printf("    - %s", proto)
	}

	// Try to open a test stream to verify relay service is actually running
	log.Println("  Checking relay service protocols...")
	relayHopProto := "/libp2p/circuit/relay/0.2.0/hop"
	relayStopProto := "/libp2p/circuit/relay/0.2.0/stop"

	// We can't directly check stream handlers, but we can verify by checking
	// if the protocols are in the supported protocols list
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
		log.Println("    ✓ Relay HOP protocol registered (relay server)")
	} else {
		log.Println("    ⚠️  Relay HOP protocol NOT FOUND - relay server will not work!")
	}

	if supportsStop {
		log.Println("    ✓ Relay STOP protocol registered (relay client)")
	} else {
		log.Println("    ⚠️  Relay STOP protocol NOT FOUND")
	}

	log.Println("  Note: Circuit v2 hop may be registered as stream handler, not visible in Mux().Protocols()")

	log.Printf("  Listening addresses (%d):", len(h.Addrs()))
	for _, addr := range h.Addrs() {
		fullAddr := fmt.Sprintf("%s/p2p/%s", addr, h.ID().String())
		log.Printf("    %s", fullAddr)
	}

	// Log network configuration
	log.Println("  Network configuration:")
	log.Printf("    - Connectedness: %v", h.Network().Connectedness(h.ID()))
	log.Printf("    - Network peers: %d", len(h.Network().Peers()))
	log.Printf("    - Network conns: %d", len(h.Network().Conns()))

	// Start DHT in server mode
	ctx := context.Background()
	log.Println("Initializing DHT...")
	log.Printf("  - Mode: Server")
	log.Printf("  - Protocol: /kad/1.0.0")

	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatalf("Failed to create DHT: %v", err)
	}

	log.Println("Bootstrapping DHT...")
	if err = kadDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("Failed to bootstrap DHT: %v", err)
	}

	// Connect to public libp2p bootstrap nodes to join the wider DHT network
	log.Println("Connecting to public DHT bootstrap nodes...")
	publicBootstrapPeers := []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
	}

	connectedToPublic := 0
	for _, addrStr := range publicBootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("  ⚠️  Invalid bootstrap address: %s", addrStr)
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Printf("  ⚠️  Failed to parse peer info: %s", err)
			continue
		}

		connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err = h.Connect(connectCtx, *peerInfo)
		cancel()

		if err != nil {
			log.Printf("  ⚠️  Failed to connect to %s: %v", peerInfo.ID.ShortString(), err)
		} else {
			connectedToPublic++
			log.Printf("  ✓ Connected to public bootstrap: %s", peerInfo.ID.ShortString())
		}
	}

	log.Println("✓ DHT bootstrap node started successfully")
	log.Printf("  - Routing table size: %d", kadDHT.RoutingTable().Size())
	log.Printf("  - Connected to public DHT nodes: %d/%d", connectedToPublic, len(publicBootstrapPeers))

	// Print configuration instructions
	printConfigurationInstructions(h)

	// Monitor relay statistics with detailed logging
	go monitorStats(h, kadDHT)

	// Wait for interrupt
	log.Println("\nRelay/Bootstrap node is running. Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("\nShutting down...")
	h.Close()
}

func loadOrGenerateIdentity() (crypto.PrivKey, error) {
	identityFile := "relay-identity.key"
	log.Printf("Looking for identity file: %s", identityFile)

	// Try to load existing identity
	if data, err := os.ReadFile(identityFile); err == nil {
		log.Printf("✓ Found existing identity file")
		log.Printf("  File size: %d bytes", len(data))

		privKeyBytes, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode identity: %w", err)
		}
		log.Printf("  Decoded %d bytes", len(privKeyBytes))

		privKey, err := crypto.UnmarshalPrivateKey(privKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
		}

		// Derive peer ID from private key
		pid, err := peer.IDFromPrivateKey(privKey)
		if err != nil {
			return nil, fmt.Errorf("failed to derive peer ID: %w", err)
		}
		log.Printf("✓ Identity loaded successfully")
		log.Printf("  Peer ID: %s", pid.String())

		return privKey, nil
	}

	// Generate new identity
	log.Printf("No existing identity found, generating new one...")
	privKey, pubKey, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	log.Printf("✓ Generated Ed25519 key pair")

	// Derive peer ID
	pid, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive peer ID: %w", err)
	}
	log.Printf("  New Peer ID: %s", pid.String())

	// Save identity
	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(privKeyBytes)
	if err := os.WriteFile(identityFile, []byte(encoded), 0600); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	log.Printf("✓ Identity saved to %s (%d bytes)", identityFile, len(encoded))
	return privKey, nil
}

func printConfigurationInstructions(h host.Host) {
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════════")
	fmt.Println("CONFIGURATION INSTRUCTIONS")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("\nAdd these lines to your docker-compose.env file:")
	fmt.Println()

	// Get the first multiaddr (typically the public one)
	if len(h.Addrs()) > 0 {
		// Try to find a non-localhost address
		var publicAddr string
		for _, addr := range h.Addrs() {
			addrStr := addr.String()
			if addrStr != "/ip4/127.0.0.1/tcp/4001" && addrStr != "/ip4/127.0.0.1/udp/4001/quic-v1" {
				publicAddr = fmt.Sprintf("%s/p2p/%s", addr, h.ID())
				break
			}
		}

		if publicAddr == "" {
			publicAddr = fmt.Sprintf("%s/p2p/%s", h.Addrs()[0], h.ID())
		}

		fmt.Printf("# Server configuration\n")
		fmt.Printf("export COLONIES_SERVER_LIBP2P_BOOTSTRAP_PEERS=\"%s\"\n", publicAddr)
		fmt.Println()
		fmt.Printf("# Client configuration\n")
		fmt.Printf("export COLONIES_CLIENT_LIBP2P_BOOTSTRAP_PEERS=\"%s\"\n", publicAddr)
	}

	fmt.Println()
	fmt.Println("NOTE: Replace the IP address with your PUBLIC IP if this relay is behind NAT")
	fmt.Printf("Example: /ip4/YOUR_PUBLIC_IP/tcp/4001/p2p/%s\n", h.ID())
	fmt.Println("═══════════════════════════════════════════════════════════════\n")
}

func setupNetworkNotifications(h host.Host) {
	log.Println("Setting up network event notifications...")

	notifee := &network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			log.Printf("🔗 CONNECTED: Peer %s connected", c.RemotePeer().ShortString())
			log.Printf("   Remote addr: %s", c.RemoteMultiaddr())
			log.Printf("   Local addr: %s", c.LocalMultiaddr())
			log.Printf("   Direction: %s", c.Stat().Direction)
			log.Printf("   Transport: %s", c.ConnState().Transport)

			// Log active streams on this connection
			streams := c.GetStreams()
			if len(streams) > 0 {
				log.Printf("   Active streams: %d", len(streams))
				for _, s := range streams {
					proto := s.Protocol()
					log.Printf("     - %s", proto)
					if strings.Contains(string(proto), "circuit") || strings.Contains(string(proto), "relay") {
						log.Printf("       ⚡ RELAY/CIRCUIT STREAM DETECTED!")
					}
				}
			}
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			log.Printf("❌ DISCONNECTED: Peer %s disconnected", c.RemotePeer().ShortString())
			log.Printf("   Remote addr: %s", c.RemoteMultiaddr())
			log.Printf("   Connection duration: %v", time.Since(c.Stat().Opened))
		},
		ListenF: func(n network.Network, addr multiaddr.Multiaddr) {
			log.Printf("👂 LISTENING on: %s", addr)
		},
		ListenCloseF: func(n network.Network, addr multiaddr.Multiaddr) {
			log.Printf("👂 LISTEN CLOSED: %s", addr)
		},
	}

	h.Network().Notify(notifee)
	log.Println("✓ Network event notifications enabled")
}

func monitorStats(h host.Host, kadDHT *dht.IpfsDHT) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		peers := h.Network().Peers()
		conns := h.Network().Conns()

		// Count relay circuits and protocol types
		relayConns := 0
		tcpConns := 0
		quicConns := 0
		protocolCounts := make(map[string]int)

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

			// Count streams by protocol
			streams := conn.GetStreams()
			for _, stream := range streams {
				protocolCounts[string(stream.Protocol())]++
			}
		}

		log.Println("═══════════════════════════════════════════════════════")
		log.Printf("📊 RELAY STATISTICS")
		log.Printf("  Peers: %d", len(peers))
		log.Printf("  Connections: %d (TCP: %d, QUIC: %d, Relay: %d)", len(conns), tcpConns, quicConns, relayConns)
		log.Printf("  DHT Routing Table: %d entries", kadDHT.RoutingTable().Size())

		if len(protocolCounts) > 0 {
			log.Printf("  Active protocols:")
			for proto, count := range protocolCounts {
				log.Printf("    - %s: %d streams", proto, count)
			}
		}

		// List connected peers with details
		if len(peers) > 0 {
			log.Printf("  Connected peers:")
			for i, p := range peers {
				if i >= 5 { // Limit to first 5 peers
					log.Printf("    ... and %d more", len(peers)-5)
					break
				}
				peerConns := h.Network().ConnsToPeer(p)
				log.Printf("    - %s (%d conns)", p.ShortString(), len(peerConns))
				for _, conn := range peerConns {
					log.Printf("      %s", conn.RemoteMultiaddr())
				}
			}
		}
		log.Println("═══════════════════════════════════════════════════════")
	}
}
