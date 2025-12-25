package controllers

import (
	"errors"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/scheduler"
	"github.com/colonyos/colonies/pkg/backends"
	backendGin "github.com/colonyos/colonies/pkg/backends/gin"
	log "github.com/sirupsen/logrus"
)

// processGraphStorageAdapter implements core.ProcessGraphStorage by combining
// ProcessDatabase and ProcessGraphDatabase interfaces
type processGraphStorageAdapter struct {
	processDB      database.ProcessDatabase
	processGraphDB database.ProcessGraphDatabase
}

func (a *processGraphStorageAdapter) GetProcessByID(processID string) (*core.Process, error) {
	return a.processDB.GetProcessByID(processID)
}

func (a *processGraphStorageAdapter) SetProcessState(processID string, state int) error {
	return a.processDB.SetProcessState(processID, state)
}

func (a *processGraphStorageAdapter) SetWaitForParents(processID string, waitForParent bool) error {
	return a.processDB.SetWaitForParents(processID, waitForParent)
}

func (a *processGraphStorageAdapter) SetProcessGraphState(processGraphID string, state int) error {
	return a.processGraphDB.SetProcessGraphState(processGraphID, state)
}

// getProcessGraphStorage creates a storage adapter for ProcessGraph
func (controller *ColoniesController) GetProcessGraphStorage() *processGraphStorageAdapter {
	return &processGraphStorageAdapter{
		processDB:      controller.processDB,
		processGraphDB: controller.processGraphDB,
	}
}

// AssignResult contains the result of a process assignment attempt
type AssignResult struct {
	Process       *core.Process
	IsPaused      bool
	ResumeChannel <-chan bool // Only set when IsPaused=true
}

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
	assignResultReplyChan  chan *AssignResult
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

type ColoniesController struct {
	databaseCore     database.DatabaseCore
	userDB           database.UserDatabase
	colonyDB         database.ColonyDatabase
	executorDB       database.ExecutorDatabase
	functionDB       database.FunctionDatabase
	processDB        database.ProcessDatabase
	attributeDB      database.AttributeDatabase
	processGraphDB   database.ProcessGraphDatabase
	generatorDB      database.GeneratorDatabase
	cronDB           database.CronDatabase
	logDB            database.LogDatabase
	fileDB           database.FileDatabase
	snapshotDB       database.SnapshotDatabase
	blueprintDB      database.BlueprintDatabase
	securityDB       database.SecurityDatabase
	cmdQueue         chan *command
	blockingCmdQueue chan *command
	scheduler        *scheduler.Scheduler
	wsSubCtrl        backends.RealtimeSubscriptionController
	relayServer      *cluster.RelayServer
	eventHandler     backends.RealtimeEventHandler
	stopFlag         bool
	stopMutex        sync.Mutex
	leaderMutex      sync.Mutex
	thisNode         cluster.Node
	clusterConfig    cluster.Config
	etcdServer       *cluster.EtcdServer
	leader           bool
	generatorPeriod  int
	cronPeriod       int
	retention        bool
	retentionPolicy  int64
	retentionPeriod  int
	// Pause channel management
	pauseChannels    map[string][]chan bool // colony -> list of waiting channels
	pauseChannelsMux sync.RWMutex
	// Channel router for bidirectional communication
	channelRouter *channel.Router
}

