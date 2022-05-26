package server

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/planner"
	"github.com/colonyos/colonies/pkg/planner/basic"
	"github.com/colonyos/colonies/pkg/rpc"
	log "github.com/sirupsen/logrus"
)

const TIMEOUT_RELEASE_INTERVALL = 1

type subscribers struct {
	processesSubscribers map[string]*processesSubscription
	processSubscribers   map[string]*processSubscription
}

type command struct {
	stop                 bool
	errorChan            chan error
	process              *core.Process
	count                int
	colony               *core.Colony
	colonyID             string
	colonyReplyChan      chan *core.Colony
	coloniesReplyChan    chan []*core.Colony
	processReplyChan     chan *core.Process
	processesReplyChan   chan []*core.Process
	processStatReplyChan chan *core.ProcessStat
	runtimeReplyChan     chan *core.Runtime
	runtimesReplyChan    chan []*core.Runtime
	attributeReplyChan   chan *core.Attribute
	handler              func(cmd *command)
}

type coloniesController struct {
	db          database.Database
	cmdQueue    chan *command
	planner     planner.Planner
	subscribers *subscribers
	stopFlag    bool
	mutex       sync.Mutex
}

func createColoniesController(db database.Database) *coloniesController {
	controller := &coloniesController{db: db}
	controller.cmdQueue = make(chan *command)
	controller.subscribers = &subscribers{}
	controller.subscribers.processesSubscribers = make(map[string]*processesSubscription)
	controller.subscribers.processSubscribers = make(map[string]*processSubscription)
	controller.planner = basic.CreatePlanner()

	go controller.masterWorker()
	go controller.timeoutChecker()

	return controller
}

func (controller *coloniesController) timeoutChecker() {
	for {
		time.Sleep(TIMEOUT_RELEASE_INTERVALL * time.Second)

		controller.mutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.mutex.Unlock()

		processes, err := controller.db.FindAllRunningProcesses()
		if err != nil {
			continue
		}
		for _, process := range processes {
			if process.ProcessSpec.MaxExecTime == -1 {
				continue
			}
			if time.Now().Unix() > process.Deadline.Unix() {
				if process.Retries >= process.ProcessSpec.MaxRetries && process.ProcessSpec.MaxRetries > -1 {
					err := controller.closeFailed(process.ID)
					if err != nil {
						log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Info("Max retries reached, but failed to close process")
						continue
					}
					log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Info("Process closed as *failed* as max retries reached")
					continue
				}

				err := controller.unassignRuntime(process.ID)
				if err != nil {
					log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error("Failed to unassign process")
				}
				log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Process was unassigned as it did not complete in time")
			}
		}
	}
}

func (controller *coloniesController) masterWorker() {
	for {
		select {
		case msg := <-controller.cmdQueue:
			if msg.stop {
				return
			}
			if msg.handler != nil {
				msg.handler(msg)
			}
		}
	}
}

func (controller *coloniesController) numberOfProcessesSubscribers() int {
	return len(controller.subscribers.processesSubscribers)
}

func (controller *coloniesController) numberOfProcessSubscribers() int {
	return len(controller.subscribers.processSubscribers)
}

func (controller *coloniesController) subscribeProcesses(runtimeID string, subscription *processesSubscription) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			controller.subscribers.processesSubscribers[runtimeID] = subscription
			cmd.errorChan <- nil
		}}
	controller.cmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) subscribeProcess(runtimeID string, subscription *processSubscription) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			controller.subscribers.processSubscribers[runtimeID] = subscription

			// Send an event immediately if process already have the state the subscriber is looking for
			// See unittest TestSubscribeChangeStateProcess2 for more info
			process, err := controller.db.GetProcessByID(subscription.processID)
			if err != nil {
				cmd.errorChan <- err
			}
			if process.State == subscription.state {
				controller.wsWriteProcessChangeEvent(process, runtimeID, subscription)
			}
			cmd.errorChan <- nil
		}}
	controller.cmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) sendProcessesEvent(process *core.Process) {
	for runtimeID, subscription := range controller.subscribers.processesSubscribers {
		if subscription.runtimeType == process.ProcessSpec.Conditions.RuntimeType && subscription.state == process.State {
			jsonString, err := process.ToJSON()
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to parse JSON when removing processes event subscription")
				delete(controller.subscribers.processesSubscribers, runtimeID)
			}
			rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to create RPC reply message when removing processes event subscription")
				delete(controller.subscribers.processSubscribers, runtimeID)
			}

			rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to generate JSON when removing processes event subcription")
				delete(controller.subscribers.processSubscribers, runtimeID)
			}
			err = subscription.wsConn.WriteMessage(subscription.wsMsgType, []byte(rpcReplyJSONString))
			if err != nil {
				log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Removing processes event subcription")
				delete(controller.subscribers.processesSubscribers, runtimeID)
			}
		}
	}
}

