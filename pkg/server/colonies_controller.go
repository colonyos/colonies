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

func (controller *ColoniesController) GetColony(colonyID string) (*core.Colony, error) {
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

func (controller *ColoniesController) AddWorker(worker *core.Worker) error {
	err := controller.db.AddWorker(worker)
	if err != nil {
		return err
	}

	return nil
}

func (controller *ColoniesController) GetWorker(workerID string) (*core.Worker, error) {
	worker, err := controller.db.GetWorkerByID(workerID)
	if err != nil {
		return nil, err
	}

	return worker, nil
}
