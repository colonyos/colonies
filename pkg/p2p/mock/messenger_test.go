package mock

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/go-playground/assert/v2"
)

func TestMessenger(t *testing.T) {
	n := CreateFakeNetwork()

	node1 := p2p.Node{HostID: "node1", Addr: "10.0.0.1:1111"}
	messenger1 := CreateMessenger(n, node1)

	node2 := p2p.Node{HostID: "node2", Addr: "10.0.0.2:1111"}
	messenger2 := CreateMessenger(n, node2)

	msgChan := make(chan p2p.Message)
	ctx := context.TODO()
	go func() {
		messenger1.ListenForever(msgChan, ctx)
	}()

	for {
		err := messenger2.Send(p2p.Message{From: node2, To: node1, Payload: []byte("Hello")}, context.TODO())
		if err != nil {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	msg := <-msgChan

	assert.Equal(t, string(msg.Payload), "Hello")
	assert.Equal(t, msg.From.HostID, "node2")
	assert.Equal(t, msg.To.HostID, "node1")
	assert.Equal(t, msg.From.Addr, "10.0.0.2:1111")
	assert.Equal(t, msg.To.Addr, "10.0.0.1:1111")

	err := messenger2.Send(p2p.Message{From: node2, To: node1, Payload: []byte("Hello 2")}, context.TODO())
	assert.Equal(t, err, nil)

	msg = <-msgChan
	assert.Equal(t, string(msg.Payload), "Hello 2")
}
