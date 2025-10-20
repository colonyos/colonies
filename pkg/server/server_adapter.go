package server

import (
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/backends/gin"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/controllers"
	attributehandlers "github.com/colonyos/colonies/pkg/server/handlers/attribute"
	"github.com/colonyos/colonies/pkg/server/handlers/colony"
	"github.com/colonyos/colonies/pkg/server/handlers/executor"
	functionhandlers "github.com/colonyos/colonies/pkg/server/handlers/function"
	generatorhandlers "github.com/colonyos/colonies/pkg/server/handlers/generator"
	loghandlers "github.com/colonyos/colonies/pkg/server/handlers/log"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/colonyos/colonies/pkg/server/handlers/processgraph"
	serverhandlers "github.com/colonyos/colonies/pkg/server/handlers/server"
	realtimehandlers "github.com/colonyos/colonies/pkg/server/handlers/realtime"
)

// ServerAdapter implements interfaces needed by handler packages
type ServerAdapter struct {
	server *Server
}

func NewServerAdapter(server *Server) *ServerAdapter {
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
	AddColony(colony *core.Colony) (*core.Colony, error)
	RemoveColony(colonyName string) error
	GetColonies() ([]*core.Colony, error)
	GetColony(colonyName string) (*core.Colony, error)
	GetColonyStatistics(colonyName string) (*core.Statistics, error)
} {
	return s.server.controller
}

func (s *ServerAdapter) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	return s.server.HandleHTTPError(c, err, errorCode)
}

func (s *ServerAdapter) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	s.server.SendHTTPReply(c, payloadType, jsonString)
}

func (s *ServerAdapter) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	s.server.SendEmptyHTTPReply(c, payloadType)
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
		AddColony(colony *core.Colony) (*core.Colony, error)
		RemoveColony(colonyName string) error
		GetColonies() ([]*core.Colony, error)
		GetColony(colonyName string) (*core.Colony, error)
		GetColonyStatistics(colonyName string) (*core.Statistics, error)
	}
}

func (c *controllerAdapter) AddColony(colony *core.Colony) (*core.Colony, error) {
	return c.controller.AddColony(colony)
}

func (c *controllerAdapter) RemoveColony(colonyName string) error {
	return c.controller.RemoveColony(colonyName)
}

func (c *controllerAdapter) GetColonies() ([]*core.Colony, error) {
	return c.controller.GetColonies()
}

func (c *controllerAdapter) GetColony(colonyName string) (*core.Colony, error) {
	return c.controller.GetColony(colonyName)
}

func (c *controllerAdapter) GetColonyStatistics(colonyName string) (*core.Statistics, error) {
	return c.controller.GetColonyStatistics(colonyName)
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
		AddExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error)
		GetExecutor(executorID string) (*core.Executor, error)
		GetExecutorByColonyName(colonyName string) ([]*core.Executor, error)
	}
}

func (c *executorControllerAdapter) AddExecutor(executor *core.Executor, allowReregister bool) (*core.Executor, error) {
	return c.controller.AddExecutor(executor, allowReregister)
}

func (c *executorControllerAdapter) GetExecutor(executorID string) (*core.Executor, error) {
	return c.controller.GetExecutor(executorID)
}

func (c *executorControllerAdapter) GetExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	return c.controller.GetExecutorByColonyName(colonyName)
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

func (s *ServerAdapter) ResourceDB() database.ResourceDatabase {
	return s.server.resourceDB
}

type attributeControllerAdapter struct {
	controller interface {
		GetProcess(processID string) (*core.Process, error)
		AddAttribute(attribute *core.Attribute) (*core.Attribute, error)
		GetAttribute(attributeID string) (*core.Attribute, error)
	}
}

func (c *attributeControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.GetProcess(processID)
}

func (c *attributeControllerAdapter) AddAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	return c.controller.AddAttribute(attribute)
}

func (c *attributeControllerAdapter) GetAttribute(attributeID string) (*core.Attribute, error) {
	return c.controller.GetAttribute(attributeID)
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
		AddFunction(function *core.Function) (*core.Function, error)
		GetFunctionByID(functionID string) (*core.Function, error)
		GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
		GetFunctionsByColonyName(colonyName string) ([]*core.Function, error)
		RemoveFunction(functionID string) error
	}
}

type generatorControllerAdapter struct {
	controller interface {
		AddGenerator(generator *core.Generator) (*core.Generator, error)
		GetGenerator(generatorID string) (*core.Generator, error)
		ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
		GetGenerators(colonyName string, count int) ([]*core.Generator, error)
		PackGenerator(generatorID string, colonyName string, arg string) error
		RemoveGenerator(generatorID string) error
		GetGeneratorPeriod() int
	}
}

func (c *functionControllerAdapter) AddFunction(function *core.Function) (*core.Function, error) {
	return c.controller.AddFunction(function)
}

func (c *functionControllerAdapter) GetFunction(functionID string) (*core.Function, error) {
	return c.controller.GetFunctionByID(functionID)
}

func (c *functionControllerAdapter) GetFunctions(colonyName string, executorName string, count int) ([]*core.Function, error) {
	return c.controller.GetFunctionsByExecutorName(colonyName, executorName)
}

