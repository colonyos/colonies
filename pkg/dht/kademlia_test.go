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
	assert.Nil(t, err)

	id1, err := crypto.GenerateID(prvKey1)
	assert.Nil(t, err)

	contact1 := CreateContact(CreateKademliaID(id1), addr)
	k, err := CreateKademlia(n, contact1)
	assert.Nil(t, err)

	return k
}

func TestKademliaFindRemoteContacts(t *testing.T) {
	n := network.CreateFakeNetwork()

	b := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	k := createKademliaNode(t, n, "localhost:8001", b.contact.Addr)
	k.Ping(b.contact.Addr, context.TODO())

	targetID := k.contact.ID

	// Ask bootstrap node for contact info to the k node
	contacts, err := k.FindRemoteContacts(b.contact.Addr, targetID.String(), 100, context.TODO())
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
	bootstrapNode.Ping(bootstrapNode.contact.Addr, context.TODO())

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Ping(bootstrapNode.contact.Addr, context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].contact.ID
	contacts, err := nodes[4].FindClosestContacts(targetID.String(), 10, context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(contacts) > 0)
	assert.Equal(t, contacts[0].ID.String(), targetID.String())

	targetID = nodes[3].contact.ID
	contacts, err = nodes[10].FindClosestContacts(targetID.String(), 10, context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(contacts) > 0)
	assert.Equal(t, contacts[0].ID.String(), targetID.String())
}

func TestKademliaFindContact(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	bootstrapNode.Ping(bootstrapNode.contact.Addr, context.TODO())

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Ping(bootstrapNode.contact.Addr, context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].contact.ID
	contact, err := nodes[0].FindContact(targetID.String(), context.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, contact)
	assert.Equal(t, contact.ID.String(), targetID.String())
}

func TestKademliaPutKVRemote(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	bootstrapNode.Ping(bootstrapNode.contact.Addr, context.TODO())

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Ping(bootstrapNode.contact.Addr, context.TODO())
		nodes = append(nodes, k)
	}

	targetNode := nodes[10]
	err := nodes[5].PutKVRemote(targetNode.contact.Addr, "/prefix/key1", "test1", context.TODO())
	assert.Nil(t, err)

	values, err := nodes[6].GetKVRemote(targetNode.contact.Addr, "/prefix/key1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(values), 1)
	assert.Equal(t, values[0], "test1")

	err = nodes[5].PutKVRemote(targetNode.contact.Addr, "/prefix/key2", "test2", context.TODO())
	assert.Nil(t, err)

	values, err = nodes[6].GetKVRemote(targetNode.contact.Addr, "/prefix", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(values), 2)

	_, err = nodes[6].GetKVRemote(targetNode.contact.Addr, "/prefix/not_found", context.TODO())
	assert.NotNil(t, err)
}
