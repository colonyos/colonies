package server

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (controller *coloniesController) isLeader() bool {
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

func (controller *coloniesController) tryBecomeLeader() bool {
	var isLeader bool
	controller.leaderMutex.Lock()
	isLeader = controller.isLeader()
	controller.leaderMutex.Unlock()

	return isLeader
}

func (controller *coloniesController) timeoutLoop() {
	for {
		time.Sleep(RELEASE_PERIOD * time.Second)

		controller.stopMutex.Lock()
		if controller.stopFlag {
			return
		}
		controller.stopMutex.Unlock()

		processes, err := controller.db.FindAllRunningProcesses()
		if err != nil {
			continue
		}
		for _, process := range processes {
			if process.ProcessSpec.MaxExecTime == -1 {
				continue
			}
			if time.Now().Unix() > process.ExecDeadline.Unix() {
				if process.Retries >= process.ProcessSpec.MaxRetries && process.ProcessSpec.MaxRetries > -1 {
					err := controller.closeFailed(process.ID, []string{"Maximum execution time limit exceeded"})
					if err != nil {
						log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Debug("Max retries reached, but failed to close process")
						continue
					}
					log.WithFields(log.Fields{"ProcessID": process.ID, "MaxExecTime": process.ProcessSpec.MaxExecTime, "MaxRetries": process.ProcessSpec.MaxRetries}).Debug("Process closed as failed as max retries reached")
					continue
				}

				err := controller.unassignRuntime(process.ID)
				if err != nil {
					log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Error("Failed to unassign process")
				}
				log.WithFields(log.Fields{"ProcessID": process.ID, "MaxExecTime": process.ProcessSpec.MaxExecTime, "MaxRetries": process.ProcessSpec.MaxRetries}).Debug("Process was unassigned as it did not complete in time")
			}
		}

		processes, err = controller.db.FindAllWaitingProcesses()
		if err != nil {
			continue
		}
		for _, process := range processes {
			if process.ProcessSpec.MaxWaitTime == -1 || process.ProcessSpec.MaxWaitTime == 0 {
				continue
			}

			if time.Now().Unix() > process.WaitDeadline.Unix() {
				err := controller.closeFailed(process.ID, []string{"Maximum waiting time limit exceeded"})
				if err != nil {
					log.WithFields(log.Fields{"ProcessID": process.ID, "Error": err}).Debug("Max waiting time reached, but failed to close process")
					continue
				}
				log.WithFields(log.Fields{"ProcessID": process.ID, "MaxWaitTime": process.ProcessSpec.MaxWaitTime}).Debug("Process closed as failed as maximum waiting time limit exceeded")
			}
		}
	}
}

func (controller *coloniesController) masterWorker() {
	for {
		select {
		case msg := <-controller.cmdQueue:
			if msg.stop {
				return
			}
			if msg.handler != nil {
				go msg.handler(msg)
			}
		}
	}
}
