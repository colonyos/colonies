package dht

import (
	"errors"
	"sync"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/dht/network"
	log "github.com/sirupsen/logrus"
)

type dispatcher struct {
	socket       network.Socket
	n            network.Network
	k            *Kademlia
	replyHandler map[string]chan *network.Message
	mutex        sync.Mutex
}

func createDispatcher(n network.Network, addr string, k *Kademlia) (*dispatcher, error) {
	socket, err := n.Listen(addr)
	if err != nil {
		return nil, err
	}

	return &dispatcher{socket: socket, n: n, k: k, replyHandler: make(map[string]chan *network.Message)}, nil
}

func (dispatcher *dispatcher) serveForever() {
	for {
		msg, err := dispatcher.socket.Receive()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to receive message")
			continue
		}
		if msg == nil {
			log.WithFields(log.Fields{"Error": "nil message"}).Error("Received nil message")
			continue
		}
		switch msg.Type {
		case network.MSG_PING_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received PING_REQ")
			dispatcher.k.handlePingReq(msg)
		case network.MSG_PING_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received PING_RESP")
			replyChan, ok := dispatcher.replyHandler[msg.ID]
			if ok {
				replyChan <- msg

				dispatcher.mutex.Lock()
				delete(dispatcher.replyHandler, msg.ID)
				dispatcher.mutex.Unlock()
			} else {
				log.WithFields(log.Fields{"Error": "No handler for message", "MsgID": msg.ID}).Error("Dropping message")
			}
		case network.MSG_FIND_CONTACTS_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_CONTACTS_REQ")
			dispatcher.k.handleFindContactsReq(msg)
		case network.MSG_FIND_CONTACTS_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_CONTACTS_RESP")
			replyChan, ok := dispatcher.replyHandler[msg.ID]
			if ok {
				replyChan <- msg

				dispatcher.mutex.Lock()
				delete(dispatcher.replyHandler, msg.ID)
				dispatcher.mutex.Unlock()
			} else {
				log.WithFields(log.Fields{"Error": "No handler for message", "MsgID": msg.ID}).Error("Dropping message")
			}
		default:
			log.WithFields(log.Fields{"Error": "Unknown message type", "Type": msg.Type}).Error("Dropping message")
		}
	}
}

func (dispatcher *dispatcher) send(msg *network.Message) (chan *network.Message, error) {
	if msg == nil {
		log.WithFields(log.Fields{"Error": "nil message"}).Error("Cannot send nil message")
		return nil, errors.New("Cannot send nil message")
	}

	msg.ID = core.GenerateRandomID()

	log.WithFields(log.Fields{"msgID": msg.ID, "From": msg.From, "To": msg.To, "Type": msg.Type}).Info("Sending message")

	replyChan := make(chan *network.Message)

	dispatcher.mutex.Lock()
	dispatcher.replyHandler[msg.ID] = replyChan
	dispatcher.mutex.Unlock()

	socket, err := dispatcher.n.Dial(msg.To)
	if err != nil {
		return nil, err
	}
	err = socket.Send(msg)
	return replyChan, err
}

func (dispatcher *dispatcher) sendReply(msg *network.Message, replyMsg *network.Message) error {
	replyMsg.ID = msg.ID
	socket, err := dispatcher.n.Dial(replyMsg.To)
	if err != nil {
		return err
	}
	err = socket.Send(replyMsg)
	return err
}
