package colony

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

// MockColonyDB implements database.ColonyDatabase
type MockColonyDB struct {
	colonies          []*core.Colony
	addColonyErr      error
	getByNameErr      error
	getByIDErr        error
	getColoniesErr    error
	removeByNameErr   error
	returnNilByName   bool
	returnNilByID     bool
}

func (m *MockColonyDB) AddColony(colony *core.Colony) error {
	if m.addColonyErr != nil {
		return m.addColonyErr
	}
	m.colonies = append(m.colonies, colony)
	return nil
}

func (m *MockColonyDB) GetColonies() ([]*core.Colony, error) {
	if m.getColoniesErr != nil {
		return nil, m.getColoniesErr
	}
	return m.colonies, nil
}

func (m *MockColonyDB) GetColonyByName(colonyName string) (*core.Colony, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, c := range m.colonies {
		if c.Name == colonyName {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockColonyDB) GetColonyByID(colonyID string) (*core.Colony, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, c := range m.colonies {
		if c.ID == colonyID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockColonyDB) RemoveColonyByName(colonyName string) error {
	if m.removeByNameErr != nil {
		return m.removeByNameErr
	}
	return nil
}

func (m *MockColonyDB) RemoveColonies() error {
	m.colonies = nil
	return nil
}

func (m *MockColonyDB) CountColonies() (int, error) {
	return len(m.colonies), nil
}

func (m *MockColonyDB) RenameColony(colonyName, newName string) error {
	return nil
}

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	countByColonyErr      error
	countByStateErr       error
	executorCount         int
	activeExecutorCount   int
	unregisteredCount     int
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                            { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error { return nil }
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                              { return nil, nil }
func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error)            { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error                        { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error                         { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error                              { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error           { return nil }
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error                  { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                                         { return 0, nil }
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName, executorName string, cap core.Capabilities) error { return nil }

func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error) {
	if m.countByColonyErr != nil {
		return 0, m.countByColonyErr
	}
	return m.executorCount, nil
}

func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	if m.countByStateErr != nil {
		return 0, m.countByStateErr
	}
	if state == core.APPROVED {
		return m.activeExecutorCount, nil
	}
	if state == core.UNREGISTERED {
		return m.unregisteredCount, nil
	}
	return 0, nil
}

// MockProcessDB implements database.ProcessDatabase
type MockProcessDB struct {
	countWaitingErr    error
	countRunningErr    error
	countSuccessErr    error
	countFailedErr     error
	waitingCount       int
	runningCount       int
	successCount       int
	failedCount        int
}

func (m *MockProcessDB) AddProcess(process *core.Process) error                      { return nil }
func (m *MockProcessDB) GetProcesses() ([]*core.Process, error)                      { return nil, nil }
func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error)      { return nil, nil }
func (m *MockProcessDB) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindWaitingProcesses(colonyName, executorType, label, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindRunningProcesses(colonyName, executorType, label, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindSuccessfulProcesses(colonyName, executorType, label, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindFailedProcesses(colonyName, executorType, label, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindAllRunningProcesses() ([]*core.Process, error)           { return nil, nil }
func (m *MockProcessDB) FindAllWaitingProcesses() ([]*core.Process, error)           { return nil, nil }
func (m *MockProcessDB) FindCandidates(colonyName, executorType, executorLocationName string, cpu, memory, storage int64, nodes, processes, processesPerNode, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) FindCandidatesByName(colonyName, executorName, executorType, executorLocationName string, cpu, memory, storage int64, nodes, processes, processesPerNode, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) RemoveProcessByID(processID string) error                    { return nil }
func (m *MockProcessDB) RemoveAllProcesses() error                                   { return nil }
func (m *MockProcessDB) RemoveAllWaitingProcessesByColonyName(string) error          { return nil }
func (m *MockProcessDB) RemoveAllRunningProcessesByColonyName(string) error          { return nil }
func (m *MockProcessDB) RemoveAllSuccessfulProcessesByColonyName(string) error       { return nil }
func (m *MockProcessDB) RemoveAllFailedProcessesByColonyName(string) error           { return nil }
func (m *MockProcessDB) RemoveAllProcessesByColonyName(string) error                 { return nil }
func (m *MockProcessDB) RemoveAllProcessesByProcessGraphID(string) error             { return nil }
func (m *MockProcessDB) RemoveAllProcessesInProcessGraphsByColonyName(string) error  { return nil }
func (m *MockProcessDB) ResetProcess(process *core.Process) error                    { return nil }
func (m *MockProcessDB) SetInput(processID string, output []interface{}) error       { return nil }
func (m *MockProcessDB) SetOutput(processID string, output []interface{}) error      { return nil }
func (m *MockProcessDB) SetErrors(processID string, errs []string) error             { return nil }
func (m *MockProcessDB) SetProcessState(processID string, state int) error           { return nil }
func (m *MockProcessDB) SetParents(processID string, parents []string) error         { return nil }
func (m *MockProcessDB) SetChildren(processID string, children []string) error       { return nil }
func (m *MockProcessDB) SetWaitForParents(processID string, waiting bool) error      { return nil }
func (m *MockProcessDB) Assign(executorID string, process *core.Process) error       { return nil }
func (m *MockProcessDB) SelectAndAssign(colonyName, executorID, executorName, executorType, executorLocation string, cpu, memory, storage int64, nodes, processes, processesPerNode, count int) (*core.Process, error) { return nil, nil }
func (m *MockProcessDB) Unassign(process *core.Process) error                        { return nil }
func (m *MockProcessDB) MarkSuccessful(processID string) (float64, float64, error)   { return 0, 0, nil }
func (m *MockProcessDB) MarkFailed(processID string, errs []string) error            { return nil }
func (m *MockProcessDB) CountProcesses() (int, error)                                { return 0, nil }
func (m *MockProcessDB) CountWaitingProcesses() (int, error)                         { return 0, nil }
func (m *MockProcessDB) CountRunningProcesses() (int, error)                         { return 0, nil }
func (m *MockProcessDB) CountSuccessfulProcesses() (int, error)                      { return 0, nil }
func (m *MockProcessDB) CountFailedProcesses() (int, error)                          { return 0, nil }

func (m *MockProcessDB) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	if m.countWaitingErr != nil {
		return 0, m.countWaitingErr
	}
	return m.waitingCount, nil
}

func (m *MockProcessDB) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	if m.countRunningErr != nil {
		return 0, m.countRunningErr
	}
	return m.runningCount, nil
}

func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	if m.countSuccessErr != nil {
		return 0, m.countSuccessErr
	}
	return m.successCount, nil
}

func (m *MockProcessDB) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	if m.countFailedErr != nil {
		return 0, m.countFailedErr
	}
	return m.failedCount, nil
}
func (m *MockProcessDB) FindCancelledProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) RemoveAllCancelledProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) MarkCancelled(string) error                           { return nil }
func (m *MockProcessDB) CountCancelledProcesses() (int, error)                { return 0, nil }
func (m *MockProcessDB) CountCancelledProcessesByColonyName(colonyName string) (int, error) {
	return 0, nil
}

// MockProcessGraphDB implements database.ProcessGraphDatabase
type MockProcessGraphDB struct {
	countWaitingErr  error
	countRunningErr  error
	countSuccessErr  error
	countFailedErr   error
	waitingCount     int
	runningCount     int
	successCount     int
	failedCount      int
}

func (m *MockProcessGraphDB) AddProcessGraph(pg *core.ProcessGraph) error                        { return nil }
func (m *MockProcessGraphDB) GetProcessGraphByID(id string) (*core.ProcessGraph, error)          { return nil, nil }
func (m *MockProcessGraphDB) SetProcessGraphState(id string, state int) error                    { return nil }
func (m *MockProcessGraphDB) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (m *MockProcessGraphDB) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (m *MockProcessGraphDB) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (m *MockProcessGraphDB) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (m *MockProcessGraphDB) RemoveProcessGraphByID(id string) error                             { return nil }
func (m *MockProcessGraphDB) RemoveAllProcessGraphsByColonyName(colonyName string) error         { return nil }
func (m *MockProcessGraphDB) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error  { return nil }
func (m *MockProcessGraphDB) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error  { return nil }
func (m *MockProcessGraphDB) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error { return nil }
func (m *MockProcessGraphDB) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error   { return nil }

func (m *MockProcessGraphDB) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	if m.countWaitingErr != nil {
		return 0, m.countWaitingErr
	}
	return m.waitingCount, nil
}

func (m *MockProcessGraphDB) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	if m.countRunningErr != nil {
		return 0, m.countRunningErr
	}
	return m.runningCount, nil
}

