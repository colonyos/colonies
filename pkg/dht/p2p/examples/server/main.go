package main

import (
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
)

func main() {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/4001"),
		libp2p.NATPortMap(), // Enables NAT traversal
	)
	if err != nil {
		log.Fatalf("Failed to create h1: %s", err)
	}

	host.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		fmt.Println("Got a new stream!")
		if _, err := s.Write([]byte("Hello, I'm a libp2p node!\n")); err != nil {
			log.Println("Error writing to stream:", err)
		}
		s.Close()
	})

	fmt.Println("Listen addresses:", host.Addrs())

	peerID := host.ID()
	fmt.Println("Host ID:", peerID)

	select {}
}
