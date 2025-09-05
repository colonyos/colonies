package server

import (
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	attributehandlers "github.com/colonyos/colonies/pkg/server/handlers/attribute"
	"github.com/colonyos/colonies/pkg/server/handlers/colony"
	"github.com/colonyos/colonies/pkg/server/handlers/executor"
	functionhandlers "github.com/colonyos/colonies/pkg/server/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/server/handlers/generator"
	loghandlers "github.com/colonyos/colonies/pkg/server/handlers/log"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/handlers/processgraph"
	serverhandlers "github.com/colonyos/colonies/pkg/server/handlers/server"
	websockethandlers "github.com/colonyos/colonies/pkg/server/handlers/websocket"
	servercommunication "github.com/colonyos/colonies/pkg/server/websocket"
	"github.com/gin-gonic/gin"
)

// ServerAdapter implements interfaces needed by handler packages
type ServerAdapter struct {
	server *ColoniesServer
}

func NewServerAdapter(server *ColoniesServer) *ServerAdapter {
	return &ServerAdapter{
		server: server,
	}
}

// User handler interface methods
func (s *ServerAdapter) GetUserDB() database.UserDatabase {
	return s.server.userDB
}

func (s *ServerAdapter) GetColonyDB() database.ColonyDatabase {
	return s.server.colonyDB
}

func (s *ServerAdapter) GetSecurityDB() database.SecurityDatabase {
	return s.server.securityDB
}

func (s *ServerAdapter) GetValidator() security.Validator {
	return s.server.validator
}

// Controller access for handlers
func (s *ServerAdapter) GetController() interface{
	addColony(colony *core.Colony) (*core.Colony, error)
	removeColony(colonyName string) error
	getColonies() ([]*core.Colony, error)
	getColony(colonyName string) (*core.Colony, error)
	getColonyStatistics(colonyName string) (*core.Statistics, error)
} {
	return s.server.controller
}

func (s *ServerAdapter) HandleHTTPError(c *gin.Context, err error, errorCode int) bool {
	return s.server.handleHTTPError(c, err, errorCode)
}

func (s *ServerAdapter) SendHTTPReply(c *gin.Context, payloadType string, jsonString string) {
	s.server.sendHTTPReply(c, payloadType, jsonString)
}

func (s *ServerAdapter) SendEmptyHTTPReply(c *gin.Context, payloadType string) {
	s.server.sendEmptyHTTPReply(c, payloadType)
}

func (s *ServerAdapter) GetServerID() (string, error) {
	return s.server.getServerID()
}

func (s *ServerAdapter) Validator() security.Validator {
	return s.server.validator
}

func (s *ServerAdapter) ColonyDB() database.ColonyDatabase {
	return s.server.colonyDB
}

type controllerAdapter struct {
	controller interface {
		addColony(colony *core.Colony) (*core.Colony, error)
		removeColony(colonyName string) error
		getColonies() ([]*core.Colony, error)
		getColony(colonyName string) (*core.Colony, error)
		getColonyStatistics(colonyName string) (*core.Statistics, error)
	}
}

func (c *controllerAdapter) AddColony(colony *core.Colony) (*core.Colony, error) {
	return c.controller.addColony(colony)
}

func (c *controllerAdapter) RemoveColony(colonyName string) error {
	return c.controller.removeColony(colonyName)
}

func (c *controllerAdapter) GetColonies() ([]*core.Colony, error) {
	return c.controller.getColonies()
}

func (c *controllerAdapter) GetColony(colonyName string) (*core.Colony, error) {
	return c.controller.getColony(colonyName)
}

func (c *controllerAdapter) GetColonyStatistics(colonyName string) (*core.Statistics, error) {
	return c.controller.getColonyStatistics(colonyName)
}

func (s *ServerAdapter) Controller() colony.Controller {
	return &controllerAdapter{controller: s.server.controller}
}

// Executor handler interface methods
func (s *ServerAdapter) ExecutorDB() database.ExecutorDatabase {
	return s.server.executorDB
}

func (s *ServerAdapter) AllowExecutorReregister() bool {
	return s.server.allowExecutorReregister
}

func (s *ServerAdapter) SetAllowExecutorReregister(allow bool) {
	s.server.allowExecutorReregister = allow
}

type executorControllerAdapter struct {
	controller interface {
		addExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error)
		getExecutor(executorID string) (*core.Executor, error)
		getExecutorByColonyName(colonyName string) ([]*core.Executor, error)
	}
}

