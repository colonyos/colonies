package network

type Socket interface {
	Send(msg *Message) error
	Receive() (*Message, error)
}
