package mock

import (
	"context"

	"github.com/colonyos/colonies/pkg/p2p"
)

type FakeSocket struct {
	conn chan p2p.Message
}

func (socket *FakeSocket) Send(msg p2p.Message) error {
	socket.conn <- msg
	return nil
}

func (socket *FakeSocket) Receive(ctx context.Context) (p2p.Message, error) {
	select {
	case msg := <-socket.conn:
		return msg, nil
	case <-ctx.Done():
		return p2p.Message{}, ctx.Err()
	}
}
