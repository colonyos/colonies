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
	k.ping(b.contact.Addr, context.TODO())

	targetID := k.contact.ID

	// Ask bootstrap node for contact info to the k node
	contacts, err := k.findRemoteContacts(b.contact.Addr, targetID.String(), 100, context.TODO())
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
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Register(bootstrapNode.contact.Addr, k.contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].contact.ID
	contacts, err := nodes[4].FindContacts(targetID.String(), 10, context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(contacts) > 0)
	assert.Equal(t, contacts[0].ID.String(), targetID.String())

	targetID = nodes[3].contact.ID
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
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Register(bootstrapNode.contact.Addr, k.contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetID := nodes[10].contact.ID
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
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Register(bootstrapNode.contact.Addr, k.contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	targetNode := nodes[10]
	err := nodes[5].putRemote(targetNode.contact.Addr, "/prefix/key1", "test1", "", context.TODO())
	assert.Nil(t, err)

	kvs, err := nodes[6].getRemote(targetNode.contact.Addr, "/prefix/key1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 1)
	assert.Equal(t, kvs[0].Value, "test1")

	err = nodes[5].putRemote(targetNode.contact.Addr, "/prefix/key2", "test2", "", context.TODO())
	assert.Nil(t, err)

	kvs, err = nodes[6].getRemote(targetNode.contact.Addr, "/prefix", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 2)

	_, err = nodes[6].getRemote(targetNode.contact.Addr, "/prefix/not_found", context.TODO())
	assert.NotNil(t, err)
}

func TestKademliaPutGet(t *testing.T) {
	n := network.CreateFakeNetwork()

	bootstrapNode := createKademliaNode(t, n, "localhost:8000", "localhost:8000")

	var nodes []*Kademlia
	for i := 1; i < 50; i++ {
		k := createKademliaNode(t, n, "localhost:800"+fmt.Sprint(i), bootstrapNode.contact.Addr)
		k.Register(bootstrapNode.contact.Addr, k.contact.ID.String(), context.TODO())
		nodes = append(nodes, k)
	}

	err := nodes[5].Put("/prefix/key1", "test1", "", 5, context.TODO())
	assert.Nil(t, err)

	kvs, err := nodes[28].Get("/prefix/key1", context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(kvs), 1)
	assert.Equal(t, kvs[0].Value, "test1")

	err = nodes[12].Put("/prefix/key2", "test2", "", 5, context.TODO())
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
