package p2p

import "context"

type Messenger interface {
	SendAdForget(msg Message, ctx context.Context) error
	SendWithReply(msg Message, replyChan *chan Message, ctx context.Context) error
	ListenForever(msgChan chan Message, ctx context.Context) error
}