func (controller *coloniesController) wsWriteProcessChangeEvent(process *core.Process, runtimeID string, subscription *processSubscription) {
	jsonString, err := process.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to parse JSON when removing process event subscription")
		delete(controller.subscribers.processSubscribers, runtimeID)
	}

	rpcReplyMsg, err := rpc.CreateRPCReplyMsg(rpc.SubscribeProcessPayloadType, jsonString)
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to create RPC reply message when removing process event subscription")
		delete(controller.subscribers.processSubscribers, runtimeID)
	}

	rpcReplyJSONString, err := rpcReplyMsg.ToJSON()
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Failed to generate JSON when removing process event subcription")
		delete(controller.subscribers.processSubscribers, runtimeID)
	}

	err = subscription.wsConn.WriteMessage(subscription.wsMsgType, []byte(rpcReplyJSONString))
	if err != nil {
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "Error": err}).Info("Removing process event subcription")
		delete(controller.subscribers.processSubscribers, runtimeID)
	}
}

func (controller *coloniesController) sendProcessChangeStateEvent(process *core.Process) {
	for runtimeID, subscription := range controller.subscribers.processSubscribers {
		if subscription.processID == process.ID && subscription.state == process.State {
			controller.wsWriteProcessChangeEvent(process, runtimeID, subscription)
		}
	}
}

func (controller *coloniesController) getColonies() ([]*core.Colony, error) {
	cmd := &command{coloniesReplyChan: make(chan []*core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies, err := controller.db.GetColonies()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.coloniesReplyChan <- colonies
		}}

	controller.cmdQueue <- cmd
	var colonies []*core.Colony
	select {
	case err := <-cmd.errorChan:
		return colonies, err
	case colonies := <-cmd.coloniesReplyChan:
		return colonies, nil
	}
}

