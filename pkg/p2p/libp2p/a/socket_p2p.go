package libp2p

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/network"
)

type P2PSocket struct {
	stream network.Stream
}

func (s *P2PSocket) Send(msg p2p.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	_, err = s.stream.Write(lengthBytes)
	if err != nil {
		return err
	}

	_, err = s.stream.Write(data)
	return err
}

func (s *P2PSocket) Receive(ctx context.Context) (p2p.Message, error) {
	var msg p2p.Message
	r := bufio.NewReader(s.stream)

	fmt.Println("Receive 1 !!!!!")
	select {
	case <-ctx.Done():
		return msg, ctx.Err()
	default:
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(r, lengthBytes)
		if err != nil {
			return msg, err
		}
		length := binary.BigEndian.Uint32(lengthBytes)

		data := make([]byte, length)
		_, err = io.ReadFull(r, data)
		if err != nil {
			return msg, err
		}

		err = json.Unmarshal(data, &msg)
		if err != nil {
			return msg, err
		}

		fmt.Println("Receive 10 !!!!!")
		return msg, nil
	}
}
