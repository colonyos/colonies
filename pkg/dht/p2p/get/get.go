package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
	ctx := context.Background()

	// Create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create host: %s", err)
	}

	// Set up a DHT for peer discovery
	kadDHT, err := dht.New(ctx, h)
	if err != nil {
		log.Fatalf("Failed to create the DHT: %s", err)
	}

	// Bootstrap the DHT. In the real world, we would need to connect to bootstrap nodes here.
	if err = kadDHT.Bootstrap(ctx); err != nil {
		log.Fatalf("Failed to bootstrap the DHT: %s", err)
	}

	// Connect to the other libp2p instance using its multiaddress
	// Replace <peerMultiAddr> with the actual multiaddress of the peer that stored the value
	peerMultiAddr := "<peerMultiAddr>"
	peerAddr, err := peer.AddrInfoFromP2pAddr(peerMultiAddr)
	if err != nil {
		log.Fatalf("Failed to parse peer multiaddr: %s", err)
	}
	if err := h.Connect(ctx, *peerAddr); err != nil {
		log.Fatalf("Failed to connect to peer: %s", err)
	}

	// Fetch the value from the DHT
	key := "myKey"
	value, err := kadDHT.GetValue(ctx, key)
	if err != nil {
		log.Fatalf("Failed to get value from DHT: %s", err)
	}

	fmt.Printf("Fetched value from DHT: %s\n", string(value))
}
