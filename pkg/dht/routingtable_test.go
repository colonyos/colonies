package dht

import (
	"fmt"
	"testing"
)

func TestRoutingTable(t *testing.T) {
	rt := createRoutingTable(CreateContact(CreateKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000"), "localhost:8000"))

	rt.addContact(CreateContact(CreateKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000"), "localhost:8001"))
	rt.addContact(CreateContact(CreateKademliaID("1111111100000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
	rt.addContact(CreateContact(CreateKademliaID("1111111200000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
	rt.addContact(CreateContact(CreateKademliaID("1111111300000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
	rt.addContact(CreateContact(CreateKademliaID("1111111400000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
	rt.addContact(CreateContact(CreateKademliaID("2111111400000000000000000000000000000000000000000000000000000000"), "localhost:8002"))

	contacts := rt.findClosestContacts(CreateKademliaID("2111111400000000000000000000000000000000000000000000000000000000"), 20)
	for i := range contacts {
		fmt.Println(contacts[i].Addr, "->", contacts[i].ID.String())
	}
}
