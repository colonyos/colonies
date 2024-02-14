package dht

import (
	"context"
	"errors"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
)

const MaxPendingRequests = 10000
const RegistrationReplicationFactor = 10

func (k *Kademlia) ping(node p2p.Node, ctx context.Context) error {
	log.WithFields(log.Fields{"To": node.String(), "From": k.Contact.Node.String()}).Info("Sending ping request")
	payload := PingReq{Header: RPCHeader{Sender: k.Contact}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	reply, err := k.dispatcher.send(p2p.Message{
		Type:    MSG_PING_REQ,
		From:    k.Contact.Node,
		To:      node,
		Payload: []byte(json)})

	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return err
	}

	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{"Node": node.String()}).Warn("Ping request timeout")
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From}).Info("Ping response received")
		rpc, err := ConvertJSONToPingResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PingResp")
			return err
		}

		contact := rpc.Header.Sender
		k.states.addContact(contact)
	}

	return nil
}

func (k *Kademlia) findRemoteContacts(node p2p.Node, kademliaID string, count int, ctx context.Context) ([]Contact, error) {
	log.WithFields(log.Fields{"To": node, "From": k.Contact.Node.String()}).Info("Sending find contacts request")
	payload := FindContactsReq{Header: RPCHeader{Sender: k.Contact}, KademliaID: kademliaID, Count: count}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return nil, err
	}

	reply, err := k.dispatcher.send(p2p.Message{
		Type:    MSG_FIND_CONTACTS_REQ,
		From:    k.Contact.Node,
		To:      node,
		Payload: []byte(json)})

	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return nil, err
	}

	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{"Address": node.String}).Warn("Find contacts timeout")
		return []Contact{}, errors.New("Find remote contacts timeout")
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From}).Info("Find contacts response received")
		resp, err := ConvertJSONToFindContactsResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindContactResp")
			return nil, err
		}

		return resp.Contacts, nil
	}
}

func (k *Kademlia) findLocalContacts(kademliaID string, count int, ctx context.Context) ([]Contact, error) {
	contactsChan, errChan := k.states.findContacts(kademliaID, count)

	select {
	case <-ctx.Done():
		log.Error("Find local closest contacts timeout")
		return []Contact{}, errors.New("Find local closest contacts timeout")
	case err := <-errChan:
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find local closest contacts")
		return []Contact{}, err
	case contacts := <-contactsChan:
		return contacts, nil
	}
}

func (k *Kademlia) FindContact(kademliaID string, ctx context.Context) (Contact, error) {
	contacts, err := k.FindContacts(kademliaID, 1, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find closest contacts")
		return Contact{}, err
	}

	foundContacts := make(map[string]Contact)

	for _, contact := range contacts {
		foundContacts[contact.ID.String()] = contact
	}
	if len(foundContacts) == 0 {
		log.Info("No contacts found")
		return Contact{}, errors.New("No contacts found")
	}

	if c, ok := foundContacts[kademliaID]; ok {
		return c, nil
	}

	return Contact{}, errors.New("No contacts found")
}

func (k *Kademlia) RegisterNetwork(bootstrapNode p2p.Node, kademliaID string, ctx context.Context) error {
	err := k.ping(bootstrapNode, ctx)
	if err != nil {
		return err
	}

	// Lookup our self in the network, this will populate remote nodes routing tables with our contact
	nodesToRegister := 20
	_, err = k.FindContacts(kademliaID, nodesToRegister, ctx)

	return err
}

func (k *Kademlia) RegisterNetworkWithAddr(bootstrapNodeAddr string, kademliaID string, ctx context.Context) error {
	bootstrapNode := p2p.CreateNode("boostrapnode", "", bootstrapNodeAddr)
	err := k.ping(bootstrapNode, ctx)
	if err != nil {
		return err
	}

	// Lookup our self in the network, this will populate remote nodes routing tables with our contact
	nodesToRegister := 20
	_, err = k.FindContacts(kademliaID, nodesToRegister, ctx)

	return err
}

func (k *Kademlia) FindContacts(kademliaID string, count int, ctx context.Context) ([]Contact, error) {
	foundContacts := make(map[string]Contact)
	pendingContactChan := make(chan Contact, 10)

	outgoingReq := make(chan struct{}, MaxPendingRequests)

	contacts, err := k.findLocalContacts(kademliaID, count, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find local contacts")
		return []Contact{}, err
	}

	if len(contacts) == 0 {
		log.Info("No local contacts found")
		return []Contact{}, nil
	}

	for _, contact := range contacts {
		pendingContactChan <- contact
	}

	var contact Contact
	for {
		if len(pendingContactChan) > 0 {
			contact = <-pendingContactChan
			if contact.ID == CreateKademliaID(kademliaID) && count == 1 {
				return []Contact{contact}, nil
			}
		} else if len(outgoingReq) == 0 && len(pendingContactChan) == 0 {
			var candidates ContactCandidates
			for _, contact := range foundContacts {
				contact.CalcDistance(CreateKademliaID(kademliaID))
				candidates.contacts = append(candidates.contacts, contact)
			}
			candidates.Sort()

			if count > candidates.Len() {
				count = candidates.Len()
			}

			return candidates.GetContacts(count), nil
		}

		if _, ok := foundContacts[contact.Node.String()]; !ok {
			foundContacts[contact.Node.String()] = contact
			outgoingReq <- struct{}{}
			go func(contact2 Contact, kademliaID2 string, count2 int, ctx2 context.Context) {
				defer func() { <-outgoingReq }()
				contacts, err := k.findRemoteContacts(contact2.Node, kademliaID2, count2, ctx2)
				if err != nil {
					log.WithFields(log.Fields{"Error": err, "Node": contact2.Node.String()}).Error("Failed to find remote contacts")
					return
				}
				for _, contact := range contacts {
					pendingContactChan <- contact
				}
			}(contact, kademliaID, count, ctx)
		}
	}
}

