package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/colonyos/colonies/pkg/p2p/dht"
	"github.com/colonyos/colonies/pkg/p2p/libp2p"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestDHTTest(t *testing.T) {
	// Create a new DHT node
	dht1, err := dht.CreateDHT(4001, "dht1")
	assert.Nil(t, err)

	// Create a second DHT node
	dht2, err := dht.CreateDHT(4002, "dht2")
	assert.Nil(t, err)

	fmt.Println(dht1.GetContact().Node.Addr)

	// Register the second DHT node with the first
	c := dht1.GetContact()
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	//err = dht2.RegisterNetwork(c.Node, c.ID.String(), ctx)
	err = dht2.RegisterNetworkWithAddr(c.Node.Addr, c.ID.String(), ctx)
	cancel()
	assert.Nil(t, err)

	// Create a ECDSA key pair to be able to publish a key-value pair to the DHT
	crypto := crypto.CreateCrypto()
	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	id, err := crypto.GenerateID(prvKey)
	assert.Nil(t, err)

	// Setup two messenger nodes
	messenger1, err := libp2p.CreateMessenger(5001, "mes1")
	assert.Nil(t, err)

	messenger2, err := libp2p.CreateMessenger(5002, "mes2")
	assert.Nil(t, err)

	msgChan := make(chan p2p.Message)
	go func() {
		ctx := context.TODO()
		messenger2.ListenForever(msgChan, ctx)
	}()

	// Register the messenger nodes with the DHT
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = dht1.RegisterNode(id, prvKey, &messenger1.Node, ctx)
	assert.Nil(t, err)
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	err = dht1.RegisterNode(id, prvKey, &messenger2.Node, ctx)
	assert.Nil(t, err)
	cancel()

	// Now messenger1 want to message messenger2. First, messenger1 needs to first find messenger2 contact information
	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	messenger2Node, err := dht2.LookupNode(id, "mes2", ctx)
	cancel()
	to := messenger2Node
	from := messenger1.Node

	ctx, cancel = context.WithTimeout(context.Background(), 1000*time.Millisecond)
	msgToSend := p2p.Message{From: from, To: *to, Payload: []byte("Hello!")}
	err = messenger1.Send(msgToSend, ctx)
	cancel()
	assert.Equal(t, err, nil)

	msg := <-msgChan
	fmt.Println(string(msg.Payload))
	assert.Equal(t, string(msg.Payload), "Hello!")
}
