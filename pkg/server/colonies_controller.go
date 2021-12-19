package server

import (
	"colonies/pkg/core"
	"colonies/pkg/database"
)

type ColoniesController struct {
	db database.Database
}

func CreateColoniesController(db database.Database) *ColoniesController {
	controller := &ColoniesController{db: db}
	return controller
}

func (controller *ColoniesController) GetColonies() ([]*core.Colony, error) {
	var colonies []*core.Colony
	colonies, err := controller.db.GetColonies()
	if err != nil {
		return colonies, err
	}

	return colonies, nil
}

func (controller *ColoniesController) GetColonyByID(colonyID string) (*core.Colony, error) {
	colony, err := controller.db.GetColonyByID(colonyID)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

func (controller *ColoniesController) AddColony(colony *core.Colony) error {
	err := controller.db.AddColony(colony)
	if err != nil {
		return err
	}

	return nil
}

func (controller *ColoniesController) AddComputer(computer *core.Computer) error {
	err := controller.db.AddComputer(computer)
	if err != nil {
		return err
	}

	return nil
}

func (controller *ColoniesController) GetComputer(computerID string) (*core.Computer, error) {
	computer, err := controller.db.GetComputerByID(computerID)
	if err != nil {
		return nil, err
	}

	return computer, nil
}

func (controller *ColoniesController) GetComputerByColonyID(colonyID string) ([]*core.Computer, error) {
	computers, err := controller.db.GetComputersByColonyID(colonyID)
	if err != nil {
		return nil, err
	}

	return computers, nil
}

func (controller *ColoniesController) AddProcess(process *core.Process) error {
	err := controller.db.AddProcess(process)
	if err != nil {
		return err
	}

	return nil
}

func (controller *ColoniesController) FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error) {
	// if count > MAX_COUNT {
	// 	return errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
	// }

	// processes, err := db.FindWaitingProcesses(colonyID, count)
	// if err != nil {
	// 	return err
	// }

	// for _, process := range processes {
	// 	fmt.Println(process)
	// }

	// TODO
	return nil, nil
}