func (m *MockProcessGraphDB) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	if m.countSuccessErr != nil {
		return 0, m.countSuccessErr
	}
	return m.successCount, nil
}

func (m *MockProcessGraphDB) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	if m.countFailedErr != nil {
		return 0, m.countFailedErr
	}
	return m.failedCount, nil
}

func (m *MockProcessGraphDB) CountWaitingProcessGraphs() (int, error)    { return 0, nil }
func (m *MockProcessGraphDB) CountRunningProcessGraphs() (int, error)    { return 0, nil }
func (m *MockProcessGraphDB) CountSuccessfulProcessGraphs() (int, error) { return 0, nil }
func (m *MockProcessGraphDB) CountFailedProcessGraphs() (int, error)     { return 0, nil }
func (m *MockProcessGraphDB) FindCancelledProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (m *MockProcessGraphDB) RemoveAllCancelledProcessGraphsByColonyName(colonyName string) error { return nil }
func (m *MockProcessGraphDB) CountCancelledProcessGraphs() (int, error)                          { return 0, nil }
func (m *MockProcessGraphDB) CountCancelledProcessGraphsByColonyName(colonyName string) (int, error) { return 0, nil }

// MockValidator implements security.Validator
type MockValidator struct {
	requireServerOwnerErr  error
	requireMembershipErr   error
}

