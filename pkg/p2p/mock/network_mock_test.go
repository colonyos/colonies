package mock

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNetworkMock(t *testing.T) {
	n := CreateFakeNetwork()

	resultChan := make(chan p2p.Message)

	socket, err := n.Listen("10.0.0.1:8080")
	assert.NotNil(t, socket)
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		message, _ := socket.Receive(ctx)
		resultChan <- message
	}()

	socket2, err := n.Dial("10.0.0.1:8080")
	assert.Nil(t, err)
	assert.NotNil(t, socket2)
	go func() {
		socket.Send(p2p.Message{
			To:      p2p.Node{HostID: "", Addr: []string{"10.0.0.1:8080"}},
			From:    p2p.Node{HostID: "", Addr: []string{"10.0.0.2:8080"}},
			Payload: []byte("test_payload")})
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timeout")
	case msg := <-resultChan:
		cancel()
		assert.Equal(t, "test_payload", string(msg.Payload))
		assert.Equal(t, "10.0.0.1:8080", msg.To.Addr[0])
		assert.Equal(t, "10.0.0.2:8080", msg.From.Addr[0])
	}
}

func TestNetworkMockCancel(t *testing.T) {
	n := CreateFakeNetwork()

	socket, err := n.Listen("10.0.0.1:8080")
	assert.NotNil(t, socket)
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	doneChan := make(chan struct{})

	go func() {
		socket.Receive(ctx)
		doneChan <- struct{}{}
	}()

	cancel()
	done := <-doneChan
	assert.NotNil(t, done)
}
