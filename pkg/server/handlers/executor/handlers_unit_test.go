package executor

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

func (m *MockContext) String(code int, format string, values ...interface{}) { m.statusCode = code }
func (m *MockContext) JSON(code int, obj interface{})                        { m.statusCode = code; m.response = obj }
func (m *MockContext) XML(code int, obj interface{})                         { m.statusCode = code }
func (m *MockContext) Data(code int, contentType string, data []byte)        { m.statusCode = code }
func (m *MockContext) Status(code int)                                       { m.statusCode = code }
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
func (m *MockContext) Get(key string) (interface{}, bool)                    { return nil, false }
func (m *MockContext) GetString(key string) string                           { return "" }
func (m *MockContext) GetBool(key string) bool                               { return false }
func (m *MockContext) GetInt(key string) int                                 { return 0 }
func (m *MockContext) GetInt64(key string) int64                             { return 0 }
func (m *MockContext) GetFloat64(key string) float64                         { return 0 }
func (m *MockContext) Abort()                                                {}
func (m *MockContext) AbortWithStatus(code int)                              { m.statusCode = code }
func (m *MockContext) AbortWithStatusJSON(code int, jsonObj interface{})     { m.statusCode = code; m.response = jsonObj }
func (m *MockContext) IsAborted() bool                                       { return false }
func (m *MockContext) Next()                                                 {}

type MockValidator struct {
	requireMembershipErr  error
	requireColonyOwnerErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID string, serverID string) error { return nil }
func (m *MockValidator) RequireColonyOwner(recoveredID string, colonyName string) error {
	return m.requireColonyOwnerErr
}
func (m *MockValidator) RequireMembership(recoveredID string, colonyName string, approved bool) error {
	return m.requireMembershipErr
}

type MockExecutorDB struct {
	executor      *core.Executor
	executors     []*core.Executor
	executorErr   error
	addErr        error
	approveErr    error
	rejectErr     error
	removeErr     error
	allocErr      error
	updateCapErr  error
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                    { return m.addErr }
func (m *MockExecutorDB) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	return m.allocErr
}
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                      { return m.executors, m.executorErr }
func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error)    { return m.executor, m.executorErr }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	return m.executors, m.executorErr
}
func (m *MockExecutorDB) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	return m.executor, m.executorErr
}
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	return nil, nil
}
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error { return m.approveErr }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error  { return m.rejectErr }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error       { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName string, executorName string) error {
	return m.removeErr
}
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                        { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error) { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	return 0, nil
}
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName string, executorName string, capabilities core.Capabilities) error {
	return m.updateCapErr
}

