package dht

import (
	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

type Kademlia struct {
	n          network.Network
	contact    Contact
	states     *states
	dispatcher *dispatcher
	socket     network.Socket
}

func CreateKademlia(n network.Network, contact Contact) (*Kademlia, error) {
	socket, err := n.Listen(contact.Addr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to listen at address: " + contact.Addr)
		return nil, err
	}

	states := createStates(contact)
	k := &Kademlia{n: n, contact: contact, states: states, socket: socket}
	dispatcher, err := createDispatcher(n, contact.Addr, k)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create dispatcher")
		return nil, err
	}
	k.dispatcher = dispatcher

	go dispatcher.serveForever()
	go states.serveForever()

	return k, nil
}

func (k *Kademlia) Shutdown() {
	k.states.shutdown()
	k.dispatcher.shutdown()
}
