package dht

import (
	"context"
	"errors"
	"fmt"
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

	reply, err := k.dispatcher.send(network.Message{Type: network.MSG_PING_REQ, From: k.contact.Addr, To: addr, Payload: []byte(json)})
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

func (k *Kademlia) FindRemoteContacts(addr string, kademliaID string, count int, ctx context.Context) ([]Contact, error) {
	log.WithFields(log.Fields{"To": addr, "From": k.contact.Addr}).Info("Sending find contacts request")
	payload := FindContactsReq{Header: RPCHeader{Sender: k.contact}, KademliaID: kademliaID}
	json, err := payload.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
		return nil, err
	}

	reply, err := k.dispatcher.send(network.Message{Type: network.MSG_FIND_CONTACTS_REQ, From: k.contact.Addr, To: addr, Payload: []byte(json)})
	defer close(reply)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
		return nil, err
	}

	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{"Address": addr}).Warn("Find contacts timeout")
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

func (k *Kademlia) FindLocalContacts(kademliaID string, count int, ctx context.Context) ([]Contact, error) {
	contactsChan := k.rtw.findContacts(kademliaID, count)
	defer close(contactsChan)

	select {
	case <-ctx.Done():
		log.Error("Find local closest contacts timeout")
		return []Contact{}, errors.New("Find local closest contacts timeout")
	case contacts := <-contactsChan:
		return contacts, nil
	}
}

// func (k *Kademlia) findRemoteContacts(addr string,
// 	kademliaID string,
// 	pendingContactChan chan Contact,
// 	count int,
// 	context context.Context) (chan bool, error) {
//
// 	doneChan := make(chan bool)
//
// 	log.WithFields(log.Fields{"To": addr, "From": k.contact.Addr}).Info("Sending find contacts request")
// 	payload := FindContactsReq{Header: RPCHeader{Sender: k.contact}, KademliaID: kademliaID}
// 	json, err := payload.ToJSON()
// 	if err != nil {
// 		log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to JSON")
// 		return doneChan, err
// 	}
//
// 	reply, err := k.dispatcher.send(network.Message{Type: network.MSG_FIND_CONTACTS_REQ, From: k.contact.Addr, To: addr, Payload: []byte(json)})
// 	if err != nil {
// 		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping request")
// 		return doneChan, err
// 	}
//
// 	select {
// 	case <-context.Done():
// 		log.Error("Find local closest contacts timeout")
// 		return doneChan, err
// 	case msg := <-reply:
// 		log.WithFields(log.Fields{"From": msg.From}).Info("Find contacts response received")
// 		resp, err := ConvertJSONToFindContactsResp(string(msg.Payload))
// 		if err != nil {
// 			log.WithFields(log.Fields{"Error": err}).Error("Failed to convert to FindContactResp")
// 			return doneChan, err
// 		}
//
// 		close(reply)
//
// 		for _, contact := range resp.Contacts {
// 			pendingContactChan <- contact
// 		}
//
// 		doneChan <- true
// 	}
//
// 	return doneChan, nil
// }

func (k *Kademlia) FindClosestContacts(kademliaID string) ([]Contact, error) {
	foundContacts := make(map[string]Contact)
	pendingContactChan := make(chan Contact, 10)
	count := 10

	maxPendingRequests := 10000
	outgoingReq := make(chan struct{}, maxPendingRequests)

	ctx := context.TODO()

	contacts, err := k.FindLocalContacts(kademliaID, count, ctx)
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

	fmt.Println(len(pendingContactChan))

	var contact Contact
	for {
		if len(pendingContactChan) > 0 {
			contact = <-pendingContactChan
		} else if len(outgoingReq) == 0 && len(pendingContactChan) == 0 {
			var candidates ContactCandidates
			for _, contact := range foundContacts {
				contact.CalcDistance(CreateKademliaID(kademliaID))
				candidates.contacts = append(candidates.contacts, contact)
			}
			candidates.Sort()
			// 100 TODO
			return candidates.GetContacts(10), nil
		}

		if _, ok := foundContacts[contact.Addr]; !ok {
			foundContacts[contact.Addr] = contact
			outgoingReq <- struct{}{}
			go func() {
				defer func() { <-outgoingReq }()
				contacts, err := k.FindRemoteContacts(contact.Addr, kademliaID, count, ctx)
				if err != nil {
					// TODO
					return
				}
				fmt.Println(len(contacts))
				for _, contact := range contacts {
					fmt.Println("adding contact", contact.Addr)
					pendingContactChan <- contact
				}
			}()
		} else {
			log.WithFields(log.Fields{"Address": contact.Addr}).Info("Contact already found")
		}
	}
}
