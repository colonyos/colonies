package dht

import (
	"context"
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
	replyHandler map[string]chan network.Message
	mutex        sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
}

func createDispatcher(n network.Network, addr string, k *Kademlia) (*dispatcher, error) {
	socket, err := n.Listen(addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &dispatcher{socket: socket,
		n:            n,
		k:            k,
		replyHandler: make(map[string]chan network.Message),
		ctx:          ctx,
		cancel:       cancel}, nil
}

func (dispatcher *dispatcher) handleResponse(msg *network.Message) {
	dispatcher.mutex.Lock()
	replyChan, ok := dispatcher.replyHandler[msg.ID]
	dispatcher.mutex.Unlock()
	if ok {
		replyChan <- *msg

		dispatcher.mutex.Lock()
		delete(dispatcher.replyHandler, msg.ID)
		dispatcher.mutex.Unlock()
	} else {
		log.WithFields(log.Fields{"Error": "No handler for message", "MsgID": msg.ID}).Error("Dropping message")
	}
}

func (dispatcher *dispatcher) serveForever() {
	for {
		msg, err := dispatcher.socket.Receive(dispatcher.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Info("Context canceled, killing dispatcher")
				return
			}
			log.WithFields(log.Fields{"Error": err}).Error("Failed to receive message")
			continue
		}

		switch msg.Type {
		case network.MSG_PING_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received PING_REQ")
			dispatcher.k.handlePingReq(msg)
		case network.MSG_PING_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received PING_RESP")
			dispatcher.handleResponse(&msg)
		case network.MSG_FIND_CONTACTS_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_CONTACTS_REQ")
			err := dispatcher.k.handleFindContactsReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle FIND_CONTACTS_REQ")
			}
		case network.MSG_FIND_CONTACTS_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_CONTACTS_RESP")
			dispatcher.handleResponse(&msg)
		case network.MSG_PUT_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_PUT_REQ")
			err := dispatcher.k.handlePutReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle PUT_REQ")
			}
		case network.MSG_PUT_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_PUT_RESP")
			dispatcher.handleResponse(&msg)
		case network.MSG_GET_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_GET_REQ")
			err := dispatcher.k.handleGetReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle GET_REQ")
			}
		case network.MSG_GET_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.contact.Addr, "From": msg.From}).Info("Received FIND_GET_RESP")
			dispatcher.handleResponse(&msg)
		default:
			log.WithFields(log.Fields{"Error": "Unknown message type", "Type": msg.Type}).Error("Dropping message")
		}
	}
}

func (dispatcher *dispatcher) send(msg network.Message) (chan network.Message, error) {
	msg.ID = core.GenerateRandomID()

	log.WithFields(log.Fields{"msgID": msg.ID, "From": msg.From, "To": msg.To, "Type": msg.Type}).Info("Sending message")

	replyChan := make(chan network.Message)

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

func (dispatcher *dispatcher) shutdown() {
	dispatcher.cancel()
}

func (dispatcher *dispatcher) sendReply(msg network.Message, replyMsg network.Message) error {
	replyMsg.ID = msg.ID
	socket, err := dispatcher.n.Dial(replyMsg.To)
	if err != nil {
		return err
	}

	err = socket.Send(replyMsg)
	return err
}
