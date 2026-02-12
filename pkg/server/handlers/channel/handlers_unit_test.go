package channel

import (
	"errors"
	"net/http"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// MockProcessDB implements database.ProcessDatabase for channel handlers
type MockProcessDB struct {
	process        *core.Process
	getProcessErr  error
	returnNil      bool
}

func (m *MockProcessDB) AddProcess(process *core.Process) error                      { return nil }
func (m *MockProcessDB) GetProcesses() ([]*core.Process, error)                      { return nil, nil }
func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error) {
	if m.getProcessErr != nil {
		return nil, m.getProcessErr
	}
	if m.returnNil {
		return nil, nil
	}
	return m.process, nil
}
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
func (m *MockProcessDB) CountWaitingProcessesByColonyName(string) (int, error)       { return 0, nil }
func (m *MockProcessDB) CountRunningProcessesByColonyName(string) (int, error)       { return 0, nil }
func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(string) (int, error)    { return 0, nil }
func (m *MockProcessDB) CountFailedProcessesByColonyName(string) (int, error)        { return 0, nil }
func (m *MockProcessDB) FindCancelledProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) RemoveAllCancelledProcessesByColonyName(string) error        { return nil }
func (m *MockProcessDB) MarkCancelled(string) error                                  { return nil }
func (m *MockProcessDB) CountCancelledProcesses() (int, error)                       { return 0, nil }
func (m *MockProcessDB) CountCancelledProcessesByColonyName(string) (int, error)     { return 0, nil }

// MockValidator implements security.Validator
type MockValidator struct {
	membershipErr   error
	colonyOwnerErr  error
	serverOwnerErr  error
}

func (m *MockValidator) RequireMembership(recoveredID string, colonyName string, executorMayJoin bool) error {
	return m.membershipErr
}

func (m *MockValidator) RequireColonyOwner(recoveredID string, colonyName string) error {
	return m.colonyOwnerErr
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
	processDB       *MockProcessDB
	channelRouter   *channel.Router
	validator       *MockValidator
	lastError       error
	lastStatusCode  int
	lastPayloadType string
	lastResponse    string
	emptyReplySent  bool
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

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	m.lastPayloadType = payloadType
	m.emptyReplySent = true
	c.JSON(http.StatusOK, nil)
}

func (m *MockServer) Validator() security.Validator {
	return m.validator
}

func (m *MockServer) ProcessDB() database.ProcessDatabase {
	return m.processDB
}

func (m *MockServer) ChannelRouter() *channel.Router {
	return m.channelRouter
}

