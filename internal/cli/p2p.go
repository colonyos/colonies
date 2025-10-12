package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	p2pCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(p2pCmd)

	lookupCmd.Flags().StringVarP(&Text, "rendezvous", "r", "colonies-server", "Rendezvous point to search for (default: colonies-server)")
	lookupCmd.Flags().IntVarP(&Timeout, "timeout", "t", 30, "Timeout in seconds for DHT discovery")
}

var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "LibP2P and DHT debugging commands",
	Long:  "LibP2P and DHT debugging commands for troubleshooting peer discovery",
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
