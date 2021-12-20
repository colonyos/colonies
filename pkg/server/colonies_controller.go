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
			} else {
				cmd.coloniesReplyChan <- colonies
			}
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
			} else {
				cmd.colonyReplyChan <- colony
			}
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case colony := <-cmd.colonyReplyChan:
		return colony, nil
	}
}

func (controller *ColoniesController) AddColony(colony *core.Colony) error {
	cmd := &command{errorChan: make(chan error, 1), handler: func(cmd *command) {
		cmd.errorChan <- controller.db.AddColony(colony)
	}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) AddComputer(computer *core.Computer) error {
	cmd := &command{errorChan: make(chan error, 1), handler: func(cmd *command) {
		cmd.errorChan <- controller.db.AddComputer(computer)
	}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) GetComputerByID(computerID string) (*core.Computer, error) {
	cmd := &command{computerReplyChan: make(chan *core.Computer),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			computer, err := controller.db.GetComputerByID(computerID)
			if err != nil {
				cmd.errorChan <- err
			} else {
				cmd.computerReplyChan <- computer
			}
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
			} else {
				cmd.computersReplyChan <- computers
			}
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

func (controller *ColoniesController) AddProcess(process *core.Process) error {
	cmd := &command{errorChan: make(chan error, 1), handler: func(cmd *command) {
		cmd.errorChan <- controller.db.AddProcess(process)
	}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) FindWaitingProcesses(computerID string, colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
			}

			processes, err := controller.db.FindWaitingProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
			} else {
				prioritizedProcesses := controller.scheduler.Prioritize(computerID, processes, count)
				cmd.processesReplyChan <- prioritizedProcesses
			}
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

func (controller *ColoniesController) AssignProcess(computerID string, colonyID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			processes, err := controller.db.FindUnassignedProcesses(colonyID, computerID, 10)

			if err != nil {
				cmd.errorChan <- err
				return
			} else {
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
			}
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
