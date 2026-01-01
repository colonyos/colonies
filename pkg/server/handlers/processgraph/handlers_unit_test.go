package processgraph

import (
	"errors"
	"net/http"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// MockController implements Controller interface
type MockController struct {
	submitErr       error
	getByIDErr      error
	findWaitingErr  error
	findRunningErr  error
	findSuccessErr  error
	findFailedErr   error
	removeErr       error
	removeAllErr    error
	addChildErr     error
	processGraph    *core.ProcessGraph
	processGraphs   []*core.ProcessGraph
	addedProcess    *core.Process
	returnNil       bool
	returnNilChild  bool
}

func (m *MockController) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, initiatorID string) (*core.ProcessGraph, error) {
	if m.submitErr != nil {
		return nil, m.submitErr
	}
	return m.processGraph, nil
}

func (m *MockController) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNil {
		return nil, nil
	}
	return m.processGraph, nil
}

func (m *MockController) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	if m.findWaitingErr != nil {
		return nil, m.findWaitingErr
	}
	return m.processGraphs, nil
}

func (m *MockController) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	if m.findRunningErr != nil {
		return nil, m.findRunningErr
	}
	return m.processGraphs, nil
}

func (m *MockController) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	if m.findSuccessErr != nil {
		return nil, m.findSuccessErr
	}
	return m.processGraphs, nil
}

func (m *MockController) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	if m.findFailedErr != nil {
		return nil, m.findFailedErr
	}
	return m.processGraphs, nil
}

func (m *MockController) RemoveProcessGraph(processGraphID string) error {
	return m.removeErr
}

func (m *MockController) RemoveAllProcessGraphs(colonyName string, state int) error {
	return m.removeAllErr
}

func (m *MockController) AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, initiatorID string, insert bool) (*core.Process, error) {
	if m.addChildErr != nil {
		return nil, m.addChildErr
	}
	if m.returnNilChild {
		return nil, nil
	}
	return m.addedProcess, nil
}

// MockValidator implements Validator interface
type MockValidator struct {
	membershipErr  error
	colonyOwnerErr error
}

func (m *MockValidator) RequireMembership(recoveredID string, colonyName string, executorMayJoin bool) error {
	return m.membershipErr
}

func (m *MockValidator) RequireColonyOwner(recoveredID string, colonyName string) error {
	return m.colonyOwnerErr
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

// MockProcessGraphDB implements database.ProcessGraphDatabase
type MockProcessGraphDB struct {
	removeErr    error
	removeAllErr error
}

func (m *MockProcessGraphDB) AddProcessGraph(pg *core.ProcessGraph) error { return nil }
func (m *MockProcessGraphDB) GetProcessGraphByID(id string) (*core.ProcessGraph, error) {
	return nil, nil
}
func (m *MockProcessGraphDB) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}
func (m *MockProcessGraphDB) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}
func (m *MockProcessGraphDB) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}
func (m *MockProcessGraphDB) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}
func (m *MockProcessGraphDB) SetProcessGraphState(id string, state int) error { return nil }
func (m *MockProcessGraphDB) RemoveProcessGraphByID(id string) error { return m.removeErr }
func (m *MockProcessGraphDB) RemoveAllProcessGraphsByColonyName(colonyName string) error {
	return m.removeAllErr
}
func (m *MockProcessGraphDB) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error {
	return nil
}
func (m *MockProcessGraphDB) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error {
	return nil
}
func (m *MockProcessGraphDB) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error {
	return nil
}
func (m *MockProcessGraphDB) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error {
	return nil
}
func (m *MockProcessGraphDB) CountWaitingProcessGraphs() (int, error)    { return 0, nil }
func (m *MockProcessGraphDB) CountRunningProcessGraphs() (int, error)    { return 0, nil }
func (m *MockProcessGraphDB) CountSuccessfulProcessGraphs() (int, error) { return 0, nil }
func (m *MockProcessGraphDB) CountFailedProcessGraphs() (int, error)     { return 0, nil }
func (m *MockProcessGraphDB) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	return 0, nil
}
func (m *MockProcessGraphDB) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	return 0, nil
}
func (m *MockProcessGraphDB) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	return 0, nil
}
func (m *MockProcessGraphDB) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	return 0, nil
}