func (c *functionControllerAdapter) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	return c.controller.GetFunctionsByColonyName(colonyName)
}

func (c *functionControllerAdapter) RemoveFunction(functionID string, initiatorID string) error {
	return c.controller.RemoveFunction(functionID)
}

func (c *generatorControllerAdapter) AddGenerator(generator *core.Generator) (*core.Generator, error) {
	return c.controller.AddGenerator(generator)
}

func (c *generatorControllerAdapter) GetGenerator(generatorID string) (*core.Generator, error) {
	return c.controller.GetGenerator(generatorID)
}

func (c *generatorControllerAdapter) ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error) {
	return c.controller.ResolveGenerator(colonyName, generatorName)
}

func (c *generatorControllerAdapter) GetGenerators(colonyName string, count int) ([]*core.Generator, error) {
	return c.controller.GetGenerators(colonyName, count)
}

func (c *generatorControllerAdapter) PackGenerator(generatorID string, colonyName string, arg string) error {
	return c.controller.PackGenerator(generatorID, colonyName, arg)
}

func (c *generatorControllerAdapter) RemoveGenerator(generatorID string) error {
	return c.controller.RemoveGenerator(generatorID)
}

func (c *generatorControllerAdapter) GetGeneratorPeriod() int {
	return c.controller.GetGeneratorPeriod()
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
		AddProcessToDB(process *core.Process) (*core.Process, error)
		AddProcess(process *core.Process) (*core.Process, error)
		GetProcess(processID string) (*core.Process, error)
		GetExecutor(executorID string) (*core.Executor, error)
		FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
		RemoveProcess(processID string) error
		RemoveAllProcesses(colonyName string, state int) error
		SetOutput(processID string, output []interface{}) error
		CloseSuccessful(processID string, executorID string, output []interface{}) error
		CloseFailed(processID string, errs []string) error
		Assign(executorID string, colonyName string, cpu int64, memory int64) (*controllers.AssignResult, error)
		UnassignExecutor(processID string) error
		PauseColonyAssignments(colonyName string) error
		ResumeColonyAssignments(colonyName string) error
		AreColonyAssignmentsPaused(colonyName string) (bool, error)
		GetEventHandler() backends.RealtimeEventHandler
		IsLeader() bool
		GetEtcdServer() *cluster.EtcdServer
	}
}

func (c *processControllerAdapter) AddProcessToDB(process *core.Process) (*core.Process, error) {
	return c.controller.AddProcessToDB(process)
}

func (c *processControllerAdapter) AddProcess(process *core.Process) (*core.Process, error) {
	return c.controller.AddProcess(process)
}

func (c *processControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.GetProcess(processID)
}

func (c *processControllerAdapter) GetExecutor(executorID string) (*core.Executor, error) {
	return c.controller.GetExecutor(executorID)
}

func (c *processControllerAdapter) FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return c.controller.FindProcessHistory(colonyName, executorID, seconds, state)
}

func (c *processControllerAdapter) RemoveProcess(processID string) error {
	return c.controller.RemoveProcess(processID)
}

func (c *processControllerAdapter) RemoveAllProcesses(colonyName string, state int) error {
	return c.controller.RemoveAllProcesses(colonyName, state)
}

func (c *processControllerAdapter) SetOutput(processID string, output []interface{}) error {
	return c.controller.SetOutput(processID, output)
}

func (c *processControllerAdapter) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	return c.controller.CloseSuccessful(processID, executorID, output)
}

func (c *processControllerAdapter) CloseFailed(processID string, errs []string) error {
	return c.controller.CloseFailed(processID, errs)
}

func (c *processControllerAdapter) Assign(executorID string, colonyName string, cpu int64, memory int64) (*process.AssignResult, error) {
	result, err := c.controller.Assign(executorID, colonyName, cpu, memory)
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
	return c.controller.UnassignExecutor(processID)
}

func (c *processControllerAdapter) PauseColonyAssignments(colonyName string) error {
	return c.controller.PauseColonyAssignments(colonyName)
}

func (c *processControllerAdapter) ResumeColonyAssignments(colonyName string) error {
	return c.controller.ResumeColonyAssignments(colonyName)
}

func (c *processControllerAdapter) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return c.controller.AreColonyAssignmentsPaused(colonyName)
}

func (c *processControllerAdapter) GetEventHandler() *process.EventHandler {
	// Convert the internal event handler to the process handler's EventHandler
	return &process.EventHandler{}
}

func (c *processControllerAdapter) IsLeader() bool {
	return c.controller.IsLeader()
}

func (c *processControllerAdapter) GetEtcdServer() process.EtcdServer {
	etcdServer := c.controller.GetEtcdServer()
	return &etcdServerAdapter{etcdServer: etcdServer}
}

type logControllerAdapter struct {
	controller interface {
		GetProcess(processID string) (*core.Process, error)
	}
}

