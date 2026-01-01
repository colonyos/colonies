package server

import (
	"errors"
	"net/http"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// MockColonyDB implements database.ColonyDatabase
type MockColonyDB struct {
	countColoniesErr   error
	coloniesCount      int
}

func (m *MockColonyDB) AddColony(colony *core.Colony) error               { return nil }
func (m *MockColonyDB) GetColonies() ([]*core.Colony, error)              { return nil, nil }
func (m *MockColonyDB) GetColonyByID(id string) (*core.Colony, error)     { return nil, nil }
func (m *MockColonyDB) GetColonyByName(name string) (*core.Colony, error) { return nil, nil }
func (m *MockColonyDB) RenameColony(colonyName string, newColonyName string) error { return nil }
func (m *MockColonyDB) RemoveColonyByName(colonyName string) error        { return nil }

func (m *MockColonyDB) CountColonies() (int, error) {
	if m.countColoniesErr != nil {
		return 0, m.countColoniesErr
	}
	return m.coloniesCount, nil
}

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	countExecutorsErr error
	executorsCount    int
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                                            { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error   { return nil }
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                                              { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error)                 { return nil, nil }
func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error)            { return nil, nil }
func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error)                            { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error)               { return nil, nil }
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error                                        { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error                                         { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error                                              { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error                           { return nil }
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error                                  { return nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error)                            { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error)         { return 0, nil }
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName, executorName string, cap core.Capabilities) error { return nil }

func (m *MockExecutorDB) CountExecutors() (int, error) {
	if m.countExecutorsErr != nil {
		return 0, m.countExecutorsErr
	}
	return m.executorsCount, nil
}

// MockProcessDB implements database.ProcessDatabase
type MockProcessDB struct {
	countWaitingErr   error
	countRunningErr   error
	countSuccessErr   error
	countFailedErr    error
	waitingCount      int
	runningCount      int
	successCount      int
	failedCount       int
}

func (m *MockProcessDB) AddProcess(process *core.Process) error                                   { return nil }
func (m *MockProcessDB) GetProcesses() ([]*core.Process, error)                                   { return nil, nil }
func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error)                  { return nil, nil }
func (m *MockProcessDB) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindAllRunningProcesses() ([]*core.Process, error)                        { return nil, nil }
func (m *MockProcessDB) FindAllWaitingProcesses() ([]*core.Process, error)                        { return nil, nil }
func (m *MockProcessDB) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) RemoveProcessByID(processID string) error                                 { return nil }
func (m *MockProcessDB) RemoveAllProcesses() error                                                { return nil }
func (m *MockProcessDB) RemoveAllWaitingProcessesByColonyName(colonyName string) error            { return nil }
func (m *MockProcessDB) RemoveAllRunningProcessesByColonyName(colonyName string) error            { return nil }
func (m *MockProcessDB) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error         { return nil }
func (m *MockProcessDB) RemoveAllFailedProcessesByColonyName(colonyName string) error             { return nil }
func (m *MockProcessDB) RemoveAllProcessesByColonyName(colonyName string) error                   { return nil }
func (m *MockProcessDB) RemoveAllProcessesByProcessGraphID(processGraphID string) error           { return nil }
func (m *MockProcessDB) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error    { return nil }
func (m *MockProcessDB) ResetProcess(process *core.Process) error                                 { return nil }
func (m *MockProcessDB) SetInput(processID string, output []interface{}) error                    { return nil }
func (m *MockProcessDB) SetOutput(processID string, output []interface{}) error                   { return nil }
func (m *MockProcessDB) SetErrors(processID string, errs []string) error                          { return nil }
func (m *MockProcessDB) SetProcessState(processID string, state int) error                        { return nil }
func (m *MockProcessDB) SetParents(processID string, parents []string) error                      { return nil }
func (m *MockProcessDB) SetChildren(processID string, children []string) error                    { return nil }
func (m *MockProcessDB) SetWaitForParents(processID string, waitingForParent bool) error          { return nil }
func (m *MockProcessDB) Assign(executorID string, process *core.Process) error                    { return nil }
func (m *MockProcessDB) SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error) { return nil, nil }
func (m *MockProcessDB) Unassign(process *core.Process) error                                     { return nil }
func (m *MockProcessDB) MarkSuccessful(processID string) (float64, float64, error)                { return 0, 0, nil }
func (m *MockProcessDB) MarkFailed(processID string, errs []string) error                         { return nil }
func (m *MockProcessDB) CountProcesses() (int, error)                                             { return 0, nil }
func (m *MockProcessDB) CountWaitingProcessesByColonyName(colonyName string) (int, error)         { return 0, nil }
func (m *MockProcessDB) CountRunningProcessesByColonyName(colonyName string) (int, error)         { return 0, nil }
func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(colonyName string) (int, error)      { return 0, nil }
func (m *MockProcessDB) CountFailedProcessesByColonyName(colonyName string) (int, error)          { return 0, nil }

