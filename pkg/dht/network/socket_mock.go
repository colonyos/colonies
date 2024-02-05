package network

type FakeSocket struct {
	conn chan *Message
}

func (socket *FakeSocket) Send(msg *Message) error {
	socket.conn <- msg
	return nil
}

func (socket *FakeSocket) Receive() (*Message, error) {
	return <-socket.conn, nil
}
