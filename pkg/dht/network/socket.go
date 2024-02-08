package network

import "context"

type Socket interface {
	Send(msg Message) error
	Receive(context.Context) (Message, error)
}
