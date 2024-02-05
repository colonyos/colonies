package network

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockNetwork(t *testing.T) {
	n := CreateFakeNetwork()

	resultChan := make(chan *Message)

	socket, err := n.Listen("10.0.0.1:8080")
	assert.NotNil(t, socket)
	assert.Nil(t, err)
	go func() {
		message, _ := socket.Receive()
		resultChan <- message
	}()

	socket2, err := n.Dial("10.0.0.1:8080")
	assert.Nil(t, err)
	assert.NotNil(t, socket2)
	go func() {
		socket.Send(&Message{ToAddr: "10.0.0.1:8080", FromAddr: "10.0.0.2:8080", Payload: []byte("test_payload")})
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timeout")
	case msg := <-resultChan:
		assert.Equal(t, "test_payload", string(msg.Payload))
		assert.Equal(t, "10.0.0.1:8080", msg.ToAddr)
		assert.Equal(t, "10.0.0.2:8080", msg.FromAddr)
	}
}
