package dht

import (
	"context"
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

	return k
}

func TestKademliaFindRemoteContacts(t *testing.T) {
	n := network.CreateFakeNetwork()

	b := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	k := createKademliaNode(t, n, "localhost:8001", b.contact.Addr)
	k.Ping(b.contact.Addr)

	targetID := k.contact.ID
	contacts, err := k.FindRemoteContacts(b.contact.Addr, targetID.String(), 100, context.TODO()) // Ask bootstrap node for contact info to the k node
	assert.Nil(t, err)

	foundContact := false
	for _, contact := range contacts {
		if contact.ID.String() == targetID.String() {
			foundContact = true
		}
	}

	assert.True(t, foundContact)

	b.Shutdown()
	k.Shutdown()
}

func TestKademliaFindClosestContacts(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	bootstrapNode.Ping(bootstrapNode.contact.Addr)

	fmt.Println("bootstrapNode: ", bootstrapNode.contact.Addr)
	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Ping(bootstrapNode.contact.Addr)
		nodes = append(nodes, k)
	}

	fmt.Println("--------------------------------------")
	//targetID := NewRandomKademliaID()

	targetID := nodes[10].contact.ID
	contacts, err := nodes[0].FindClosestContacts(targetID.String())
	assert.Nil(t, err)

	for _, contact := range contacts {
		fmt.Println(contact.Addr + "->" + contact.ID.String())
	}
	fmt.Println("targetID: ", targetID.String())
	fmt.Println("--------------------------------------")

	targetID = nodes[3].contact.ID
	contacts, err = nodes[10].FindClosestContacts(targetID.String())
	assert.Nil(t, err)

	for _, contact := range contacts {
		fmt.Println(contact.Addr + "->" + contact.ID.String())
	}
	fmt.Println("targetID: ", targetID.String())
}
