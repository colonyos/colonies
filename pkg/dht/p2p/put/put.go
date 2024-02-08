package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

func formatKey(originalKey string) string {
	hash := sha256.Sum256([]byte(originalKey))
	return hex.EncodeToString(hash[:])
}

func main() {
	ctx := context.Background()

	// Create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create h: %s", err)
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

	fmt.Println("This node's addresses:")
	for _, addr := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", addr, h.ID())
	}

	// Use a hash function to convert the key to a suitable format for the DHT
	key := "myKey"
	//hashedKey := sha256.Sum256([]byte(key))
	//keyString := fmt.Sprintf("/record/%x", hashedKey)
	formattedKey := formatKey(key)

	value := []byte("Hello World")
	err = kadDHT.PutValue(ctx, formattedKey, value)
	if err != nil {
		log.Fatalf("Failed to put value in DHT: %s", err)
	}

	fmt.Println("Successfully stored value in DHT")

	// Keep the host alive until the user interrupts the program
	select {}
}
