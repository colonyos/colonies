package dht

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

type Kademlia struct {
	n            network.Network
	contact      *Contact
	routingTable *RoutingTable
	socket       network.Socket
}

func CreateKademlia(n network.Network, contact *Contact) (*Kademlia, error) {
	routingTable := CreateRoutingTable(contact)
	socket, err := n.Listen(contact.Addr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to listen")
		return nil, err
	}

	return &Kademlia{n: n, contact: contact, routingTable: routingTable, socket: socket}, nil
}

func (k *Kademlia) ReceiveMsg() (*network.Message, error) {
	msg, err := k.socket.Receive()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to receive message")
		return nil, err
	}
	if msg == nil {
		log.WithFields(log.Fields{"Error": "nil message"}).Error("Received nil message")
		return nil, err
	}

	return msg, nil
}

func (k *Kademlia) handlePingRequest(msg *network.Message) error {
	log.WithFields(log.Fields{"FromAddr": msg.FromAddr}).Info("Received ping request")
	socket, err := k.n.Dial(msg.FromAddr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "FromAddr": msg.FromAddr}).Error("Failed to dial")
		return err
	}
	if socket == nil {
		log.WithFields(log.Fields{"Error": "nil socket", "FromAddr": msg.FromAddr}).Error("Failed to dial")
		return errors.New("Failed to dial")
	}

	log.WithFields(log.Fields{"Contact": k.contact.Addr}).Info("Sending ping response")
	payload := PingResponse{Header: RPCHeader{Sender: k.contact}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}
	return socket.Send(&network.Message{Type: network.MSG_PING_RESP, FromAddr: k.contact.Addr, ToAddr: msg.FromAddr, Payload: []byte(json)})
}

func (k *Kademlia) handlePingResponse(msg *network.Message) error {
	fmt.Println("Received ping response 2")
	response, err := ConvertJSONToPingResponse(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to ping response")
		return err
	}

	log.WithFields(log.Fields{"Contact": response.Header.Sender.Addr, "KademliaID": response.Header.Sender.ID}).Info("Adding contact to routing table")
	k.routingTable.AddContact(response.Header.Sender)
	return nil
}

func (k *Kademlia) ServerForEver() error {
	for {
		msg, err := k.ReceiveMsg()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to receive message")
			continue
		}

		switch msg.Type {
		case network.MSG_PING_REQ:
			k.handlePingRequest(msg)
		case network.MSG_PING_RESP:
			k.handlePingResponse(msg)
		}
	}
}

func (kademlia *Kademlia) Ping(addr string) error {
	socket, err := kademlia.n.Dial(addr)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to dial")
		return err

	}
	if socket == nil {
		log.WithFields(log.Fields{"Error": "nil socket"}).Error("Failed to dial")
		return errors.New("Failed to dial")
	}

	payload := PingRequest{Header: RPCHeader{Sender: kademlia.contact}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}
	return socket.Send(&network.Message{Type: network.MSG_PING_REQ, FromAddr: kademlia.contact.Addr, ToAddr: addr, Payload: []byte(json)})
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
