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

	contact1 := CreateContact(CreateKademliaID(id1), addr)
	k, err := CreateKademlia(n, contact1)

	assert.Nil(t, err)

	err = k.Ping(bootstrapAddr)
	assert.Nil(t, err)

	return k
}

func TestNewKademlia(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 3; i++ {
		nodes = append(nodes, createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr))
	}

	fmt.Println("-------------------------------------")
	targetID := nodes[1].contact.ID
	fmt.Println("targetID: ", targetID)
	//contacts, err := nodes[0].FindContacts(nodes[1].contact.Addr, targetID.String())
	contacts, err := nodes[0].FindContacts(bootstrapNode.contact.Addr, targetID.String())
	assert.Nil(t, err)

	for _, contact := range contacts {
		fmt.Println(contact.Addr)
	}

	select {}
}
