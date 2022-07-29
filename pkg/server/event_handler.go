package server

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/colonyos/colonies/pkg/core"
)

type eventHandler struct {
	listeners  map[string]map[string]chan *core.Process
	processIDs map[string]string
	msgQueue   chan *message
	idCounter  int
	stopped    bool
	mutex      sync.Mutex
}

type message struct {
	stop    bool // Just for testing purposes
	handler func(msg *message)
	reply   chan replyMessage
}

type replyMessage struct {
	processChan  chan *core.Process
	listenerID   string
	allListeners int  // Just for testing purposes
	listeners    int  // Just for testing purposes
	processIDs   int  // Just for testing purposes
	stopped      bool // Just for testing purposes
}

func createEventHandler() *eventHandler {
	handler := &eventHandler{}
	handler.listeners = make(map[string]map[string]chan *core.Process)
	handler.processIDs = make(map[string]string)
	handler.msgQueue = make(chan *message)

	handler.mutex.Lock()
	handler.stopped = true
	handler.mutex.Unlock()

	// Start master worker
	go handler.masterWorker()

	return handler
}

func (handler *eventHandler) masterWorker() {
	handler.mutex.Lock()
	handler.stopped = false
	handler.mutex.Unlock()

	for {
		select {
		case msg := <-handler.msgQueue:
			if msg.stop {
				handler.mutex.Lock()
				handler.stopped = true
				handler.mutex.Unlock()
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

func (handler *eventHandler) register(runtimeType string, state int, processID string) (string, chan *core.Process) {
	t := handler.target(runtimeType, state)
	if _, ok := handler.listeners[t]; !ok {
		handler.listeners[t] = make(map[string]chan *core.Process)
	}

	c := make(chan *core.Process)
	listenerID := strconv.Itoa(handler.idCounter)
	handler.listeners[t][listenerID] = c
	if processID != "" {
		handler.processIDs[listenerID] = processID
	}
	handler.idCounter++
	return listenerID, c
}

func (handler *eventHandler) unregister(runtimeType string, state int, listenerID string) {
	t := handler.target(runtimeType, state)
	if _, ok := handler.listeners[t]; ok {
		delete(handler.listeners[t], listenerID)
		delete(handler.processIDs, listenerID)
	}

	if len(handler.listeners[t]) == 0 {
		delete(handler.listeners, t)
	}
}

func (handler *eventHandler) signal(process *core.Process) {
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		t := handler.target(process.ProcessSpec.Conditions.RuntimeType, process.State)
		if _, ok := handler.listeners[t]; ok {
			for listenerID, c := range handler.listeners[t] {
				if processID, ok := handler.processIDs[listenerID]; ok {
					if process.ID == processID {
						c <- process.Clone() // Send a copy of the process to all listeners interested in this particular processID
					}
				} else {
					c <- process.Clone() // Send a copy of the process to all listeners
				}
			}
		}
	}}
	handler.msgQueue <- msg // Send the message to the masterworker
}

func (handler *eventHandler) waitForProcess(runtimeType string, state int, processID string, ctx context.Context) (*core.Process, error) {
	// Register
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		listenerID, c := handler.register(runtimeType, state, processID)
		r := replyMessage{processChan: c, listenerID: listenerID}
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
			return nil, errors.New("timeout")
		case process := <-r.processChan:
			return process, nil
		}
	}
}

func (handler *eventHandler) subscribe(runtimeType string, state int, processID string, ctx context.Context) (chan *core.Process, chan error) {
	// Register
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		listenerID, c := handler.register(runtimeType, state, processID)
		r := replyMessage{processChan: c, listenerID: listenerID}
		msg.reply <- r
	}}
	handler.msgQueue <- msg

	// Wait for the masterworker to execute the handler code
	r := <-msg.reply

	processChan := make(chan *core.Process)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Unregister
				msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
					handler.unregister(runtimeType, state, r.listenerID)
				}}
				handler.msgQueue <- msg
				errChan <- errors.New("timeout")
				return
			case process := <-r.processChan:
				processChan <- process
			}
		}
	}()

	return processChan, errChan
}

func (handler *eventHandler) stop() {
	handler.msgQueue <- &message{stop: true}
}

func (handler *eventHandler) numberOfListeners(runtimeType string, state int) (int, int, int) { // Just for testing purposes
	msg := &message{reply: make(chan replyMessage, 1), handler: func(msg *message) {
		allListeners := len(handler.listeners)
		processIDs := len(handler.processIDs)
		listeners := len(handler.listeners[handler.target(runtimeType, state)])
		r := replyMessage{allListeners: allListeners, listeners: listeners, processIDs: processIDs}
		msg.reply <- r
	}}

	handler.msgQueue <- msg
	r := <-msg.reply

	return r.allListeners, r.listeners, r.processIDs
}

func (handler *eventHandler) hasStopped() bool { // Just for testing purposes
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	return handler.stopped
}
