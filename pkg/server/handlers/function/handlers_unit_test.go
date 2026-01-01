package function

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

// MockFunctionDB implements database.FunctionDatabase
type MockFunctionDB struct {
	functions         []*core.Function
	addFunctionErr    error
	getFunctionErr    error
	getByColonyErr    error
	getByExecutorErr  error
	removeErr         error
	returnNilByID     bool
}

func (m *MockFunctionDB) AddFunction(function *core.Function) error {
	if m.addFunctionErr != nil {
		return m.addFunctionErr
	}
	m.functions = append(m.functions, function)
	return nil
}

func (m *MockFunctionDB) GetFunctionByID(functionID string) (*core.Function, error) {
	if m.getFunctionErr != nil {
		return nil, m.getFunctionErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, f := range m.functions {
		if f.FunctionID == functionID {
			return f, nil
		}
	}
	return nil, nil
}

func (m *MockFunctionDB) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	if m.getByColonyErr != nil {
		return nil, m.getByColonyErr
	}
	var result []*core.Function
	for _, f := range m.functions {
		if f.ColonyName == colonyName {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *MockFunctionDB) GetFunctionsByExecutorName(colonyName, executorName string) ([]*core.Function, error) {
	if m.getByExecutorErr != nil {
		return nil, m.getByExecutorErr
	}
	var result []*core.Function
	for _, f := range m.functions {
		if f.ColonyName == colonyName && f.ExecutorName == executorName {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *MockFunctionDB) RemoveFunctionByID(functionID string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func (m *MockFunctionDB) RemoveFunctionsByColonyName(colonyName string) error { return nil }
func (m *MockFunctionDB) RemoveFunctionsByExecutorName(colonyName, executorName string) error {
	return nil
}
func (m *MockFunctionDB) RemoveFunctionByName(colonyName, executorName, name string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}
func (m *MockFunctionDB) UpdateFunctionStats(colonyName, executorName, name string, counter int, minWaitTime, maxWaitTime, minExecTime, maxExecTime, avgWaitTime, avgExecTime float64) error {
	return nil
}
func (m *MockFunctionDB) RemoveFunctions() error { return nil }
func (m *MockFunctionDB) GetFunctionsByExecutorAndName(colonyName, executorName, name string) (*core.Function, error) {
	for _, f := range m.functions {
		if f.ColonyName == colonyName && f.ExecutorName == executorName && f.FuncName == name {
			return f, nil
		}
	}
	return nil, nil
}

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	executors       []*core.Executor
	getByNameErr    error
	returnNilByName bool
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                            { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error { return nil }
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                              { return nil, nil }
func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error)            { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error                        { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error                         { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error                              { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error           { return nil }
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error                  { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                                         { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error)            { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(string, int) (int, error)          { return 0, nil }
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName, name string, cap core.Capabilities) error { return nil }

func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, e := range m.executors {
		if e.ColonyName == colonyName && e.Name == executorName {
			return e, nil
		}
	}
	return nil, nil
}

// MockUserDB implements database.UserDatabase
type MockUserDB struct{}

func (m *MockUserDB) AddUser(user *core.User) error                                          { return nil }
func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error)           { return nil, nil }
func (m *MockUserDB) GetUserByID(colonyName, userID string) (*core.User, error)              { return nil, nil }
func (m *MockUserDB) GetUserByName(colonyName, name string) (*core.User, error)              { return nil, nil }
func (m *MockUserDB) RemoveUserByID(colonyName, userID string) error                         { return nil }
func (m *MockUserDB) RemoveUserByName(colonyName, name string) error                         { return nil }
func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error                        { return nil }

// MockValidator implements security.Validator
type MockValidator struct {
	requireMembershipErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID, serverID string) error { return nil }
func (m *MockValidator) RequireColonyOwner(recoveredID, colonyName string) error { return nil }

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
	functionDB           *MockFunctionDB
	executorDB           *MockExecutorDB
	userDB               *MockUserDB
	validator            *MockValidator
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

func (m *MockServer) FunctionDB() database.FunctionDatabase {
	return m.functionDB
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) UserDB() database.UserDatabase {
	return m.userDB
}

// Helper functions
func createTestMocks() (*MockServer, *MockFunctionDB, *MockExecutorDB, *MockValidator, *MockContext) {
	functionDB := &MockFunctionDB{}
	executorDB := &MockExecutorDB{}
	validator := &MockValidator{}
	ctx := &MockContext{}
	server := &MockServer{
		functionDB: functionDB,
		executorDB: executorDB,
		userDB:     &MockUserDB{},
		validator:  validator,
	}
	return server, functionDB, executorDB, validator, ctx
}

func createTestFunction(functionID, funcName, executorName, colonyName string) *core.Function {
	return &core.Function{
		FunctionID:   functionID,
		FuncName:     funcName,
		ExecutorName: executorName,
		ColonyName:   colonyName,
	}
}

func createTestExecutor(id, name, colonyName string) *core.Executor {
	return &core.Executor{
		ID:         id,
		Name:       name,
		ColonyName: colonyName,
	}
}

// =============================================
// Tests for HandleAddFunction
// =============================================

func TestHandleAddFunction_Success(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
	assert.Len(t, functionDB.functions, 1)
}

func TestHandleAddFunction_InvalidJSON(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddFunction_MsgTypeMismatch(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddFunction_NilFunction(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := &rpc.AddFunctionMsg{
		MsgType:  rpc.AddFunctionPayloadType,
		Function: nil,
	}
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddFunction_AuthError(t *testing.T) {
	server, _, executorDB, validator, ctx := createTestMocks()
	validator.requireMembershipErr = errors.New("not a member")
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleAddFunction_ExecutorDBError(t *testing.T) {
	server, _, executorDB, _, ctx := createTestMocks()
	executorDB.getByNameErr = errors.New("database error")
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleAddFunction_ExecutorNotFound(t *testing.T) {
	server, _, executorDB, _, ctx := createTestMocks()
	executorDB.returnNilByName = true
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleAddFunction_WrongExecutor(t *testing.T) {
	server, _, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("different-exec", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleAddFunction_AddFunctionDBError(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	functionDB.addFunctionErr = errors.New("add error")
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleAddFunction_GetAddedFunctionError(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	// First call to AddFunction succeeds, but GetFunctionByID fails
	functionDB.getFunctionErr = errors.New("get error")

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleAddFunction_GeneratesIDIfEmpty(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("", "test-func", "test-executor", "test-colony")
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, _ := msg.ToJSON()

	h.HandleAddFunction(ctx, "exec-123", rpc.AddFunctionPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
	assert.Len(t, functionDB.functions, 1)
	assert.NotEmpty(t, functionDB.functions[0].FunctionID)
}

// =============================================
// Tests for HandleGetFunctions
// =============================================

func TestHandleGetFunctions_ByColony_Success(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
}

func TestHandleGetFunctions_ByExecutor_Success(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "test-executor")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, jsonString)

	assert.True(t, server.httpReplyCalled)
}

func TestHandleGetFunctions_InvalidJSON(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetFunctions_MsgTypeMismatch(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleGetFunctions_AuthError(t *testing.T) {
	server, _, _, validator, ctx := createTestMocks()
	validator.requireMembershipErr = errors.New("not a member")
	h := NewHandlers(server)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleGetFunctions_GetByColonyError(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	functionDB.getByColonyErr = errors.New("database error")
	h := NewHandlers(server)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleGetFunctions_GetByExecutorError(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	functionDB.getByExecutorErr = errors.New("database error")
	h := NewHandlers(server)

	msg := rpc.CreateGetFunctionsMsg("test-colony", "test-executor")
	jsonString, _ := msg.ToJSON()

	h.HandleGetFunctions(ctx, "member-123", rpc.GetFunctionsPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

// =============================================
// Tests for HandleRemoveFunction
// =============================================

func TestHandleRemoveFunction_Success(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.emptyReplyCalled)
}

func TestHandleRemoveFunction_InvalidJSON(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, "invalid json")

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveFunction_MsgTypeMismatch(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", "wrong_type", jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveFunction_EmptyFunctionID(t *testing.T) {
	server, _, _, _, ctx := createTestMocks()
	h := NewHandlers(server)

	msg := rpc.CreateRemoveFunctionMsg("")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveFunction_FunctionNotFound(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	functionDB.returnNilByID = true
	h := NewHandlers(server)

	msg := rpc.CreateRemoveFunctionMsg("non-existent")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusNotFound, server.httpErrorCode)
}

func TestHandleRemoveFunction_GetFunctionError(t *testing.T) {
	server, functionDB, _, _, ctx := createTestMocks()
	functionDB.getFunctionErr = errors.New("database error")
	h := NewHandlers(server)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusBadRequest, server.httpErrorCode)
}

func TestHandleRemoveFunction_ExecutorDBError(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	executorDB.getByNameErr = errors.New("database error")
	h := NewHandlers(server)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

func TestHandleRemoveFunction_AuthError(t *testing.T) {
	server, functionDB, executorDB, validator, ctx := createTestMocks()
	validator.requireMembershipErr = errors.New("not a member")
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleRemoveFunction_WrongExecutor(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	h := NewHandlers(server)

	executor := createTestExecutor("different-exec", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusForbidden, server.httpErrorCode)
}

func TestHandleRemoveFunction_RemoveError(t *testing.T) {
	server, functionDB, executorDB, _, ctx := createTestMocks()
	functionDB.removeErr = errors.New("remove error")
	h := NewHandlers(server)

	executor := createTestExecutor("exec-123", "test-executor", "test-colony")
	executorDB.executors = append(executorDB.executors, executor)

	function := createTestFunction("func-123", "test-func", "test-executor", "test-colony")
	functionDB.functions = append(functionDB.functions, function)

	msg := rpc.CreateRemoveFunctionMsg("func-123")
	jsonString, _ := msg.ToJSON()

	h.HandleRemoveFunction(ctx, "exec-123", rpc.RemoveFunctionPayloadType, jsonString)

	assert.True(t, server.httpErrorCalled)
	assert.Equal(t, http.StatusInternalServerError, server.httpErrorCode)
}

// =============================================
// Tests for NewHandlers and RegisterHandlers
// =============================================

func TestNewHandlers(t *testing.T) {
	server, _, _, _, _ := createTestMocks()
	h := NewHandlers(server)
	assert.NotNil(t, h)
}

func TestRegisterHandlers_Success(t *testing.T) {
	server, _, _, _, _ := createTestMocks()
	h := NewHandlers(server)

	reg := registry.NewHandlerRegistry()
	err := h.RegisterHandlers(reg)

	assert.NoError(t, err)
}

func TestRegisterHandlers_DuplicateError(t *testing.T) {
	server, _, _, _, _ := createTestMocks()
	h := NewHandlers(server)

	reg := registry.NewHandlerRegistry()

	// Register once - should succeed
	err := h.RegisterHandlers(reg)
	assert.NoError(t, err)

	// Register again - should fail on first duplicate
	err = h.RegisterHandlers(reg)
	assert.Error(t, err)
}
