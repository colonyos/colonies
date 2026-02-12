package controllers

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

func (controller *ColoniesController) AddGenerator(generator *core.Generator) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.generatorDB.AddGenerator(generator)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedGenerator, err := controller.generatorDB.GetGeneratorByID(generator.ID)
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

func (controller *ColoniesController) TriggerGenerators() {
	cmd := &command{threaded: true, handler: func(cmd *command) {
		generatorsFromDB, err := controller.generatorDB.FindAllGenerators()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed get all generators from db")
			return
		}
		for _, generator := range generatorsFromDB {
			counter, err := controller.generatorDB.CountGeneratorArgs(generator.ID)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed count generator args from db")
				continue
			}
			now := time.Now()
			timeout := false
			if generator.LastRun.Unix() <= 0 { // Generator has never run
				if generator.FirstPack.Unix() <= 0 { // Generator has never been packed
					timeout = false
				} else { // Generator has been packed, calulcate deadline based first pack
					timeoutDeadline := generator.FirstPack.Add(time.Duration(generator.Timeout) * time.Second)
					timeout = now.Unix() > timeoutDeadline.Unix()
				}
			} else { // Generator has run before
				timeoutDeadline := generator.LastRun.Add(time.Duration(generator.Timeout) * time.Second)
				timeout = now.Unix() > timeoutDeadline.Unix()
			}
			if counter >= generator.Trigger {
				timesToSubmit := counter / generator.Trigger
				for i := 0; i < timesToSubmit; i++ {
					log.WithFields(log.Fields{
						"GeneratorId": generator.ID,
						"Counter":     counter}).
						Debug("Generator threshold reached, submitting workflow")
					controller.SubmitWorkflow(generator, generator.Trigger, generator.InitiatorID)
				}
			} else if counter >= 1 && generator.Timeout > 0 && timeout {
				log.WithFields(log.Fields{
					"GeneratorId": generator.ID,
					"Counter":     counter}).
					Debug("Generator threshold reached, submitting workflow")
				controller.SubmitWorkflow(generator, counter, generator.InitiatorID)
			}
		}
	}}

	controller.cmdQueue <- cmd
}

func (controller *ColoniesController) GetGenerator(generatorID string) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generator, err := controller.generatorDB.GetGeneratorByID(generatorID)
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

func (controller *ColoniesController) ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error) {
	cmd := &command{generatorReplyChan: make(chan *core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generator, err := controller.generatorDB.GetGeneratorByName(colonyName, generatorName)
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

func (controller *ColoniesController) GetGenerators(colonyName string, count int) ([]*core.Generator, error) {
	cmd := &command{generatorsReplyChan: make(chan []*core.Generator, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generators, err := controller.generatorDB.FindGeneratorsByColonyName(colonyName, count)
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

func (controller *ColoniesController) PackGenerator(generatorID string, colonyName, arg string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			generatorArg := core.CreateGeneratorArg(generatorID, colonyName, arg)
			err := controller.generatorDB.AddGeneratorArg(generatorArg)
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed add generator args")
				cmd.errorChan <- err
			}
			count, err := controller.generatorDB.CountGeneratorArgs(generatorID)
			log.WithFields(log.Fields{"Arg": arg, "Count": count, "GeneratorId": generatorID}).Debug("Added args to generator")

			generator, err := controller.generatorDB.GetGeneratorByID(generatorID)
			if err != nil {
				log.WithFields(log.Fields{"Error": err, "GeneratorId": generatorID}).Error("Failed to get generator")
				cmd.errorChan <- err
			}

			if generator.FirstPack.Unix() < 0 {
				err = controller.generatorDB.SetGeneratorFirstPack(generatorID)
				if err != nil {
					log.WithFields(log.Fields{"Error": err, "GeneratorId": generatorID}).Error("Failed to set generator first pack")
					cmd.errorChan <- err
				}
			}

			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return err
	}
}

func (controller *ColoniesController) GeneratorTriggerLoop() {
	for {
		time.Sleep(time.Duration(controller.generatorPeriod) * time.Millisecond)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		isLeader := controller.TryBecomeLeader()
		if isLeader {
			controller.TriggerGenerators()
		}
	}
}

func (controller *ColoniesController) SubmitWorkflow(generator *core.Generator, counter int, recoveredID string) {
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(generator.WorkflowSpec)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to parse workflow spec")
		return
	}

	generatorArgs, err := controller.generatorDB.GetGeneratorArgs(generator.ID, counter)
	var args []string
	for _, generatorArg := range generatorArgs {
		args = append(args, generatorArg.Arg)
	}

	log.WithFields(log.Fields{
		"GeneratorId": generator.ID,
		"Trigger":     generator.Trigger,
		"Counter":     counter,
		"Args":        args}).
		Debug("Generator submitting workflow")

	argsif := make([]interface{}, len(args))
	for i, v := range args {
		argsif[i] = v
	}

	_, err = controller.CreateProcessGraph(workflowSpec, argsif, make(map[string]interface{}), make([]interface{}, 0), recoveredID)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err}).
			Error("Failed to create generator processgraph")
		return
	}

	// Now it safe to remove the args since they are now attached to a process graph
	for _, generatorArg := range generatorArgs {
		count, err := controller.generatorDB.CountGeneratorArgs(generator.ID)
		log.WithFields(log.Fields{
			"GeneratorId": generator.ID,
			"Trigger":     generator.Trigger,
			"Count":       count,
			"Arg":         generatorArg.Arg}).
			Debug("Removing generator arg")

		err = controller.generatorDB.RemoveGeneratorArgByID(generatorArg.ID)
		if err != nil {
			log.WithFields(log.Fields{
				"Error": err}).
				Error("Failed to remove generator arg")
			return
		}
	}

	err = controller.generatorDB.SetGeneratorLastRun(generator.ID)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed mark generator as run")
	}
}