func (controller *coloniesController) getColonyByID(colonyID string) (*core.Colony, error) {
	cmd := &command{colonyReplyChan: make(chan *core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colony, err := controller.db.GetColonyByID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.colonyReplyChan <- colony
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case colony := <-cmd.colonyReplyChan:
		return colony, nil
	}
}

func (controller *coloniesController) addColony(colony *core.Colony) (*core.Colony, error) {
	cmd := &command{colonyReplyChan: make(chan *core.Colony, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddColony(colony)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedColony, err := controller.db.GetColonyByID(colony.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.colonyReplyChan <- addedColony
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedColony := <-cmd.colonyReplyChan:
		return addedColony, nil
	}
}

func (controller *coloniesController) deleteColony(colonyID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteColonyByID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) addRuntime(runtime *core.Runtime) (*core.Runtime, error) {
	cmd := &command{runtimeReplyChan: make(chan *core.Runtime, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddRuntime(runtime)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedRuntime, err := controller.db.GetRuntimeByID(runtime.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.runtimeReplyChan <- addedRuntime
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedRuntime := <-cmd.runtimeReplyChan:
		return addedRuntime, nil
	}
}

func (controller *coloniesController) getRuntimeByID(runtimeID string) (*core.Runtime, error) {
	cmd := &command{runtimeReplyChan: make(chan *core.Runtime),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			runtime, err := controller.db.GetRuntimeByID(runtimeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.runtimeReplyChan <- runtime
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case runtime := <-cmd.runtimeReplyChan:
		return runtime, nil
	}
}

func (controller *coloniesController) getRuntimeByColonyID(colonyID string) ([]*core.Runtime, error) {
	cmd := &command{runtimesReplyChan: make(chan []*core.Runtime),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			runtimes, err := controller.db.GetRuntimesByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.runtimesReplyChan <- runtimes
		}}

	controller.cmdQueue <- cmd
	var runtimes []*core.Runtime
	select {
	case err := <-cmd.errorChan:
		return runtimes, err
	case runtimes := <-cmd.runtimesReplyChan:
		return runtimes, nil
	}

	return runtimes, nil
}

func (controller *coloniesController) approveRuntime(runtimeID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			runtime, err := controller.db.GetRuntimeByID(runtimeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.ApproveRuntime(runtime)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) rejectRuntime(runtimeID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			runtime, err := controller.db.GetRuntimeByID(runtimeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.RejectRuntime(runtime)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteRuntime(runtimeID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteRuntimeByID(runtimeID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) addProcess(process *core.Process) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddProcess(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			addedProcess, err := controller.db.GetProcessByID(process.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			controller.sendProcessesEvent(process)
			cmd.processReplyChan <- addedProcess
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case process := <-cmd.processReplyChan:
		return process, nil
	}
}

func (controller *coloniesController) getProcessByID(processID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processReplyChan <- process
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case process := <-cmd.processReplyChan:
		return process, nil
	}
}

func (controller *coloniesController) findProcessHistory(colonyID string, runtimeID string, seconds int, state int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			var err error
			if runtimeID == "" {
				processes, err = controller.db.FindProcessesForColony(colonyID, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			} else {
				processes, err = controller.db.FindProcessesForRuntime(colonyID, runtimeID, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) findPrioritizedProcesses(runtimeID string, colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindWaitingProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			prioritizedProcesses := controller.planner.Prioritize(runtimeID, processes, count)
			cmd.processesReplyChan <- prioritizedProcesses
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) findWaitingProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindWaitingProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) findRunningProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindRunningProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) findSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindSuccessfulProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) findFailedProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindFailedProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) deleteProcess(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteProcessByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteAllProcesses(colonyID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteAllProcessesForColony(colonyID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) closeSuccessful(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.MarkSuccessful(process)
			controller.sendProcessChangeStateEvent(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) closeFailed(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.MarkFailed(process)
			controller.sendProcessChangeStateEvent(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) assignRuntime(runtimeID string, colonyID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			runtime, err := controller.db.GetRuntimeByID(runtimeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if runtime == nil {
				cmd.errorChan <- errors.New("runtime with id <" + runtimeID + "> could not be found")
				return
			}

			err = controller.db.MarkAlive(runtime)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			var processes []*core.Process
			processes, err = controller.db.FindUnassignedProcesses(colonyID, runtimeID, runtime.RuntimeType, 10)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			selectedProcess, err := controller.planner.Select(runtimeID, processes)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			err = controller.db.AssignRuntime(runtimeID, selectedProcess)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			maxExecTime := selectedProcess.ProcessSpec.MaxExecTime
			if maxExecTime > 0 {
				err := controller.db.SetDeadline(selectedProcess, time.Now().Add(time.Duration(maxExecTime)*time.Second))
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}

			cmd.processReplyChan <- selectedProcess
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processes := <-cmd.processReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) unassignRuntime(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.UnassignRuntime(process)
			controller.sendProcessesEvent(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) getProcessStat(colonyID string) (*core.ProcessStat, error) {
	cmd := &command{processStatReplyChan: make(chan *core.ProcessStat),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			waiting, err := controller.db.NrWaitingProcessesForColony(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			running, err := controller.db.NrRunningProcessesForColony(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			success, err := controller.db.NrSuccessfulProcessesForColony(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failed, err := controller.db.NrFailedProcessesForColony(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.processStatReplyChan <- core.CreateProcessStat(waiting, running, success, failed)
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processStat := <-cmd.processStatReplyChan:
		return processStat, nil
	}
}

func (controller *coloniesController) addAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	cmd := &command{attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddAttribute(attribute)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedAttribute, err := controller.db.GetAttributeByID(attribute.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.attributeReplyChan <- addedAttribute
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedAttribute := <-cmd.attributeReplyChan:
		return addedAttribute, nil
	}
}

func (controller *coloniesController) getAttribute(attributeID string) (*core.Attribute, error) {
	cmd := &command{attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			attribute, err := controller.db.GetAttributeByID(attributeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.attributeReplyChan <- attribute
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case attribute := <-cmd.attributeReplyChan:
		return attribute, nil
	}
}

func (controller *coloniesController) stop() {
	controller.mutex.Lock()
	controller.stopFlag = true
	controller.mutex.Unlock()
	controller.cmdQueue <- &command{stop: true}
}
