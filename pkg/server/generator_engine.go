package server

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	log "github.com/sirupsen/logrus"
)

const MaxCount = 1000000

type state struct {
	generator    *core.Generator
	workflowSpec *core.WorkflowSpec
}

type generatorEngine struct {
	controller *coloniesController
	db         database.Database
	states     map[string]*state
}

func createGeneratorEngine(db database.Database, controller *coloniesController) *generatorEngine {
	engine := &generatorEngine{}
	engine.controller = controller
	engine.db = db
	engine.states = make(map[string]*state)

	return engine
}

func (engine *generatorEngine) syncStatesFromDB() {
	colonies, err := engine.db.GetColonies()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to list colonies")
	}
	for _, colony := range colonies {
		generatorsFromDB, err := engine.db.FindGeneratorsByColonyID(colony.ID, MaxCount)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed list generators by id")
		}
		tempMap := make(map[string]bool)
		// Add state objects from db not found in states
		for _, generator := range generatorsFromDB {
			if _, ok := engine.states[generator.ID]; !ok {
				workflowSpec, err := core.ConvertJSONToWorkflowSpec(generator.WorkflowSpec)
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("Failed to parse workflow spec")
				} else {
					log.WithFields(log.Fields{"GeneratorId": generator.ID}).Info("Adding generator to engine and submitting workflow")
					generator.LastRun = time.Now()
					generator.Counter = 0
					state := &state{generator: generator, workflowSpec: workflowSpec}
					engine.states[generator.ID] = state
				}
			}
			tempMap[generator.ID] = true
		}
		// Delete state objects from states not found on db
		for _, state := range engine.states {
			generator := state.generator
			if _, ok := tempMap[generator.ID]; !ok {
				log.WithFields(log.Fields{
					"GeneratorId": generator.ID}).
					Info("Deleting generator from engine")
				if colony.ID == generator.ColonyID {
					delete(engine.states, generator.ID)
				}
			}
		}
	}
}

func (engine *generatorEngine) submitWorkflow(state *state) {
	if engine.controller != nil {
		state.generator.Counter = 0
		state.generator.LastRun = time.Now()
		_, err := engine.controller.createProcessGraph(state.workflowSpec)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err}).
				Error("Failed to create processgraph")
		}
	} else {
		log.Error("Failed to submit workflow in generator engine as coloniesController is nil")
	}
}

func (engine *generatorEngine) getGenerator(generatorID string) *core.Generator {
	if state, ok := engine.states[generatorID]; ok {
		return state.generator
	}
	return nil
}

func (engine *generatorEngine) increaseCounter(generatorID string) error {
	if state, ok := engine.states[generatorID]; ok {
		state.generator.Counter++
		engine.triggerGenerators()
	} else {
		log.WithFields(log.Fields{
			"GeneratorId": generatorID}).
			Error("Generator does not exists")
		return errors.New("Invalid generator Id")
	}

	return nil
}

func (engine *generatorEngine) triggerGenerators() {
	for _, state := range engine.states {
		now := time.Now()
		deadline := state.generator.LastRun.Add(time.Duration(state.generator.Timeout) * time.Second)
		if state.generator.Counter > 0 && now.Unix() > deadline.Unix() {
			log.WithFields(log.Fields{
				"GeneratorId": state.generator.ID,
				"Timeout":     state.generator.Timeout}).
				Info("Generator timed out, submitting workflow")
			engine.submitWorkflow(state)
		}
		if state.generator.Counter > state.generator.Trigger {
			log.WithFields(log.Fields{
				"GeneratorId": state.generator.ID,
				"Timeout":     state.generator.Timeout}).
				Info("Generator counter exceeded trigger value, submitting workflow")
			engine.submitWorkflow(state)
		}
	}
}