type MockServer struct {
	validator       *MockValidator
	executorDB      *MockExecutorDB
	httpError       bool
	lastErrCode     int
	lastResponse    string
	allowReregister bool
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

func (m *MockServer) AllowExecutorReregister() bool {
	return m.allowReregister
}

func createMockServer() *MockServer {
	return &MockServer{
		validator:  &MockValidator{},
		executorDB: &MockExecutorDB{},
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

// HandleAddExecutor tests

func TestHandleAddExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	executor := &core.Executor{Name: "test-executor", ColonyName: "test-colony"}
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorNilExecutorUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateAddExecutorMsg(nil)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorColonyOwnerErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	executor := &core.Executor{Name: "test-executor", ColonyName: "test-colony"}
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleAddExecutorDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.addErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	executor := &core.Executor{Name: "test-executor", ColonyName: "test-colony"}
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleAddExecutorNilResultUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	executor := &core.Executor{Name: "test-executor", ColonyName: "test-colony"}
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusInternalServerError, server.lastErrCode)
}

func TestHandleAddExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	// Need to allow reregister since GetExecutorByName will return the same executor
	server.allowReregister = true
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
		Type:       "test-type",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	executor := &core.Executor{Name: "test-executor", ColonyName: "test-colony"}
	msg := rpc.CreateAddExecutorMsg(executor)
	jsonStr, _ := msg.ToJSON()

	handlers.HandleAddExecutor(ctx, "test-id", rpc.AddExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleGetExecutors tests

func TestHandleGetExecutorsInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleGetExecutors(ctx, "test-id", rpc.GetExecutorsPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorsMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutors(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorsMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutors(ctx, "test-id", rpc.GetExecutorsPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleGetExecutorsDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executorErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutors(ctx, "test-id", rpc.GetExecutorsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorsSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executors = []*core.Executor{}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorsMsg("test-colony")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutors(ctx, "test-id", rpc.GetExecutorsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleGetExecutor tests

func TestHandleGetExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleGetExecutor(ctx, "test-id", rpc.GetExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executorErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutor(ctx, "test-id", rpc.GetExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutor(ctx, "test-id", rpc.GetExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusInternalServerError, server.lastErrCode)
}

func TestHandleGetExecutorMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutor(ctx, "test-id", rpc.GetExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleGetExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutor(ctx, "test-id", rpc.GetExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleGetExecutorByID tests

func TestHandleGetExecutorByIDInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleGetExecutorByID(ctx, "test-id", rpc.GetExecutorByIDPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorByIDMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorByIDMsg("test-colony", "executor-id")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutorByID(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorByIDDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executorErr = errors.New("db error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorByIDMsg("test-colony", "executor-id")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutorByID(ctx, "test-id", rpc.GetExecutorByIDPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleGetExecutorByIDNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorByIDMsg("test-colony", "executor-id")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutorByID(ctx, "test-id", rpc.GetExecutorByIDPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusNotFound, server.lastErrCode)
}

func TestHandleGetExecutorByIDMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorByIDMsg("test-colony", "executor-id")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutorByID(ctx, "test-id", rpc.GetExecutorByIDPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleGetExecutorByIDSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateGetExecutorByIDMsg("test-colony", "executor-id")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleGetExecutorByID(ctx, "test-id", rpc.GetExecutorByIDPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleApproveExecutor tests

func TestHandleApproveExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleApproveExecutor(ctx, "test-id", rpc.ApproveExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleApproveExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateApproveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleApproveExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleApproveExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateApproveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleApproveExecutor(ctx, "test-id", rpc.ApproveExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusInternalServerError, server.lastErrCode)
}

func TestHandleApproveExecutorColonyOwnerErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateApproveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleApproveExecutor(ctx, "test-id", rpc.ApproveExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleApproveExecutorDBErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.executorDB.approveErr = errors.New("approve error")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateApproveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleApproveExecutor(ctx, "test-id", rpc.ApproveExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleApproveExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateApproveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleApproveExecutor(ctx, "test-id", rpc.ApproveExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleRejectExecutor tests

func TestHandleRejectExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleRejectExecutor(ctx, "test-id", rpc.RejectExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleRejectExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRejectExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRejectExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleRejectExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRejectExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRejectExecutor(ctx, "test-id", rpc.RejectExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusInternalServerError, server.lastErrCode)
}

func TestHandleRejectExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRejectExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRejectExecutor(ctx, "test-id", rpc.RejectExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleRemoveExecutor tests

func TestHandleRemoveExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleRemoveExecutor(ctx, "test-id", rpc.RemoveExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleRemoveExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRemoveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRemoveExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleRemoveExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRemoveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRemoveExecutor(ctx, "test-id", rpc.RemoveExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusInternalServerError, server.lastErrCode)
}

func TestHandleRemoveExecutorColonyOwnerErrorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	server.validator.requireColonyOwnerErr = errors.New("not colony owner")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRemoveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRemoveExecutor(ctx, "test-id", rpc.RemoveExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleRemoveExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateRemoveExecutorMsg("test-colony", "test-executor")
	jsonStr, _ := msg.ToJSON()

	handlers.HandleRemoveExecutor(ctx, "test-id", rpc.RemoveExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleReportAllocations tests

func TestHandleReportAllocationsInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleReportAllocations(ctx, "test-id", rpc.ReportAllocationsPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleReportAllocationsMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateReportAllocationsMsg("test-colony", "test-executor", core.Allocations{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleReportAllocations(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleReportAllocationsMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateReportAllocationsMsg("test-colony", "test-executor", core.Allocations{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleReportAllocations(ctx, "test-id", rpc.ReportAllocationsPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleReportAllocationsExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateReportAllocationsMsg("test-colony", "test-executor", core.Allocations{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleReportAllocations(ctx, "test-id", rpc.ReportAllocationsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleReportAllocationsNotOwnExecutorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "different-executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateReportAllocationsMsg("test-colony", "test-executor", core.Allocations{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleReportAllocations(ctx, "test-id", rpc.ReportAllocationsPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleReportAllocationsSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateReportAllocationsMsg("test-colony", "test-executor", core.Allocations{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleReportAllocations(ctx, "test-id", rpc.ReportAllocationsPayloadType, jsonStr)
	assert.False(t, server.httpError)
}

// HandleUpdateExecutor tests

func TestHandleUpdateExecutorInvalidJSONUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	handlers.HandleUpdateExecutor(ctx, "test-id", rpc.UpdateExecutorPayloadType, "invalid json")
	assert.True(t, server.httpError)
}

func TestHandleUpdateExecutorMsgTypeMismatchUnit(t *testing.T) {
	server := createMockServer()
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateUpdateExecutorMsg("test-colony", "test-executor", core.Capabilities{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleUpdateExecutor(ctx, "test-id", "wrong-type", jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleUpdateExecutorMembershipErrorUnit(t *testing.T) {
	server := createMockServer()
	server.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateUpdateExecutorMsg("test-colony", "test-executor", core.Capabilities{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleUpdateExecutor(ctx, "test-id", rpc.UpdateExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
	assert.Equal(t, http.StatusForbidden, server.lastErrCode)
}

func TestHandleUpdateExecutorNilUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = nil
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateUpdateExecutorMsg("test-colony", "test-executor", core.Capabilities{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleUpdateExecutor(ctx, "test-id", rpc.UpdateExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleUpdateExecutorNotOwnExecutorUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "different-executor-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateUpdateExecutorMsg("test-colony", "test-executor", core.Capabilities{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleUpdateExecutor(ctx, "test-id", rpc.UpdateExecutorPayloadType, jsonStr)
	assert.True(t, server.httpError)
}

func TestHandleUpdateExecutorSuccessUnit(t *testing.T) {
	server := createMockServer()
	server.executorDB.executor = &core.Executor{
		ID:         "test-id",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	handlers := NewHandlers(server)
	ctx := &MockContext{}

	msg := rpc.CreateUpdateExecutorMsg("test-colony", "test-executor", core.Capabilities{})
	jsonStr, _ := msg.ToJSON()

	handlers.HandleUpdateExecutor(ctx, "test-id", rpc.UpdateExecutorPayloadType, jsonStr)
	assert.False(t, server.httpError)
}
