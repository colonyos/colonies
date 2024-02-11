package p2p

import "context"

type Messenger interface {
	Send(msg Message, ctx context.Context) error
	ListenForever(msgChan chan Message, ctx context.Context) error
}
