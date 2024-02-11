package mock

import (
	"context"

	"github.com/colonyos/colonies/pkg/p2p"
)

type Socket interface {
	Send(msg p2p.Message) error
	Receive(context.Context) (p2p.Message, error)
}
