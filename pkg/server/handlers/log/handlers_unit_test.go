package log

import (
	"errors"
	"net/http"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// Mock implementations

type MockContext struct {
	statusCode int
	response   interface{}
}

func (m *MockContext) String(code int, format string, values ...interface{}) {
	m.statusCode = code
}

func (m *MockContext) JSON(code int, obj interface{}) {
	m.statusCode = code
	m.response = obj
}

func (m *MockContext) XML(code int, obj interface{}) {
	m.statusCode = code
}

func (m *MockContext) Data(code int, contentType string, data []byte) {
	m.statusCode = code
}

func (m *MockContext) Status(code int) {
	m.statusCode = code
}

func (m *MockContext) Request() *http.Request {
	return nil
}

func (m *MockContext) ReadBody() ([]byte, error) {
	return nil, nil
}

func (m *MockContext) GetHeader(key string) string {
	return ""
}

func (m *MockContext) Header(key, value string) {}

func (m *MockContext) Param(key string) string {
	return ""
}

func (m *MockContext) Query(key string) string {
	return ""
}

func (m *MockContext) DefaultQuery(key, defaultValue string) string {
	return defaultValue
}

func (m *MockContext) PostForm(key string) string {
	return ""
}

func (m *MockContext) DefaultPostForm(key, defaultValue string) string {
	return defaultValue
}

func (m *MockContext) Bind(obj interface{}) error {
	return nil
}

func (m *MockContext) ShouldBind(obj interface{}) error {
	return nil
}

func (m *MockContext) BindJSON(obj interface{}) error {
	return nil
}

func (m *MockContext) ShouldBindJSON(obj interface{}) error {
	return nil
}

func (m *MockContext) Set(key string, value interface{}) {}

func (m *MockContext) Get(key string) (interface{}, bool) {
	return nil, false
}

func (m *MockContext) GetString(key string) string {
	return ""
}

func (m *MockContext) GetBool(key string) bool {
	return false
}

func (m *MockContext) GetInt(key string) int {
	return 0
}

func (m *MockContext) GetInt64(key string) int64 {
	return 0
}

func (m *MockContext) GetFloat64(key string) float64 {
	return 0
}

func (m *MockContext) Abort() {}

func (m *MockContext) AbortWithStatus(code int) {
	m.statusCode = code
}

func (m *MockContext) AbortWithStatusJSON(code int, jsonObj interface{}) {
	m.statusCode = code
	m.response = jsonObj
}

func (m *MockContext) IsAborted() bool {
	return false
}

func (m *MockContext) Next() {}

type MockValidator struct {
	requireMembershipErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID string, serverID string) error {
	return nil
}

func (m *MockValidator) RequireColonyOwner(recoveredID string, colonyName string) error {
	return nil
}

func (m *MockValidator) RequireMembership(recoveredID string, colonyName string, approved bool) error {
	return m.requireMembershipErr
}

type MockExecutorDB struct {
	executor    *core.Executor
	executorErr error
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	return nil
}

func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error) {
	return nil, nil
}

func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error) {
	return m.executor, m.executorErr
}

func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	return nil, nil
}

func (m *MockExecutorDB) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	return m.executor, m.executorErr
}

func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	return nil, nil
}

func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) RemoveExecutorByName(colonyName string, executorName string) error {
	return nil
}

func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error {
	return nil
}

func (m *MockExecutorDB) CountExecutors() (int, error) {
	return 0, nil
}

func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error) {
	return 0, nil
}

func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	return 0, nil
}

func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName string, executorName string, capabilities core.Capabilities) error {
	return nil
}

type MockProcessDB struct {
	process    *core.Process
	processErr error
}

func (m *MockProcessDB) AddProcess(process *core.Process) error {
	return nil
}

func (m *MockProcessDB) GetProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error) {
	return m.process, m.processErr
}

func (m *MockProcessDB) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindAllRunningProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindAllWaitingProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) RemoveProcessByID(processID string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllProcesses() error {
	return nil
}

func (m *MockProcessDB) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllProcessesByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	return nil
}

func (m *MockProcessDB) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (m *MockProcessDB) ResetProcess(process *core.Process) error {
	return nil
}

func (m *MockProcessDB) SetInput(processID string, output []interface{}) error {
	return nil
}

func (m *MockProcessDB) SetOutput(processID string, output []interface{}) error {
	return nil
}

func (m *MockProcessDB) SetErrors(processID string, errs []string) error {
	return nil
}

func (m *MockProcessDB) SetProcessState(processID string, state int) error {
	return nil
}

func (m *MockProcessDB) SetParents(processID string, parents []string) error {
	return nil
}

func (m *MockProcessDB) SetChildren(processID string, children []string) error {
	return nil
}

func (m *MockProcessDB) SetWaitForParents(processID string, waitingForParent bool) error {
	return nil
}

func (m *MockProcessDB) Assign(executorID string, process *core.Process) error {
	return nil
}

func (m *MockProcessDB) SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error) {
	return nil, nil
}

func (m *MockProcessDB) Unassign(process *core.Process) error {
	return nil
}

func (m *MockProcessDB) MarkSuccessful(processID string) (float64, float64, error) {
	return 0, 0, nil
}

