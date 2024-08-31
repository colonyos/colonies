package server

import (
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/scheduler"
	log "github.com/sirupsen/logrus"
)

type command struct {
	stop                   bool
	errorChan              chan error
	process                *core.Process
	count                  int
	colony                 *core.Colony
	colonyName             string
	colonyReplyChan        chan *core.Colony
	coloniesReplyChan      chan []*core.Colony
	processReplyChan       chan *core.Process
	processesReplyChan     chan []*core.Process
	processGraphReplyChan  chan *core.ProcessGraph
	processGraphsReplyChan chan []*core.ProcessGraph
	statisticsReplyChan    chan *core.Statistics
	logsReplyChan          chan []core.Log
	executorReplyChan      chan *core.Executor
	executorsReplyChan     chan []*core.Executor
	attributeReplyChan     chan *core.Attribute
	generatorReplyChan     chan *core.Generator
	generatorsReplyChan    chan []*core.Generator
	cronReplyChan          chan *core.Cron
	cronsReplyChan         chan []*core.Cron
	functionReplyChan      chan *core.Function
	functionsReplyChan     chan []*core.Function
	threaded               bool
	handler                func(cmd *command)
}

type coloniesController struct {
	db               database.Database
	cmdQueue         chan *command
	blockingCmdQueue chan *command
	scheduler        *scheduler.Scheduler
	wsSubCtrl        *wsSubscriptionController
	thisNode         cluster.Node
	clusterConfig    cluster.Config
	clusterManager   *cluster.ClusterManager
	relay            *cluster.Relay
	eventHandler     *eventHandler
	stopFlag         bool
	stopMutex        sync.Mutex
	leaderMutex      sync.Mutex
	leader           bool
	generatorPeriod  int
	cronPeriod       int
	retention        bool
	retentionPolicy  int64
	retentionPeriod  int
}

func createColoniesController(db database.Database,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int,
	retention bool,
	retentionPolicy int64,
	retentionPeriod int) *coloniesController {

	controller := &coloniesController{}
	controller.db = db
	controller.generatorPeriod = generatorPeriod
	controller.cronPeriod = cronPeriod
	controller.retention = retention
	controller.retentionPolicy = retentionPolicy
	controller.retentionPeriod = retentionPeriod

	controller.clusterConfig = clusterConfig
	controller.leader = false
	controller.thisNode = thisNode
	controller.clusterManager = cluster.CreateClusterManager(controller.thisNode, controller.clusterConfig, etcdDataPath)
	controller.relay = controller.clusterManager.Relay()
	controller.eventHandler = createEventHandler(controller.relay)
	controller.wsSubCtrl = createWSSubscriptionController(controller.eventHandler)
	controller.scheduler = scheduler.CreateScheduler(controller.db)

	controller.cmdQueue = make(chan *command)
	controller.blockingCmdQueue = make(chan *command)

	log.Info("Waiting for cluster to be ready...")
	controller.clusterManager.BlockUntilReady()
	log.Info("Cluster is ready")

	controller.tryBecomeLeader()
	go controller.blockingCmdQueueWorker()
	go controller.cmdQueueWorker()
	go controller.timeoutLoop()
	go controller.generatorTriggerLoop()
	go controller.cronTriggerLoop()
	go controller.retentionWorker()

	return controller
}

func (controller *coloniesController) getClusterManager() *cluster.ClusterManager {
	return controller.clusterManager
}

func (controller *coloniesController) getCronPeriod() int {
	return controller.cronPeriod
}

func (controller *coloniesController) getGeneratorPeriod() int {
	return controller.generatorPeriod
}

func (controller *coloniesController) getEventHandler() *eventHandler {
	return controller.eventHandler
}

func (controller *coloniesController) getThisNode() cluster.Node {
	return controller.thisNode
}

