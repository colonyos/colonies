package libp2p

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type P2PNetwork struct {
	host host.Host
}

func CreateClient() (*P2PNetwork, error) {
	h, err := libp2p.New(
		libp2p.NATPortMap(),
		libp2p.EnableRelay())

	if err != nil {
		return nil, err
	}

	return &P2PNetwork{host: h}, nil
}

func CreateServer(addr string) (*P2PNetwork, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(addr),
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, err
	}

	return &P2PNetwork{host: h}, nil
}

func (n *P2PNetwork) Serve(addr string) error {
	n.host.SetStreamHandler("/colonies/1.0.0", func(stream network.Stream) {
		socket := &P2PSocket{stream: stream}
		msg, err := socket.Receive(context.TODO())
		fmt.Println("received message: ", string(msg.Payload), err)
	})

	return nil
}

func (n *P2PNetwork) Dial(peerID string, ctx context.Context) (*P2PSocket, error) {
	p, err := peer.Decode(peerID)
	if err != nil {
		return nil, err
	}

	hostAddrs := []string{
		"/ip4/10.0.0.201/tcp/4001", "/ip4/127.0.0.1/tcp/4001",
	}

	for _, addrStr := range hostAddrs {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			log.Fatalf("Failed to parse multiaddr \"%s\": %s", addrStr, err)
		}
		n.host.Peerstore().AddAddr(p, addr, peerstore.PermanentAddrTTL)
	}

	s, err := n.host.NewStream(ctx, p, "/colonies/1.0.0")
	if err != nil {
		return nil, err
	}

	return &P2PSocket{stream: s}, nil
}

func (n *P2PNetwork) ID() string {
	return n.host.ID().String()
}
