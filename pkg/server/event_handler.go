package server

import (
	"context"
	"errors"
	"strconv"
)

type eventHandler struct {
	listeners map[string]map[string]chan struct{}
	msgQueue  chan *message
	idCounter int
}

type message struct {
	stop    bool
	handler func(msg *message)
	reply   chan replyMessage
}

type replyMessage struct {
	c            chan struct{}
	listenerID   string
	allListeners int // Just for testing purposes
	listeners    int // Just for testing purposes
}

func createEventHandler() *eventHandler {
	eventHandler := &eventHandler{}
	eventHandler.listeners = make(map[string]map[string]chan struct{})
	eventHandler.msgQueue = make(chan *message)

	go eventHandler.masterWorker()

	return eventHandler
}

func (handler *eventHandler) masterWorker() {
	for {
		select {
		case msg := <-handler.msgQueue:
			if msg.stop {
				return
			}
			if msg.handler != nil {
				msg.handler(msg)
			}
		}
	}
}

func (handler *eventHandler) target(runtimeType string, state int) string {
	return runtimeType + strconv.Itoa(state)
}

func (handler *eventHandler) register(runtimeType string, state int) (string, chan struct{}) {
	t := handler.target(runtimeType, state)
	if _, ok := handler.listeners[t]; !ok {
		handler.listeners[t] = make(map[string]chan struct{})
	}

	c := make(chan struct{})
	listenerID := strconv.Itoa(handler.idCounter)
	handler.listeners[t][listenerID] = c
	handler.idCounter++
	return listenerID, c
}

func (handler *eventHandler) unregister(runtimeType string, state int, listenerID string) {
	t := handler.target(runtimeType, state)
	if _, ok := handler.listeners[t]; ok {
		delete(handler.listeners[t], listenerID)
	}

	if len(handler.listeners[t]) == 0 {
		delete(handler.listeners, t)
	}
}

func (handler *eventHandler) signal(runtimeType string, state int) {
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		t := handler.target(runtimeType, state)
		if _, ok := handler.listeners[t]; ok {
			for _, c := range handler.listeners[t] {
				c <- struct{}{} // Wake up listeners
			}
		}
	}}
	handler.msgQueue <- msg // Send the message to the masterworker
}

func (handler *eventHandler) wait(runtimeType string, state int, ctx context.Context) error {
	// Register
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		listenerID, c := handler.register(runtimeType, state)
		r := replyMessage{c: c, listenerID: listenerID}
		msg.reply <- r
	}}
	handler.msgQueue <- msg

	// Wait for the masterworker to execute the handler code
	r := <-msg.reply

	// Unregister
	defer func() {
		msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
			handler.unregister(runtimeType, state, r.listenerID)
		}}
		handler.msgQueue <- msg
	}()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout")
		case <-r.c:
			return nil
		}
	}
}

func (handler *eventHandler) numberOfListeners(runtimeType string, state int) (int, int) { // Just for testing purposes
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		allListeners := len(handler.listeners)
		listeners := len(handler.listeners[handler.target(runtimeType, state)])
		r := replyMessage{allListeners: allListeners, listeners: listeners}
		msg.reply <- r
	}}

	handler.msgQueue <- msg
	r := <-msg.reply

	return r.allListeners, r.listeners
}
