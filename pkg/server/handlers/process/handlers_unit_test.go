package process

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

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	executors     []*core.Executor
	getByIDErr    error
	returnNilByID bool
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error {
	return nil
}
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error) { return m.executors, nil }
func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, e := range m.executors {
		if e.ID == executorID {
			return e, nil
		}
	}
	return nil, nil
}
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	return nil, nil
}
func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error) {
	return nil, nil
}
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	return nil, nil
}
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error  { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error       { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error {
	return nil
}
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                        { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error) {
	return 0, nil
}
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	return 0, nil
}
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName, executorName string, capabilities core.Capabilities) error {
	return nil
}

// MockUserDB implements database.UserDatabase
type MockUserDB struct {
	users         []*core.User
	getByIDErr    error
	returnNilByID bool
}

func (m *MockUserDB) AddUser(user *core.User) error { return nil }
func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	return nil, nil
}
func (m *MockUserDB) GetUserByID(colonyName, userID string) (*core.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, u := range m.users {
		if u.ID == userID {
			return u, nil
		}
	}
	return nil, nil
}
func (m *MockUserDB) GetUserByName(colonyName, name string) (*core.User, error) { return nil, nil }
func (m *MockUserDB) RemoveUserByID(colonyName, userID string) error            { return nil }
func (m *MockUserDB) RemoveUserByName(colonyName, name string) error            { return nil }
func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error           { return nil }

// MockProcessDB implements database.ProcessDatabase
type MockProcessDB struct {
	processes     []*core.Process
	getByIDErr    error
	findErr       error
	returnNilByID bool
	setOutputErr  error
}

func (m *MockProcessDB) AddProcess(process *core.Process) error { return nil }
func (m *MockProcessDB) GetProcesses() ([]*core.Process, error) { return m.processes, nil }
func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, p := range m.processes {
		if p.ID == processID {
			return p, nil
		}
	}
	return nil, nil
}
func (m *MockProcessDB) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []*core.Process
	for _, p := range m.processes {
		if p.FunctionSpec.Conditions.ColonyName == colonyName && p.State == state {
			result = append(result, p)
		}
	}
	return result, nil
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
func (m *MockProcessDB) FindAllRunningProcesses() ([]*core.Process, error)  { return nil, nil }
func (m *MockProcessDB) FindAllWaitingProcesses() ([]*core.Process, error)  { return nil, nil }
func (m *MockProcessDB) RemoveProcessByID(processID string) error           { return nil }
func (m *MockProcessDB) RemoveAllProcesses() error                          { return nil }
func (m *MockProcessDB) RemoveAllWaitingProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllRunningProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllSuccessfulProcessesByColonyName(string) error {
	return nil
}
func (m *MockProcessDB) RemoveAllFailedProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllProcessesByColonyName(string) error       { return nil }
func (m *MockProcessDB) RemoveAllProcessesByProcessGraphID(string) error   { return nil }
func (m *MockProcessDB) RemoveAllProcessesInProcessGraphsByColonyName(string) error {
	return nil
}
func (m *MockProcessDB) ResetProcess(process *core.Process) error { return nil }
func (m *MockProcessDB) SetInput(processID string, output []interface{}) error {
	return nil
}
func (m *MockProcessDB) SetOutput(processID string, output []interface{}) error {
	if m.setOutputErr != nil {
		return m.setOutputErr
	}
	return nil
}
func (m *MockProcessDB) SetErrors(processID string, errs []string) error { return nil }
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
func (m *MockProcessDB) Unassign(process *core.Process) error { return nil }
func (m *MockProcessDB) MarkSuccessful(processID string) (float64, float64, error) {
	return 0, 0, nil
}
func (m *MockProcessDB) MarkFailed(processID string, errs []string) error { return nil }
func (m *MockProcessDB) CountProcesses() (int, error)                     { return 0, nil }
func (m *MockProcessDB) CountWaitingProcesses() (int, error)              { return 0, nil }
func (m *MockProcessDB) CountRunningProcesses() (int, error)              { return 0, nil }
func (m *MockProcessDB) CountSuccessfulProcesses() (int, error)           { return 0, nil }
func (m *MockProcessDB) CountFailedProcesses() (int, error)               { return 0, nil }
func (m *MockProcessDB) CountWaitingProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountRunningProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountFailedProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

