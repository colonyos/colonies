package dht

import (
	"context"
	"sync"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/p2p"
	log "github.com/sirupsen/logrus"
)

type dispatcher struct {
	k            *Kademlia
	replyHandler map[string]chan p2p.Message
	mutex        sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	msgChan      chan p2p.Message
}

func createDispatcher(k *Kademlia) (*dispatcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	msgChan := make(chan p2p.Message, 1000)
	go k.messenger.ListenForever(msgChan, ctx)

	return &dispatcher{
		k:            k,
		replyHandler: make(map[string]chan p2p.Message),
		ctx:          ctx,
		cancel:       cancel,
		msgChan:      msgChan}, nil
}

func (dispatcher *dispatcher) handleResponse(msg *p2p.Message) {
	dispatcher.mutex.Lock()
	replyChan, ok := dispatcher.replyHandler[msg.ID]
	dispatcher.mutex.Unlock()
	if ok {
		replyChan <- *msg

		dispatcher.mutex.Lock()
		close(replyChan)
		delete(dispatcher.replyHandler, msg.ID)
		dispatcher.mutex.Unlock()
	} else {
		log.WithFields(log.Fields{"Error": "No handler for message", "MsgID": msg.ID}).Error("Dropping message")
	}
}

func (dispatcher *dispatcher) serveForever() {
	for {
		msg := <-dispatcher.msgChan

		switch msg.Type {
		case MSG_PING_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received PING_REQ")
			dispatcher.k.handlePingReq(msg)
		case MSG_PING_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received PING_RESP")
			dispatcher.handleResponse(&msg)
		case MSG_FIND_CONTACTS_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_CONTACTS_REQ")
			err := dispatcher.k.handleFindContactsReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle FIND_CONTACTS_REQ")
			}
		case MSG_FIND_CONTACTS_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_CONTACTS_RESP")
			dispatcher.handleResponse(&msg)
		case MSG_PUT_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_PUT_REQ")
			err := dispatcher.k.handlePutReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle PUT_REQ")
			}
		case MSG_PUT_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_PUT_RESP")
			dispatcher.handleResponse(&msg)
		case MSG_GET_REQ:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_GET_REQ")
			err := dispatcher.k.handleGetReq(msg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to handle GET_REQ")
			}
		case MSG_GET_RESP:
			log.WithFields(log.Fields{"MsgID": msg.ID, "MyAddr": dispatcher.k.Contact.Node.String(), "From": msg.From.String()}).Info("Received FIND_GET_RESP")
			dispatcher.handleResponse(&msg)
		default:
			log.WithFields(log.Fields{"Error": "Unknown message type", "Type": msg.Type}).Error("Dropping message")
		}
	}
}

func (dispatcher *dispatcher) send(msg p2p.Message) (chan p2p.Message, error) {
	msg.ID = core.GenerateRandomID()

	log.WithFields(log.Fields{"msgID": msg.ID, "From": msg.From.String(), "To": msg.To.String(), "Type": msg.Type}).Info("Sending message")

	replyChan := make(chan p2p.Message)

	dispatcher.mutex.Lock()
	dispatcher.replyHandler[msg.ID] = replyChan
	dispatcher.mutex.Unlock()

	err := dispatcher.k.messenger.Send(msg, dispatcher.ctx)
	return replyChan, err
}

func (dispatcher *dispatcher) shutdown() {
	dispatcher.cancel()
}

func (dispatcher *dispatcher) sendReply(msg p2p.Message, replyMsg p2p.Message) error {
	replyMsg.ID = msg.ID
	err := dispatcher.k.messenger.Send(replyMsg, dispatcher.ctx)
	return err
}