func (m *MockProcessDB) CountWaitingProcesses() (int, error) {
	if m.countWaitingErr != nil {
		return 0, m.countWaitingErr
	}
	return m.waitingCount, nil
}

func (m *MockProcessDB) CountRunningProcesses() (int, error) {
	if m.countRunningErr != nil {
		return 0, m.countRunningErr
	}
	return m.runningCount, nil
}

func (m *MockProcessDB) CountSuccessfulProcesses() (int, error) {
	if m.countSuccessErr != nil {
		return 0, m.countSuccessErr
	}
	return m.successCount, nil
}

func (m *MockProcessDB) CountFailedProcesses() (int, error) {
	if m.countFailedErr != nil {
		return 0, m.countFailedErr
	}
	return m.failedCount, nil
}

// MockEtcdServer implements EtcdServer interface
type MockEtcdServer struct {
	clusterConfig cluster.Config
}

func (m *MockEtcdServer) CurrentCluster() cluster.Config {
	return m.clusterConfig
}

// MockController implements Controller interface
type MockController struct {
	etcdServer *MockEtcdServer
}

func (m *MockController) GetEtcdServer() EtcdServer {
	return m.etcdServer
}

// MockValidator implements Validator interface
type MockValidator struct {
	serverOwnerErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID string, serverID string) error {
	return m.serverOwnerErr
}

// MockContext implements backends.Context
type MockContext struct {
	aborted               bool
	abortedWithStatus     int
	abortedWithStatusJSON int
	jsonResponse          interface{}
}

func (m *MockContext) String(code int, format string, values ...interface{}) {}
func (m *MockContext) JSON(code int, obj interface{})                        { m.jsonResponse = obj }
func (m *MockContext) XML(code int, obj interface{})                         {}
func (m *MockContext) Data(code int, contentType string, data []byte)        {}
func (m *MockContext) Status(code int)                                       {}
func (m *MockContext) Request() *http.Request                                { return nil }
func (m *MockContext) ReadBody() ([]byte, error)                             { return nil, nil }
func (m *MockContext) GetHeader(key string) string                           { return "" }
func (m *MockContext) Header(key, value string)                              {}
func (m *MockContext) Param(key string) string                               { return "" }
func (m *MockContext) Query(key string) string                               { return "" }
func (m *MockContext) DefaultQuery(key, defaultValue string) string          { return defaultValue }
func (m *MockContext) PostForm(key string) string                            { return "" }
func (m *MockContext) DefaultPostForm(key, defaultValue string) string       { return defaultValue }
func (m *MockContext) Bind(obj interface{}) error                            { return nil }
func (m *MockContext) ShouldBind(obj interface{}) error                      { return nil }
func (m *MockContext) BindJSON(obj interface{}) error                        { return nil }
func (m *MockContext) ShouldBindJSON(obj interface{}) error                  { return nil }
func (m *MockContext) Set(key string, value interface{})                     {}
func (m *MockContext) Get(key string) (value interface{}, exists bool)       { return nil, false }
func (m *MockContext) GetString(key string) string                           { return "" }
func (m *MockContext) GetBool(key string) bool                               { return false }
func (m *MockContext) GetInt(key string) int                                 { return 0 }
func (m *MockContext) GetInt64(key string) int64                             { return 0 }
func (m *MockContext) GetFloat64(key string) float64                         { return 0 }
func (m *MockContext) Abort()                                                { m.aborted = true }
func (m *MockContext) AbortWithStatus(code int) {
	m.abortedWithStatus = code
	m.aborted = true
}
func (m *MockContext) AbortWithStatusJSON(code int, jsonObj interface{}) {
	m.abortedWithStatusJSON = code
	m.jsonResponse = jsonObj
	m.aborted = true
}
func (m *MockContext) IsAborted() bool { return m.aborted }
func (m *MockContext) Next()           {}

