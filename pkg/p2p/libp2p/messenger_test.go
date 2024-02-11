package libp2p

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/go-playground/assert/v2"
)

func TestMessenger(t *testing.T) {
	messenger1, err := CreateMessenger([]string{"/ip4/10.0.0.201/tcp/4001", "/ip4/127.0.0.1/tcp/4001"})
	assert.Equal(t, err, nil)

	messenger2, err := CreateMessenger([]string{"/ip4/10.0.0.201/tcp/4002", "/ip4/127.0.0.1/tcp/4002"})
	assert.Equal(t, err, nil)

	msgChan := make(chan p2p.Message)
	ctx := context.TODO()
	go func() {
		messenger1.ListenForever(msgChan, ctx)
	}()

	for {
		err := messenger2.Send(p2p.Message{From: messenger2.Node, To: messenger1.Node, Payload: []byte("Hello")}, context.TODO())
		if err != nil {
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}

	msg := <-msgChan

	assert.Equal(t, string(msg.Payload), "Hello")
	assert.Equal(t, msg.From.HostID, messenger2.Node.HostID)
	assert.Equal(t, msg.To.HostID, messenger1.Node.HostID)

	err = messenger2.Send(p2p.Message{From: messenger2.Node, To: messenger1.Node, Payload: []byte("Hello 2")}, context.TODO())
	assert.Equal(t, err, nil)

	msg = <-msgChan
	assert.Equal(t, string(msg.Payload), "Hello 2")
}
