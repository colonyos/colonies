package controllers

import (
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/constants"
	log "github.com/sirupsen/logrus"
)

func (controller *ColoniesController) IsLeader() bool {
	areWeLeader := controller.etcdServer.Leader() == controller.thisNode.Name
	if areWeLeader && !controller.leader {
		log.WithFields(log.Fields{"EtcdNode": controller.thisNode.Name}).Debug("ColoniesServer became leader")
		controller.leader = true
	}

	if !areWeLeader && controller.leader {
		log.WithFields(log.Fields{"EtcdNode": controller.thisNode.Name}).Debug("ColoniesServer is no longer leader")
		controller.leader = false
	}

	return areWeLeader
}

func (controller *ColoniesController) TryBecomeLeader() bool {
	var isLeader bool
	controller.leaderMutex.Lock()
	isLeader = controller.IsLeader()
	controller.leaderMutex.Unlock()

	return isLeader
}

func (controller *ColoniesController) TimeoutLoop() {
	for {
		time.Sleep(constants.RELEASE_PERIOD * time.Second)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		processes, err := controller.processDB.FindAllRunningProcesses()
		if err != nil {
			log.Error(err)
			continue
		}
		for _, process := range processes {
			if process.FunctionSpec.MaxExecTime == -1 {
				continue
			}
			if time.Now().Unix() > process.ExecDeadline.Unix() {
				if process.Retries >= process.FunctionSpec.MaxRetries && process.FunctionSpec.MaxRetries > -1 {
					err := controller.CloseFailed(process.ID, []string{"Maximum execution time limit exceeded"})
					if err != nil {
						log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Debug("Max retries reached, but failed to close process")
						continue
					}
					log.WithFields(log.Fields{"ProcessId": process.ID, "MaxExecTime": process.FunctionSpec.MaxExecTime, "MaxRetries": process.FunctionSpec.MaxRetries}).Debug("Process closed as failed as max retries reached")
					continue
				}

				err := controller.UnassignExecutor(process.ID)
				if err != nil {
					log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to unassign process")
				}
				log.WithFields(log.Fields{"ProcessId": process.ID, "MaxExecTime": process.FunctionSpec.MaxExecTime, "MaxRetries": process.FunctionSpec.MaxRetries}).Debug("Process was unassigned as it did not complete in time")
			}
		}

		// TODO: FindAllWaitingProcesses will only return max 1000 processes, this is to avoid dumping the entire database
		// However, the means that maxWaitTime may not work correctly if there are more than 1000 waiting processes.
		processes, err = controller.processDB.FindAllWaitingProcesses()
		if err != nil {
			continue
		}
		for _, process := range processes {
			if process.FunctionSpec.MaxWaitTime == -1 || process.FunctionSpec.MaxWaitTime == 0 {
				continue
			}

			if time.Now().Unix() > process.WaitDeadline.Unix() {
				err := controller.CloseFailed(process.ID, []string{"Maximum waiting time limit exceeded"})
				if err != nil {
					log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Debug("Max waiting time reached, but failed to close process")
					continue
				}
				log.WithFields(log.Fields{"ProcessId": process.ID, "MaxWaitTime": process.FunctionSpec.MaxWaitTime}).Debug("Process closed as failed as maximum waiting time limit exceeded")
			}
		}
	}
}

func (controller *ColoniesController) BlockingCmdQueueWorker() {
	for {
		select {
		case cmd := <-controller.blockingCmdQueue:
			if cmd.stop {
				return
			}
			if cmd.handler != nil {
				if cmd.threaded {
					go cmd.handler(cmd)
				} else {
					cmd.handler(cmd)
				}
			}
		}
	}
}

func (controller *ColoniesController) RetentionWorker() {
	for {
		isLeader := controller.TryBecomeLeader()

		if isLeader && controller.retention {
			log.Debug("Appling retention policy")
			controller.databaseCore.ApplyRetentionPolicy(controller.retentionPolicy)
		}

		time.Sleep(time.Duration(controller.retentionPeriod) * time.Millisecond)
	}
}

func (controller *ColoniesController) CmdQueueWorker() {
	for {
		select {
		case cmd := <-controller.cmdQueue:
			if cmd.stop {
				return
			}
			if cmd.handler != nil {
				if cmd.threaded {
					go cmd.handler(cmd)
				} else {
					cmd.handler(cmd)
				}
			}
		}
	}
}

func (controller *ColoniesController) CleanupWorker() {
	cleanupInterval := 60 * time.Second // 60 second interval

	// Read cleanup interval from environment if set
	if envInterval := os.Getenv("COLONIES_CLEANUP_INTERVAL"); envInterval != "" {
		if interval, err := strconv.Atoi(envInterval); err == nil {
			cleanupInterval = time.Duration(interval) * time.Second
		}
	}

	log.WithFields(log.Fields{
		"CleanupInterval":       cleanupInterval,
		"StaleExecutorDuration": controller.staleExecutorDuration,
	}).Info("Starting cleanup worker")

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			controller.stopMutex.Lock()
			shouldStop := controller.stopFlag
			controller.stopMutex.Unlock()

			if shouldStop {
				return
			}

			// Only run cleanup if this node is the leader
			controller.leaderMutex.Lock()
			isLeader := controller.leader
			controller.leaderMutex.Unlock()

			if isLeader {
				controller.cleanupStaleExecutors()
			}
		}
	}
}

func (controller *ColoniesController) cleanupStaleExecutors() {
	executors, err := controller.executorDB.GetExecutors()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Failed to get executors for cleanup")
		return
	}

	now := time.Now()
	cleanedCount := 0

	for _, executor := range executors {
		// Skip executors that have never communicated (LastHeardFromTime is zero)
		// This prevents removing newly registered executors before they have a chance to communicate
		if executor.LastHeardFromTime.IsZero() {
			continue
		}

		timeSinceLastHeard := now.Sub(executor.LastHeardFromTime)
		if timeSinceLastHeard > controller.staleExecutorDuration {
			log.WithFields(log.Fields{
				"ExecutorName":       executor.Name,
				"ExecutorID":         executor.ID,
				"ColonyName":         executor.ColonyName,
				"TimeSinceLastHeard": timeSinceLastHeard,
			}).Info("Auto-removing stale executor")

			err := controller.executorDB.RemoveExecutorByName(executor.ColonyName, executor.Name)
			if err != nil {
				log.WithFields(log.Fields{
					"Error":        err,
					"ExecutorName": executor.Name,
				}).Warn("Failed to remove stale executor")
			} else {
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		log.WithFields(log.Fields{"CleanedExecutors": cleanedCount}).Info("Completed executor cleanup")
	}
}
