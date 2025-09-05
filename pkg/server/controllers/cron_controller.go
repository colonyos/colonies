package controllers

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
	cronlib "github.com/colonyos/colonies/pkg/cron"
	log "github.com/sirupsen/logrus"
)

func (controller *ColoniesController) AddCron(cron *core.Cron) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.cronDB.AddCron(cron)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedCron, err := controller.cronDB.GetCronByID(cron.ID)
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

func (controller *ColoniesController) RemoveGenerator(generatorID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cmd.errorChan <- controller.generatorDB.RemoveGeneratorByID(generatorID)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) GetCron(cronID string) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cron, err := controller.cronDB.GetCronByID(cronID)
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

func (controller *ColoniesController) GetCrons(colonyName string, count int) ([]*core.Cron, error) {
	cmd := &command{cronsReplyChan: make(chan []*core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			crons, err := controller.cronDB.FindCronsByColonyName(colonyName, count)
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

func (controller *ColoniesController) RunCron(cronID string) (*core.Cron, error) {
	cmd := &command{cronReplyChan: make(chan *core.Cron, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			cron, err := controller.cronDB.GetCronByID(cronID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			controller.StartCron(cron)
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

func (controller *ColoniesController) RemoveCron(cronID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.cronDB.RemoveCronByID(cronID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) CalcNextRun(cron *core.Cron) time.Time {
	nextRun := time.Time{}
	var err error
	if cron.Interval > 0 {
		nextRun, err = cronlib.NextInterval(cron.Interval)
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

func (controller *ColoniesController) StartCron(cron *core.Cron) {
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(cron.WorkflowSpec)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to parsing WorkflowSpec")
		return
	}

	var rootInput []interface{}
	// Pick all outputs from the leaves of the previous processgraph and
	// then use it as input to the root process in the next processgraph
	if cron.PrevProcessGraphID != "" {
		processGraph, err := controller.processGraphDB.GetProcessGraphByID(cron.PrevProcessGraphID)
		if err == nil && processGraph != nil {
			processGraph.SetStorage(controller.GetProcessGraphStorage())
			leafIDs, err := processGraph.Leaves()
			if err == nil {
				for _, leafID := range leafIDs {
					leaf, err := controller.processDB.GetProcessByID(leafID)
					if err == nil {
						rootInput = append(rootInput, leaf.Output...)
					}
				}

			}
		}
	}

	processGraph, err := controller.CreateProcessGraph(workflowSpec, make([]interface{}, 0), make(map[string]interface{}), rootInput, cron.InitiatorID)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "CronId": cron.ID}).Error("Failed to create cron processgraph")
		return
	}

	nextRun := controller.CalcNextRun(cron)
	controller.cronDB.UpdateCron(cron.ID, nextRun, time.Now(), processGraph.ID)
}

func (controller *ColoniesController) TriggerCrons() {
	cmd := &command{handler: func(cmd *command) {
		crons, err := controller.cronDB.FindAllCrons()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed getting all crons")
			return
		}
		for _, cron := range crons {
			t := time.Time{}
			if t.Unix() == cron.NextRun.Unix() { // This if-statement will be true the first time the cron is evaluted
				nextRun := controller.CalcNextRun(cron)
				controller.cronDB.UpdateCron(cron.ID, nextRun, time.Time{}, "")
				cron.NextRun = nextRun
				continue
			}
			if cron.HasExpired() {
				processgraph, err := controller.processGraphDB.GetProcessGraphByID(cron.PrevProcessGraphID)
				if err != nil {
					log.WithFields(log.Fields{"Error": err, "PrevProcessGraphId": cron.PrevProcessGraphID}).Error("Failed getting all crons")
					continue
				}
				if processgraph == nil {
					controller.StartCron(cron)
					continue
				}
				if cron.WaitForPrevProcessGraph {
					if processgraph.State == core.SUCCESS || processgraph.State == core.FAILED {
						log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Triggering cron workflow")
						controller.StartCron(cron)
					}
				} else {
					log.WithFields(log.Fields{"CronId": cron.ID}).Debug("Triggering cron workflow")
					controller.StartCron(cron)
				}
			}
		}
	}}

	controller.cmdQueue <- cmd
}

func (controller *ColoniesController) CronTriggerLoop() {
	for {
		time.Sleep(time.Duration(controller.cronPeriod) * time.Millisecond)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		isLeader := controller.TryBecomeLeader()
		if isLeader {
			controller.TriggerCrons()
		}
	}
}
