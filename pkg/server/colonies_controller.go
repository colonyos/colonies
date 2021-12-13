package server

import "colonies/pkg/database"

type ColoniesController struct {
	db database.Database
}

func CreateColoniesController(db database.Database) *ColoniesController {
	controller := &ColoniesController{db: db}
	return controller
}

func (controller *ColoniesController) AddColony() {
}