// MockServer implements Server interface
type MockServer struct {
	controller      *MockController
	validator       *MockValidator
	processGraphDB  *MockProcessGraphDB
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

func (m *MockServer) Validator() Validator {
	return m.validator
}

func (m *MockServer) Controller() Controller {
	return m.controller
}

func (m *MockServer) ProcessGraphDB() database.ProcessGraphDatabase {
	return m.processGraphDB
}

// Helper to create test process graph
func createTestProcessGraph() *core.ProcessGraph {
	return &core.ProcessGraph{
		ID:         "processgraph-123",
		ColonyName: "test-colony",
		State:      core.WAITING,
	}
}

// Helper to create test process
func createTestProcess() *core.Process {
	funcSpec := &core.FunctionSpec{
		NodeName: "test-node",
		FuncName: "test-func",
		Conditions: core.Conditions{
			ColonyName: "test-colony",
		},
	}
	return core.CreateProcess(funcSpec)
}

// Helper to create test workflow spec
func createTestWorkflowSpec() *core.WorkflowSpec {
	return &core.WorkflowSpec{
		ColonyName: "test-colony",
		FunctionSpecs: []core.FunctionSpec{
			{
				NodeName: "task1",
				FuncName: "test-func",
				Conditions: core.Conditions{
					ColonyName: "test-colony",
				},
			},
		},
	}
}

// Helper to create test function spec
func createTestFunctionSpec() *core.FunctionSpec {
	return &core.FunctionSpec{
		NodeName: "child-node",
		FuncName: "child-func",
		Conditions: core.Conditions{
			ColonyName: "test-colony",
		},
	}
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	processGraph := createTestProcessGraph()
	process := createTestProcess()
	controller := &MockController{
		processGraph:  processGraph,
		processGraphs: []*core.ProcessGraph{processGraph},
		addedProcess:  process,
	}
	validator := &MockValidator{}

	server := &MockServer{
		controller:     controller,
		validator:      validator,
		processGraphDB: &MockProcessGraphDB{},
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

// Tests for HandleSubmitWorkflow
func TestHandleSubmitWorkflow_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	workflowSpec := createTestWorkflowSpec()
	msg := rpc.CreateSubmitWorkflowSpecMsg(workflowSpec)
	jsonString, _ := msg.ToJSON()

	handlers.HandleSubmitWorkflow(ctx, "user-123", rpc.SubmitWorkflowSpecPayloadType, jsonString)

	assert.Equal(t, rpc.SubmitWorkflowSpecPayloadType, server.lastPayloadType)
}

func TestHandleSubmitWorkflow_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleSubmitWorkflow(ctx, "user-123", rpc.SubmitWorkflowSpecPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleSubmitWorkflow_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	workflowSpec := createTestWorkflowSpec()
	msg := rpc.CreateSubmitWorkflowSpecMsg(workflowSpec)
	jsonString, _ := msg.ToJSON()

	handlers.HandleSubmitWorkflow(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleSubmitWorkflow_NilWorkflowSpec(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateSubmitWorkflowSpecMsg(nil)
	jsonString, _ := msg.ToJSON()

	handlers.HandleSubmitWorkflow(ctx, "user-123", rpc.SubmitWorkflowSpecPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleSubmitWorkflow_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	workflowSpec := createTestWorkflowSpec()
	msg := rpc.CreateSubmitWorkflowSpecMsg(workflowSpec)
	jsonString, _ := msg.ToJSON()

	handlers.HandleSubmitWorkflow(ctx, "user-123", rpc.SubmitWorkflowSpecPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleSubmitWorkflow_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.submitErr = errors.New("controller error")
	handlers := NewHandlers(server)

	workflowSpec := createTestWorkflowSpec()
	msg := rpc.CreateSubmitWorkflowSpecMsg(workflowSpec)
	jsonString, _ := msg.ToJSON()

	handlers.HandleSubmitWorkflow(ctx, "user-123", rpc.SubmitWorkflowSpecPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetProcessGraph
func TestHandleGetProcessGraph_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraph(ctx, "user-123", rpc.GetProcessGraphPayloadType, jsonString)

	assert.Equal(t, rpc.GetProcessGraphPayloadType, server.lastPayloadType)
}

func TestHandleGetProcessGraph_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetProcessGraph(ctx, "user-123", rpc.GetProcessGraphPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraph_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraph(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraph_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraph(ctx, "user-123", rpc.GetProcessGraphPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleGetProcessGraph_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraph(ctx, "user-123", rpc.GetProcessGraphPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleGetProcessGraphs
func TestHandleGetProcessGraphs_Waiting_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.WAITING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, rpc.GetProcessGraphsPayloadType, server.lastPayloadType)
}

func TestHandleGetProcessGraphs_Running_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.RUNNING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, rpc.GetProcessGraphsPayloadType, server.lastPayloadType)
}

func TestHandleGetProcessGraphs_Success_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.SUCCESS)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, rpc.GetProcessGraphsPayloadType, server.lastPayloadType)
}

func TestHandleGetProcessGraphs_Failed_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.FAILED)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, rpc.GetProcessGraphsPayloadType, server.lastPayloadType)
}

func TestHandleGetProcessGraphs_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.WAITING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.WAITING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_InvalidState(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, -999)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_WaitingError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.findWaitingErr = errors.New("waiting error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.WAITING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_RunningError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.findRunningErr = errors.New("running error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.RUNNING)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_SuccessError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.findSuccessErr = errors.New("success error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.SUCCESS)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetProcessGraphs_FailedError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.findFailedErr = errors.New("failed error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetProcessGraphsMsg("test-colony", 10, core.FAILED)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetProcessGraphs(ctx, "user-123", rpc.GetProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleRemoveProcessGraph
func TestHandleRemoveProcessGraph_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveProcessGraph(ctx, "user-123", rpc.RemoveProcessGraphPayloadType, jsonString)

	assert.Equal(t, rpc.RemoveProcessGraphPayloadType, server.lastPayloadType)
	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveProcessGraph_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveProcessGraph(ctx, "user-123", rpc.RemoveProcessGraphPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveProcessGraph_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveProcessGraph(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveProcessGraph_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveProcessGraph(ctx, "user-123", rpc.RemoveProcessGraphPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleRemoveProcessGraph_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveProcessGraph(ctx, "user-123", rpc.RemoveProcessGraphPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveProcessGraph_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.removeErr = errors.New("remove error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveProcessGraphMsg("processgraph-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveProcessGraph(ctx, "user-123", rpc.RemoveProcessGraphPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleRemoveAllProcessGraphs
func TestHandleRemoveAllProcessGraphs_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllProcessGraphsMsg("test-colony")
	msg.State = core.WAITING
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllProcessGraphs(ctx, "owner-123", rpc.RemoveAllProcessGraphsPayloadType, jsonString)

	assert.Equal(t, rpc.RemoveAllProcessGraphsPayloadType, server.lastPayloadType)
	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveAllProcessGraphs_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveAllProcessGraphs(ctx, "owner-123", rpc.RemoveAllProcessGraphsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveAllProcessGraphs_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllProcessGraphsMsg("test-colony")
	msg.State = core.WAITING
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllProcessGraphs(ctx, "owner-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveAllProcessGraphs_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("colony owner error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllProcessGraphsMsg("test-colony")
	msg.State = core.WAITING
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllProcessGraphs(ctx, "owner-123", rpc.RemoveAllProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveAllProcessGraphs_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.removeAllErr = errors.New("remove all error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllProcessGraphsMsg("test-colony")
	msg.State = core.WAITING
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllProcessGraphs(ctx, "owner-123", rpc.RemoveAllProcessGraphsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleAddChild
func TestHandleAddChild_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, rpc.AddChildPayloadType, server.lastPayloadType)
}

func TestHandleAddChild_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddChild_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddChild_NilFunctionSpec(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", nil, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddChild_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleAddChild_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.addChildErr = errors.New("add child error")
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddChild_ReturnsNil(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.returnNilChild = true
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleAddChild_WithInsert(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	funcSpec := createTestFunctionSpec()
	msg := rpc.CreateAddChildMsg("processgraph-123", "parent-123", "child-123", funcSpec, true)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddChild(ctx, "user-123", rpc.AddChildPayloadType, jsonString)

	assert.Equal(t, rpc.AddChildPayloadType, server.lastPayloadType)
}