func CreateColoniesController(db database.Database,
	thisNode cluster.Node,
	clusterConfig cluster.Config,
	etcdDataPath string,
	generatorPeriod int,
	cronPeriod int,
	retention bool,
	retentionPolicy int64,
	retentionPeriod int) *ColoniesController {

	controller := &ColoniesController{}
	// Set all the specific database interfaces
	controller.databaseCore = db
	controller.userDB = db
	controller.colonyDB = db
	controller.executorDB = db
	controller.functionDB = db
	controller.processDB = db
	controller.attributeDB = db
	controller.processGraphDB = db
	controller.generatorDB = db
	controller.cronDB = db
	controller.logDB = db
	controller.fileDB = db
	controller.snapshotDB = db
	controller.blueprintDB = db
	controller.securityDB = db
	controller.thisNode = thisNode
	controller.clusterConfig = clusterConfig
	controller.etcdServer = cluster.CreateEtcdServer(controller.thisNode, controller.clusterConfig, etcdDataPath)
	controller.etcdServer.Start()
	controller.etcdServer.WaitToStart()
	controller.leader = false
	controller.generatorPeriod = generatorPeriod
	controller.cronPeriod = cronPeriod
	controller.retention = retention
	controller.retentionPolicy = retentionPolicy
	controller.retentionPeriod = retentionPeriod
	controller.pauseChannels = make(map[string][]chan bool)
	controller.channelRouter = channel.NewRouter()

	controller.relayServer = cluster.CreateRelayServer(controller.thisNode, controller.clusterConfig)

	factory := backendGin.NewFactory()
	controller.eventHandler = factory.CreateEventHandler(controller.relayServer)
	controller.wsSubCtrl = factory.CreateSubscriptionController(controller.eventHandler)
	controller.scheduler = scheduler.CreateScheduler(controller.processDB)

	controller.cmdQueue = make(chan *command)
	controller.blockingCmdQueue = make(chan *command)

	controller.TryBecomeLeader()
	go controller.BlockingCmdQueueWorker()
	go controller.CmdQueueWorker()
	go controller.TimeoutLoop()
	go controller.GeneratorTriggerLoop()
	go controller.CronTriggerLoop()
	go controller.RetentionWorker()
	go controller.CleanupWorker()

	return controller
}

func (controller *ColoniesController) GetCronPeriod() int {
	return controller.cronPeriod
}

func (controller *ColoniesController) GetGeneratorPeriod() int {
	return controller.generatorPeriod
}

func (controller *ColoniesController) GetEtcdServer() *cluster.EtcdServer {
	return controller.etcdServer
}

func (controller *ColoniesController) GetEventHandler() backends.RealtimeEventHandler {
	return controller.eventHandler
}

func (controller *ColoniesController) GetThisNode() cluster.Node {
	return controller.thisNode
}

