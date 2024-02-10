package dht

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

const MAX_COUNT = 100

func (k *Kademlia) handlePingReq(msg network.Message) error {
	log.WithFields(log.Fields{"Me": k.contact.Addr}).Info("Handling ping request")

	req, err := ConvertJSONToPingReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PingReq")
		return err
	}

	k.states.addContact(req.Header.Sender)

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
	log.WithFields(log.Fields{"Me": k.contact.Addr}).Info("Handling find contacts request")

	req, err := ConvertJSONToFindContactsReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindNodeReq")
		return err
	}

	kademliaID := req.KademliaID

	k.states.addContact(req.Header.Sender)

	count := req.Count
	if count > MAX_COUNT {
		count = MAX_COUNT
	}

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Me": k.contact.Addr, "MyID": k.contact.ID, "TargetID": kademliaID}).Error("Failed to send closest contacts")
		return errors.New("Failed to find closest contacts")
	case contacts := <-k.states.findContacts(kademliaID, count):
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

func (k *Kademlia) handlePutReq(msg network.Message) error {
	log.WithFields(log.Fields{"Me": k.contact.Addr}).Info("Handling put request")

	req, err := ConvertJSONToPutReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to GetReq")
		return err
	}

	k.states.addContact(req.Header.Sender)
	errChan := k.states.put(req.Key, req.Value)

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Me": k.contact.Addr, "Key": req.Key, "Value": req.Value}).Error("Failed to put key-value pair")
		return errors.New("Failed to put key-value pair")
	case e := <-errChan:
		err = e
	}

	close(errChan)

	errMsg := ""
	status := PUT_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to put key-value")
		errMsg = err.Error()
		status = PUT_STATUS_ERROR
	}

	payload := PutResp{Header: RPCHeader{Sender: k.contact}, Status: status, Error: errMsg}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	log.WithFields(log.Fields{"Me": k.contact.Addr, "Key": req.Key, "Value": req.Value}).Info("Sending put response")

	return k.dispatcher.sendReply(msg, network.Message{
		Type:    network.MSG_PUT_RESP,
		From:    k.contact.Addr,
		To:      msg.From,
		Payload: []byte(json)})
}

func (k *Kademlia) handleGetReq(msg network.Message) error {
	log.WithFields(log.Fields{"Me": k.contact.Addr}).Info("Handling get request")

	req, err := ConvertJSONToPutReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to GetReq")
		return err
	}

	k.states.addContact(req.Header.Sender)
	valuesChan, errChan := k.states.get(req.Key)

	var values []string

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Me": k.contact.Addr, "Key": req.Key, "Value": req.Value}).Error("Failed to get key-value pair")
		return errors.New("Failed to get key-value pair")
	case v := <-valuesChan:
		values = v
	case e := <-errChan:
		err = e
	}

	close(valuesChan)
	close(errChan)

	errMsg := ""
	status := GET_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to get key-value")
		errMsg = err.Error()
		status = GET_STATUS_ERROR
	}

	payload := GetResp{Header: RPCHeader{Sender: k.contact}, Status: status, Error: errMsg, Values: values}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	return k.dispatcher.sendReply(msg, network.Message{
		Type:    network.MSG_GET_RESP,
		From:    k.contact.Addr,
		To:      msg.From,
		Payload: []byte(json)})
}