func (c *executorControllerAdapter) AddExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error) {
	return c.controller.addExecutor(executor, allowReregister)
}

func (c *executorControllerAdapter) GetExecutor(executorID string) (*core.Executor, error) {
	return c.controller.getExecutor(executorID)
}

func (c *executorControllerAdapter) GetExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	return c.controller.getExecutorByColonyName(colonyName)
}

func (s *ServerAdapter) ExecutorController() executor.Controller {
	return &executorControllerAdapter{controller: s.server.controller}
}

// Process handler interface methods
func (s *ServerAdapter) ProcessDB() database.ProcessDatabase {
	return s.server.processDB
}

func (s *ServerAdapter) LogDB() database.LogDatabase {
	return s.server.logDB
}

func (s *ServerAdapter) SnapshotDB() database.SnapshotDatabase {
	return s.server.snapshotDB
}

type attributeControllerAdapter struct {
	controller interface {
		getProcess(processID string) (*core.Process, error)
		addAttribute(attribute *core.Attribute) (*core.Attribute, error)
		getAttribute(attributeID string) (*core.Attribute, error)
	}
}

func (c *attributeControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.getProcess(processID)
}

func (c *attributeControllerAdapter) AddAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	return c.controller.addAttribute(attribute)
}

func (c *attributeControllerAdapter) GetAttribute(attributeID string) (*core.Attribute, error) {
	return c.controller.getAttribute(attributeID)
}

func (s *ServerAdapter) AttributeController() attributehandlers.Controller {
	return &attributeControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) FunctionDB() database.FunctionDatabase {
	return s.server.functionDB
}

func (s *ServerAdapter) GeneratorDB() database.GeneratorDatabase {
	return s.server.generatorDB
}

type functionControllerAdapter struct {
	controller interface {
		addFunction(function *core.Function) (*core.Function, error)
		getFunctionByID(functionID string) (*core.Function, error)
		getFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
		getFunctionsByColonyName(colonyName string) ([]*core.Function, error)
		removeFunction(functionID string) error
	}
}

type generatorControllerAdapter struct {
	controller interface {
		addGenerator(generator *core.Generator) (*core.Generator, error)
		getGenerator(generatorID string) (*core.Generator, error)
		resolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
		getGenerators(colonyName string, count int) ([]*core.Generator, error)
		packGenerator(generatorID string, colonyName string, arg string) error
		removeGenerator(generatorID string) error
		getGeneratorPeriod() int
	}
}

func (c *functionControllerAdapter) AddFunction(function *core.Function) (*core.Function, error) {
	return c.controller.addFunction(function)
}

func (c *functionControllerAdapter) GetFunction(functionID string) (*core.Function, error) {
	return c.controller.getFunctionByID(functionID)
}

func (c *functionControllerAdapter) GetFunctions(colonyName string, executorName string, count int) ([]*core.Function, error) {
	return c.controller.getFunctionsByExecutorName(colonyName, executorName)
}

func (c *functionControllerAdapter) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	return c.controller.getFunctionsByColonyName(colonyName)
}

func (c *functionControllerAdapter) RemoveFunction(functionID string, initiatorID string) error {
	return c.controller.removeFunction(functionID)
}

func (c *generatorControllerAdapter) AddGenerator(generator *core.Generator) (*core.Generator, error) {
	return c.controller.addGenerator(generator)
}

func (c *generatorControllerAdapter) GetGenerator(generatorID string) (*core.Generator, error) {
	return c.controller.getGenerator(generatorID)
}

func (c *generatorControllerAdapter) ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error) {
	return c.controller.resolveGenerator(colonyName, generatorName)
}

func (c *generatorControllerAdapter) GetGenerators(colonyName string, count int) ([]*core.Generator, error) {
	return c.controller.getGenerators(colonyName, count)
}

func (c *generatorControllerAdapter) PackGenerator(generatorID string, colonyName string, arg string) error {
	return c.controller.packGenerator(generatorID, colonyName, arg)
}

func (c *generatorControllerAdapter) RemoveGenerator(generatorID string) error {
	return c.controller.removeGenerator(generatorID)
}

func (c *generatorControllerAdapter) GetGeneratorPeriod() int {
	return c.controller.getGeneratorPeriod()
}

func (s *ServerAdapter) FunctionController() functionhandlers.Controller {
	return &functionControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) GeneratorController() generatorhandlers.Controller {
	return &generatorControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) ExclusiveAssign() bool {
	return s.server.exclusiveAssign
}