func (controller *ColoniesController) SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error {
	cmd := &command{threaded: false, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.executorDB.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if executor == nil {
				cmd.errorChan <- errors.New("executor not found")
				return
			}
			err = controller.wsSubCtrl.AddProcessesSubscriber(executorID, subscription)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}
	controller.blockingCmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *ColoniesController) SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error {
	cmd := &command{threaded: false, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.processDB.GetProcessByID(subscription.ProcessID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if process == nil {
				cmd.errorChan <- errors.New("Invalid process with Id " + subscription.ProcessID)
				return
			}

			err = controller.wsSubCtrl.AddProcessSubscriber(executorID, process, subscription)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}
	controller.blockingCmdQueue <- cmd

	return <-cmd.errorChan
}

func (controller *ColoniesController) GetColonies() ([]*core.Colony, error) {
	cmd := &command{threaded: true, coloniesReplyChan: make(chan []*core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies, err := controller.colonyDB.GetColonies()
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

func (controller *ColoniesController) GetColony(colonyName string) (*core.Colony, error) {
	cmd := &command{threaded: true, colonyReplyChan: make(chan *core.Colony),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colony, err := controller.colonyDB.GetColonyByName(colonyName)
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

func (controller *ColoniesController) AddColony(colony *core.Colony) (*core.Colony, error) {
	cmd := &command{threaded: true, colonyReplyChan: make(chan *core.Colony, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.colonyDB.AddColony(colony)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if colony == nil {
				cmd.errorChan <- errors.New("Invalid colony, colony is nil")
				return
			}

			addedColony, err := controller.colonyDB.GetColonyByID(colony.ID)
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

func (controller *ColoniesController) RemoveColony(colonyName string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.colonyDB.RemoveColonyByName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) AddExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error) {
	cmd := &command{threaded: true, executorReplyChan: make(chan *core.Executor, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if executor == nil {
				cmd.errorChan <- errors.New("Invalid executor, executor is nil")
				return
			}
			executorFromDB, err := controller.executorDB.GetExecutorByName(executor.ColonyName, executor.Name)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if executorFromDB != nil {
				if allowExecutorReregister {
					err := controller.executorDB.RemoveExecutorByName(executor.ColonyName, executorFromDB.Name)
					if err != nil {
						cmd.errorChan <- err
						return
					}
				} else {
					cmd.errorChan <- errors.New("Executor with name <" + executorFromDB.Name + "> in Colony <" + executorFromDB.ColonyName + "> already exists")
					return
				}
			}
			err = controller.executorDB.AddExecutor(executor)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedExecutor, err := controller.executorDB.GetExecutorByID(executor.ID)
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

func (controller *ColoniesController) GetExecutor(executorID string) (*core.Executor, error) {
	cmd := &command{threaded: true, executorReplyChan: make(chan *core.Executor),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.executorDB.GetExecutorByID(executorID)
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

func (controller *ColoniesController) GetExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	cmd := &command{threaded: true, executorsReplyChan: make(chan []*core.Executor),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executors, err := controller.executorDB.GetExecutorsByColonyName(colonyName)
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

func (controller *ColoniesController) AddProcessToDB(process *core.Process) (*core.Process, error) {
	if process == nil {
		return nil, errors.New("Invalid process, process is nil")
	}

	err := controller.processDB.AddProcess(process)
	if err != nil {
		return nil, err
	}

	addedProcess, err := controller.processDB.GetProcessByID(process.ID)
	if err != nil {
		return nil, err
	}

	// Create channels defined in FunctionSpec
	// Use deterministic IDs (processID_channelName) so channels can be created
	// consistently across cluster servers (lazy creation on any server)
	if addedProcess.FunctionSpec.Channels != nil {
		for _, channelName := range addedProcess.FunctionSpec.Channels {
			ch := &channel.Channel{
				ID:          addedProcess.ID + "_" + channelName, // Deterministic ID for cluster consistency
				ProcessID:   addedProcess.ID,
				Name:        channelName,
				SubmitterID: addedProcess.InitiatorID,
				ExecutorID:  "", // Will be set when process is assigned
			}
			if err := controller.channelRouter.Create(ch); err != nil {
				log.WithFields(log.Fields{"Error": err, "ProcessID": addedProcess.ID, "Channel": channelName}).Error("Failed to create channel")
			}
		}
	}

	return addedProcess, nil
}

// GetChannelRouter returns the channel router
func (controller *ColoniesController) GetChannelRouter() *channel.Router {
	return controller.channelRouter
}

func (controller *ColoniesController) AddProcess(process *core.Process) (*core.Process, error) {
	cmd := &command{threaded: true, processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			addedProcess, err := controller.AddProcessToDB(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if !addedProcess.WaitForParents {
				controller.eventHandler.Signal(addedProcess)
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

func (controller *ColoniesController) AddChild(
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
			parentProcess, err := controller.processDB.GetProcessByID(parentProcessID)
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
			addedProcess, err := controller.AddProcessToDB(process)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if insert {
				parentsChildren := parentProcess.Children
				controller.processDB.SetChildren(process.ID, parentsChildren)
				controller.processDB.SetChildren(parentProcessID, []string{process.ID})
				for _, parentsChildID := range parentsChildren {
					parentChild, err := controller.processDB.GetProcessByID(parentsChildID)
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
					controller.processDB.SetParents(parentsChildID, newParents)
				}
			} else {
				parentsChildren := parentProcess.Children
				parentsChildren = append(parentsChildren, process.ID)
				controller.processDB.SetChildren(parentProcessID, parentsChildren)
				if childProcessID != "" {
					controller.processDB.SetChildren(process.ID, []string{childProcessID})
					childProcess, err := controller.processDB.GetProcessByID(childProcessID)
					if err != nil {
						cmd.errorChan <- err
						return
					}
					newParents := childProcess.Parents
					newParents = append(newParents, process.ID)
					controller.processDB.SetParents(childProcessID, newParents)
				}
			}

			if !addedProcess.WaitForParents {
				controller.eventHandler.Signal(addedProcess)
			}

			updatedProcess, err := controller.processDB.GetProcessByID(addedProcess.ID)
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

func (controller *ColoniesController) GetProcess(processID string) (*core.Process, error) {
	cmd := &command{threaded: true, processReplyChan: make(chan *core.Process, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.processDB.GetProcessByID(processID)
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

func (controller *ColoniesController) FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	cmd := &command{threaded: true, processesReplyChan: make(chan []*core.Process),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			var processes []*core.Process
			var err error
			if executorID == "" {
				processes, err = controller.processDB.FindProcessesByColonyName(colonyName, seconds, state)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			} else {
				processes, err = controller.processDB.FindProcessesByExecutorID(colonyName, executorID, seconds, state)
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

func (controller *ColoniesController) UpdateProcessGraph(graph *core.ProcessGraph) error {
	graph.SetStorage(controller.GetProcessGraphStorage())
	return graph.UpdateProcessIDs()
}

func (controller *ColoniesController) CreateProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error) {
	processgraph, err := core.CreateProcessGraph(workflowSpec.ColonyName)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Error("Failed to create processgraph")
		return nil, err
	}

	initiatorName, err := resolveInitiator(workflowSpec.ColonyName, recoveredID, controller.executorDB, controller.userDB)
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

	err = controller.processGraphDB.AddProcessGraph(processgraph)
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
		addedProcess, err := controller.AddProcessToDB(process)
		if err != nil {
			msg := "Failed to submit workflow, failed to add process"
			log.WithFields(log.Fields{"Error": err}).Error(msg)
			return nil, errors.New(msg)
		}
		if !addedProcess.WaitForParents {
			controller.eventHandler.Signal(addedProcess)
		}
	}

	return processgraph, nil
}

func (controller *ColoniesController) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error) {
	cmd := &command{threaded: false, processGraphReplyChan: make(chan *core.ProcessGraph, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			addedProcessGraph, err := controller.CreateProcessGraph(workflowSpec, make([]interface{}, 0), make(map[string]interface{}), make([]interface{}, 0), recoveredID)
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

func (controller *ColoniesController) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphReplyChan: make(chan *core.ProcessGraph, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			graph, err := controller.processGraphDB.GetProcessGraphByID(processGraphID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if graph == nil {
				cmd.processGraphReplyChan <- nil
				return
			}
			err = controller.UpdateProcessGraph(graph)
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

func (controller *ColoniesController) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > constants.MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(constants.MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.processGraphDB.FindWaitingProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.UpdateProcessGraph(graph)
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

func (controller *ColoniesController) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > constants.MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(constants.MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.processGraphDB.FindRunningProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.UpdateProcessGraph(graph)
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

func (controller *ColoniesController) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > constants.MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(constants.MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.processGraphDB.FindSuccessfulProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.UpdateProcessGraph(graph)
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

func (controller *ColoniesController) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	cmd := &command{threaded: true, processGraphsReplyChan: make(chan []*core.ProcessGraph),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if count > constants.MAX_COUNT {
				cmd.errorChan <- errors.New("Count is larger than MaxCount limit <" + strconv.Itoa(constants.MAX_COUNT) + ">")
				return
			}
			graphs, err := controller.processGraphDB.FindFailedProcessGraphs(colonyName, count)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			for _, graph := range graphs {
				err = controller.UpdateProcessGraph(graph)
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

func (controller *ColoniesController) RemoveProcess(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.processDB.RemoveProcessByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) RemoveAllProcesses(colonyName string, state int) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			switch state {
			case core.WAITING:
				cmd.errorChan <- controller.processDB.RemoveAllWaitingProcessesByColonyName(colonyName)
			case core.RUNNING:
				cmd.errorChan <- errors.New("Not possible to remove running processes")
			case core.SUCCESS:
				cmd.errorChan <- controller.processDB.RemoveAllSuccessfulProcessesByColonyName(colonyName)
			case core.FAILED:
				cmd.errorChan <- controller.processDB.RemoveAllFailedProcessesByColonyName(colonyName)
			case core.NOTSET:
				cmd.errorChan <- controller.processDB.RemoveAllProcessesByColonyName(colonyName)
			default:
				cmd.errorChan <- errors.New("Invalid state when deleting all processes")
			}
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) RemoveProcessGraph(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.processGraphDB.RemoveProcessGraphByID(processID)
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) RemoveAllProcessGraphs(colonyName string, state int) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			switch state {
			case core.WAITING:
				cmd.errorChan <- controller.processGraphDB.RemoveAllWaitingProcessGraphsByColonyName(colonyName)
			case core.RUNNING:
				cmd.errorChan <- errors.New("not possible to remove running processgraphs")
			case core.SUCCESS:
				cmd.errorChan <- controller.processGraphDB.RemoveAllSuccessfulProcessGraphsByColonyName(colonyName)
			case core.FAILED:
				cmd.errorChan <- controller.processGraphDB.RemoveAllFailedProcessGraphsByColonyName(colonyName)
			case core.NOTSET:
				cmd.errorChan <- controller.processGraphDB.RemoveAllProcessGraphsByColonyName(colonyName)
			default:
				cmd.errorChan <- errors.New("invalid state when deleting all processgraphs")
			}
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) SetOutput(processID string, output []interface{}) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			if len(output) > 0 {
				err := controller.processDB.SetOutput(processID, output)
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

func (controller *ColoniesController) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.processDB.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if len(output) > 0 {
				err = controller.processDB.SetOutput(processID, output)
			}

			waitingTime, processingTime, err := controller.processDB.MarkSuccessful(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if process.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": process.ProcessGraphID}).Debug("Resolving processgraph (close successful)")
				processGraph, err := controller.processGraphDB.GetProcessGraphByID(process.ProcessGraphID)
				if err != nil {
					cmd.errorChan <- err
					return
				}
				processGraph.SetStorage(controller.GetProcessGraphStorage())
				err = processGraph.Resolve()
				if err != nil {
					err2 := controller.HandleDefunctProcessgraph(processGraph.ID, process.ID, err)
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
				err = controller.NotifyChildren(process)
				if err != nil {
					cmd.errorChan <- err
					return
				}
			}

			process.State = core.SUCCESS

			executor, err := controller.executorDB.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			function, err := controller.functionDB.GetFunctionsByExecutorAndName(process.FunctionSpec.Conditions.ColonyName, executor.Name, process.FunctionSpec.FuncName)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if function != nil {
				process, err = controller.processDB.GetProcessByID(processID)
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

				controller.functionDB.UpdateFunctionStats(
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

			// Cleanup channels for this process
			controller.channelRouter.CleanupProcess(processID)

			controller.eventHandler.Signal(process)
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) NotifyChildren(process *core.Process) error {
	// First check if parent processes are completed
	counter := 0
	for _, parentProcessID := range process.Parents {
		parentProcess, err := controller.processDB.GetProcessByID(parentProcessID)
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
			childProcess, err := controller.processDB.GetProcessByID(childProcessID)
			if err != nil {
				return err
			}
			controller.eventHandler.Signal(childProcess)
		}
	}

	return nil
}

func (controller *ColoniesController) CloseFailed(processID string, errs []string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.processDB.MarkFailed(processID, errs)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			process, err := controller.processDB.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if process.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": process.ProcessGraphID}).Debug("Resolving processgraph (close failed)")
				processGraph, err := controller.processGraphDB.GetProcessGraphByID(process.ProcessGraphID)
				if err != nil {
					cmd.errorChan <- err
					return
				}
				processGraph.SetStorage(controller.GetProcessGraphStorage())
				err = processGraph.Resolve()
				if err != nil {
					err2 := controller.HandleDefunctProcessgraph(processGraph.ID, process.ID, err)
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

			// Cleanup channels for this process
			controller.channelRouter.CleanupProcess(processID)

			controller.eventHandler.Signal(process)
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) HandleDefunctProcessgraph(processGraphID string, processID string, err error) error {
	err2 := controller.processDB.MarkFailed(processID, []string{err.Error()})
	if err2 != nil {
		return err2
	}
	err2 = controller.processGraphDB.SetProcessGraphState(processGraphID, core.FAILED)
	if err2 != nil {
		return err2
	}

	return nil
}

func (controller *ColoniesController) Assign(executorID string, colonyName string, cpu int64, mem int64) (*AssignResult, error) {
	cmd := &command{threaded: false, assignResultReplyChan: make(chan *AssignResult),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			executor, err := controller.executorDB.GetExecutorByID(executorID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			if executor == nil {
				cmd.errorChan <- errors.New("Executor with Id <" + executorID + "> could not be found")
				return
			}

			err = controller.executorDB.MarkAlive(executor)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			// Check if assignments are paused for this colony first
			paused, err := controller.AreColonyAssignmentsPaused(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			if paused {
				// Create a resume channel for this colony
				resumeChannel := controller.CreateResumeChannel(colonyName)
				result := &AssignResult{
					Process:       nil,
					IsPaused:      true,
					ResumeChannel: resumeChannel,
				}
				cmd.assignResultReplyChan <- result
				return
			}

			// Not paused - proceed with normal assignment
			selectedProcess, err := controller.scheduler.Select(colonyName, executor, cpu, mem)
			if err != nil {
				// If no processes can be selected, return a result with nil process (not an error)
				if err.Error() == "No processes can be selected for executor with Id <"+executorID+">" {
					result := &AssignResult{
						Process:       nil,
						IsPaused:      false,
						ResumeChannel: nil,
					}
					cmd.assignResultReplyChan <- result
					return
				}
				// For other errors, return as error
				cmd.errorChan <- err
				return
			}

			err = controller.processDB.Assign(executorID, selectedProcess)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			// Update executor ID on all channels for this process
			controller.channelRouter.SetExecutorIDForProcess(selectedProcess.ID, executorID)

			if selectedProcess.ProcessGraphID != "" {
				log.WithFields(log.Fields{"ProcessGraphId": selectedProcess.ProcessGraphID}).Debug("Resolving processgraph (assigned)")
				processGraph, err := controller.processGraphDB.GetProcessGraphByID(selectedProcess.ProcessGraphID)
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
				processGraph.SetStorage(controller.GetProcessGraphStorage())

				// One Colonies server might have added a processgraph, and another colonies directly get an assign request
				// This means that all processes part of the graph might not yet have been added, consequently the
				// processgraph.Resolve() call might fail.
				// The solution is to retry a couple of times.
				maxRetries := 10
				timeBetweenRetries := 500 * time.Millisecond // We will wait what max 10 * 0.5 = 5 seconds
				retries := 0

				for {
					if retries >= maxRetries {
						err2 := controller.HandleDefunctProcessgraph(processGraph.ID, selectedProcess.ID, err)
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
					parentProcess, err := controller.processDB.GetProcessByID(parentID)
					if err != nil {
						log.Error(err)
						cmd.errorChan <- err
						return
					}
					output = append(output, parentProcess.Output...)
				}
				if len(selectedProcess.Parents) > 0 {
					controller.processDB.SetInput(selectedProcess.ID, output)
					selectedProcess.Input = output
				}
			}

			// Signal that the process is now RUNNING so subscribers are notified
			controller.eventHandler.Signal(selectedProcess)

			result := &AssignResult{
				Process:       selectedProcess,
				IsPaused:      false,
				ResumeChannel: nil,
			}
			cmd.assignResultReplyChan <- result
		}}

	controller.blockingCmdQueue <- cmd
	select {
	case err := <-cmd.errorChan:
		return nil, err
	case result := <-cmd.assignResultReplyChan:
		return result, nil
	}
}

func (controller *ColoniesController) UnassignExecutor(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.processDB.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.errorChan <- controller.processDB.Unassign(process)
			controller.eventHandler.Signal(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) ResetProcess(processID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			process, err := controller.processDB.GetProcessByID(processID)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.errorChan <- controller.processDB.ResetProcess(process)
			controller.eventHandler.Signal(process)
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) GetColonyStatistics(colonyName string) (*core.Statistics, error) {
	cmd := &command{threaded: true, statisticsReplyChan: make(chan *core.Statistics),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies := 1
			executors, err := controller.executorDB.CountExecutorsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			activeExecutors, err := controller.executorDB.CountExecutorsByColonyNameAndState(colonyName, core.APPROVED)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			unregisteredExecutors, err := controller.executorDB.CountExecutorsByColonyNameAndState(colonyName, core.UNREGISTERED)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingProcesses, err := controller.processDB.CountWaitingProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningProcesses, err := controller.processDB.CountRunningProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successProcesses, err := controller.processDB.CountSuccessfulProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedProcesses, err := controller.processDB.CountFailedProcessesByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingWorkflows, err := controller.processGraphDB.CountWaitingProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningWorkflows, err := controller.processGraphDB.CountRunningProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successWorkflows, err := controller.processGraphDB.CountSuccessfulProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedWorkflows, err := controller.processGraphDB.CountFailedProcessGraphsByColonyName(colonyName)
			if err != nil {
				cmd.errorChan <- err
				return
			}

			cmd.statisticsReplyChan <- core.CreateStatistics(colonies,
				executors,
				activeExecutors,
				unregisteredExecutors,
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

func (controller *ColoniesController) GetStatistics() (*core.Statistics, error) {
	cmd := &command{threaded: true, statisticsReplyChan: make(chan *core.Statistics),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			colonies, err := controller.colonyDB.CountColonies()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			executors, err := controller.executorDB.CountExecutors()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingProcesses, err := controller.processDB.CountWaitingProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningProcesses, err := controller.processDB.CountRunningProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successProcesses, err := controller.processDB.CountSuccessfulProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedProcesses, err := controller.processDB.CountFailedProcesses()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			waitingWorkflows, err := controller.processGraphDB.CountWaitingProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			runningWorkflows, err := controller.processGraphDB.CountRunningProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			successWorkflows, err := controller.processGraphDB.CountSuccessfulProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			failedWorkflows, err := controller.processGraphDB.CountFailedProcessGraphs()
			if err != nil {
				cmd.errorChan <- err
				return
			}

			// For global statistics, active/unregistered counts are not meaningful across colonies
			// These fields are populated at the colony level
			cmd.statisticsReplyChan <- core.CreateStatistics(colonies,
				executors,
				0, // activeExecutors (not applicable for global stats)
				0, // unregisteredExecutors (not applicable for global stats)
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

func (controller *ColoniesController) AddAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	cmd := &command{threaded: true, attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.attributeDB.AddAttribute(*attribute)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedAttribute, err := controller.attributeDB.GetAttributeByID(attribute.ID)
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

func (controller *ColoniesController) GetAttribute(attributeID string) (*core.Attribute, error) {
	cmd := &command{threaded: true, attributeReplyChan: make(chan *core.Attribute, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			attribute, err := controller.attributeDB.GetAttributeByID(attributeID)
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

func (controller *ColoniesController) AddFunction(function *core.Function) (*core.Function, error) {
	cmd := &command{threaded: true, functionReplyChan: make(chan *core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.functionDB.AddFunction(function)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			addedFunction, err := controller.functionDB.GetFunctionByID(function.FunctionID)
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

func (controller *ColoniesController) GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) {
	cmd := &command{threaded: true, functionsReplyChan: make(chan []*core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			functions, err := controller.functionDB.GetFunctionsByExecutorName(colonyName, executorName)
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

func (controller *ColoniesController) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	cmd := &command{threaded: true, functionsReplyChan: make(chan []*core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			functions, err := controller.functionDB.GetFunctionsByColonyName(colonyName)
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

func (controller *ColoniesController) GetFunctionByID(functionID string) (*core.Function, error) {
	cmd := &command{threaded: true, functionReplyChan: make(chan *core.Function, 1),
		errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			function, err := controller.functionDB.GetFunctionByID(functionID)
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

func (controller *ColoniesController) RemoveFunction(functionID string) error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.functionDB.RemoveFunctionByID(functionID)
			if err != nil {
				cmd.errorChan <- err
				return
			}
			cmd.errorChan <- nil
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) ResetDatabase() error {
	cmd := &command{threaded: true, errorChan: make(chan error, 1),
		handler: func(cmd *command) {
			err := controller.databaseCore.Drop()
			if err != nil {
				cmd.errorChan <- err
				return
			}
			err = controller.databaseCore.Initialize()
			cmd.errorChan <- err
		}}

	controller.cmdQueue <- cmd
	return <-cmd.errorChan
}

func (controller *ColoniesController) PauseColonyAssignments(colonyName string) error {
	return controller.etcdServer.PauseColonyAssignments(colonyName)
}

func (controller *ColoniesController) ResumeColonyAssignments(colonyName string) error {
	err := controller.etcdServer.ResumeColonyAssignments(colonyName)
	if err != nil {
		return err
	}
	
	// Wake up all waiting executors for this colony
	controller.WakeupPausedAssignments(colonyName)
	return nil
}

func (controller *ColoniesController) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return controller.etcdServer.AreColonyAssignmentsPaused(colonyName)
}

// createResumeChannel creates a channel that will be signaled when assignments are resumed for a colony
func (controller *ColoniesController) CreateResumeChannel(colonyName string) <-chan bool {
	controller.pauseChannelsMux.Lock()
	defer controller.pauseChannelsMux.Unlock()
	
	resumeChannel := make(chan bool, 1)
	if controller.pauseChannels[colonyName] == nil {
		controller.pauseChannels[colonyName] = make([]chan bool, 0)
	}
	controller.pauseChannels[colonyName] = append(controller.pauseChannels[colonyName], resumeChannel)
	
	return resumeChannel
}

// wakeupPausedAssignments signals all waiting executors that assignments have been resumed
func (controller *ColoniesController) WakeupPausedAssignments(colonyName string) {
	controller.pauseChannelsMux.Lock()
	defer controller.pauseChannelsMux.Unlock()
	
	if channels, exists := controller.pauseChannels[colonyName]; exists {
		for _, ch := range channels {
			select {
			case ch <- true:
			default:
				// Channel already has a value or is closed, skip
			}
		}
		// Clear the channels slice after waking everyone up
		delete(controller.pauseChannels, colonyName)
	}
}

func (controller *ColoniesController) Stop() {
	controller.stopMutex.Lock()
	controller.stopFlag = true
	controller.stopMutex.Unlock()
	controller.cmdQueue <- &command{stop: true}
	controller.eventHandler.Stop()
	controller.relayServer.Shutdown()
	controller.etcdServer.Stop()
	controller.etcdServer.WaitToStop()
	os.RemoveAll(controller.etcdServer.StorageDir())
}
