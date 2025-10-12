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
	h, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	log.Printf("✓ Relay host created")
	log.Printf("  Peer ID: %s", h.ID().String())

	// Verify relay protocols are registered
	protocols := h.Mux().Protocols()
	hasCircuitV2 := false
	for _, proto := range protocols {
		if proto == "/libp2p/circuit/relay/0.2.0/hop" {
			hasCircuitV2 = true
			log.Printf("  ✓ Circuit v2 relay protocol registered: %s", proto)
		}
	}
	if !hasCircuitV2 {
		log.Println("  ⚠️  WARNING: Circuit v2 relay protocol NOT found!")
	}

	log.Printf("  Listening on:")
	for _, addr := range h.Addrs() {
		log.Printf("    %s/p2p/%s", addr, h.ID().String())
	}

	// Start DHT in server mode
	ctx := context.Background()
	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		log.Fatalf("Failed to create DHT: %v", err)
	}

	if err = kadDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("Failed to bootstrap DHT: %v", err)
	}

	log.Println("✓ DHT bootstrap node started")

	// Print configuration instructions
	printConfigurationInstructions(h)

	// Monitor relay statistics
	go monitorStats(h)

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

	// Try to load existing identity
	if data, err := os.ReadFile(identityFile); err == nil {
		log.Printf("Loading existing identity from %s", identityFile)
		privKeyBytes, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode identity: %w", err)
		}
		return crypto.UnmarshalPrivateKey(privKeyBytes)
	}

	// Generate new identity
	log.Printf("Generating new identity...")
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Save identity
	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(privKeyBytes)
	if err := os.WriteFile(identityFile, []byte(encoded), 0600); err != nil {
		return nil, fmt.Errorf("failed to save identity: %w", err)
	}

	log.Printf("✓ Identity saved to %s", identityFile)
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

func monitorStats(h host.Host) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		peers := h.Network().Peers()
		conns := h.Network().Conns()

		// Count relay circuits
		relayConns := 0
		for _, conn := range conns {
			// Check if this is a relay connection by looking at the remote address
			remoteAddr := conn.RemoteMultiaddr().String()
			if strings.Contains(remoteAddr, "/p2p-circuit") {
				relayConns++
			}
		}

		log.Printf("Stats: %d peers, %d connections (%d relay circuits)", len(peers), len(conns), relayConns)
	}
}