func (m *MockValidator) RequireServerOwner(recoveredID, serverID string) error {
	return m.requireServerOwnerErr
}

func (m *MockValidator) RequireColonyOwner(recoveredID, colonyName string) error {
	return nil
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
func (m *MockContext) AbortWithStatus(code int)                              { m.abortedWithStatus = code; m.aborted = true }
func (m *MockContext) AbortWithStatusJSON(code int, jsonObj interface{})     { m.abortedWithStatusJSON = code; m.jsonResponse = jsonObj; m.aborted = true }
func (m *MockContext) IsAborted() bool                                       { return m.aborted }
func (m *MockContext) Next()                                                 {}

// MockServer implements the Server interface
type MockServer struct {
	colonyDB             *MockColonyDB
	executorDB           *MockExecutorDB
	processDB            *MockProcessDB
	processGraphDB       *MockProcessGraphDB
	validator            *MockValidator
	serverID             string
	serverIDErr          error
	httpErrorCalled      bool
	httpErrorCode        int
	httpReplyCalled      bool
	httpReplyPayloadType string
	httpReplyJSON        string
	emptyReplyCalled     bool
}

func (m *MockServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	if err != nil {
		m.httpErrorCalled = true
		m.httpErrorCode = errorCode
		c.AbortWithStatusJSON(errorCode, map[string]string{"error": err.Error()})
		return true
	}
	return false
}

func (m *MockServer) GetServerID() (string, error) {
	if m.serverIDErr != nil {
		return "", m.serverIDErr
	}
	return m.serverID, nil
}

func (m *MockServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
	m.httpReplyCalled = true
	m.httpReplyPayloadType = payloadType
	m.httpReplyJSON = jsonString
	c.JSON(http.StatusOK, jsonString)
}

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	m.emptyReplyCalled = true
	m.httpReplyPayloadType = payloadType
}