func (s *ServerAdapter) TLS() bool {
	return s.server.tls
}

type processControllerAdapter struct {
	controller interface {
		addProcessToDB(process *core.Process) (*core.Process, error)
		addProcess(process *core.Process) (*core.Process, error)
		getProcess(processID string) (*core.Process, error)
		getExecutor(executorID string) (*core.Executor, error)
		findProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
		removeProcess(processID string) error
		removeAllProcesses(colonyName string, state int) error
		setOutput(processID string, output []interface{}) error
		closeSuccessful(processID string, executorID string, output []interface{}) error
		closeFailed(processID string, errs []string) error
		assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error)
		unassignExecutor(processID string) error
		pauseColonyAssignments(colonyName string) error
		resumeColonyAssignments(colonyName string) error
		areColonyAssignmentsPaused(colonyName string) (bool, error)
		getEventHandler() *servercommunication.EventHandler
		isLeader() bool
		getEtcdServer() *cluster.EtcdServer
	}
}

func (c *processControllerAdapter) AddProcessToDB(process *core.Process) (*core.Process, error) {
	return c.controller.addProcessToDB(process)
}

func (c *processControllerAdapter) AddProcess(process *core.Process) (*core.Process, error) {
	return c.controller.addProcess(process)
}

func (c *processControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.getProcess(processID)
}

func (c *processControllerAdapter) GetExecutor(executorID string) (*core.Executor, error) {
	return c.controller.getExecutor(executorID)
}

func (c *processControllerAdapter) FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return c.controller.findProcessHistory(colonyName, executorID, seconds, state)
}

func (c *processControllerAdapter) RemoveProcess(processID string) error {
	return c.controller.removeProcess(processID)
}

func (c *processControllerAdapter) RemoveAllProcesses(colonyName string, state int) error {
	return c.controller.removeAllProcesses(colonyName, state)
}

func (c *processControllerAdapter) SetOutput(processID string, output []interface{}) error {
	return c.controller.setOutput(processID, output)
}

func (c *processControllerAdapter) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	return c.controller.closeSuccessful(processID, executorID, output)
}

func (c *processControllerAdapter) CloseFailed(processID string, errs []string) error {
	return c.controller.closeFailed(processID, errs)
}

func (c *processControllerAdapter) Assign(executorID string, colonyName string, cpu int64, memory int64) (*process.AssignResult, error) {
	result, err := c.controller.assign(executorID, colonyName, cpu, memory)
	if err != nil {
		return nil, err
	}
	// Convert the internal assign result to the process handler's AssignResult
	return &process.AssignResult{
		Process:       result.Process,
		IsPaused:      result.IsPaused,
		ResumeChannel: result.ResumeChannel,
	}, nil
}

func (c *processControllerAdapter) UnassignExecutor(processID string) error {
	return c.controller.unassignExecutor(processID)
}

func (c *processControllerAdapter) PauseColonyAssignments(colonyName string) error {
	return c.controller.pauseColonyAssignments(colonyName)
}

func (c *processControllerAdapter) ResumeColonyAssignments(colonyName string) error {
	return c.controller.resumeColonyAssignments(colonyName)
}

func (c *processControllerAdapter) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return c.controller.areColonyAssignmentsPaused(colonyName)
}

func (c *processControllerAdapter) GetEventHandler() *process.EventHandler {
	// Convert the internal event handler to the process handler's EventHandler
	return &process.EventHandler{}
}

func (c *processControllerAdapter) IsLeader() bool {
	return c.controller.isLeader()
}

func (c *processControllerAdapter) GetEtcdServer() process.EtcdServer {
	etcdServer := c.controller.getEtcdServer()
	return &etcdServerAdapter{etcdServer: etcdServer}
}

type logControllerAdapter struct {
	controller interface {
		getProcess(processID string) (*core.Process, error)
	}
}

func (c *logControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.getProcess(processID)
}

type etcdServerAdapter struct {
	etcdServer *cluster.EtcdServer
}

func (e *etcdServerAdapter) CurrentCluster() process.Cluster {
	clusterConfig := e.etcdServer.CurrentCluster()
	return &clusterAdapter{cluster: &clusterConfig}
}

type clusterAdapter struct {
	cluster *cluster.Config
}

