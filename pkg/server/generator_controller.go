package server

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

func (controller *coloniesController) addGenerator(generator *core.Generator) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddGenerator(generator)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedGenerator, err := controller.db.GetGeneratorByID(generator.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.generatorReplyChan <- addedGenerator
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedGenerator := <-cmd.generatorReplyChan:
		return addedGenerator, nil
	}
}

func (controller *coloniesController) triggerGenerators() {
	cmd := &command{handler: func(cmd *command) {
		generatorsFromDB, err := controller.db.FindAllGenerators()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed get all generators from db")
			return
		}
		for _, generator := range generatorsFromDB {
			counter, err := controller.db.CountGeneratorArgs(generator.ID)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed count generator args from db")
				continue
			}
			if counter >= generator.Trigger {
				timesToSubmit := counter / generator.Trigger
				for i := 0; i < timesToSubmit; i++ {
					log.WithFields(log.Fields{
						"GeneratorId": generator.ID,
						"Counter":     counter}).
						Info("Generator threshold reached, submitting workflow")
					controller.submitWorkflow(generator)
				}
			}
		}
	}}

	controller.cmdQueue <- cmd
}

func (controller *coloniesController) getGenerator(generatorID string) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generator, err := controller.db.GetGeneratorByID(generatorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.generatorReplyChan <- generator
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case generator := <-cmd.generatorReplyChan:
		return generator, nil
	}
}

func (controller *coloniesController) resolveGenerator(generatorName string) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generator, err := controller.db.GetGeneratorByName(generatorName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.generatorReplyChan <- generator
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case generator := <-cmd.generatorReplyChan:
		return generator, nil
	}
}

func (controller *coloniesController) getGenerators(colonyID string, count int) ([]*core.Generator, error) {
	cmd := &command{generatorsReplyChan: make(chan []*core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generators, err := controller.db.FindGeneratorsByColonyID(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.generatorsReplyChan <- generators
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case generators := <-cmd.generatorsReplyChan:
		return generators, nil
	}
}

func (controller *coloniesController) packGenerator(generatorID string, colonyID, arg string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generatorArg := core.CreateGeneratorArg(generatorID, colonyID, arg)
			cmd.errorChan <- controller.db.AddGeneratorArg(generatorArg)
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return err
	}
}

func (controller *coloniesController) generatorTriggerLoop() {
	for {
		time.Sleep(TIMEOUT_GENERATOR_TRIGGER_INTERVALL * time.Second)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		isLeader := controller.tryBecomeLeader()
		if isLeader {
			controller.triggerGenerators()
		}
	}
}

func (controller *coloniesController) submitWorkflow(generator *core.Generator) {
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(generator.WorkflowSpec)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to parse workflow spec")
		return
	}

	generatorArgs, err := controller.db.GetGeneratorArgs(generator.ID, generator.Trigger)
	var args []string
	for _, generatorArg := range generatorArgs {
		args = append(args, generatorArg.Arg)
	}

	log.WithFields(log.Fields{
		"GeneratorId": generator.ID,
		"Trigger":     generator.Trigger,
		"Args":        args}).
		Debug("Generator submitting workflow")

	_, err = controller.createProcessGraph(workflowSpec, args, []string{})
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err}).
			Error("Failed to create processgraph")
		return
	}

	// Now it safe to remove the args since they are now attached to a process graph
	for _, generatorArg := range generatorArgs {
		controller.db.DeleteGeneratorArgByID(generatorArg.ID)
	}

	err = controller.db.SetGeneratorLastRun(generator.ID)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed mark generator as run")
	}
}
