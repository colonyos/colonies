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
	cleanupInterval := 60 * time.Second      // 60 second interval
	staleExecutorDuration := 10 * time.Minute // Executors not heard from in 10 minutes are stale
	staleNodeDuration := 24 * time.Hour       // Nodes with no executors for 24 hours are stale

	// Read cleanup intervals from environment if set
	if envInterval := os.Getenv("COLONIES_CLEANUP_INTERVAL"); envInterval != "" {
		if interval, err := strconv.Atoi(envInterval); err == nil {
			cleanupInterval = time.Duration(interval) * time.Second
		}
	}
	if envStaleExecutor := os.Getenv("COLONIES_STALE_EXECUTOR_DURATION"); envStaleExecutor != "" {
		if duration, err := strconv.Atoi(envStaleExecutor); err == nil {
			staleExecutorDuration = time.Duration(duration) * time.Second
		}
	}
	if envStaleNode := os.Getenv("COLONIES_STALE_NODE_DURATION"); envStaleNode != "" {
		if duration, err := strconv.Atoi(envStaleNode); err == nil {
			staleNodeDuration = time.Duration(duration) * time.Second
		}
	}

	log.WithFields(log.Fields{
		"CleanupInterval":       cleanupInterval,
		"StaleExecutorDuration": staleExecutorDuration,
		"StaleNodeDuration":     staleNodeDuration,
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
				controller.cleanupStaleExecutors(staleExecutorDuration)
				controller.cleanupStaleNodes(staleNodeDuration)
			}
		}
	}
}

func (controller *ColoniesController) cleanupStaleExecutors(staleDuration time.Duration) {
	executors, err := controller.executorDB.GetExecutors()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Failed to get executors for cleanup")
		return
	}

	now := time.Now()
	cleanedCount := 0

	for _, executor := range executors {
		timeSinceLastHeard := now.Sub(executor.LastHeardFromTime)
		if timeSinceLastHeard > staleDuration {
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

func (controller *ColoniesController) cleanupStaleNodes(staleDuration time.Duration) {
	colonies, err := controller.colonyDB.GetColonies()
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn("Failed to get colonies for node cleanup")
		return
	}

	cleanedCount := 0
	now := time.Now()

	for _, colony := range colonies {
		nodes, err := controller.nodeDB.GetNodes(colony.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":      err,
				"ColonyName": colony.Name,
			}).Warn("Failed to get nodes for cleanup")
			continue
		}

		executors, err := controller.executorDB.GetExecutorsByColonyName(colony.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":      err,
				"ColonyName": colony.Name,
			}).Warn("Failed to get executors for node cleanup")
			continue
		}

		// Build map of nodeID -> executor count
		nodeExecutorCount := make(map[string]int)
		for _, executor := range executors {
			if executor.NodeID != "" {
				nodeExecutorCount[executor.NodeID]++
			}
		}

		// Remove nodes with no executors that haven't been seen in a while
		for _, node := range nodes {
			if nodeExecutorCount[node.ID] == 0 {
				nodeAge := now.Sub(node.LastSeen)
				if nodeAge > staleDuration {
					log.WithFields(log.Fields{
						"NodeName":   node.Name,
						"NodeID":     node.ID,
						"ColonyName": colony.Name,
						"Age":        nodeAge,
					}).Info("Auto-removing stale node")

					err := controller.nodeDB.RemoveNodeByID(node.ID)
					if err != nil {
						log.WithFields(log.Fields{
							"Error":    err,
							"NodeName": node.Name,
						}).Warn("Failed to remove stale node")
					} else {
						cleanedCount++
					}
				}
			}
		}
	}

	if cleanedCount > 0 {
		log.WithFields(log.Fields{"CleanedNodes": cleanedCount}).Info("Completed node cleanup")
	}
}
