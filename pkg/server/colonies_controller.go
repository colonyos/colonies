package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
	"colonies/pkg/logging"
	"colonies/pkg/scheduler"
	"colonies/pkg/scheduler/basic"
	"errors"
	"strconv"
)

type command struct {
	stop               bool
	errorChan          chan error
	process            *core.Process
	count              int
	colony             *core.Colony
	colonyID           string
	colonyReplyChan    chan *core.Colony
	coloniesReplyChan  chan []*core.Colony
	processReplyChan   chan *core.Process
	processesReplyChan chan []*core.Process
	computerReplyChan  chan *core.Computer
	computersReplyChan chan []*core.Computer
	attributeReplyChan chan *core.Attribute
	handler            func(cmd *command)
}

type ColoniesController struct {
	db        database.Database
	cmdQueue  chan *command
	scheduler scheduler.Scheduler
}

func CreateColoniesController(db database.Database) *ColoniesController {
	controller := &ColoniesController{db: db}
	controller.cmdQueue = make(chan *command)
	controller.scheduler = basic.CreateScheduler()
	go controller.masterWorker()
	return controller
}

func (controller *ColoniesController) masterWorker() {
	for {
		select {
		case msg := <-controller.cmdQueue:
			if msg.stop {
				logging.Log().Info("Stopping Colonies controller")
				return
			}
			if msg.handler != nil {
				msg.handler(msg)
			}
		}
	}
}

func (controller *ColoniesController) GetColonies() ([]*core.Colony, error) {
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

func (controller *ColoniesController) GetColonyByID(colonyID string) (*core.Colony, error) {
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

func (controller *ColoniesController) AddColony(colony *core.Colony) (*core.Colony, error) {
	cmd := &command{colonyReplyChan: make(chan *core.Colony, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddColony(colony)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedColony, err := controller.db.GetColonyByID(colony.ID())
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

func (controller *ColoniesController) AddComputer(computer *core.Computer) (*core.Computer, error) {
	cmd := &command{computerReplyChan: make(chan *core.Computer, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddComputer(computer)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedComputer, err := controller.db.GetComputerByID(computer.ID())
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.computerReplyChan <- addedComputer
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedComputer := <-cmd.computerReplyChan:
		return addedComputer, nil
	}
}

func (controller *ColoniesController) GetComputerByID(computerID string) (*core.Computer, error) {
	cmd := &command{computerReplyChan: make(chan *core.Computer),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			computer, err := controller.db.GetComputerByID(computerID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.computerReplyChan <- computer
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case computer := <-cmd.computerReplyChan:
		return computer, nil
	}
}

func (controller *ColoniesController) GetComputerByColonyID(colonyID string) ([]*core.Computer, error) {
	cmd := &command{computersReplyChan: make(chan []*core.Computer),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			computers, err := controller.db.GetComputersByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.computersReplyChan <- computers
		}}

	controller.cmdQueue <- cmd
	var computers []*core.Computer
	select {
	case err := <-cmd.errorChan:
		return computers, err
	case computers := <-cmd.computersReplyChan:
		return computers, nil
	}

	return computers, nil
}

func (controller *ColoniesController) ApproveComputer(computerID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			computer, err := controller.db.GetComputerByID(computerID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.ApproveComputer(computer)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) RejectComputer(computerID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			computer, err := controller.db.GetComputerByID(computerID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.RejectComputer(computer)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) AddProcess(process *core.Process) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddProcess(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedProcess, err := controller.db.GetProcessByID(process.ID())
			if err != nil {
				cmd.errorChan <- err
				return
			}
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

func (controller *ColoniesController) GetProcessByID(colonyID string, processID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if process.TargetColonyID() != colonyID { // TODO: These kinds of checks should be done by security
				cmd.errorChan <- errors.New("Process not bound to specifid colony id <" + colonyID + ">")
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

func (controller *ColoniesController) FindWaitingProcesses(computerID string, colonyID string, count int) ([]*core.Process, error) {
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
			prioritizedProcesses := controller.scheduler.Prioritize(computerID, processes, count)
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

func (controller *ColoniesController) MarkSuccessful(computerID string, processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if process.AssignedComputerID() != computerID { // TODO: Move to security
				cmd.errorChan <- errors.New("Computer is not assigned to process, cannot mark as succesful")
				return
			}
			cmd.errorChan <- controller.db.MarkSuccessful(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) MarkFailed(computerID string, processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if process.AssignedComputerID() != computerID { // TODO: Move to security
				cmd.errorChan <- errors.New("Computer is not assigned to process, cannot mark as succesful")
				return
			}
			cmd.errorChan <- controller.db.MarkFailed(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) AssignProcess(computerID string, colonyID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			processes, err := controller.db.FindUnassignedProcesses(colonyID, computerID, 10)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			selectedProcesses, err := controller.scheduler.Select(computerID, processes)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			err = controller.db.AssignComputer(computerID, selectedProcesses)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processReplyChan <- selectedProcesses
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processes := <-cmd.processReplyChan:
		return processes, nil
	}
}

func (controller *ColoniesController) Stop() {
	controller.cmdQueue <- &command{stop: true}
}

func (controller *ColoniesController) AddAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	cmd := &command{attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddAttribute(attribute)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedAttribute, err := controller.db.GetAttributeByID(attribute.ID())
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

func (controller *ColoniesController) GetAttribute(attributeID string) (*core.Attribute, error) {
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
