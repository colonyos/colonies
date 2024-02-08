package dht

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

const COUNT = 10

func (k *Kademlia) handlePingReq(msg network.Message) error {
	log.WithFields(log.Fields{"Me": k.contact.Addr}).Info("Sending ping response")

	req, err := ConvertJSONToPingReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PingReq")
		return err
	}

	k.rtw.addContact(req.Header.Sender)

	payload := PingResp{Header: RPCHeader{Sender: k.contact}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	return k.dispatcher.sendReply(msg, network.Message{
		Type:    network.MSG_PING_RESP,
		From:    k.contact.Addr,
		To:      msg.From,
		Payload: []byte(json)})
}

func (k *Kademlia) handleFindContactsReq(msg network.Message) error {
	req, err := ConvertJSONToFindContactsReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindNodeReq")
		return err
	}

	kademliaID := req.KademliaID

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Me": k.contact.Addr, "MyID": k.contact.ID, "TargetID": kademliaID}).Error("Failed to send closest contacts")
		return errors.New("Failed to find closest contacts")
	case contacts := <-k.rtw.findContacts(kademliaID, COUNT):
		payload := FindContactsResp{Header: RPCHeader{Sender: k.contact}, Contacts: contacts}
		json, err := payload.ToJSON()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
			return err
		}

		log.WithFields(log.Fields{"Me": k.contact.Addr, "MyID": k.contact.ID.String(), "TargetID": kademliaID}).Info("Sending closest contacts")
		return k.dispatcher.sendReply(msg, network.Message{
			Type:    network.MSG_FIND_CONTACTS_RESP,
			From:    k.contact.Addr,
			To:      msg.From,
			Payload: []byte(json)})
	}
}
