package main

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	ctx := context.Background()

	// Create a new libp2p Host for the client.
	clientHost, err := libp2p.New(
		libp2p.NATPortMap(),
		libp2p.EnableRelay(),
	//	libp2p.EnableAutoRelay(), // EnableAutoRelayWithStaticRelays or EnableAutoRelayWithPeerSource
	)

	// The host's Peer ID as a string.
	peerIDStr := "12D3KooWPFWVJaBcQypmAzke9Qoy3vHRUCacstwgdxVUxbG6s4aJ"

	// Convert the Peer ID string to a peer.ID.
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		log.Fatalf("Failed to parse peer ID: %s", err)
	}

	// Known multiaddresses of the host (replace these with actual addresses).
	hostAddrs := []string{
		"/ip4/10.0.0.201/tcp/4001", "/ip4/127.0.0.1/tcp/4001",
	}

	// Convert the string addresses to multiaddr and add them to the peerstore.
	for _, addrStr := range hostAddrs {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			log.Fatalf("Failed to parse multiaddr \"%s\": %s", addrStr, err)
		}
		clientHost.Peerstore().AddAddr(peerID, addr, peerstore.PermanentAddrTTL)
	}

	// Establish a new stream to the host using the echo protocol.
	stream, err := clientHost.NewStream(ctx, peerID, "/echo/1.0.0")
	if err != nil {
		log.Fatalf("Failed to establish a new stream: %s", err)
	}

	// Send a message to the host.
	message := "Hello, host!"
	fmt.Fprintf(stream, message+"\n")

	// Read the host's response (which should echo the message).
	response, err := bufio.NewReader(stream).ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read response: %s", err)
	}

	fmt.Printf("Received response: %s", response)

	// Close the stream.
	stream.Close()

	// Close the host.
	clientHost.Close()
}
