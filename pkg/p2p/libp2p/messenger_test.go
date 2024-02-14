package libp2p

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/go-playground/assert/v2"
)

func TestMessenger(t *testing.T) {
	messenger1, err := CreateMessenger(4001, "mes1")
	assert.Equal(t, err, nil)

	messenger2, err := CreateMessenger(4002, "mes2")
	assert.Equal(t, err, nil)

	msgChan := make(chan p2p.Message)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		messenger1.ListenForever(msgChan, ctx)
		cancel()
	}()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		err := messenger2.Send(p2p.Message{From: messenger2.Node, To: messenger1.Node, Payload: []byte("Hello")}, ctx)
		cancel()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	msg := <-msgChan

	assert.Equal(t, string(msg.Payload), "Hello")

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = messenger2.Send(p2p.Message{From: messenger2.Node, To: messenger1.Node, Payload: []byte("Hello 2")}, ctx)
	cancel()
	assert.Equal(t, err, nil)

	msg = <-msgChan
	assert.Equal(t, string(msg.Payload), "Hello 2")
}
