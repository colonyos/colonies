package libp2p

import (
	"context"
	"testing"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/stretchr/testify/assert"
)

func TestNetworkP2P(t *testing.T) {
	serverAddr := "/ip4/0.0.0.0/tcp/4001"
	server, err := CreateServer(serverAddr)
	assert.Nil(t, err)

	go func() {
		for {
			err := server.Serve(serverAddr)
			assert.Nil(t, err)
		}
	}()

	client, err := CreateClient()
	assert.Nil(t, err)

	socket, err := client.Dial(server.ID(), context.TODO())
	assert.Nil(t, err)
	socket.Send(p2.pMessage{To: "10.0.0.1:8080", From: "10.0.0.2:8080", Payload: []byte("Hello World 1!")})

	socket, err = client.Dial(server.ID(), context.TODO())
	assert.Nil(t, err)
	socket.Send(p2p.Message{To: "10.0.0.1:8080", From: "10.0.0.2:8080", Payload: []byte("Hello World 2!")})

	select {}
}
