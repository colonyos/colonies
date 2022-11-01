package server

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	cronlib "github.com/colonyos/colonies/pkg/cron"
	log "github.com/sirupsen/logrus"
)

func (controller *coloniesController) addCron(cron *core.Cron) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddCron(cron)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedCron, err := controller.db.GetCronByID(cron.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.cronReplyChan <- addedCron
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedCron := <-cmd.cronReplyChan:
		return addedCron, nil
	}
}

func (controller *coloniesController) deleteGenerator(generatorID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cmd.errorChan <- controller.db.DeleteGeneratorByID(generatorID)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) getCron(cronID string) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cron, err := controller.db.GetCronByID(cronID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.cronReplyChan <- cron
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case cron := <-cmd.cronReplyChan:
		return cron, nil
	}
}

func (controller *coloniesController) getCrons(colonyID string, count int) ([]*core.Cron, error) {
	cmd := &command{cronsReplyChan: make(chan []*core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			crons, err := controller.db.FindCronsByColonyID(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.cronsReplyChan <- crons
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case crons := <-cmd.cronsReplyChan:
		return crons, nil
	}
}

func (controller *coloniesController) runCron(cronID string) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cron, err := controller.db.GetCronByID(cronID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			controller.startCron(cron)
			cmd.cronReplyChan <- cron
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case cron := <-cmd.cronReplyChan:
		return cron, nil
	}
}

func (controller *coloniesController) deleteCron(cronID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteCronByID(cronID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) calcNextRun(cron *core.Cron) time.Time {
	nextRun := time.Time{}
	var err error
	if cron.Interval > 0 {
		nextRun, err = cronlib.NextIntervall(cron.Interval)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed generate random next run")
		}
	} else if cron.Interval > 0 && cron.Random {
		nextRun, err = cronlib.Random(cron.Interval)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed generate random next run")
		}
	} else {
		nextRun, err = cronlib.Next(cron.CronExpression)
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed generate next run based on cron expression")
		}
	}

	return nextRun
}

func (controller *coloniesController) startCron(cron *core.Cron) {
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(cron.WorkflowSpec)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to parsing WorkflowSpec")
		return
	}

	rootInput := []string{}
	// Pick all outputs from the leaves of the previos processgraph, and
	// then use it as input to the root process in the next processgraph
	if cron.PrevProcessGraphID != "" {
		processGraph, err := controller.db.GetProcessGraphByID(cron.PrevProcessGraphID)
		if err == nil && processGraph != nil {
			processGraph.SetStorage(controller.db)
			leafIDs, err := processGraph.Leaves()
			if err == nil {
				for _, leafID := range leafIDs {
					leaf, err := controller.db.GetProcessByID(leafID)
					if err == nil {
						rootInput = append(rootInput, leaf.Output...)
					}
				}

			}
		}
	}

	processGraph, err := controller.createProcessGraph(workflowSpec, []string{}, rootInput)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "CronId": cron.ID}).Error("Failed to create cron processgraph")
		return
	}

	nextRun := controller.calcNextRun(cron)
	controller.db.UpdateCron(cron.ID, nextRun, time.Now(), processGraph.ID)
}

func (controller *coloniesController) triggerCrons() {
	cmd := &command{handler: func(cmd *command) {
		crons, err := controller.db.FindAllCrons()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed getting all crons")
			return
		}
		for _, cron := range crons {
			t := time.Time{}
			if t.Unix() == cron.NextRun.Unix() { // This if-statement will be true the first time the cron is evaluted
				nextRun := controller.calcNextRun(cron)
				controller.db.UpdateCron(cron.ID, nextRun, time.Time{}, "")
				cron.NextRun = nextRun
				continue
			}
			if cron.HasExpired() {
				processgraph, err := controller.db.GetProcessGraphByID(cron.PrevProcessGraphID)
				if err != nil {
					log.WithFields(log.Fields{"Error": err, "PrevProcessGraphID": cron.PrevProcessGraphID}).Error("Failed getting all crons")
					continue
				}
				if processgraph == nil {
					controller.startCron(cron)
					continue
				}
				if cron.WaitForPrevProcessGraph {
					if processgraph.State == core.SUCCESS || processgraph.State == core.FAILED {
						log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Triggering cron workflow")
						controller.startCron(cron)
					}
				} else {
					log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Triggering cron workflow")
					controller.startCron(cron)
				}
			}
		}
	}}

	controller.cmdQueue <- cmd
}

func (controller *coloniesController) cronTriggerLoop() {
	for {
		time.Sleep(TIMEOUT_CRON_TRIGGER_INTERVALL * time.Second)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		isLeader := controller.tryBecomeLeader()
		if isLeader {
			controller.triggerCrons()
		}
	}
}
