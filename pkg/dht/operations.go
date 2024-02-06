package dht

import (
	"time"

	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

func (k *Kademlia) Ping(addr string) error {
	log.WithFields(log.Fields{"To": addr, "From": k.contact.Addr}).Info("Sending ping request")
	payload := PingReq{Header: RPCHeader{Sender: k.contact}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	reply, err := k.dispatcher.send(&network.Message{Type: network.MSG_PING_REQ, From: k.contact.Addr, To: addr, Payload: []byte(json)})
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return err
	}

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Address": addr}).Warn("Ping timeout")
		// TODO: handle timeout
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From}).Info("Ping response received")
		rpc, err := ConvertJSONToPingResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PingResp")
			return err
		}

		contact := rpc.Header.Sender
		k.rtw.addContact(contact)
	}

	return nil
}

func (k *Kademlia) FindContacts(addr string, kademliaID string) ([]Contact, error) {
	log.WithFields(log.Fields{"To": addr, "From": k.contact.Addr}).Info("Sending find contacts request")
	payload := FindContactsReq{Header: RPCHeader{Sender: k.contact}, KademliaID: kademliaID}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return nil, err
	}

	reply, err := k.dispatcher.send(&network.Message{Type: network.MSG_FIND_CONTACTS_REQ, From: k.contact.Addr, To: addr, Payload: []byte(json)})
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return nil, err
	}

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Address": addr}).Warn("Find contacts timeout")
		// TODO: handle timeout
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From}).Info("Find contacts response received")
		resp, err := ConvertJSONToFindContactsResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindContactResp")
			return nil, err
		}

		return resp.Contacts, nil
	}

	return make([]Contact, 0), nil
}