func (k *Kademlia) putRemote(node p2p.Node, id string, prvKey string, key string, value string, ctx context.Context) error {
	log.WithFields(log.Fields{"To": node, "From": k.Contact.Node.String()}).Info("Sending put request")

	crypto := crypto.CreateCrypto()
	hash := crypto.GenerateHash(value)
	sig, err := crypto.GenerateSignature(hash, prvKey)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to generate signature")
		return err
	}

	payload := PutReq{Header: RPCHeader{Sender: k.Contact}, KV: KV{ID: id, Key: "/" + id + key, Value: value, Sig: sig}}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return err
	}

	reply, err := k.dispatcher.send(p2p.Message{
		Type:    MSG_PUT_REQ,
		From:    k.Contact.Node,
		To:      node,
		Payload: []byte(json)})

	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return err
	}

	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{"Node": node.String()}).Warn("Put request timeout")
		return errors.New("Put request timeout")
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From}).Info("Put response received")
		rpc, err := ConvertJSONToPutResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to PutResp")
			return err
		}

		if rpc.Status != PUT_STATUS_SUCCESS {
			return errors.New(rpc.Error)
		}

		return nil
	}
}

func (k *Kademlia) getRemote(node p2p.Node, id string, key string, ctx context.Context) ([]KV, error) {
	log.WithFields(log.Fields{"To": node, "From": k.Contact.Node.String()}).Info("Sending get request")
	payload := GetReq{Header: RPCHeader{Sender: k.Contact}, Key: "/" + id + key}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return nil, err
	}

	reply, err := k.dispatcher.send(p2p.Message{
		Type:    MSG_GET_REQ,
		From:    k.Contact.Node,
		To:      node,
		Payload: []byte(json)})

	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send get request")
		return nil, err
	}

	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{"Node": node.String()}).Warn("Get request timeout")
		return nil, errors.New("Put request timeout")
	case msg := <-reply:
		log.WithFields(log.Fields{"From": msg.From.String()}).Info("Get response received")
		rpc, err := ConvertJSONToGetResp(string(msg.Payload))
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to GetResp")
			return nil, err
		}

		if rpc.Status != GET_STATUS_SUCCESS {
			return nil, errors.New(rpc.Error)
		}

		return rpc.KVS, nil
	}
}

func (k *Kademlia) Put(id string, prvKey string, key string, value string, replicationFactor int, ctx context.Context) error {
	crypto := crypto.CreateCrypto()
	hash := crypto.GenerateHash(id)

	contacts, err := k.FindContacts(hash, replicationFactor, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find closest contacts")
		return err
	}

	if len(contacts) == 0 {
		log.Info("No contacts found")
		return errors.New("No contacts found")
	}

	for _, contact := range contacts {
		err := k.putRemote(contact.Node, id, prvKey, key, value, ctx)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to put value")
			return err
		}
	}

	for _, contact := range contacts {
		log.WithFields(log.Fields{"Contact": contact.Node.String(), "Key": key}).Info("Contact has stored value")
	}

	return nil
}

func (k *Kademlia) Get(id string, key string, replicationFactor int, ctx context.Context) ([]KV, error) {
	crypto := crypto.CreateCrypto()
	hash := crypto.GenerateHash(id)

	contacts, err := k.FindContacts(hash, replicationFactor, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to find closest contacts")
		return nil, err
	}

	if len(contacts) == 0 {
		log.Info("No contacts found")
		return nil, errors.New("No contacts found")
	}

	kvsMap := make(map[string]KV)
	for _, contact := range contacts {
		kvs, err := k.getRemote(contact.Node, id, key, ctx)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to get value")
		}
		for _, kv := range kvs {
			kvsMap[kv.String()] = kv
		}
	}

	for _, contact := range contacts {
		log.WithFields(log.Fields{"Contact": contact.Node.String(), "Key": key}).Info("Contact should have value")
	}

	var result []KV
	for _, v := range kvsMap {
		result = append(result, v)
	}

	return result, nil
}

func (k *Kademlia) RegisterNode(id string, prvKey string, node *p2p.Node, ctx context.Context) error {
	nodeJSON, err := node.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert node to JSON")
		return err
	}
	err = k.Put(id, prvKey, "/nodes/"+node.Name, nodeJSON, RegistrationReplicationFactor, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to put node")
		return err
	}

	return nil
}

func (k *Kademlia) LookupNode(id string, name string, ctx context.Context) (*p2p.Node, error) {
	kvs, err := k.Get(id, "/nodes/"+name, RegistrationReplicationFactor, ctx)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to put node")
		return nil, err
	}

	if len(kvs) == 0 {
		return nil, errors.New("No nodes found")
	}

	if len(kvs) > 1 {
		return nil, errors.New("Multiple nodes found")
	}

	node, err := p2p.ConvertJSONToNode(kvs[0].Value)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to node")
		return nil, err
	}

	return &node, nil
}
