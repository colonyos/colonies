package dht

import (
	"github.com/colonyos/colonies/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

type Kademlia struct {
	messenger  p2p.Messenger
	Contact    Contact
	states     *states
	dispatcher *dispatcher
}

func CreateKademlia(messenger p2p.Messenger, contact Contact) (*Kademlia, error) {
	states := createStates(contact)
	k := &Kademlia{messenger: messenger, Contact: contact, states: states}
	dispatcher, err := createDispatcher(k)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create dispatcher")
		return nil, err
	}
	k.dispatcher = dispatcher

	go dispatcher.serveForever()
	go states.serveForever()

	return k, nil
}

func (k *Kademlia) GetContact() Contact {
	return k.Contact
}

func (k *Kademlia) Shutdown() {
	k.states.shutdown()
	k.dispatcher.shutdown()
}