func (controller *coloniesController) subscribeProcesses(executorID string, subscription *subscription) error {
	cmd := &command{threaded: false, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			controller.wsSubCtrl.addProcessesSubscriber(executorID, subscription)
			cmd.errorChan <- nil
		}}
	controller.blockingCmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) subscribeProcess(executorID string, subscription *subscription) error {
	cmd := &command{threaded: false, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(subscription.processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if process == nil {
				cmd.errorChan <- errors.New("Invalid process with Id " + subscription.processID)
				return
			}

			controller.wsSubCtrl.addProcessSubscriber(executorID, process, subscription)
			cmd.errorChan <- nil
		}}
	controller.blockingCmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) getColonies() ([]*core.Colony, error) {
	cmd := &command{threaded: true, coloniesReplyChan: make(chan []*core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies, err := controller.db.GetColonies()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.coloniesReplyChan <- colonies
		}}

	controller.cmdQueue <- cmd
	var colonies []*core.Colony
	select {
	case err := <-cmd.errorChan:
		return colonies, err
	case colonies := <-cmd.coloniesReplyChan:
		return colonies, nil
	}
}

func (controller *coloniesController) getColony(colonyName string) (*core.Colony, error) {
	cmd := &command{threaded: true, colonyReplyChan: make(chan *core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colony, err := controller.db.GetColonyByName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.colonyReplyChan <- colony
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case colony := <-cmd.colonyReplyChan:
		return colony, nil
	}
}

func (controller *coloniesController) addColony(colony *core.Colony) (*core.Colony, error) {
	cmd := &command{threaded: true, colonyReplyChan: make(chan *core.Colony, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddColony(colony)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if colony == nil {
				cmd.errorChan <- errors.New("Invalid colony, colony is nil")
				return
			}

			addedColony, err := controller.db.GetColonyByID(colony.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.colonyReplyChan <- addedColony
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedColony := <-cmd.colonyReplyChan:
		return addedColony, nil
	}
}

func (controller *coloniesController) removeColony(colonyName string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.RemoveColonyByName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) addExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error) {
	cmd := &command{threaded: true, executorReplyChan: make(chan *core.Executor, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if executor == nil {
				cmd.errorChan <- errors.New("Invalid executor, executor is nil")
				return
			}
			executorFromDB, err := controller.db.GetExecutorByName(executor.ColonyName, executor.Name)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if allowExecutorReregister {
				if executorFromDB != nil {
					err := controller.db.RemoveExecutorByName(executor.ColonyName, executorFromDB.Name)
					if err != nil {
						cmd.errorChan <- err
						return
					}
				}
			} else {
				if executorFromDB != nil {
					cmd.errorChan <- errors.New("Executor with name <" + executorFromDB.Name + "> in Colony <" + executorFromDB.ColonyName + "> already exists")
					return
				}

			}
			err = controller.db.AddExecutor(executor)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedExecutor, err := controller.db.GetExecutorByID(executor.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.executorReplyChan <- addedExecutor
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedExecutor := <-cmd.executorReplyChan:
		return addedExecutor, nil
	}
}

func (controller *coloniesController) getExecutor(executorID string) (*core.Executor, error) {
	cmd := &command{threaded: true, executorReplyChan: make(chan *core.Executor),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.executorReplyChan <- executor
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case executor := <-cmd.executorReplyChan:
		return executor, nil
	}
}

func (controller *coloniesController) getExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	cmd := &command{threaded: true, executorsReplyChan: make(chan []*core.Executor),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executors, err := controller.db.GetExecutorsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.executorsReplyChan <- executors
		}}

	controller.cmdQueue <- cmd
	var executors []*core.Executor
	select {
	case err := <-cmd.errorChan:
		return executors, err
	case executors := <-cmd.executorsReplyChan:
		return executors, nil
	}
}

func (controller *coloniesController) addProcessToDB(process *core.Process) (*core.Process, error) {
	if process == nil {
		return nil, errors.New("Invalid process, process is nil")
	}

	err := controller.db.AddProcess(process)
	if err != nil {
		return nil, err
	}

	addedProcess, err := controller.db.GetProcessByID(process.ID)
	if err != nil {
		return nil, err
	}

	return addedProcess, nil
}

func (controller *coloniesController) addProcess(process *core.Process) (*core.Process, error) {
	cmd := &command{threaded: true, processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			addedProcess, err := controller.addProcessToDB(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if !addedProcess.WaitForParents {
				controller.eventHandler.signal(addedProcess)
			}
			cmd.processReplyChan <- addedProcess
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case process := <-cmd.processReplyChan:
		return process, nil
	}
}

func (controller *coloniesController) addChild(
	processGraphID string,
	parentProcessID string,
	childProcessID string,
	process *core.Process,
	executorID string,
	insert bool) (*core.Process, error) {
	cmd := &command{threaded: false, processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process.WaitForParents = true
			parentProcess, err := controller.db.GetProcessByID(parentProcessID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if parentProcess.State != core.RUNNING {
				cmd.errorChan <- errors.New("Process with Id " + parentProcessID + " is not running")
				return
			}

			if parentProcess.AssignedExecutorID != executorID {
				cmd.errorChan <- errors.New("Process with Id " + parentProcessID + " is not assigned to executor with Id " + executorID)
				return
			}

			if parentProcess.ProcessGraphID == "" {
				cmd.errorChan <- errors.New("Process with Id " + parentProcessID + " does not belong to a processgraph")
				return
			}

			process.Parents = []string{parentProcess.ID}
			process.ProcessGraphID = processGraphID
			addedProcess, err := controller.addProcessToDB(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if insert {
				parentsChildren := parentProcess.Children
				controller.db.SetChildren(process.ID, parentsChildren)
				controller.db.SetChildren(parentProcessID, []string{process.ID})
				for _, parentsChildID := range parentsChildren {
					parentChild, err := controller.db.GetProcessByID(parentsChildID)
					if err != nil {
						cmd.errorChan <- err
						return
					}
					parentChildParents := parentChild.Parents
					var newParents []string
					for _, p := range parentChildParents {
						if p != parentProcessID {
							newParents = append(newParents, p)
						}
					}
					newParents = append(newParents, process.ID)
					controller.db.SetParents(parentsChildID, newParents)
				}
			} else {
				parentsChildren := parentProcess.Children
				parentsChildren = append(parentsChildren, process.ID)
				controller.db.SetChildren(parentProcessID, parentsChildren)
				if childProcessID != "" {
					controller.db.SetChildren(process.ID, []string{childProcessID})
					childProcess, err := controller.db.GetProcessByID(childProcessID)
					if err != nil {
						cmd.errorChan <- err
						return
					}
					newParents := childProcess.Parents
					newParents = append(newParents, process.ID)
					controller.db.SetParents(childProcessID, newParents)
				}
			}

			if !addedProcess.WaitForParents {
				controller.eventHandler.signal(addedProcess)
			}

			updatedProcess, err := controller.db.GetProcessByID(addedProcess.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.processReplyChan <- updatedProcess
		}}

	controller.blockingCmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case process := <-cmd.processReplyChan:
		return process, nil
	}
}

func (controller *coloniesController) getProcess(processID string) (*core.Process, error) {
	cmd := &command{threaded: true, processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.processReplyChan <- process
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case process := <-cmd.processReplyChan:
		return process, nil
	}
}

func (controller *coloniesController) findProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	cmd := &command{threaded: true, processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			var err error
			if executorID == "" {
				processes, err = controller.db.FindProcessesByColonyName(colonyName, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			} else {
				processes, err = controller.db.FindProcessesByExecutorID(colonyName, executorID, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.processesReplyChan <- processes
		}}

	controller.cmdQueue <- cmd
	var processes []*core.Process
	select {
	case err := <-cmd.errorChan:
		return processes, err
	case processes := <-cmd.processesReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) updateProcessGraph(graph *core.ProcessGraph) error {
	graph.SetStorage(controller.db)
	return graph.UpdateProcessIDs()
}

func (controller *coloniesController) createProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error) {
	processgraph, err := core.CreateProcessGraph(workflowSpec.ColonyName)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create processgraph")
		return nil, err
	}

	initiatorName, err := resolveInitiator(workflowSpec.ColonyName, recoveredID, controller.db)
	if err != nil {
		return nil, err
	}

	// Create all processes
	processMap := make(map[string]*core.Process)
	var processIDs []string
	for _, funcSpec := range workflowSpec.FunctionSpecs {
		if funcSpec.NodeName == "" {
			return nil, errors.New("Nodename is missing, check function spec for errors")
		}

		if funcSpec.MaxExecTime == 0 {
			log.WithFields(log.Fields{"NodeName": funcSpec.NodeName}).Debug("MaxExecTime was set to 0, resetting to -1")
			funcSpec.MaxExecTime = -1
		}
		process := core.CreateProcess(&funcSpec)
		log.WithFields(log.Fields{"ProcessId": process.ID, "MaxExecTime": process.FunctionSpec.MaxExecTime, "MaxRetries": process.FunctionSpec.MaxRetries}).Debug("Creating new process")
		if len(funcSpec.Conditions.Dependencies) == 0 {
			// The process is a root process, let it start immediately
			process.WaitForParents = false
			if len(args) > 0 {
				// NOTE, overwrite the args, this will only happen when using Generators
				argsif := make([]interface{}, len(args))
				for i, k := range args {
					argsif[i] = k
				}
				process.FunctionSpec.Args = argsif

				kwargsif := make(map[string]interface{}, len(kwargs))
				for v, k := range kwargs {
					kwargsif[v] = k
				}
				process.FunctionSpec.KwArgs = kwargsif
			}

			if len(rootInput) > 0 {
				rootInputIf := make([]interface{}, len(rootInput))
				for k, v := range rootInput {
					rootInputIf[k] = v
				}
				process.Input = rootInputIf
				if err != nil {
					return nil, err
				}
			}

			processgraph.AddRoot(process.ID)
		} else {
			// The process has to wait for its parents
			process.WaitForParents = true
		}
		processIDs = append(processIDs, process.ID)
		process.ProcessGraphID = processgraph.ID
		process.FunctionSpec.Conditions.ColonyName = workflowSpec.ColonyName

		process.InitiatorID = recoveredID
		process.InitiatorName = initiatorName

		_, exists := processMap[process.FunctionSpec.NodeName]
		if exists {
			return nil, errors.New("Duplicate nodename: " + process.FunctionSpec.NodeName)
		}

		processMap[process.FunctionSpec.NodeName] = process
	}

	processgraph.ProcessIDs = processIDs

	processgraph.InitiatorID = recoveredID
	processgraph.InitiatorName = initiatorName

	err = controller.db.AddProcessGraph(processgraph)
	if err != nil {
		msg := "Failed to create processgraph, failed to add processgraph"
		log.WithFields(log.Fields{"Error": err}).Error(msg)
		return nil, errors.New(msg)
	}

	log.WithFields(log.Fields{"ProcessGraphId": processgraph.ID}).Debug("Submitting workflow")

	// Create dependencies
	for _, process := range processMap {
		for _, dependsOn := range process.FunctionSpec.Conditions.Dependencies {
			parentProcess := processMap[dependsOn]
			if parentProcess == nil {
				msg := "Failed to submit workflow, invalid dependencies, are you depending on a nodename that does not exits?"
				log.WithFields(log.Fields{"Error": err}).Error(msg)
				return nil, errors.New(msg)
			}
			process.AddParent(parentProcess.ID)
			parentProcess.AddChild(process.ID)
		}
	}

	// Now, start all processes
	for _, process := range processMap {
		// This function is called from the controller, so it OK to use the database layer directly, in fact
		// we will cause a deadlock if we call controller.addProcess
		log.WithFields(log.Fields{"ProcessId": process.ID}).Debug("Submitting process part of processgraph")
		addedProcess, err := controller.addProcessToDB(process)
		if err != nil {
			msg := "Failed to submit workflow, failed to add process"
			log.WithFields(log.Fields{"Error": err}).Error(msg)
			return nil, errors.New(msg)
		}
		if !addedProcess.WaitForParents {
			controller.eventHandler.signal(addedProcess)
		}
	}

	return processgraph, nil
}

func (controller *coloniesController) submitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error) {
	cmd := &command{threaded: false, processGraphReplyChan: make(chan *core.ProcessGraph, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			addedProcessGraph, err := controller.createProcessGraph(workflowSpec, make([]interface{}, 0), make(map[string]interface{}), make([]interface{}, 0), recoveredID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.processGraphReplyChan <- addedProcessGraph
		}}

	controller.blockingCmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processGraph := <-cmd.processGraphReplyChan:
		return processGraph, nil
	}
}

func (controller *coloniesController) getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphReplyChan: make(chan *core.ProcessGraph, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			graph, err := controller.db.GetProcessGraphByID(processGraphID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if graph == nil {
				cmd.processGraphReplyChan <- nil
				return
			}
			err = controller.updateProcessGraph(graph)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.processGraphReplyChan <- graph
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case graph := <-cmd.processGraphReplyChan:
		return graph, nil
	}
}

func (controller *coloniesController) findWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindWaitingProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.updateProcessGraph(graph)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}

			cmd.processGraphsReplyChan <- graphs
		}}

	controller.cmdQueue <- cmd
	var graphs []*core.ProcessGraph
	select {
	case err := <-cmd.errorChan:
		return graphs, err
	case graphs := <-cmd.processGraphsReplyChan:
		return graphs, nil
	}
}

