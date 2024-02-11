package dht

import (
	"context"
	"fmt"
	"testing"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/dht/network"
	"github.com/stretchr/testify/assert"
)

func createKademliaNode(t *testing.T, n network.Network, addr string, bootstrapAddr string) *Kademlia {
	id, err := crypto.CreateIdendity()
	assert.Nil(t, err)

	prvKey := id.PrivateKeyAsHex()

	contact1, err := CreateContact(id.ID(), prvKey)
	assert.Nil(t, err)

	k, err := CreateKademlia(n, contact1)
	assert.Nil(t, err)

	return k
}

func TestKademliaFindRemoteContacts(t *testing.T) {
	n := network.CreateFakeNetwork()

	b := createKademliaNode(t, n, "localhost:8000", "localhost:8000")
	k := createKademliaNode(t, n, "localhost:8001", b.Contact.Addr)
	k.ping(b.Contact.Addr, context.TODO())

	targetID := k.Contact.ID

	// Ask bootstrap node for contact info to the k node
	contacts, err := k.findRemoteContacts(b.Contact.Addr, targetID.String(), 100, context.TODO())
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

func TestKademliaFindContacts(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.Contact.Addr)
		k.Register(bootstrapNode.Contact.Addr, k.Contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].Contact.ID
	contacts, err := nodes[4].FindContacts(targetID.String(), 10, context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(contacts) > 0)
	assert.Equal(t, contacts[0].ID.String(), targetID.String())

	targetID = nodes[3].Contact.ID
	contacts, err = nodes[10].FindContacts(targetID.String(), 10, context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(contacts) > 0)
	assert.Equal(t, contacts[0].ID.String(), targetID.String())
}

func TestKademliaFindContact(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.Contact.Addr)
		k.Register(bootstrapNode.Contact.Addr, k.Contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].Contact.ID
	contact, err := nodes[0].FindContact(targetID.String(), context.TODO())
	assert.Nil(t, err)
	assert.NotNil(t, contact)
	assert.Equal(t, contact.ID.String(), targetID.String())
}

func TestKademliaPutGetRemote(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 20; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.Contact.Addr)
		k.Register(bootstrapNode.Contact.Addr, k.Contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetNode := nodes[10]
	err := nodes[5].putRemote(targetNode.Contact.Addr, "/prefix/key1", "test1", context.TODO())
	assert.Nil(t, err)

	kvs, err := nodes[6].getRemote(targetNode.Contact.Addr, "/prefix/key1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 1)
	assert.Equal(t, kvs[0].Value, "test1")

	err = nodes[5].putRemote(targetNode.Contact.Addr, "/prefix/key2", "test2", context.TODO())
	assert.Nil(t, err)

	kvs, err = nodes[6].getRemote(targetNode.Contact.Addr, "/prefix", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 2)

	_, err = nodes[6].getRemote(targetNode.Contact.Addr, "/prefix/not_found", context.TODO())
	assert.NotNil(t, err)
}

func TestKademliaPutGet(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 50; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.Contact.Addr)
		k.Register(bootstrapNode.Contact.Addr, k.Contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	err := nodes[5].Put("/prefix/key1", "test1", 5, context.TODO())
	assert.Nil(t, err)

	kvs, err := nodes[28].Get("/prefix/key1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 1)
	assert.Equal(t, kvs[0].Value, "test1")

	err = nodes[12].Put("/prefix/key2", "test2", 5, context.TODO())
	assert.Nil(t, err)

	kvs, err = nodes[40].Get("/prefix", context.TODO())
	assert.Nil(t, err)

	count := 0
	for _, kv := range kvs {
		if kv.Value == "test1" || kv.Value == "test2" {
			count++
		}
	}
	assert.Equal(t, count, 2)
}