// MockServer implements Server interface
type MockServer struct {
	colonyDB        *MockColonyDB
	executorDB      *MockExecutorDB
	processDB       *MockProcessDB
	controller      *MockController
	validator       *MockValidator
	serverID        string
	serverIDErr     error
	lastError       error
	lastStatusCode  int
	lastPayloadType string
	lastResponse    string
}

func (m *MockServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	if err != nil {
		m.lastError = err
		m.lastStatusCode = errorCode
		c.AbortWithStatusJSON(errorCode, map[string]string{"error": err.Error()})
		return true
	}
	return false
}

func (m *MockServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	m.lastPayloadType = payloadType
	m.lastResponse = jsonString
	c.JSON(http.StatusOK, map[string]string{"response": jsonString})
}

func (m *MockServer) GetServerID() (string, error) {
	return m.serverID, m.serverIDErr
}

func (m *MockServer) Validator() Validator {
	return m.validator
}

func (m *MockServer) Controller() Controller {
	return m.controller
}

func (m *MockServer) ColonyDB() database.ColonyDatabase {
	return m.colonyDB
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) ProcessDB() database.ProcessDatabase {
	return m.processDB
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	colonyDB := &MockColonyDB{coloniesCount: 5}
	executorDB := &MockExecutorDB{executorsCount: 10}
	processDB := &MockProcessDB{
		waitingCount: 20,
		runningCount: 15,
		successCount: 100,
		failedCount:  5,
	}
	etcdServer := &MockEtcdServer{
		clusterConfig: cluster.Config{},
	}
	controller := &MockController{etcdServer: etcdServer}
	validator := &MockValidator{}

	server := &MockServer{
		colonyDB:   colonyDB,
		executorDB: executorDB,
		processDB:  processDB,
		controller: controller,
		validator:  validator,
		serverID:   "server-123",
	}

	ctx := &MockContext{}
	return server, ctx
}

// Tests for RegisterHandlers
func TestRegisterHandlers(t *testing.T) {
	server, _ := createMockServer()
	handlers := NewHandlers(server)
	reg := registry.NewHandlerRegistry()

	err := handlers.RegisterHandlers(reg)
	assert.Nil(t, err)
}

// Tests for HandleStatistics
func TestHandleStatistics_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, rpc.GetStatisiticsPayloadType, server.lastPayloadType)
}

func TestHandleStatistics_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleStatistics_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleStatistics_GetServerIDError(t *testing.T) {
	server, ctx := createMockServer()
	server.serverIDErr = errors.New("server id error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_ServerOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.serverOwnerErr = errors.New("server owner error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleStatistics_CountColoniesError(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.countColoniesErr = errors.New("count colonies error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_CountExecutorsError(t *testing.T) {
	server, ctx := createMockServer()
	server.executorDB.countExecutorsErr = errors.New("count executors error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_CountWaitingProcessesError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.countWaitingErr = errors.New("count waiting error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_CountRunningProcessesError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.countRunningErr = errors.New("count running error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_CountSuccessfulProcessesError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.countSuccessErr = errors.New("count success error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleStatistics_CountFailedProcessesError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.countFailedErr = errors.New("count failed error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetStatisticsMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleStatistics(ctx, "server-owner", rpc.GetStatisiticsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetCluster
func TestHandleGetCluster_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetClusterMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetCluster(ctx, "server-owner", rpc.GetClusterPayloadType, jsonString)

	assert.Equal(t, rpc.GetClusterPayloadType, server.lastPayloadType)
}

func TestHandleGetCluster_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetCluster(ctx, "server-owner", rpc.GetClusterPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetCluster_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetClusterMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetCluster(ctx, "server-owner", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetCluster_GetServerIDError(t *testing.T) {
	server, ctx := createMockServer()
	server.serverIDErr = errors.New("server id error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetClusterMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetCluster(ctx, "server-owner", rpc.GetClusterPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleGetCluster_ServerOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.serverOwnerErr = errors.New("server owner error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetClusterMsg()
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetCluster(ctx, "server-owner", rpc.GetClusterPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleVersion
func TestHandleVersion_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateVersionMsg("", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleVersion(ctx, rpc.VersionPayloadType, jsonString)

	assert.Equal(t, rpc.VersionPayloadType, server.lastPayloadType)
}

func TestHandleVersion_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleVersion(ctx, rpc.VersionPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleVersion_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateVersionMsg("", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleVersion(ctx, "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}
