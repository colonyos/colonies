package gin

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

// DefaultEventHandler implements the backends.RealtimeEventHandler interface.
//
// Signal Distribution Strategy (Thundering Herd Prevention):
//
// This event handler uses a two-pass signal distribution to prevent the "thundering herd"
// problem that can cause massive database load amplification.
//
// The Problem:
// When a new process is submitted, executors waiting for work need to be notified.
// If ALL waiting executors wake up simultaneously (broadcast), they all call Assign()
// at once, creating a stampede of database operations. With 8 executors and buffered
// channels allowing rapid re-wake, this caused 1,840x amplification in production
// (19.7M database updates for 1,341 processes instead of ~10K expected).
//
// The Solution - Two-Pass Distribution:
//
// Pass 1 (Broadcast): Notify ALL listeners waiting for a SPECIFIC processID.
// This is used for state change notifications (e.g., process completed) where
// the caller needs to know about their specific process.
//
// Pass 2 (Single Wake-up with Round-Robin): Wake exactly ONE general listener.
// General listeners are executors waiting for ANY process of their type.
// Only one executor should wake up per signal to attempt assignment.
// Round-robin ensures fair distribution across executors.
//
// Round-Robin Load Balancing:
// The nextExecutor map tracks which executor should receive the next signal
// for each target (executorType+state combination). This ensures all executors
// get equal opportunity to receive work, preventing any single executor from
// being starved or overloaded.
type DefaultEventHandler struct {
	listeners         map[string]map[string]chan *core.Process // target -> listenerID -> channel
	processIDs        map[string]string                        // listenerID -> processID (for specific process listeners)
	nextExecutor      map[string]int                           // target -> round-robin index for fair distribution
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

// CreateEventHandler creates a new DefaultEventHandler
func CreateEventHandler(relayServer *cluster.RelayServer) backends.RealtimeEventHandler {
	handler := &DefaultEventHandler{}
	handler.listeners = make(map[string]map[string]chan *core.Process)
	handler.processIDs = make(map[string]string)
	handler.nextExecutor = make(map[string]int)
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

// CreateTestableEventHandler creates a new DefaultEventHandler that implements TestableRealtimeEventHandler for testing
func CreateTestableEventHandler(relayServer *cluster.RelayServer) backends.TestableRealtimeEventHandler {
	handler := &DefaultEventHandler{}
	handler.listeners = make(map[string]map[string]chan *core.Process)
	handler.processIDs = make(map[string]string)
	handler.nextExecutor = make(map[string]int)
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

func (handler *DefaultEventHandler) relayListener() {
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

func (handler *DefaultEventHandler) masterWorker() {
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

func (handler *DefaultEventHandler) target(executorType string, state int) string {
	return executorType + strconv.Itoa(state)
}

func (handler *DefaultEventHandler) register(executorType string, state int, processID string) (string, chan *core.Process) {
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

func (handler *DefaultEventHandler) unregister(executorType string, state int, listenerID string) {
	t := handler.target(executorType, state)
	if _, ok := handler.listeners[t]; ok {
		delete(handler.listeners[t], listenerID)
		delete(handler.processIDs, listenerID)
	}

	if len(handler.listeners[t]) == 0 {
		delete(handler.listeners, t)
	}
}

// sendSignal distributes a process event to registered listeners using two-pass distribution.
//
// This is the core of the thundering herd prevention mechanism. See DefaultEventHandler
// documentation for the full explanation of the problem and solution.
//
// Algorithm:
//  1. Find all listeners for this process's executorType and state
//  2. Pass 1: Broadcast to all listeners waiting for this specific processID
//  3. Pass 2: Wake exactly ONE general listener using round-robin selection
//
// The round-robin selection ensures fair distribution:
//   - Listener IDs are sorted for deterministic ordering
//   - nextExecutor[target] tracks which listener is next in rotation
//   - If the selected listener's channel is full, try the next one
//   - After successful send, advance the index for next time
func (handler *DefaultEventHandler) sendSignal(process *core.Process) {
	msg := &message{reply: make(chan replyMessage, 100), handler: func(msg *message) {
		t := handler.target(process.FunctionSpec.Conditions.ExecutorType, process.State)
		if _, ok := handler.listeners[t]; ok {
			// Pass 1: Broadcast to all listeners waiting for this specific processID.
			// These are callers waiting for state changes on a process they submitted/own.
			for listenerID, c := range handler.listeners[t] {
				if processID, ok := handler.processIDs[listenerID]; ok {
					if process.ID == processID {
						select {
						case c <- process.Clone():
						default:
							// Channel full, skip
						}
					}
				}
			}

			// Pass 2: Wake ONE general listener using round-robin.
			// General listeners are executors waiting for ANY process of their type.
			// Collect general listener IDs (those without specific processID)
			var generalListeners []string
			for listenerID := range handler.listeners[t] {
				if _, ok := handler.processIDs[listenerID]; !ok {
					generalListeners = append(generalListeners, listenerID)
				}
			}

			if len(generalListeners) == 0 {
				return
			}

			// Sort for deterministic round-robin order (Go maps iterate randomly)
			sort.Strings(generalListeners)

			// Get current round-robin index and ensure it's valid
			idx := handler.nextExecutor[t] % len(generalListeners)

			// Try each listener starting from idx, wrapping around if channel is full
			for i := 0; i < len(generalListeners); i++ {
				listenerID := generalListeners[(idx+i)%len(generalListeners)]
				c := handler.listeners[t][listenerID]
				select {
				case c <- process.Clone():
					// Success - advance round-robin index for next signal
					handler.nextExecutor[t] = (idx + i + 1) % len(generalListeners)
					return
				default:
					continue // Channel full, try next listener
				}
			}
		}
	}}

	handler.msgQueue <- msg // Send the message to the masterworker
}

func (handler *DefaultEventHandler) signalNoRelay(process *core.Process) {
	handler.sendSignal(process)
}

// Signal implements EventHandler interface
func (handler *DefaultEventHandler) Signal(process *core.Process) {
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

// WaitForProcess implements EventHandler interface
func (handler *DefaultEventHandler) WaitForProcess(executorType string, state int, processID string, ctx context.Context) (*core.Process, error) {
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

// Subscribe implements EventHandler interface
func (handler *DefaultEventHandler) Subscribe(executorType string, state int, processID string, ctx context.Context) (chan *core.Process, chan error) {
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

// Stop implements EventHandler interface
func (handler *DefaultEventHandler) Stop() {
	handler.msgQueue <- &message{stop: true}
	if handler.relayServer != nil {
		handler.stopRelayListener <- struct{}{}
		handler.relayServer.Shutdown()
	}
}

// Additional methods for testing (not part of interface)
func (handler *DefaultEventHandler) NumberOfListeners(executorType string, state int) (int, int, int) {
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

func (handler *DefaultEventHandler) HasStopped() bool {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()
	return handler.stopped
}