func (m *MockProcessDB) MarkFailed(processID string, errs []string) error {
	return nil
}

func (m *MockProcessDB) CountProcesses() (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountWaitingProcesses() (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountRunningProcesses() (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountSuccessfulProcesses() (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountFailedProcesses() (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	return 0, nil
}

func (m *MockProcessDB) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	return 0, nil
}

type MockLogDB struct {
	logs      []*core.Log
	addLogErr error
	getLogErr error
}

func (m *MockLogDB) AddLog(processID string, colonyName string, executorName string, timestamp int64, message string) error {
	return m.addLogErr
}

func (m *MockLogDB) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) GetLogsByProcessIDLatest(processID string, limit int) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) GetLogsByExecutorLatest(executorName string, limit int) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

func (m *MockLogDB) RemoveLogsByColonyName(colonyName string) error {
	return nil
}

func (m *MockLogDB) CountLogs(colonyName string) (int, error) {
	return 0, nil
}

func (m *MockLogDB) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	return m.logs, m.getLogErr
}

type MockServer struct {
	validator    *MockValidator
	executorDB   *MockExecutorDB
	processDB    *MockProcessDB
	logDB        *MockLogDB
	httpError    bool
	lastErrCode  int
	lastResponse string
}

func (m *MockServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	if err != nil {
		m.lastErrCode = errorCode
		m.httpError = true
		return true
	}
	return false
}

func (m *MockServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	m.lastResponse = jsonString
}

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	m.lastResponse = ""
}

func (m *MockServer) Validator() security.Validator {
	return m.validator
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) ProcessDB() database.ProcessDatabase {
	return m.processDB
}

func (m *MockServer) LogDB() database.LogDatabase {
	return m.logDB
}

func createMockServer() *MockServer {
	return &MockServer{
		validator:  &MockValidator{},
		executorDB: &MockExecutorDB{},
		processDB:  &MockProcessDB{},
		logDB:      &MockLogDB{},
	}
}

// Tests

func TestNewHandlersUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	assert.NotNil(t, handlers)
}

func TestRegisterHandlersUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	handlerRegistry := registry.NewHandlerRegistry()

	err := handlers.RegisterHandlers(handlerRegistry)
	assert.Nil(t, err)
}

func TestHandleAddLogInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleAddLogMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddLogProcessNilUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddLogProcessNotRunningUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID:    "process-id",
		State: core.WAITING,
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddLogNotAssignedExecutorUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID:                 "process-id",
		State:              core.RUNNING,
		AssignedExecutorID: "other-executor-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddLogSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID:                 "process-id",
		State:              core.RUNNING,
		AssignedExecutorID: "test-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleAddLogMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	server.processDB.process = &core.Process{
		ID:    "process-id",
		State: core.RUNNING,
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddLogExecutorDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID:    "process-id",
		State: core.RUNNING,
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.executorDB.executorErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddLogDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID:                 "process-id",
		State:              core.RUNNING,
		AssignedExecutorID: "test-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	server.logDB.addLogErr = errors.New("failed to add log")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddLogMsg("process-id", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddLog(ctx, "test-id", rpc.AddLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorLogInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorLogMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorLogMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddExecutorLogExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddExecutorLogNameMismatchUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "different-executor",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddExecutorLogSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleAddExecutorLogDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:   "test-id",
		Name: "test-executor",
	}
	server.logDB.addLogErr = errors.New("failed to add log")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorLogMsg("test-colony", "test-executor", "test message")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutorLog(ctx, "test-id", rpc.AddExecutorLogPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleGetLogsMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsByExecutorNotFoundUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "", 100, 0)
	msg.ExecutorName = "test-executor"
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsByProcessNotFoundUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsExceedsMaxCountUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", MAX_LOG_COUNT+1, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsByProcessSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleGetLogsByExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "", 100, 0)
	msg.ExecutorName = "test-executor"
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleGetLogsByExecutorMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "", 100, 0)
	msg.ExecutorName = "test-executor"
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleGetLogsByProcessMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleGetLogsDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.logDB.getLogErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetLogsByProcessLatestUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 0)
	msg.Latest = true
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleGetLogsByProcessSinceUnit(t *testing.T) {
	server := createMockServer()
	server.processDB.process = &core.Process{
		ID: "process-id",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "process-id", 100, 12345)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleGetLogsByExecutorLatestUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "", 100, 0)
	msg.ExecutorName = "test-executor"
	msg.Latest = true
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleGetLogsByExecutorSinceUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetLogsMsg("test-colony", "", 100, 12345)
	msg.ExecutorName = "test-executor"
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetLogs(ctx, "test-id", rpc.GetLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleSearchLogsInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleSearchLogsMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", 1, 10)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleSearchLogsMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", 1, 10)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleSearchLogsExceedsMaxCountUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", 1, MAX_COUNT+1)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleSearchLogsExceedsMaxDaysUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", MAX_DAYS+1, 10)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleSearchLogsSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.logDB.logs = []*core.Log{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", 1, 10)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

func TestHandleSearchLogsDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.logDB.getLogErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateSearchLogsMsg("test-colony", "ERROR", 1, 10)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleSearchLogs(ctx, "test-id", rpc.SearchLogsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}
