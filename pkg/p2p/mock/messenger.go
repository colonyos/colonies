package mock

import (
	"context"

	"github.com/colonyos/colonies/pkg/p2p"
)

type MockMessenger struct {
	network Network
	node    p2p.Node
}

func CreateMessenger(network Network, node p2p.Node) *MockMessenger {
	return &MockMessenger{network: network, node: node}
}

func (m *MockMessenger) Send(msg p2p.Message, ctx context.Context) error {
	socket, err := m.network.Dial(msg.To.String())
	if err != nil {
		return err
	}
	msg.From = m.node
	return socket.Send(msg)
}

func (m *MockMessenger) ListenForever(msgChan chan p2p.Message, ctx context.Context) error {
	socket, err := m.network.Listen(m.node.String())
	if err != nil {
		return err
	}

	for {
		msg, _ := socket.Receive(ctx)
		select {
		case <-ctx.Done():
			return nil
		default:
			msgChan <- msg
		}
	}
}
