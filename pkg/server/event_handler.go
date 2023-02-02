package server

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type eventHandler struct {
	listeners         map[string]map[string]chan *core.Process
	processIDs        map[string]string
	msgQueue          chan *message
	idCounter         int
	stopped           bool
	mutex             sync.Mutex
	relayServer       *cluster.RelayServer
	relayChan         chan []byte
	stopRelayListener chan struct{}
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

func createEventHandler(relayServer *cluster.RelayServer) *eventHandler {
	handler := &eventHandler{}
	handler.listeners = make(map[string]map[string]chan *core.Process)
	handler.processIDs = make(map[string]string)
	handler.msgQueue = make(chan *message)
	handler.relayServer = relayServer

	handler.mutex.Lock()
	handler.stopped = true
	handler.mutex.Unlock()

	// Start master worker
	go handler.masterWorker()

	if relayServer != nil {
		handler.stopRelayListener = make(chan struct{})
		handler.relayChan = relayServer.Receive()
		go handler.relayListener()
	}

	return handler
}

func (handler *eventHandler) relayListener() {
	for {
		select {
		case msg := <-handler.relayChan:
			process, err := core.ConvertJSONToProcess(string(msg))
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Warning("relayListener received invalid process JSON")
			} else {
				handler.signalNoRelay(process)
			}
		case <-handler.stopRelayListener:
			return
		}
	}

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

func (handler *eventHandler) target(executorType string, state int) string {
	return executorType + strconv.Itoa(state)
}

func (handler *eventHandler) register(executorType string, state int, processID string) (string, chan *core.Process) {
	t := handler.target(executorType, state)
	if _, ok := handler.listeners[t]; !ok {
		handler.listeners[t] = make(map[string]chan *core.Process)
	}

	c := make(chan *core.Process, 100)
	listenerID := strconv.Itoa(handler.idCounter)
	handler.listeners[t][listenerID] = c
	if processID != "" {
		handler.processIDs[listenerID] = processID
	}
	handler.idCounter++
	return listenerID, c
}

func (handler *eventHandler) unregister(executorType string, state int, listenerID string) {
	t := handler.target(executorType, state)
	if _, ok := handler.listeners[t]; ok {
		delete(handler.listeners[t], listenerID)
		delete(handler.processIDs, listenerID)
	}

	if len(handler.listeners[t]) == 0 {
		delete(handler.listeners, t)
	}
}

func (handler *eventHandler) sendSignal(process *core.Process) {
	msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
		t := handler.target(process.ProcessSpec.Conditions.ExecutorType, process.State)
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

func (handler *eventHandler) signalNoRelay(process *core.Process) {
	handler.sendSignal(process)
}

func (handler *eventHandler) signal(process *core.Process) {
	handler.sendSignal(process)

	// broadcast the msg to the relayServer
	go func() {
		if handler.relayServer != nil {
			jsonStr, err := process.ToJSON()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to parse JSON in signal")
			}
			handler.relayServer.Broadcast([]byte(jsonStr))
		}
	}()
}

func (handler *eventHandler) waitForProcess(executorType string, state int, processID string, ctx context.Context) (*core.Process, error) {
	// Register
	msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
		listenerID, c := handler.register(executorType, state, processID)
		r := replyMessage{processChan: c, listenerID: listenerID}
		msg.reply <- r
	}}
	handler.msgQueue <- msg

	// Wait for the masterworker to execute the handler code
	r := <-msg.reply

	// Unregister
	defer func() {
		msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
			handler.unregister(executorType, state, r.listenerID)
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

func (handler *eventHandler) subscribe(executorType string, state int, processID string, ctx context.Context) (chan *core.Process, chan error) {
	// Register
	msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
		listenerID, c := handler.register(executorType, state, processID)
		r := replyMessage{processChan: c, listenerID: listenerID}
		msg.reply <- r
	}}
	handler.msgQueue <- msg

	// Wait for the masterworker to execute the handler code
	r := <-msg.reply

	processChan := make(chan *core.Process, 100)
	errChan := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Unregister
				msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
					handler.unregister(executorType, state, r.listenerID)
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
	if handler.relayServer != nil {
		handler.stopRelayListener <- struct{}{}
		handler.relayServer.Shutdown()
	}
}

func (handler *eventHandler) numberOfListeners(executorType string, state int) (int, int, int) { // Just for testing purposes
	msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
		allListeners := len(handler.listeners)
		processIDs := len(handler.processIDs)
		listeners := len(handler.listeners[handler.target(executorType, state)])
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
