package dht

import (
	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

type Kademlia struct {
	n          network.Network
	contact    Contact
	rtw        *routingTableWorker
	dispatcher *dispatcher
	socket     network.Socket
}

func CreateKademlia(n network.Network, contact Contact) (*Kademlia, error) {
	socket, err := n.Listen(contact.Addr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to listen at address: " + contact.Addr)
		return nil, err
	}

	rtw := createRoutingTableWorker(contact)
	k := &Kademlia{n: n, contact: contact, rtw: rtw, socket: socket}
	dispatcher, err := createDispatcher(n, contact.Addr, k)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create dispatcher")
		return nil, err
	}
	k.dispatcher = dispatcher

	go dispatcher.serveForever()
	go rtw.serveForever()

	return k, nil
}

func (k *Kademlia) Shutdown() {
	k.rtw.shutdown()
	k.dispatcher.shutdown()
}