func (c *clusterAdapter) GetLeader() *process.Leader {
	leader := c.cluster.Leader
	return &process.Leader{
		Host:    leader.Host,
		APIPort: leader.APIPort,
	}
}

func (s *ServerAdapter) ProcessController() process.Controller {
	return &processControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) LogProcessController() loghandlers.Controller {
	return &logControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) UserDB() database.UserDatabase {
	return s.server.userDB
}

func (s *ServerAdapter) FileDB() database.FileDatabase {
	return s.server.fileDB
}

func (s *ServerAdapter) SecurityDB() database.SecurityDatabase {
	return s.server.securityDB
}

func (s *ServerAdapter) WSController() websockethandlers.WSController {
	return &wsControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) ParseSignature(payload string, signature string) (string, error) {
	return s.server.parseSignature(payload, signature)
}

func (s *ServerAdapter) GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
	return s.server.generateRPCErrorMsg(err, errorCode)
}

// WSController interface for WebSocket handlers
type WSController interface {
	SubscribeProcesses(executorID string, subscription *websockethandlers.Subscription) error
	SubscribeProcess(executorID string, subscription *websockethandlers.Subscription) error
}

// wsControllerAdapter adapter for WebSocket handlers
type wsControllerAdapter struct {
	controller interface {
		subscribeProcesses(executorID string, subscription *websockethandlers.Subscription) error
		subscribeProcess(executorID string, subscription *websockethandlers.Subscription) error
	}
}

func (c *wsControllerAdapter) SubscribeProcesses(executorID string, subscription *websockethandlers.Subscription) error {
	return c.controller.subscribeProcesses(executorID, subscription)
}

func (c *wsControllerAdapter) SubscribeProcess(executorID string, subscription *websockethandlers.Subscription) error {
	return c.controller.subscribeProcess(executorID, subscription)
}

// Cron controller adapter
type cronControllerAdapter struct {
	controller interface {
		addCron(cron *core.Cron) (*core.Cron, error)
		getCron(cronID string) (*core.Cron, error)
		getCrons(colonyName string, count int) ([]*core.Cron, error)
		runCron(cronID string) (*core.Cron, error)
		removeCron(cronID string) error
		getCronPeriod() int
	}
}

func (c *cronControllerAdapter) AddCron(cron *core.Cron) (*core.Cron, error) {
	return c.controller.addCron(cron)
}

func (c *cronControllerAdapter) GetCron(cronID string) (*core.Cron, error) {
	return c.controller.getCron(cronID)
}

func (c *cronControllerAdapter) GetCrons(colonyName string, count int) ([]*core.Cron, error) {
	return c.controller.getCrons(colonyName, count)
}

func (c *cronControllerAdapter) RunCron(cronID string) (*core.Cron, error) {
	return c.controller.runCron(cronID)
}

func (c *cronControllerAdapter) RemoveCron(cronID string) error {
	return c.controller.removeCron(cronID)
}

func (c *cronControllerAdapter) GetCronPeriod() int {
	return c.controller.getCronPeriod()
}

// CronController returns the server's controller interface for cron operations
func (s *ServerAdapter) CronController() interface {
	AddCron(cron *core.Cron) (*core.Cron, error)
	GetCron(cronID string) (*core.Cron, error)
	GetCrons(colonyName string, count int) ([]*core.Cron, error)
	RunCron(cronID string) (*core.Cron, error)
	RemoveCron(cronID string) error
	GetCronPeriod() int
} {
	return &cronControllerAdapter{controller: s.server.controller}
}

// ProcessGraph controller adapter
type processgraphControllerAdapter struct {
	controller interface {
		submitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error)
		getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
		findWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		findRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		findSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		findFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		removeProcessGraph(processGraphID string) error
		removeAllProcessGraphs(colonyName string, state int) error
		addChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error)
	}
}

func (c *processgraphControllerAdapter) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error) {
	return c.controller.submitWorkflowSpec(workflowSpec, initiatorID)
}

func (c *processgraphControllerAdapter) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return c.controller.getProcessGraphByID(processGraphID)
}

func (c *processgraphControllerAdapter) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.findWaitingProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.findRunningProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.findSuccessfulProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.findFailedProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) RemoveProcessGraph(processGraphID string) error {
	return c.controller.removeProcessGraph(processGraphID)
}

func (c *processgraphControllerAdapter) RemoveAllProcessGraphs(colonyName string, state int) error {
	return c.controller.removeAllProcessGraphs(colonyName, state)
}

