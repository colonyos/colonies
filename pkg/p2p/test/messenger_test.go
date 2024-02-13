package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/colonyos/colonies/pkg/p2p/dht"
	"github.com/colonyos/colonies/pkg/p2p/libp2p"
	"github.com/stretchr/testify/assert"
)

func TestDHTTest(t *testing.T) {
	// Create a new DHT node
	dht1, err := dht.CreateDHT(4001)
	assert.Nil(t, err)

	// Create a second DHT node
	dht2, err := dht.CreateDHT(4002)
	assert.Nil(t, err)

	// Register the second DHT node with the first
	c := dht1.GetContact()
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = dht2.Register(c.Node, c.ID.String(), ctx)
	cancel()
	assert.Nil(t, err)

	// Setup two messenger nodes
	messenger1, err := libp2p.CreateMessenger(5001)
	assert.Nil(t, err)
	messenger1Node := messenger1.Node
	messenger1NodeJSON, err := messenger1Node.ToJSON()

	messenger2, err := libp2p.CreateMessenger(5002)
	assert.Nil(t, err)
	messenger2Node := messenger2.Node
	messenger2NodeJSON, err := messenger2Node.ToJSON()

	msgChan := make(chan p2p.Message)
	go func() {
		ctx := context.TODO()
		messenger2.ListenForever(msgChan, ctx)
	}()

	// Register the messenger nodes with the DHT
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = dht1.Put("/m1", messenger1NodeJSON, 2, ctx)
	assert.Nil(t, err)
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = dht1.Put("/m2", messenger2NodeJSON, 2, ctx)
	assert.Nil(t, err)
	cancel()

	// Now messenger1 want to message messenger2. First, messenger1 needs to first find messenger2 contact information
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	kv, err := dht1.Get("/m2", 2, ctx)
	cancel()
	assert.Len(t, kv, 1)
	to, err := p2p.ConvertJSONToNode(kv[0].Value)
	from := messenger1.Node

	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = messenger1.Send(p2p.Message{From: from, To: to, Payload: []byte("Hello!")}, ctx)
	cancel()
	assert.Equal(t, err, nil)

	msg := <-msgChan
	fmt.Println(string(msg.Payload))
}
