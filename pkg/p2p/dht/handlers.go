package dht

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

const MAX_COUNT = 100

func (k *Kademlia) addContact(contact *Contact) error {
	errChan := k.states.addContact(*contact)

	select {
	case <-time.After(1 * time.Second):
		log.WithFields(log.Fields{"Node": k.Contact.Node.String()}).Error("Failed to add contact")
		return errors.New("Failed to add contact")
	case e := <-errChan:
		return e
	}
}

func (k *Kademlia) handlePingReq(msg p2p.Message) error {
	log.WithFields(log.Fields{"Node": k.Contact.Node.String()}).Info("Handling ping request")

	req, err := ConvertJSONToPingReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PingReq")
		return err
	}

	err = k.addContact(&req.Header.Sender)

	errMsg := ""
	status := PING_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to add contact")
		errMsg = err.Error()
		status = PING_STATUS_ERROR
	}

	payload := PingResp{Header: RPCHeader{Sender: k.Contact}, Status: status, Error: errMsg}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	return k.dispatcher.sendReply(msg, p2p.Message{
		Type:    MSG_PING_RESP,
		From:    k.Contact.Node,
		To:      msg.From,
		Payload: []byte(json)})
}

func (k *Kademlia) handleFindContactsReq(msg p2p.Message) error {
	log.WithFields(log.Fields{"Node": k.Contact.Node.String()}).Info("Handling find contacts request")

	req, err := ConvertJSONToFindContactsReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindNodeReq")
		return err
	}

	kademliaID := req.KademliaID

	var contacts []Contact
	err = k.addContact(&req.Header.Sender)
	if err == nil {
		count := req.Count
		if count > MAX_COUNT {
			count = MAX_COUNT
		}

		contactsChan, errChan := k.states.findContacts(kademliaID, count)

		select {
		case <-time.After(1 * time.Second):
			log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "ID": k.Contact.ID.String(), "TargetID": kademliaID}).Error("Failed to send closest contacts (timeout)")
			return errors.New("Failed to find closest contacts")
		case e := <-errChan:
			err = e
		case c := <-contactsChan:
			contacts = c
		}
	}

	errMsg := ""
	status := FIND_CONTACTS_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find contacts")
		errMsg = err.Error()
		status = FIND_CONTACTS_STATUS_ERROR
	}

	payload := FindContactsResp{Header: RPCHeader{Sender: k.Contact}, Contacts: contacts, Status: status, Error: errMsg}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "ID": k.Contact.ID.String(), "TargetID": kademliaID}).Info("Sending closest contacts")
	return k.dispatcher.sendReply(msg, p2p.Message{
		Type:    MSG_FIND_CONTACTS_RESP,
		From:    k.Contact.Node,
		To:      msg.From,
		Payload: []byte(json)})
}

func (k *Kademlia) handlePutReq(msg p2p.Message) error {
	log.WithFields(log.Fields{"Node": k.Contact.Node.String()}).Info("Handling put request")

	req, err := ConvertJSONToPutReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to GetReq")
		return err
	}

	isValueValid, err := ValidateValue(req.KV)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to validate value")
		return err
	}

	if !isValueValid {
		log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "Key": req.KV.Key, "Value": req.KV.Value}).Error("Failed to validate value")
		return errors.New("Failed to validate value")
	}

	err = k.addContact(&req.Header.Sender)
	if err == nil {
		errChan := k.states.put(req.KV.ID, req.KV.Key, req.KV.Value, req.KV.Sig)

		select {
		case <-time.After(1 * time.Second):
			log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "Key": req.KV.Key, "Value": req.KV.Value}).Error("Failed to put key-value pair")
			return errors.New("Failed to put key-value pair")
		case e := <-errChan:
			err = e
		}
	}

	errMsg := ""
	status := PUT_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to put key-value")
		errMsg = err.Error()
		status = PUT_STATUS_ERROR
	}

	payload := PutResp{Header: RPCHeader{Sender: k.Contact}, Status: status, Error: errMsg}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "Key": req.KV.Key, "Value": req.KV.Value}).Info("Sending put response")

	return k.dispatcher.sendReply(msg, p2p.Message{
		Type:    MSG_PUT_RESP,
		From:    k.Contact.Node,
		To:      msg.From,
		Payload: []byte(json)})
}

func (k *Kademlia) handleGetReq(msg p2p.Message) error {
	log.WithFields(log.Fields{"Node": k.Contact.Node.String()}).Info("Handling get request")

	req, err := ConvertJSONToGetReq(string(msg.Payload))
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to GetReq")
		return err
	}

	var kvs []KV
	err = k.addContact(&req.Header.Sender)
	if err == nil {
		kvsChan, errChan := k.states.get(req.Key)

		select {
		case <-time.After(1 * time.Second):
			log.WithFields(log.Fields{"Node": k.Contact.Node.String(), "Key": req.Key}).Error("Failed to get key-value pair")
			return errors.New("Failed to get key-value pair")
		case k := <-kvsChan:
			kvs = k
		case e := <-errChan:
			err = e
		}
	}

	errMsg := ""
	status := GET_STATUS_SUCCESS
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to get key-value")
		errMsg = err.Error()
		status = GET_STATUS_ERROR
	}

	payload := GetResp{Header: RPCHeader{Sender: k.Contact}, Status: status, Error: errMsg, KVS: kvs}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	return k.dispatcher.sendReply(msg, p2p.Message{
		Type:    MSG_GET_RESP,
		From:    k.Contact.Node,
		To:      msg.From,
		Payload: []byte(json)})
}
