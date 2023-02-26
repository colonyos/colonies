package server

import (
	"errors"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/planner"
	"github.com/colonyos/colonies/pkg/planner/basic"
	log "github.com/sirupsen/logrus"
)

type command struct {
	stop                   bool
	errorChan              chan error
	process                *core.Process
	count                  int
	colony                 *core.Colony
	colonyID               string
	colonyReplyChan        chan *core.Colony
	coloniesReplyChan      chan []*core.Colony
	processReplyChan       chan *core.Process
	processesReplyChan     chan []*core.Process
	processGraphReplyChan  chan *core.ProcessGraph
	processGraphsReplyChan chan []*core.ProcessGraph
	statisticsReplyChan    chan *core.Statistics
	executorReplyChan      chan *core.Executor
	executorsReplyChan     chan []*core.Executor
	attributeReplyChan     chan *core.Attribute
	generatorReplyChan     chan *core.Generator
	generatorsReplyChan    chan []*core.Generator
	cronReplyChan          chan *core.Cron
	cronsReplyChan         chan []*core.Cron
	functionReplyChan      chan *core.Function
	functionsReplyChan     chan []*core.Function
	handler                func(cmd *command)
}

type coloniesController struct {
	db              database.Database
	cmdQueue        chan *command
	planner         planner.Planner
	wsSubCtrl       *wsSubscriptionController
	relayServer     *cluster.RelayServer
	eventHandler    *eventHandler
	stopFlag        bool
	stopMutex       sync.Mutex
	leaderMutex     sync.Mutex
	assignMutex     sync.Mutex
	thisNode        cluster.Node
	clusterConfig   cluster.Config
	etcdServer      *cluster.EtcdServer
	leader          bool
	generatorPeriod int
	cronPeriod      int
}

func createColoniesController(db database.Database,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int) *coloniesController {

	controller := &coloniesController{}
	controller.db = db
	controller.thisNode = thisNode
	controller.clusterConfig = clusterConfig
	controller.etcdServer = cluster.CreateEtcdServer(controller.thisNode, controller.clusterConfig, etcdDataPath)
	controller.etcdServer.Start()
	controller.etcdServer.WaitToStart()
	controller.leader = false
	controller.generatorPeriod = generatorPeriod
	controller.cronPeriod = cronPeriod

	controller.relayServer = cluster.CreateRelayServer(controller.thisNode, controller.clusterConfig)
	controller.eventHandler = createEventHandler(controller.relayServer)
	controller.wsSubCtrl = createWSSubscriptionController(controller.eventHandler)
	controller.planner = basic.CreatePlanner()

	controller.cmdQueue = make(chan *command)

	controller.tryBecomeLeader()
	go controller.masterWorker()
	go controller.timeoutLoop()
	go controller.generatorTriggerLoop()
	go controller.cronTriggerLoop()

	return controller
}

