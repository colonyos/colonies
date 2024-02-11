package p2p

import (
	"context"
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/colonyos/colonies/pkg/p2p/dht"
	"github.com/colonyos/colonies/pkg/p2p/libp2p"
	"github.com/stretchr/testify/assert"
)

func TestDHTTest(t *testing.T) {
	// Create a new DHT node
	dht1, err := dht.CreateDHT([]string{"/ip4/10.0.0.201/tcp/4001", "/ip4/127.0.0.1/tcp/4001"})
	assert.Nil(t, err)

	// Create a second DHT node
	dht2, err := dht.CreateDHT([]string{"/ip4/10.0.0.201/tcp/4002", "/ip4/127.0.0.1/tcp/4002"})
	assert.Nil(t, err)

	// Register the second DHT node with the first
	c := dht1.GetContact()
	err = dht2.Register(c.Node, c.ID.String(), context.TODO())
	assert.Nil(t, err)

	// Setup two messenger nodes
	messenger1, err := libp2p.CreateMessenger([]string{"/ip4/10.0.0.201/tcp/5001", "/ip4/127.0.0.1/tcp/5001"})
	assert.Nil(t, err)
	messenger1Node := messenger1.Node
	messenger1NodeJSON, err := messenger1Node.ToJSON()

	messenger2, err := libp2p.CreateMessenger([]string{"/ip4/10.0.0.201/tcp/5002", "/ip4/127.0.0.1/tcp/5002"})
	assert.Nil(t, err)
	messenger2Node := messenger2.Node
	messenger2NodeJSON, err := messenger2Node.ToJSON()

	msgChan := make(chan p2p.Message)
	ctx := context.TODO()
	go func() {
		messenger2.ListenForever(msgChan, ctx)
	}()

	// Register the messenger nodes with the DHT
	err = dht1.Put("/m1", messenger1NodeJSON, 2, context.TODO())
	assert.Nil(t, err)

	err = dht1.Put("/m2", messenger2NodeJSON, 2, context.TODO())
	assert.Nil(t, err)

	// Now messenger1 want to message messenger2. First, messenger1 needs to first find messenger2 contact information
	kv, err := dht1.Get("/m2", context.TODO())
	assert.Len(t, kv, 1)
	to, err := p2p.ConvertJSONToNode(kv[0].Value)
	from := messenger1.Node

	err = messenger1.Send(p2p.Message{From: from, To: *to, Payload: []byte("Hello!")}, context.TODO()) // TODO: fix reference, see the strange *
	assert.Equal(t, err, nil)

	msg := <-msgChan
	fmt.Println(string(msg.Payload))
}