func (c *logControllerAdapter) GetProcess(processID string) (*core.Process, error) {
	return c.controller.GetProcess(processID)
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

func (s *ServerAdapter) WSController() gin.WSController {
	return &wsControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) WSControllerCompat() WSController {
	return &wsControllerAdapter{controller: s.server.controller}
}

func (s *ServerAdapter) RealtimeHandler() realtimehandlers.RealtimeHandler {
	return s.server.realtimeHandler
}

func (s *ServerAdapter) ParseSignature(payload string, signature string) (string, error) {
	return s.server.parseSignature(payload, signature)
}

func (s *ServerAdapter) GenerateRPCErrorMsg(err error, errorCode int) (*rpc.RPCReplyMsg, error) {
	return s.server.generateRPCErrorMsg(err, errorCode)
}


// wsControllerAdapter adapter for WebSocket handlers
type wsControllerAdapter struct {
	controller interface {
		SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error
		SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error
	}
}

func (c *wsControllerAdapter) SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error {
	return c.controller.SubscribeProcesses(executorID, subscription)
}

func (c *wsControllerAdapter) SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error {
	return c.controller.SubscribeProcess(executorID, subscription)
}

// Cron controller adapter
type cronControllerAdapter struct {
	controller interface {
		AddCron(cron *core.Cron) (*core.Cron, error)
		GetCron(cronID string) (*core.Cron, error)
		GetCrons(colonyName string, count int) ([]*core.Cron, error)
		RunCron(cronID string) (*core.Cron, error)
		RemoveCron(cronID string) error
		GetCronPeriod() int
	}
}

func (c *cronControllerAdapter) AddCron(cron *core.Cron) (*core.Cron, error) {
	return c.controller.AddCron(cron)
}

func (c *cronControllerAdapter) GetCron(cronID string) (*core.Cron, error) {
	return c.controller.GetCron(cronID)
}

func (c *cronControllerAdapter) GetCrons(colonyName string, count int) ([]*core.Cron, error) {
	return c.controller.GetCrons(colonyName, count)
}

func (c *cronControllerAdapter) RunCron(cronID string) (*core.Cron, error) {
	return c.controller.RunCron(cronID)
}

func (c *cronControllerAdapter) RemoveCron(cronID string) error {
	return c.controller.RemoveCron(cronID)
}

func (c *cronControllerAdapter) GetCronPeriod() int {
	return c.controller.GetCronPeriod()
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
		SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error)
		GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
		FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
		RemoveProcessGraph(processGraphID string) error
		RemoveAllProcessGraphs(colonyName string, state int) error
		AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error)
	}
}

func (c *processgraphControllerAdapter) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error) {
	return c.controller.SubmitWorkflowSpec(workflowSpec, initiatorID)
}

func (c *processgraphControllerAdapter) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return c.controller.GetProcessGraphByID(processGraphID)
}

func (c *processgraphControllerAdapter) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.FindWaitingProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.FindRunningProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.FindSuccessfulProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return c.controller.FindFailedProcessGraphs(colonyName, count)
}

func (c *processgraphControllerAdapter) RemoveProcessGraph(processGraphID string) error {
	return c.controller.RemoveProcessGraph(processGraphID)
}

func (c *processgraphControllerAdapter) RemoveAllProcessGraphs(colonyName string, state int) error {
	return c.controller.RemoveAllProcessGraphs(colonyName, state)
}

func (c *processgraphControllerAdapter) AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error) {
	return c.controller.AddChild(processGraphID, parentProcessID, childProcessID, process, initiatorID, insert)
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
	server *Server
	adapter *ServerAdapter
}

func (s *processgraphServerAdapter) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	return s.server.HandleHTTPError(c, err, errorCode)
}

func (s *processgraphServerAdapter) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	s.server.SendHTTPReply(c, payloadType, jsonString)
}

func (s *processgraphServerAdapter) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	s.server.SendEmptyHTTPReply(c, payloadType)
}

func (s *processgraphServerAdapter) Validator() processgraph.Validator {
	return s.adapter.ProcessgraphValidator()
}

func (s *processgraphServerAdapter) Controller() processgraph.Controller {
	return s.adapter.ProcessgraphController()
}

func (s *ServerAdapter) ProcessgraphServer() processgraph.Server {
	return &processgraphServerAdapter{
		server: s.server,
		adapter: s,
	}
}

// Server handler controller adapter
type serverControllerAdapter struct {
	controller interface {
		GetStatistics() (*core.Statistics, error)
		GetEtcdServer() *cluster.EtcdServer
	}
}

func (c *serverControllerAdapter) GetStatistics() (*core.Statistics, error) {
	return c.controller.GetStatistics()
}

func (c *serverControllerAdapter) GetEtcdServer() serverhandlers.EtcdServer {
	etcdServer := c.controller.GetEtcdServer()
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
	server *Server
	adapter *ServerAdapter
}

func (s *serverServerAdapter) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	return s.server.HandleHTTPError(c, err, errorCode)
}

func (s *serverServerAdapter) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	s.server.SendHTTPReply(c, payloadType, jsonString)
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

func (s *ServerAdapter) ServerServer() serverhandlers.Server {
	return &serverServerAdapter{
		server: s.server,
		adapter: s,
	}
}