// Helper to create test process
func createTestProcess() *core.Process {
	funcSpec := core.FunctionSpec{
		Conditions: core.Conditions{
			ColonyName: "test-colony",
		},
		Channels: []string{"test-channel"},
	}
	process := core.CreateProcess(&funcSpec)
	process.ID = "test-process-id"
	process.InitiatorID = "test-initiator"
	process.AssignedExecutorID = "test-executor"
	process.State = core.RUNNING
	return process
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	processDB := &MockProcessDB{process: createTestProcess()}
	channelRouter := channel.NewRouterWithoutRateLimit()
	validator := &MockValidator{}

	// Create a channel in the router
	ch := &channel.Channel{
		ID:          "test-process-id_test-channel",
		ProcessID:   "test-process-id",
		Name:        "test-channel",
		SubmitterID: "test-initiator",
		ExecutorID:  "test-executor",
	}
	channelRouter.Create(ch)

	server := &MockServer{
		processDB:     processDB,
		channelRouter: channelRouter,
		validator:     validator,
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

// Tests for HandleChannelAppend
func TestHandleChannelAppend_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
	assert.Nil(t, server.lastError)
}

func TestHandleChannelAppend_SuccessWithType(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsgWithType("test-process-id", "test-channel", 1, 0, []byte("test data"), "end")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
	assert.Nil(t, server.lastError)
}

func TestHandleChannelAppend_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelAppend_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelAppend_ProcessDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.getProcessErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelAppend_ProcessNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

func TestHandleChannelAppend_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChannelAppend_ChannelNotFound(t *testing.T) {
	server, ctx := createMockServer()
	// Remove the channel from the router
	server.channelRouter.CleanupProcess("test-process-id")
	handlers := NewHandlers(server)

	// Process does not have the channel in its spec for lazy creation
	server.processDB.process.FunctionSpec.Channels = []string{}

	msg := rpc.CreateChannelAppendMsg("test-process-id", "nonexistent-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

func TestHandleChannelAppend_LazyChannelCreation(t *testing.T) {
	server, ctx := createMockServer()
	// Remove the channel from the router
	server.channelRouter.CleanupProcess("test-process-id")
	handlers := NewHandlers(server)

	// Process has the channel in its spec for lazy creation
	server.processDB.process.FunctionSpec.Channels = []string{"test-channel"}

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
	assert.Nil(t, server.lastError)
}

func TestHandleChannelAppend_UnauthorizedAppend(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	// Use an unauthorized caller
	handlers.HandleChannelAppend(ctx, "unauthorized-user", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleChannelRead
func TestHandleChannelRead_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	// First append some data
	appendMsg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	appendJSON, _ := appendMsg.ToJSON()
	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, appendJSON)

	// Reset context
	ctx = &MockContext{}
	server.emptyReplySent = false
	server.lastResponse = ""

	// Now read
	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, rpc.ChannelReadPayloadType, server.lastPayloadType)
	assert.NotEmpty(t, server.lastResponse)
}

func TestHandleChannelRead_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelRead_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelRead_ProcessDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.getProcessErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChannelRead_ProcessNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.processDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

func TestHandleChannelRead_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChannelRead_ChannelNotFound(t *testing.T) {
	server, ctx := createMockServer()
	// Remove the channel
	server.channelRouter.CleanupProcess("test-process-id")
	// Process does not have the channel in its spec for lazy creation
	server.processDB.process.FunctionSpec.Channels = []string{}
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "nonexistent-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

func TestHandleChannelRead_LazyChannelCreation(t *testing.T) {
	server, ctx := createMockServer()
	// Remove the channel
	server.channelRouter.CleanupProcess("test-process-id")
	// Process has the channel in its spec for lazy creation
	server.processDB.process.FunctionSpec.Channels = []string{"test-channel"}
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelRead(ctx, "test-initiator", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, rpc.ChannelReadPayloadType, server.lastPayloadType)
}

func TestHandleChannelRead_UnauthorizedRead(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelReadMsg("test-process-id", "test-channel", 0, 10)
	jsonString, _ := msg.ToJSON()

	// Use an unauthorized caller
	handlers.HandleChannelRead(ctx, "unauthorized-user", rpc.ChannelReadPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Test helper functions
func TestGetCallerID_Submitter(t *testing.T) {
	process := createTestProcess()
	callerID := getCallerID(process.InitiatorID, process)
	assert.Equal(t, process.InitiatorID, callerID)
}

func TestGetCallerID_Executor(t *testing.T) {
	process := createTestProcess()
	callerID := getCallerID(process.AssignedExecutorID, process)
	assert.Equal(t, process.AssignedExecutorID, callerID)
}

func TestGetCallerID_Other(t *testing.T) {
	process := createTestProcess()
	callerID := getCallerID("other-user", process)
	assert.Equal(t, "other-user", callerID)
}

// Test ensureChannelExists for closed processes
func TestEnsureChannelExists_ClosedProcess(t *testing.T) {
	server, ctx := createMockServer()
	// Remove the channel
	server.channelRouter.CleanupProcess("test-process-id")
	// Process is in SUCCESS state
	server.processDB.process.State = core.SUCCESS
	server.processDB.process.FunctionSpec.Channels = []string{"test-channel"}
	handlers := NewHandlers(server)

	msg := rpc.CreateChannelAppendMsg("test-process-id", "test-channel", 1, 0, []byte("test data"))
	jsonString, _ := msg.ToJSON()

	handlers.HandleChannelAppend(ctx, "test-initiator", rpc.ChannelAppendPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}