func (controller *coloniesController) subscribeProcesses(executorID string, subscription *subscription) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			controller.wsSubCtrl.addProcessesSubscriber(executorID, subscription)
			cmd.errorChan <- nil
		}}
	controller.cmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) subscribeProcess(executorID string, subscription *subscription) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.db.GetProcessByID(subscription.processID)
			if err != nil {
				cmd.errorChan <- err
			}

			controller.wsSubCtrl.addProcessSubscriber(executorID, process, subscription)
			cmd.errorChan <- nil
		}}
	controller.cmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *coloniesController) getColonies() ([]*core.Colony, error) {
	cmd := &command{coloniesReplyChan: make(chan []*core.Colony),
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

func (controller *coloniesController) getColony(colonyID string) (*core.Colony, error) {
	cmd := &command{colonyReplyChan: make(chan *core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colony, err := controller.db.GetColonyByID(colonyID)
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
	cmd := &command{colonyReplyChan: make(chan *core.Colony, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddColony(colony)
			if err != nil {
				cmd.errorChan <- err
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

func (controller *coloniesController) deleteColony(colonyID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteColonyByID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) addExecutor(executor *core.Executor) (*core.Executor, error) {
	cmd := &command{executorReplyChan: make(chan *core.Executor, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.AddExecutor(executor)
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
	cmd := &command{executorReplyChan: make(chan *core.Executor),
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

func (controller *coloniesController) getExecutorByColonyID(colonyID string) ([]*core.Executor, error) {
	cmd := &command{executorsReplyChan: make(chan []*core.Executor),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executors, err := controller.db.GetExecutorsByColonyID(colonyID)
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

	return executors, nil
}

func (controller *coloniesController) approveExecutor(executorID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.ApproveExecutor(executor)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) rejectExecutor(executorID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- controller.db.RejectExecutor(executor)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteExecutor(executorID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteExecutorByID(executorID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) addProcessToDB(process *core.Process) (*core.Process, error) {
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
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
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

func (controller *coloniesController) addChild(processGraphID string, parentProcessID string, process *core.Process, executorID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
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

			parentsChildren := parentProcess.Children
			parentsChildren = append(parentsChildren, process.ID)
			controller.db.SetChildren(parentProcessID, parentsChildren)

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

func (controller *coloniesController) getProcess(processID string) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process, 1),
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

func (controller *coloniesController) findProcessHistory(colonyID string, executorID string, seconds int, state int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			var err error
			if executorID == "" {
				processes, err = controller.db.FindProcessesByColonyID(colonyID, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			} else {
				processes, err = controller.db.FindProcessesByExecutorID(colonyID, executorID, seconds, state)
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

func (controller *coloniesController) findPrioritizedProcesses(executorID string, colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindWaitingProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			prioritizedProcesses := controller.planner.Prioritize(executorID, processes, count, false)
			cmd.processesReplyChan <- prioritizedProcesses
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

func (controller *coloniesController) findWaitingProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindWaitingProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
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

func (controller *coloniesController) findRunningProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindRunningProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
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

func (controller *coloniesController) findSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindSuccessfulProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
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

func (controller *coloniesController) findFailedProcesses(colonyID string, count int) ([]*core.Process, error) {
	cmd := &command{processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			processes, err := controller.db.FindFailedProcesses(colonyID, count)
			if err != nil {
				cmd.errorChan <- err
				return
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

func (controller *coloniesController) createProcessGraph(workflowSpec *core.WorkflowSpec, args []string, rootInput []string) (*core.ProcessGraph, error) {
	processgraph, err := core.CreateProcessGraph(workflowSpec.ColonyID)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create processgraph")
		return nil, err
	}

	// Create all processes
	processMap := make(map[string]*core.Process)
	var rootProcesses []*core.Process
	for _, funcSpec := range workflowSpec.FunctionSpecs {
		if funcSpec.MaxExecTime == 0 {
			log.WithFields(log.Fields{"Name": funcSpec.Name}).Debug("MaxExecTime was set to 0, resetting to -1")
			funcSpec.MaxExecTime = -1
		}
		process := core.CreateProcess(&funcSpec)
		log.WithFields(log.Fields{"ProcessId": process.ID, "MaxExecTime": process.FunctionSpec.MaxExecTime, "MaxRetries": process.FunctionSpec.MaxRetries}).Debug("Creating new process")
		if len(funcSpec.Conditions.Dependencies) == 0 {
			// The process is a root process, let it start immediately
			process.WaitForParents = false
			if len(args) > 0 {
				// NOTE, overwrite the args, this will only happen when using Generators
				process.FunctionSpec.Args = args
			}
			if len(rootInput) > 0 {
				process.Input = rootInput
				if err != nil {
					return nil, err
				}
			}
			rootProcesses = append(rootProcesses, process)

			processgraph.AddRoot(process.ID)
		} else {
			// The process has to wait for its parents
			process.WaitForParents = true
		}
		process.ProcessGraphID = processgraph.ID
		process.FunctionSpec.Conditions.ColonyID = workflowSpec.ColonyID
		processMap[process.FunctionSpec.Name] = process
	}

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
				msg := "Failed to submit workflow, invalid dependencies, are you depending on a process spec name that does not exits?"
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

func (controller *coloniesController) submitWorkflowSpec(workflowSpec *core.WorkflowSpec) (*core.ProcessGraph, error) {
	cmd := &command{processGraphReplyChan: make(chan *core.ProcessGraph, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			addedProcessGraph, err := controller.createProcessGraph(workflowSpec, []string{}, []string{})
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.processGraphReplyChan <- addedProcessGraph
		}}

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processGraph := <-cmd.processGraphReplyChan:
		return processGraph, nil
	}
}

func (controller *coloniesController) getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	cmd := &command{processGraphReplyChan: make(chan *core.ProcessGraph, 1),
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

func (controller *coloniesController) findWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindWaitingProcessGraphs(colonyID, count)
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

func (controller *coloniesController) findRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindRunningProcessGraphs(colonyID, count)
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

func (controller *coloniesController) findSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindSuccessfulProcessGraphs(colonyID, count)
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

func (controller *coloniesController) findFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.db.FindFailedProcessGraphs(colonyID, count)
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

func (controller *coloniesController) deleteProcess(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteProcessByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteAllProcesses(colonyID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteAllProcessesByColonyID(colonyID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteProcessGraph(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteProcessGraphByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) deleteAllProcessGraphs(colonyID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteAllProcessGraphsByColonyID(colonyID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *coloniesController) closeSuccessful(processID string, executorID string, output []string) error {
	cmd := &command{errorChan: make(chan error, 1),
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

			function, err := controller.db.GetFunctionsByExecutorIDAndName(executorID, process.FunctionSpec.Name)
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
					maxWaitTime = math.Max(function.MinWaitTime, waitingTime)
				} else {
					maxWaitTime = waitingTime
				}
				minExecTime := 0.0
				if function.MinExecTime > 0 {
					minExecTime = math.Min(function.MinWaitTime, processingTime)
				} else {
					minExecTime = processingTime
				}
				maxExecTime := 0.0
				if function.MaxExecTime > 0 {
					maxExecTime = math.Max(function.MinExecTime, processingTime)
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

				controller.db.UpdateFunctionStats(function.ExecutorID,
					function.Name,
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
	cmd := &command{errorChan: make(chan error, 1),
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

func (controller *coloniesController) assign(executorID string, colonyID string, latest bool) (*core.Process, error) {
	cmd := &command{processReplyChan: make(chan *core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			controller.assignMutex.Lock()

			executor, err := controller.db.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				controller.assignMutex.Unlock()
				return
			}
			if executor == nil {
				cmd.errorChan <- errors.New("Executor with Id <" + executorID + "> could not be found")
				controller.assignMutex.Unlock()
				return
			}

			err = controller.db.MarkAlive(executor)
			if err != nil {
				cmd.errorChan <- err
				controller.assignMutex.Unlock()
				return
			}

			var processes []*core.Process
			processes, err = controller.db.FindUnassignedProcesses(colonyID, executorID, executor.Type, 10, latest)
			if err != nil {
				cmd.errorChan <- err
				controller.assignMutex.Unlock()
				return
			}

			selectedProcess, err := controller.planner.Select(executorID, processes, latest)
			if err != nil {
				cmd.errorChan <- err
				controller.assignMutex.Unlock()
				return
			}

			err = controller.db.Assign(executorID, selectedProcess)
			if err != nil {
				cmd.errorChan <- err
				controller.assignMutex.Unlock()
				return
			}

			controller.assignMutex.Unlock()

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
				err = processGraph.Resolve()
				if err != nil {
					log.Error(err)
					cmd.errorChan <- err
					return
				}

				// Now, we need to collect the output from the parents and use ut as our input
				output := []string{}
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

	controller.cmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case processes := <-cmd.processReplyChan:
		return processes, nil
	}
}

func (controller *coloniesController) unassignExecutor(processID string) error {
	cmd := &command{errorChan: make(chan error, 1),
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
	cmd := &command{errorChan: make(chan error, 1),
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

func (controller *coloniesController) getColonyStatistics(colonyID string) (*core.Statistics, error) {
	cmd := &command{statisticsReplyChan: make(chan *core.Statistics),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies := 1
			executors, err := controller.db.CountExecutorsByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingProcesses, err := controller.db.CountWaitingProcessesByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningProcesses, err := controller.db.CountRunningProcessesByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successProcesses, err := controller.db.CountSuccessfulProcessesByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedProcesses, err := controller.db.CountFailedProcessesByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingWorkflows, err := controller.db.CountWaitingProcessGraphsByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningWorkflows, err := controller.db.CountRunningProcessGraphsByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successWorkflows, err := controller.db.CountSuccessfulProcessGraphsByColonyID(colonyID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedWorkflows, err := controller.db.CountFailedProcessGraphsByColonyID(colonyID)
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
	cmd := &command{statisticsReplyChan: make(chan *core.Statistics),
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
	cmd := &command{attributeReplyChan: make(chan *core.Attribute, 1),
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
	cmd := &command{attributeReplyChan: make(chan *core.Attribute, 1),
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
	cmd := &command{functionReplyChan: make(chan *core.Function, 1),
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

func (controller *coloniesController) getFunctionByExecutorID(executorID string) ([]*core.Function, error) {
	cmd := &command{functionsReplyChan: make(chan []*core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			functions, err := controller.db.GetFunctionsByExecutorID(executorID)
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
	cmd := &command{functionReplyChan: make(chan *core.Function, 1),
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

func (controller *coloniesController) deleteFunction(functionID string) error {
	cmd := &command{errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.db.DeleteFunctionByID(functionID)
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
	cmd := &command{errorChan: make(chan error, 1),
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
	controller.relayServer.Shutdown()
	controller.etcdServer.Stop()
	controller.etcdServer.WaitToStop()
	os.RemoveAll(controller.etcdServer.StorageDir())
}