func (m *MockServer) Validator() security.Validator {
	return m.validator
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

func (m *MockServer) ProcessGraphDB() database.ProcessGraphDatabase {
	return m.processGraphDB
}

// Helper functions
func createTestMocks() (*MockServer, *MockColonyDB, *MockValidator, *MockContext) {
	colonyDB := &MockColonyDB{}
	validator := &MockValidator{}
	ctx := &MockContext{}
	server := &MockServer{
		colonyDB:       colonyDB,
		executorDB:     &MockExecutorDB{},
		processDB:      &MockProcessDB{},
		processGraphDB: &MockProcessGraphDB{},
		validator:      validator,
		serverID:       "server-123",
	}
	return server, colonyDB, validator, ctx
}

func createTestColony(id, name string) *core.Colony {
	return core.CreateColony(id, name)
}

// =============================================
// Tests for HandleAddColony
// =============================================

func TestHandleAddColony_Success(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
	assert.Len(t, colonyDB.colonies, 1)
}

func TestHandleAddColony_InvalidJSON(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_MsgTypeMismatch(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_ServerIDError(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	server.serverIDErr = errors.New("server ID error")
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleAddColony_AuthError(t *testing.T) {
	server, _, validator, ctx := createTestMocks()
	validator.requireServerOwnerErr = errors.New("not authorized")
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "wrong-id", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleAddColony_NilColony(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := &rpc.AddColonyMsg{
		MsgType: rpc.AddColonyPayloadType,
		Colony:  nil,
	}
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_InvalidIDLength(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("short-id", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_ColonyAlreadyExists(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	existingColony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, existingColony)

	colony := createTestColony("abcdef0123456789abcdef0123456789abcdef0123456789abcdef01234567", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_AddColonyDBError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.addColonyErr = errors.New("database error")
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddColony_GetAddedColonyError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.getByIDErr = errors.New("get error")
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleAddColony_AddedColonyNil(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.returnNilByID = true
	h := NewHandlers(server)

	colony := createTestColony("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "test-colony")
	msg := rpc.CreateAddColonyMsg(colony)
	jsonString, _ := msg.ToJSON()

	h.HandleAddColony(ctx, "server-123", rpc.AddColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

// =============================================
// Tests for HandleRemoveColony
// =============================================

func TestHandleRemoveColony_Success(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)

	msg := rpc.CreateRemoveColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveColony(ctx, "server-123", rpc.RemoveColonyPayloadType, jsonString)

	assert.True(t, server.emptyReplyCalled)
}

func TestHandleRemoveColony_InvalidJSON(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleRemoveColony(ctx, "server-123", rpc.RemoveColonyPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveColony_MsgTypeMismatch(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateRemoveColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveColony(ctx, "server-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveColony_AuthError(t *testing.T) {
	server, _, validator, ctx := createTestMocks()
	validator.requireServerOwnerErr = errors.New("not authorized")
	h := NewHandlers(server)

	msg := rpc.CreateRemoveColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveColony(ctx, "wrong-id", rpc.RemoveColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleRemoveColony_ColonyNotFound(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.returnNilByName = true
	h := NewHandlers(server)

	msg := rpc.CreateRemoveColonyMsg("non-existent")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveColony(ctx, "server-123", rpc.RemoveColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveColony_RemoveError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)
	colonyDB.removeByNameErr = errors.New("remove failed")

	msg := rpc.CreateRemoveColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveColony(ctx, "server-123", rpc.RemoveColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

// =============================================
// Tests for HandleGetColonies
// =============================================

func TestHandleGetColonies_Success(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)

	msg := rpc.CreateGetColoniesMsg()
	jsonString, _ := msg.ToJSON()

	h.HandleGetColonies(ctx, "server-123", rpc.GetColoniesPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
}

func TestHandleGetColonies_InvalidJSON(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleGetColonies(ctx, "server-123", rpc.GetColoniesPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetColonies_MsgTypeMismatch(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateGetColoniesMsg()
	jsonString, _ := msg.ToJSON()

	h.HandleGetColonies(ctx, "server-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetColonies_AuthError(t *testing.T) {
	server, _, validator, ctx := createTestMocks()
	validator.requireServerOwnerErr = errors.New("not authorized")
	h := NewHandlers(server)

	msg := rpc.CreateGetColoniesMsg()
	jsonString, _ := msg.ToJSON()

	h.HandleGetColonies(ctx, "wrong-id", rpc.GetColoniesPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleGetColonies_DBError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.getColoniesErr = errors.New("database error")
	h := NewHandlers(server)

	msg := rpc.CreateGetColoniesMsg()
	jsonString, _ := msg.ToJSON()

	h.HandleGetColonies(ctx, "server-123", rpc.GetColoniesPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

// =============================================
// Tests for HandleGetColony
// =============================================

func TestHandleGetColony_Success(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)

	msg := rpc.CreateGetColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleGetColony(ctx, "member-123", rpc.GetColonyPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
}

func TestHandleGetColony_InvalidJSON(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleGetColony(ctx, "member-123", rpc.GetColonyPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetColony_MsgTypeMismatch(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateGetColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleGetColony(ctx, "member-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetColony_AuthError(t *testing.T) {
	server, _, validator, ctx := createTestMocks()
	validator.requireMembershipErr = errors.New("not a member")
	h := NewHandlers(server)

	msg := rpc.CreateGetColonyMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleGetColony(ctx, "wrong-id", rpc.GetColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleGetColony_ColonyNil(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.returnNilByName = true
	h := NewHandlers(server)

	msg := rpc.CreateGetColonyMsg("non-existent")
	jsonString, _ := msg.ToJSON()

	h.HandleGetColony(ctx, "member-123", rpc.GetColonyPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

// =============================================
// Tests for HandleColonyStatistics
// =============================================

func TestHandleColonyStatistics_Success(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)
	server.executorDB.executorCount = 5
	server.executorDB.activeExecutorCount = 3
	server.processDB.waitingCount = 10
	server.processDB.runningCount = 5

	msg := rpc.CreateGetColonyStatisticsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
}

func TestHandleColonyStatistics_InvalidJSON(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleColonyStatistics_MsgTypeMismatch(t *testing.T) {
	server, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateGetColonyStatisticsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleColonyStatistics_ColonyNotFound(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	colonyDB.returnNilByName = true
	h := NewHandlers(server)

	msg := rpc.CreateGetColonyStatisticsMsg("non-existent")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleColonyStatistics_ExecutorCountError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)
	server.executorDB.countByColonyErr = errors.New("count error")

	msg := rpc.CreateGetColonyStatisticsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleColonyStatistics_ProcessCountError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)
	server.processDB.countWaitingErr = errors.New("count error")

	msg := rpc.CreateGetColonyStatisticsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleColonyStatistics_WorkflowCountError(t *testing.T) {
	server, colonyDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	colony := createTestColony("colony-id", "test-colony")
	colonyDB.colonies = append(colonyDB.colonies, colony)
	server.processGraphDB.countWaitingErr = errors.New("count error")

	msg := rpc.CreateGetColonyStatisticsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	h.HandleColonyStatistics(ctx, "member-123", rpc.GetColonyStatisticsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

// =============================================
// Tests for NewHandlers and RegisterHandlers
// =============================================

func TestNewHandlers(t *testing.T) {
	server, _, _, _ := createTestMocks()
	h := NewHandlers(server)
	assert.NotNil(t, h)
}

func TestRegisterHandlers_Success(t *testing.T) {
	server, _, _, _ := createTestMocks()
	h := NewHandlers(server)

	reg := registry.NewHandlerRegistry()
	err := h.RegisterHandlers(reg)

	assert.NoError(t, err)
}

func TestRegisterHandlers_DuplicateError(t *testing.T) {
	server, _, _, _ := createTestMocks()
	h := NewHandlers(server)

	reg := registry.NewHandlerRegistry()

	// Register once - should succeed
	err := h.RegisterHandlers(reg)
	assert.NoError(t, err)

	// Register again - should fail on first duplicate
	err = h.RegisterHandlers(reg)
	assert.Error(t, err)
}