func (controller *coloniesController) findRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindRunningProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.updateProcessGraph(graph)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.processGraphsReplyChan <- graphs
		}}

	controller.cmdQueue <- cmd
	var graphs []*core.ProcessGraph
	select {
	case err := <-cmd.errorChan:
		return graphs, err
	case graphs := <-cmd.processGraphsReplyChan:
		return graphs, nil
	}
}

func (controller *coloniesController) findSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindSuccessfulProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.updateProcessGraph(graph)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.processGraphsReplyChan <- graphs
		}}

	controller.cmdQueue <- cmd
	var graphs []*core.ProcessGraph
	select {
	case err := <-cmd.errorChan:
		return graphs, err
	case graphs := <-cmd.processGraphsReplyChan:
		return graphs, nil
	}
}

func (controller *coloniesController) findFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindFailedProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.updateProcessGraph(graph)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.processGraphsReplyChan <- graphs
		}}

	controller.cmdQueue <- cmd
	var graphs []*core.ProcessGraph
	select {
	case err := <-cmd.errorChan:
		return graphs, err
	case graphs := <-cmd.processGraphsReplyChan:
		return graphs, nil
	}
}

func (controller *coloniesController) removeProcess(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.RemoveProcessByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) removeAllProcesses(colonyName string, state int) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			switch state {
			case core.WAITING:
				cmd.errorChan <- controller.db.RemoveAllWaitingProcessesByColonyName(colonyName)
			case core.RUNNING:
				cmd.errorChan <- errors.New("Not possible to remove running processes")
			case core.SUCCESS:
				cmd.errorChan <- controller.db.RemoveAllSuccessfulProcessesByColonyName(colonyName)
			case core.FAILED:
				cmd.errorChan <- controller.db.RemoveAllFailedProcessesByColonyName(colonyName)
			case core.NOTSET:
				cmd.errorChan <- controller.db.RemoveAllProcessesByColonyName(colonyName)
			default:
				cmd.errorChan <- errors.New("Invalid state when deleting all processes")
			}
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) removeProcessGraph(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.RemoveProcessGraphByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) removeAllProcessGraphs(colonyName string, state int) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			switch state {
			case core.WAITING:
				cmd.errorChan <- controller.db.RemoveAllWaitingProcessGraphsByColonyName(colonyName)
			case core.RUNNING:
				cmd.errorChan <- errors.New("not possible to remove running processgraphs")
			case core.SUCCESS:
				cmd.errorChan <- controller.db.RemoveAllSuccessfulProcessGraphsByColonyName(colonyName)
			case core.FAILED:
				cmd.errorChan <- controller.db.RemoveAllFailedProcessGraphsByColonyName(colonyName)
			case core.NOTSET:
				cmd.errorChan <- controller.db.RemoveAllProcessGraphsByColonyName(colonyName)
			default:
				cmd.errorChan <- errors.New("invalid state when deleting all processgraphs")
			}
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) setOutput(processID string, output []interface{}) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if len(output) > 0 {
				err := controller.db.SetOutput(processID, output)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) closeSuccessful(processID string, executorID string, output []interface{}) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if len(output) > 0 {
				err = controller.db.SetOutput(processID, output)
			}

			waitingTime, processingTime, err := controller.db.MarkSuccessful(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if process.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": process.ProcessGraphID}).Debug("Resolving processgraph (close successful)")
				processGraph, err := controller.db.GetProcessGraphByID(process.ProcessGraphID)
				if err != nil {
					cmd.errorChan <- err
					return
				}
				processGraph.SetStorage(controller.db)
				err = processGraph.Resolve()
				if err != nil {
					err2 := controller.handleDefunctProcessgraph(processGraph.ID, process.ID, err)
					if err2 != nil {
						log.Error(err2)
						cmd.errorChan <- err2
						return
					}

					log.Error(err)
					cmd.errorChan <- err
					return
				}

				// This is process is now closed. This means that children processes can now execute,
				// assuming all their parents are closed successfully
				err = controller.notifyChildren(process)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}

			process.State = core.SUCCESS

			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			function, err := controller.db.GetFunctionsByExecutorAndName(process.FunctionSpec.Conditions.ColonyName, executor.Name, process.FunctionSpec.FuncName)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if function != nil {
				process, err = controller.db.GetProcessByID(processID)
				if err != nil {
					cmd.errorChan <- err
					return
				}
				minWaitTime := 0.0
				if function.MinWaitTime > 0 {
					minWaitTime = math.Min(function.MinWaitTime, waitingTime)
				} else {
					minWaitTime = waitingTime
				}
				maxWaitTime := 0.0
				if function.MaxWaitTime > 0 {
					maxWaitTime = math.Max(function.MaxWaitTime, waitingTime)
				} else {
					maxWaitTime = waitingTime
				}
				minExecTime := 0.0
				if function.MinExecTime > 0 {
					minExecTime = math.Min(function.MinExecTime, processingTime)
				} else {
					minExecTime = processingTime
				}
				maxExecTime := 0.0
				if function.MaxExecTime > 0 {
					maxExecTime = math.Max(function.MaxExecTime, processingTime)
				} else {
					maxExecTime = processingTime
				}
				avgWaitTime := 0.0
				if function.AvgWaitTime > 0.0 {
					avgWaitTime = (function.AvgWaitTime + waitingTime) / 2
				} else {
					avgWaitTime = waitingTime
				}
				avgExecTime := 0.0
				if function.AvgExecTime > 0.0 {
					avgExecTime = (function.AvgExecTime + processingTime) / 2
				} else {
					avgExecTime = processingTime
				}

				controller.db.UpdateFunctionStats(
					process.FunctionSpec.Conditions.ColonyName,
					function.ExecutorName,
					function.FuncName,
					function.Counter+1,
					minWaitTime,
					maxWaitTime,
					minExecTime,
					maxExecTime,
					avgWaitTime,
					avgExecTime)
			}

			controller.eventHandler.signal(process)
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) notifyChildren(process *core.Process) error {
	// First check if parent processes are completed
	counter := 0
	for _, parentProcessID := range process.Parents {
		parentProcess, err := controller.db.GetProcessByID(parentProcessID)
		if err != nil {
			return err
		}
		if parentProcess.State == core.SUCCESS {
			counter++
		}
	}

	// Notiify children of parents are completed
	if counter == len(process.Parents) {
		for _, childProcessID := range process.Children {
			childProcess, err := controller.db.GetProcessByID(childProcessID)
			if err != nil {
				return err
			}
			controller.eventHandler.signal(childProcess)
		}
	}

	return nil
}