// MockBlueprintDB implements database.BlueprintDatabase
type MockBlueprintDB struct{}

func (m *MockBlueprintDB) AddBlueprintDefinition(sd *core.BlueprintDefinition) error { return nil }
func (m *MockBlueprintDB) GetBlueprintDefinitionByID(id string) (*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintDefinitionByName(namespace, name string) (*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintDefinitions() ([]*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintDefinitionsByNamespace(namespace string) ([]*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintDefinitionsByGroup(group string) ([]*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintDefinitionByKind(kind string) (*core.BlueprintDefinition, error) {
	return nil, nil
}
func (m *MockBlueprintDB) UpdateBlueprintDefinition(sd *core.BlueprintDefinition) error { return nil }
func (m *MockBlueprintDB) RemoveBlueprintDefinitionByID(id string) error                { return nil }
func (m *MockBlueprintDB) RemoveBlueprintDefinitionByName(namespace, name string) error { return nil }
func (m *MockBlueprintDB) CountBlueprintDefinitions() (int, error)                      { return 0, nil }
func (m *MockBlueprintDB) AddBlueprint(blueprint *core.Blueprint) error                 { return nil }
func (m *MockBlueprintDB) GetBlueprintByID(id string) (*core.Blueprint, error)          { return nil, nil }
func (m *MockBlueprintDB) GetBlueprintByName(namespace, name string) (*core.Blueprint, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprints() ([]*core.Blueprint, error) { return nil, nil }
func (m *MockBlueprintDB) GetBlueprintsByNamespace(namespace string) ([]*core.Blueprint, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintsByKind(kind string) ([]*core.Blueprint, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintsByNamespaceAndKind(namespace, kind string) ([]*core.Blueprint, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintsByNamespaceKindAndLocation(namespace, kind, locationName string) ([]*core.Blueprint, error) {
	return nil, nil
}
func (m *MockBlueprintDB) UpdateBlueprint(blueprint *core.Blueprint) error { return nil }
func (m *MockBlueprintDB) UpdateBlueprintStatus(id string, status map[string]interface{}) error {
	return nil
}
func (m *MockBlueprintDB) RemoveBlueprintByID(id string) error                       { return nil }
func (m *MockBlueprintDB) RemoveBlueprintByName(namespace, name string) error        { return nil }
func (m *MockBlueprintDB) RemoveBlueprintsByNamespace(namespace string) error        { return nil }
func (m *MockBlueprintDB) CountBlueprints() (int, error)                             { return 0, nil }
func (m *MockBlueprintDB) CountBlueprintsByNamespace(namespace string) (int, error)  { return 0, nil }
func (m *MockBlueprintDB) AddBlueprintHistory(history *core.BlueprintHistory) error  { return nil }
func (m *MockBlueprintDB) GetBlueprintHistory(blueprintID string, limit int) ([]*core.BlueprintHistory, error) {
	return nil, nil
}
func (m *MockBlueprintDB) GetBlueprintHistoryByGeneration(blueprintID string, generation int64) (*core.BlueprintHistory, error) {
	return nil, nil
}
func (m *MockBlueprintDB) RemoveBlueprintHistory(blueprintID string) error { return nil }

// MockController implements Controller interface
type MockController struct {
	addProcessFunc              func(*core.Process) (*core.Process, error)
	removeProcessErr            error
	removeAllProcessesErr       error
	closeSuccessfulErr          error
	closeFailedErr              error
	pauseAssignmentsErr         error
	resumeAssignmentsErr        error
	pauseStatusResult           bool
	pauseStatusErr              error
	isLeader                    bool
}

func (m *MockController) AddProcessToDB(process *core.Process) (*core.Process, error) {
	return process, nil
}

func (m *MockController) AddProcess(process *core.Process) (*core.Process, error) {
	if m.addProcessFunc != nil {
		return m.addProcessFunc(process)
	}
	process.ID = "process-123"
	return process, nil
}

func (m *MockController) RemoveProcess(processID string) error {
	return m.removeProcessErr
}

func (m *MockController) RemoveAllProcesses(colonyName string, state int) error {
	return m.removeAllProcessesErr
}

func (m *MockController) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	return m.closeSuccessfulErr
}

func (m *MockController) CloseFailed(processID string, errs []string) error {
	return m.closeFailedErr
}

func (m *MockController) Assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error) {
	return nil, nil
}

func (m *MockController) DistributedAssign(executor *core.Executor, colonyName string, cpu int64, memory int64, storage int64) (*AssignResult, error) {
	return nil, nil
}

func (m *MockController) UnassignExecutor(processID string) error { return nil }

func (m *MockController) PauseColonyAssignments(colonyName string) error {
	return m.pauseAssignmentsErr
}

func (m *MockController) ResumeColonyAssignments(colonyName string) error {
	return m.resumeAssignmentsErr
}

func (m *MockController) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return m.pauseStatusResult, m.pauseStatusErr
}

func (m *MockController) GetEventHandler() *EventHandler {
	return nil
}

func (m *MockController) IsLeader() bool {
	return m.isLeader
}

func (m *MockController) GetEtcdServer() EtcdServer {
	return nil
}

// MockValidator implements security.Validator
type MockValidator struct {
	requireMembershipErr  error
	requireColonyOwnerErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID, serverID string) error { return nil }
func (m *MockValidator) RequireColonyOwner(recoveredID, colonyName string) error {
	return m.requireColonyOwnerErr
}
func (m *MockValidator) RequireMembership(recoveredID, colonyName string, approved bool) error {
	return m.requireMembershipErr
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

// MockServer implements the Server interface for process handlers
type MockServer struct {
	executorDB        *MockExecutorDB
	userDB            *MockUserDB
	processDB         *MockProcessDB
	blueprintDB       *MockBlueprintDB
	validator         *MockValidator
	controller        *MockController
	httpErrorCode     int
	replyPayload      string
	replyType         string
	exclusiveAssign   bool
	tls               bool
}

func (m *MockServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	if err != nil {
		m.httpErrorCode = errorCode
		c.AbortWithStatusJSON(errorCode, map[string]string{"error": err.Error()})
		return true
	}
	return false
}

func (m *MockServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	m.replyType = payloadType
	m.replyPayload = jsonString
	c.JSON(http.StatusOK, jsonString)
}

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	m.replyType = payloadType
	m.replyPayload = ""
}

func (m *MockServer) Validator() security.Validator {
	return m.validator
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) UserDB() database.UserDatabase {
	return m.userDB
}

func (m *MockServer) ProcessDB() database.ProcessDatabase {
	return m.processDB
}

func (m *MockServer) BlueprintDB() database.BlueprintDatabase {
	return m.blueprintDB
}

func (m *MockServer) ProcessController() Controller {
	return m.controller
}

func (m *MockServer) ExclusiveAssign() bool {
	return m.exclusiveAssign
}

func (m *MockServer) TLS() bool {
	return m.tls
}

// Test setup helpers
func createTestProcess() *core.Process {
	return &core.Process{
		ID:                 "process-123",
		State:              core.RUNNING,
		AssignedExecutorID: "executor-123",
		ProcessGraphID:     "graph-123",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName:   "test-colony",
				ExecutorType: "test-type",
			},
		},
	}
}

func createTestFunctionSpec() *core.FunctionSpec {
	return &core.FunctionSpec{
		Conditions: core.Conditions{
			ColonyName:   "test-colony",
			ExecutorType: "test-type",
		},
		Priority: 0,
	}
}

func createMockServer() *MockServer {
	executor := &core.Executor{
		ID:         "executor-123",
		Name:       "test-executor",
		ColonyName: "test-colony",
		Type:       "test-type",
	}

	process := createTestProcess()

	executorDB := &MockExecutorDB{executors: []*core.Executor{executor}}
	userDB := &MockUserDB{}
	processDB := &MockProcessDB{processes: []*core.Process{process}}
	blueprintDB := &MockBlueprintDB{}
	validator := &MockValidator{}
	controller := &MockController{isLeader: true}

	return &MockServer{
		executorDB:  executorDB,
		userDB:      userDB,
		processDB:   processDB,
		blueprintDB: blueprintDB,
		validator:   validator,
		controller:  controller,
	}
}

// HandleSubmit tests
func TestHandleSubmit_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.SubmitFunctionSpecPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleSubmit_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSubmit_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSubmit_NilFunctionSpec(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSubmitFunctionSpecMsg(nil)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSubmit_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleSubmit_InvalidPriority(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	funcSpec.Priority = -100000 // Invalid priority (MIN_PRIORITY is -50000)
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSubmit_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.addProcessFunc = func(p *core.Process) (*core.Process, error) {
		return nil, errors.New("controller error")
	}
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSubmit_ControllerReturnsNil(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.addProcessFunc = func(p *core.Process) (*core.Process, error) {
		return nil, nil
	}
	handlers := NewHandlers(mockServer)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateSubmitFunctionSpecMsg(funcSpec)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSubmit(ctx, "executor-123", rpc.SubmitFunctionSpecPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// HandleGetProcess tests
func TestHandleGetProcess_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", rpc.GetProcessPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.GetProcessPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleGetProcess_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", rpc.GetProcessPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetProcess_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetProcess_DBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", rpc.GetProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetProcess_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetProcessMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", rpc.GetProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleGetProcess_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetProcess(ctx, "executor-123", rpc.GetProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

// HandleRemoveProcess tests
func TestHandleRemoveProcess_Success(t *testing.T) {
	mockServer := createMockServer()
	// Process must not be part of a workflow for successful removal
	mockServer.processDB.processes[0].ProcessGraphID = ""
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.RemoveProcessPayloadType, mockServer.replyType)
}

func TestHandleRemoveProcess_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveProcess_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveProcess_ProcessDBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveProcess_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleRemoveProcess_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleRemoveProcess_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	// Process must not be part of a workflow to reach controller
	mockServer.processDB.processes[0].ProcessGraphID = ""
	mockServer.controller.removeProcessErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveProcessMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveProcess(ctx, "executor-123", rpc.RemoveProcessPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

// HandleRemoveAllProcesses tests
func TestHandleRemoveAllProcesses_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveAllProcessesMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveAllProcesses(ctx, "executor-123", rpc.RemoveAllProcessesPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.RemoveAllProcessesPayloadType, mockServer.replyType)
}

func TestHandleRemoveAllProcesses_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleRemoveAllProcesses(ctx, "executor-123", rpc.RemoveAllProcessesPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveAllProcesses_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveAllProcessesMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveAllProcesses(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveAllProcesses_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveAllProcessesMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveAllProcesses(ctx, "executor-123", rpc.RemoveAllProcessesPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleRemoveAllProcesses_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.removeAllProcessesErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveAllProcessesMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveAllProcesses(ctx, "executor-123", rpc.RemoveAllProcessesPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

// HandleCloseSuccessful tests
func TestHandleCloseSuccessful_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", rpc.CloseSuccessfulPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.CloseSuccessfulPayloadType, mockServer.replyType)
}

func TestHandleCloseSuccessful_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", rpc.CloseSuccessfulPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleCloseSuccessful_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleCloseSuccessful_ProcessNotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", rpc.CloseSuccessfulPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleCloseSuccessful_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", rpc.CloseSuccessfulPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleCloseSuccessful_WrongExecutor(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "wrong-executor", rpc.CloseSuccessfulPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleCloseSuccessful_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.closeSuccessfulErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseSuccessfulMsg("process-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseSuccessful(ctx, "executor-123", rpc.CloseSuccessfulPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

// HandleCloseFailed tests
func TestHandleCloseFailed_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("process-123", []string{"error message"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", rpc.CloseFailedPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.CloseFailedPayloadType, mockServer.replyType)
}

func TestHandleCloseFailed_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", rpc.CloseFailedPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleCloseFailed_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("process-123", []string{"error"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleCloseFailed_ProcessNotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("non-existent", []string{"error"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", rpc.CloseFailedPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleCloseFailed_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("process-123", []string{"error"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", rpc.CloseFailedPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleCloseFailed_WrongExecutor(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("process-123", []string{"error"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "wrong-executor", rpc.CloseFailedPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleCloseFailed_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.closeFailedErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateCloseFailedMsg("process-123", []string{"error"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleCloseFailed(ctx, "executor-123", rpc.CloseFailedPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

// HandleSetOutput tests
func TestHandleSetOutput_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("process-123", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", rpc.SetOutputPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.SetOutputPayloadType, mockServer.replyType)
}

func TestHandleSetOutput_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", rpc.SetOutputPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSetOutput_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("process-123", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleSetOutput_ProcessNotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("non-existent", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", rpc.SetOutputPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleSetOutput_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("process-123", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", rpc.SetOutputPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleSetOutput_WrongExecutor(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("process-123", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "wrong-executor", rpc.SetOutputPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleSetOutput_DBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.setOutputErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateSetOutputMsg("process-123", []interface{}{"output"})
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleSetOutput(ctx, "executor-123", rpc.SetOutputPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

// HandlePauseAssignments tests
func TestHandlePauseAssignments_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreatePauseAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandlePauseAssignments(ctx, "executor-123", rpc.PauseAssignmentsPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.PauseAssignmentsPayloadType, mockServer.replyType)
}

func TestHandlePauseAssignments_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandlePauseAssignments(ctx, "executor-123", rpc.PauseAssignmentsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandlePauseAssignments_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreatePauseAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandlePauseAssignments(ctx, "executor-123", rpc.PauseAssignmentsPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandlePauseAssignments_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.pauseAssignmentsErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreatePauseAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandlePauseAssignments(ctx, "executor-123", rpc.PauseAssignmentsPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// HandleResumeAssignments tests
func TestHandleResumeAssignments_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateResumeAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleResumeAssignments(ctx, "executor-123", rpc.ResumeAssignmentsPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.ResumeAssignmentsPayloadType, mockServer.replyType)
}

func TestHandleResumeAssignments_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleResumeAssignments(ctx, "executor-123", rpc.ResumeAssignmentsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleResumeAssignments_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateResumeAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleResumeAssignments(ctx, "executor-123", rpc.ResumeAssignmentsPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleResumeAssignments_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.resumeAssignmentsErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateResumeAssignmentsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleResumeAssignments(ctx, "executor-123", rpc.ResumeAssignmentsPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// HandleGetPauseStatus tests
func TestHandleGetPauseStatus_Success(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.pauseStatusResult = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetPauseStatusMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetPauseStatus(ctx, "executor-123", rpc.GetPauseStatusPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.PauseStatusReplyPayloadType, mockServer.replyType)
	assert.Contains(t, mockServer.replyPayload, "true")
}

func TestHandleGetPauseStatus_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleGetPauseStatus(ctx, "executor-123", rpc.GetPauseStatusPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetPauseStatus_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetPauseStatusMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetPauseStatus(ctx, "executor-123", rpc.GetPauseStatusPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleGetPauseStatus_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.controller.pauseStatusErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetPauseStatusMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetPauseStatus(ctx, "executor-123", rpc.GetPauseStatusPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// Handler registration tests
func TestRegisterHandlers_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	handlerRegistry := registry.NewHandlerRegistry()
	err := handlers.RegisterHandlers(handlerRegistry)

	assert.Nil(t, err)
}

func TestRegisterHandlers_DuplicateError(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	handlerRegistry := registry.NewHandlerRegistry()
	err := handlers.RegisterHandlers(handlerRegistry)
	assert.Nil(t, err)

	// Try to register again
	err = handlers.RegisterHandlers(handlerRegistry)
	assert.NotNil(t, err)
}
