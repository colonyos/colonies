package network

import "context"

type FakeSocket struct {
	conn chan Message
}

func (socket *FakeSocket) Send(msg Message) error {
	socket.conn <- msg
	return nil
}

func (socket *FakeSocket) Receive(ctx context.Context) (Message, error) {
	select {
	case msg := <-socket.conn:
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}