func (controller *coloniesController) closeFailed(processID string, errs []string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.MarkFailed(processID, errs)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if process.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": process.ProcessGraphID}).Debug("Resolving processgraph (close failed)")
				processGraph, err := controller.db.GetProcessGraphByID(process.ProcessGraphID)
				if err != nil {
					cmd.errorChan <- err
					return
				}
				processGraph.SetStorage(controller.db)
				err = processGraph.Resolve()
				if err != nil {
					err2 := controller.handleDefunctProcessgraph(processGraph.ID, process.ID, err)
					if err2 != nil {
						log.Error(err2)
						cmd.errorChan <- err2
						return
					}

					log.Error(err)
					cmd.errorChan <- err
					return
				}

			}

			process.State = core.FAILED

			controller.eventHandler.signal(process)
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) handleDefunctProcessgraph(processGraphID string, processID string, err error) error {
	err2 := controller.db.MarkFailed(processID, []string{err.Error()})
	if err2 != nil {
		return err2
	}
	err2 = controller.db.SetProcessGraphState(processGraphID, core.FAILED)
	if err2 != nil {
		return err2
	}

	return nil
}

func (controller *coloniesController) assign(executorID string, colonyName string, cpu int64, mem int64) (*core.Process, error) {
	cmd := &command{threaded: false, processReplyChan: make(chan *core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if executor == nil {
				cmd.errorChan <- errors.New("Executor with Id <" + executorID + "> could not be found")
				return
			}

			err = controller.db.MarkAlive(executor)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			selectedProcess, err := controller.scheduler.Select(colonyName, executor, cpu, mem)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			err = controller.db.Assign(executorID, selectedProcess)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if selectedProcess.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": selectedProcess.ProcessGraphID}).Debug("Resolving processgraph (assigned)")
				processGraph, err := controller.db.GetProcessGraphByID(selectedProcess.ProcessGraphID)
				if err != nil {
					log.Error(err)
					cmd.errorChan <- err
					return
				}
				if processGraph == nil {
					errMsg := "Failed to resolve processgraph from controller, processGraph is nil (should not be)"
					log.Error(errMsg)
					cmd.errorChan <- errors.New(errMsg)
				}
				processGraph.SetStorage(controller.db)

				// One Colonies server might have added a processgraph, and another colonies directly get an assign request
				// This means that all processes part of the graph might not yet have been added, consequently the
				// processgraph.Resolve() call might fail.
				// The solution is to retry a couple of times.
				maxRetries := 10
				timeBetweenRetries := 500 * time.Millisecond // We will wait what max 10 * 0.5 = 5 seconds
				retries := 0

				for {
					if retries >= maxRetries {
						err2 := controller.handleDefunctProcessgraph(processGraph.ID, selectedProcess.ID, err)
						if err2 != nil {
							log.Error(err2)
							cmd.errorChan <- err2
							return
						}

						log.Error(err)
						cmd.errorChan <- err
						return
					}
					err = processGraph.Resolve()
					if err != nil {
						retries++
						time.Sleep(timeBetweenRetries)
						continue
					} else {
						break
					}
				}

				// Now, we need to collect the output from the parents and use ut as our input
				var output []interface{}
				for _, parentID := range selectedProcess.Parents {
					parentProcess, err := controller.db.GetProcessByID(parentID)
					if err != nil {
						log.Error(err)
						cmd.errorChan <- err
						return
					}
					output = append(output, parentProcess.Output...)
				}
				if len(selectedProcess.Parents) > 0 {
					controller.db.SetInput(selectedProcess.ID, output)
					selectedProcess.Input = output
				}
			}

			cmd.processReplyChan <- selectedProcess
		}}

	controller.blockingCmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processes := <-cmd.processReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) unassignExecutor(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.errorChan <- controller.db.Unassign(process)
			controller.eventHandler.signal(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) resetProcess(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.errorChan <- controller.db.ResetProcess(process)
			controller.eventHandler.signal(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) getColonyStatistics(colonyName string) (*core.Statistics, error) {
	cmd := &command{threaded: true, statisticsReplyChan: make(chan *core.Statistics),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies := 1
			executors, err := controller.db.CountExecutorsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingProcesses, err := controller.db.CountWaitingProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningProcesses, err := controller.db.CountRunningProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successProcesses, err := controller.db.CountSuccessfulProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedProcesses, err := controller.db.CountFailedProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingWorkflows, err := controller.db.CountWaitingProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningWorkflows, err := controller.db.CountRunningProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successWorkflows, err := controller.db.CountSuccessfulProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedWorkflows, err := controller.db.CountFailedProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.statisticsReplyChan <- core.CreateStatistics(colonies,
				executors,
				waitingProcesses,
				runningProcesses,
				successProcesses,
				failedProcesses,
				waitingWorkflows,
				runningWorkflows,
				successWorkflows,
				failedWorkflows)
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case stat := <-cmd.statisticsReplyChan:
		return stat, nil
	}
}

func (controller *coloniesController) getStatistics() (*core.Statistics, error) {
	cmd := &command{threaded: true, statisticsReplyChan: make(chan *core.Statistics),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies, err := controller.db.CountColonies()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			executors, err := controller.db.CountExecutors()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingProcesses, err := controller.db.CountWaitingProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningProcesses, err := controller.db.CountRunningProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successProcesses, err := controller.db.CountSuccessfulProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedProcesses, err := controller.db.CountFailedProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingWorkflows, err := controller.db.CountWaitingProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningWorkflows, err := controller.db.CountRunningProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successWorkflows, err := controller.db.CountSuccessfulProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedWorkflows, err := controller.db.CountFailedProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.statisticsReplyChan <- core.CreateStatistics(colonies,
				executors,
				waitingProcesses,
				runningProcesses,
				successProcesses,
				failedProcesses,
				waitingWorkflows,
				runningWorkflows,
				successWorkflows,
				failedWorkflows)
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case stat := <-cmd.statisticsReplyChan:
		return stat, nil
	}
}

func (controller *coloniesController) addAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	cmd := &command{threaded: true, attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddAttribute(*attribute)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedAttribute, err := controller.db.GetAttributeByID(attribute.ID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.attributeReplyChan <- &addedAttribute
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedAttribute := <-cmd.attributeReplyChan:
		return addedAttribute, nil
	}
}

func (controller *coloniesController) getAttribute(attributeID string) (*core.Attribute, error) {
	cmd := &command{threaded: true, attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			attribute, err := controller.db.GetAttributeByID(attributeID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.attributeReplyChan <- &attribute
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case attribute := <-cmd.attributeReplyChan:
		return attribute, nil
	}
}

func (controller *coloniesController) addFunction(function *core.Function) (*core.Function, error) {
	cmd := &command{threaded: true, functionReplyChan: make(chan *core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddFunction(function)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedFunction, err := controller.db.GetFunctionByID(function.FunctionID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.functionReplyChan <- addedFunction
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case addedFunction := <-cmd.functionReplyChan:
		return addedFunction, nil
	}
}

func (controller *coloniesController) getFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) {
	cmd := &command{threaded: true, functionsReplyChan: make(chan []*core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			functions, err := controller.db.GetFunctionsByExecutorName(colonyName, executorName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.functionsReplyChan <- functions
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case functions := <-cmd.functionsReplyChan:
		return functions, nil
	}
}

func (controller *coloniesController) getFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	cmd := &command{threaded: true, functionsReplyChan: make(chan []*core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			functions, err := controller.db.GetFunctionsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.functionsReplyChan <- functions
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case functions := <-cmd.functionsReplyChan:
		return functions, nil
	}
}

func (controller *coloniesController) getFunctionByID(functionID string) (*core.Function, error) {
	cmd := &command{threaded: true, functionReplyChan: make(chan *core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			function, err := controller.db.GetFunctionByID(functionID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.functionReplyChan <- function
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case function := <-cmd.functionReplyChan:
		return function, nil
	}
}

func (controller *coloniesController) removeFunction(functionID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.RemoveFunctionByID(functionID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) resetDatabase() error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.Drop()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			err = controller.db.Initialize()
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) stop() {
	controller.stopMutex.Lock()
	controller.stopFlag = true
	controller.stopMutex.Unlock()
	controller.cmdQueue <- &command{stop: true}
	controller.eventHandler.stop()
	controller.clusterManager.Shutdown()
}
