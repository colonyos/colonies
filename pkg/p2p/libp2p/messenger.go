package libp2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

type Messenger struct {
	Node p2p.Node
	host host.Host
}

func CreateMessenger(addr []string) (*Messenger, error) {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(addr[0]), // TODO: bind to all addresses
		libp2p.NATPortMap(),
	)
	if err != nil {
		return nil, err
	}

	return &Messenger{host: host, Node: p2p.Node{HostID: host.ID().String(), Addr: addr}}, nil
}

func (m *Messenger) Send(msg p2p.Message, ctx context.Context) error {
	host, err := libp2p.New(
		libp2p.NATPortMap(),
		libp2p.EnableRelay())
	if err != nil {
		return err
	}

	p, err := peer.Decode(msg.To.HostID)
	if err != nil {
		return err
	}

	for _, addrStr := range msg.To.Addr {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			log.Fatalf("Failed to parse multiaddr \"%s\": %s", addrStr, err)
		}
		host.Peerstore().AddAddr(p, addr, peerstore.PermanentAddrTTL)
	}

	stream, err := host.NewStream(ctx, p, "/colonies/1.0.0")
	if err != nil {
		return err
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	_, err = stream.Write(lengthBytes)
	if err != nil {
		return err
	}

	_, err = stream.Write(data)

	return err
}

func (m *Messenger) ListenForever(msgChan chan p2p.Message, ctx context.Context) error {
	m.host.SetStreamHandler("/colonies/1.0.0", func(stream network.Stream) {
		var msg p2p.Message
		r := bufio.NewReader(stream)

		select {
		case <-ctx.Done():
			return
		default:
			lengthBytes := make([]byte, 4)
			_, err := io.ReadFull(r, lengthBytes)
			if err != nil {
				return
			}
			length := binary.BigEndian.Uint32(lengthBytes)

			data := make([]byte, length)
			_, err = io.ReadFull(r, data)
			if err != nil {
				return
			}

			err = json.Unmarshal(data, &msg)
			if err != nil {
				return
			}

			msgChan <- msg
		}
	})

	return nil
}
