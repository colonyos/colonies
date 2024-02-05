package dht

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/dht/network"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func createKademliaNode(t *testing.T, n network.Network, addr string, bootstrapAddr string) *Kademlia {
	crypto := crypto.CreateCrypto()

	prvKey1, err := crypto.GeneratePrivateKey()
	if err != nil {
		fmt.Println(err)
	}

	id1, err := crypto.GenerateID(prvKey1)
	if err != nil {
		fmt.Println(err)
	}

	contact1 := CreateContact(CreateKademliaID(id1), "localhost:8001")
	k, err := CreateKademlia(n, contact1)

	go k.ServerForEver()

	assert.Nil(t, err)

	err = k.Ping(bootstrapAddr)
	assert.Nil(t, err)

	return k
}

func TestNewKademlia(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8001", "localhost:8001")

	var nodes []*Kademlia
	for i := 0; i < 3; i++ {
		nodes = append(nodes, createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr))
	}

	select {}
}
