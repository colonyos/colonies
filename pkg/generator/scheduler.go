package generator

import (
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	log "github.com/sirupsen/logrus"
)

const MaxCount = 1000000

type state struct {
	generator    *core.Generator
	workflowSpec *core.WorkflowSpec
}

type GeneratorScheduler struct {
	db     database.Database
	states map[string]*state
}

func CreateGeneratorScheduler(db database.Database) *GeneratorScheduler {
	scheduler := &GeneratorScheduler{}
	scheduler.db = db
	scheduler.states = make(map[string]*state)

	return scheduler
}

func (scheduler *GeneratorScheduler) syncStates() {
	colonies, err := scheduler.db.GetColonies()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to list colonies")
	}
	for _, colony := range colonies {
		generatorsFromDB, err := scheduler.db.FindGeneratorsByColonyID(colony.ID, MaxCount)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed list generators by id")
		}
		tempMap := make(map[string]bool)
		// add state objects from db not found in states
		for _, generator := range generatorsFromDB {
			if _, ok := scheduler.states[generator.ID]; !ok {
				log.WithFields(log.Fields{"GeneratorId": generator.ID}).Info("Adding generator to scheduler)")
				scheduler.states[generator.ID] = &state{generator: generator}
			}
			tempMap[generator.ID] = true
		}
		// delete state objects from states not found on db
		for _, state := range scheduler.states {
			generator := state.generator
			if _, ok := tempMap[generator.ID]; !ok {
				log.WithFields(log.Fields{"GeneratorId": generator.ID}).Info("Deleting generator from scheduler)")
				delete(scheduler.states, generator.ID)
			}
		}
	}
}