func (c *processgraphControllerAdapter) AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error) {
	return c.controller.addChild(processGraphID, parentProcessID, childProcessID, process, initiatorID, insert)
}

func (s *ServerAdapter) ProcessgraphController() processgraph.Controller {
	return &processgraphControllerAdapter{controller: s.server.controller}
}

// ProcessGraph handler validator adapter
type processgraphValidatorAdapter struct {
	validator security.Validator
}

func (v *processgraphValidatorAdapter) RequireMembership(recoveredID string, colonyName string, executorMayJoin bool) error {
	return v.validator.RequireMembership(recoveredID, colonyName, executorMayJoin)
}

func (v *processgraphValidatorAdapter) RequireColonyOwner(recoveredID string, colonyName string) error {
	return v.validator.RequireColonyOwner(recoveredID, colonyName)
}

func (s *ServerAdapter) ProcessgraphValidator() processgraph.Validator {
	return &processgraphValidatorAdapter{validator: s.server.validator}
}

// ProcessGraph server adapter
type processgraphServerAdapter struct {
	server *ColoniesServer
	adapter *ServerAdapter
}

func (s *processgraphServerAdapter) HandleHTTPError(c *gin.Context, err error, errorCode int) bool {
	return s.server.handleHTTPError(c, err, errorCode)
}

func (s *processgraphServerAdapter) SendHTTPReply(c *gin.Context, payloadType string, jsonString string) {
	s.server.sendHTTPReply(c, payloadType, jsonString)
}

func (s *processgraphServerAdapter) SendEmptyHTTPReply(c *gin.Context, payloadType string) {
	s.server.sendEmptyHTTPReply(c, payloadType)
}

func (s *processgraphServerAdapter) Validator() processgraph.Validator {
	return s.adapter.ProcessgraphValidator()
}

func (s *processgraphServerAdapter) Controller() processgraph.Controller {
	return s.adapter.ProcessgraphController()
}

func (s *ServerAdapter) ProcessgraphServer() processgraph.ColoniesServer {
	return &processgraphServerAdapter{
		server: s.server,
		adapter: s,
	}
}

// Server handler controller adapter
type serverControllerAdapter struct {
	controller interface {
		getStatistics() (*core.Statistics, error)
		getEtcdServer() *cluster.EtcdServer
	}
}

func (c *serverControllerAdapter) GetStatistics() (*core.Statistics, error) {
	return c.controller.getStatistics()
}

func (c *serverControllerAdapter) GetEtcdServer() serverhandlers.EtcdServer {
	etcdServer := c.controller.getEtcdServer()
	return &serverEtcdServerAdapter{etcdServer: etcdServer}
}

type serverEtcdServerAdapter struct {
	etcdServer *cluster.EtcdServer
}

func (e *serverEtcdServerAdapter) CurrentCluster() cluster.Config {
	return e.etcdServer.CurrentCluster()
}

func (s *ServerAdapter) ServerController() serverhandlers.Controller {
	return &serverControllerAdapter{controller: s.server.controller}
}

// Server handler validator adapter
type serverValidatorAdapter struct {
	validator security.Validator
}

func (v *serverValidatorAdapter) RequireServerOwner(recoveredID string, serverID string) error {
	return v.validator.RequireServerOwner(recoveredID, serverID)
}

func (s *ServerAdapter) ServerValidator() serverhandlers.Validator {
	return &serverValidatorAdapter{validator: s.server.validator}
}

// Server handler server adapter
type serverServerAdapter struct {
	server *ColoniesServer
	adapter *ServerAdapter
}

func (s *serverServerAdapter) HandleHTTPError(c *gin.Context, err error, errorCode int) bool {
	return s.server.handleHTTPError(c, err, errorCode)
}

func (s *serverServerAdapter) SendHTTPReply(c *gin.Context, payloadType string, jsonString string) {
	s.server.sendHTTPReply(c, payloadType, jsonString)
}

func (s *serverServerAdapter) GetServerID() (string, error) {
	return s.server.getServerID()
}

func (s *serverServerAdapter) Validator() serverhandlers.Validator {
	return s.adapter.ServerValidator()
}

func (s *serverServerAdapter) Controller() serverhandlers.Controller {
	return s.adapter.ServerController()
}

func (s *ServerAdapter) ServerServer() serverhandlers.ColoniesServer {
	return &serverServerAdapter{
		server: s.server,
		adapter: s,
	}
}