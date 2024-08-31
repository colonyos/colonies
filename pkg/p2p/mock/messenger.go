package mock

import (
	"context"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/p2p"
)

type MockMessenger struct {
	network     Network
	node        p2p.Node
	pendingMsgs map[string]chan *p2p.Message
}

func CreateMessenger(network Network, node p2p.Node) *MockMessenger {
	return &MockMessenger{network: network, node: node, pendingMsgs: make(map[string]chan *p2p.Message)}
}

func (m *MockMessenger) SendAndForget(msg p2p.Message, ctx context.Context) error {
	socket, err := m.network.Dial(msg.To.String())
	if err != nil {
		return err
	}
	msg.From = m.node
	return socket.Send(msg)

}

func (m *MockMessenger) SendWithReply(msg p2p.Message, replyChan chan *p2p.Message, ctx context.Context) error {
	socket, err := m.network.Dial(msg.To.String())
	if err != nil {
		return err
	}
	msg.From = m.node
	msg.ID = core.GenerateRandomID()

	fmt.Println("adding ", msg.ID, " to pending messages")
	m.pendingMsgs[msg.ID] = replyChan

	err = socket.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func (m *MockMessenger) Reply(msg p2p.Message, reply string, ctx context.Context) error {
	replyMsg := p2p.Message{From: m.node, To: msg.From, Payload: []byte(reply)}

	socket, err := m.network.Dial(replyMsg.To.String())
	if err != nil {
		return err
	}
	msg.From = m.node
	fmt.Println("Replying to", replyMsg.To.String())
	return socket.Send(replyMsg)

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
			fmt.Println(msg.ID, "Received message from ", msg.From.String())
			fmt.Println("----------------------")
			for k, v := range m.pendingMsgs {
				fmt.Println(k, v)
			}
			fmt.Println("----------------------")
			if replyChan, ok := m.pendingMsgs[msg.ID]; ok {
				fmt.Println("Received reply for ", msg.ID)
				replyChan <- &msg
				delete(m.pendingMsgs, msg.ID)
			} else {
				msgChan <- msg
			}
		}
	}
}
