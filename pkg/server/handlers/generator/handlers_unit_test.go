package generator

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

// MockGeneratorDB implements database.GeneratorDatabase
type MockGeneratorDB struct {
	generators       []*core.Generator
	generatorArgs    []*core.GeneratorArg
	addErr           error
	getByIDErr       error
	getByNameErr     error
	findByColonyErr  error
	countArgsErr     error
	returnNilByID    bool
	returnNilByName  bool
}

func (m *MockGeneratorDB) AddGenerator(generator *core.Generator) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.generators = append(m.generators, generator)
	return nil
}

func (m *MockGeneratorDB) SetGeneratorLastRun(generatorID string) error {
	return nil
}

func (m *MockGeneratorDB) SetGeneratorFirstPack(generatorID string) error {
	return nil
}

func (m *MockGeneratorDB) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, g := range m.generators {
		if g.ID == generatorID {
			return g, nil
		}
	}
	if len(m.generators) > 0 {
		return m.generators[0], nil
	}
	return nil, nil
}

func (m *MockGeneratorDB) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, g := range m.generators {
		if g.Name == name && g.ColonyName == colonyName {
			return g, nil
		}
	}
	if len(m.generators) > 0 {
		return m.generators[0], nil
	}
	return nil, nil
}

func (m *MockGeneratorDB) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	if m.findByColonyErr != nil {
		return nil, m.findByColonyErr
	}
	return m.generators, nil
}

func (m *MockGeneratorDB) FindAllGenerators() ([]*core.Generator, error) {
	return m.generators, nil
}

func (m *MockGeneratorDB) RemoveGeneratorByID(generatorID string) error {
	return nil
}

func (m *MockGeneratorDB) RemoveAllGeneratorsByColonyName(colonyName string) error {
	return nil
}

func (m *MockGeneratorDB) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	m.generatorArgs = append(m.generatorArgs, generatorArg)
	return nil
}

func (m *MockGeneratorDB) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	return m.generatorArgs, nil
}

func (m *MockGeneratorDB) CountGeneratorArgs(generatorID string) (int, error) {
	if m.countArgsErr != nil {
		return 0, m.countArgsErr
	}
	return len(m.generatorArgs), nil
}

func (m *MockGeneratorDB) RemoveGeneratorArgByID(generatorArgsID string) error {
	return nil
}

func (m *MockGeneratorDB) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error {
	return nil
}

func (m *MockGeneratorDB) RemoveAllGeneratorArgsByColonyName(generatorID string) error {
	return nil
}

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	executors     []*core.Executor
	getByIDErr    error
	returnNilByID bool
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                            { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error { return nil }
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                              { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) { return nil, nil }
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error                        { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error                         { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error                              { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error           { return nil }
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error                  { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                                         { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error)            { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) { return 0, nil }
func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName, executorName string, cap core.Capabilities) error { return nil }

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

// MockUserDB implements database.UserDatabase
type MockUserDB struct {
	users         []*core.User
	getByIDErr    error
	returnNilByID bool
}

func (m *MockUserDB) AddUser(user *core.User) error                            { return nil }
func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error) { return nil, nil }
func (m *MockUserDB) GetUserByName(colonyName string, name string) (*core.User, error) { return nil, nil }
func (m *MockUserDB) RemoveUserByID(colonyName string, userID string) error    { return nil }
func (m *MockUserDB) RemoveUserByName(colonyName string, name string) error    { return nil }
func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error          { return nil }
func (m *MockUserDB) CountUsers() (int, error)                                 { return 0, nil }
func (m *MockUserDB) CountUsersByColonyName(colonyName string) (int, error)    { return 0, nil }

func (m *MockUserDB) GetUserByID(colonyName string, userID string) (*core.User, error) {
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

// MockController implements Controller interface
type MockController struct {
	addErr      error
	packErr     error
	removeErr   error
	addedGen    *core.Generator
	period      int
	returnNil   bool
}

func (m *MockController) AddGenerator(generator *core.Generator) (*core.Generator, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	if m.returnNil {
		return nil, nil
	}
	m.addedGen = generator
	return generator, nil
}

func (m *MockController) PackGenerator(generatorID string, colonyName string, arg string) error {
	if m.packErr != nil {
		return m.packErr
	}
	return nil
}

func (m *MockController) RemoveGenerator(generatorID string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func (m *MockController) GetGeneratorPeriod() int {
	return m.period
}

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
	generatorDB     *MockGeneratorDB
	executorDB      *MockExecutorDB
	userDB          *MockUserDB
	controller      *MockController
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

func (m *MockServer) GeneratorController() Controller {
	return m.controller
}

func (m *MockServer) GeneratorDB() database.GeneratorDatabase {
	return m.generatorDB
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) UserDB() database.UserDatabase {
	return m.userDB
}

// Helper to create valid workflow spec JSON
func createValidWorkflowSpecJSON() string {
	return `{"colonyname": "test-colony", "functionspecs": [{"nodename": "task1", "funcname": "test", "conditions": {"colonyname": "test-colony"}}]}`
}

// Helper to create test generator
func createTestGenerator() *core.Generator {
	return &core.Generator{
		ID:           "generator-123",
		ColonyName:   "test-colony",
		Name:         "test-generator",
		WorkflowSpec: createValidWorkflowSpecJSON(),
	}
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	generator := createTestGenerator()
	generatorDB := &MockGeneratorDB{generators: []*core.Generator{generator}}
	executorDB := &MockExecutorDB{executors: []*core.Executor{{ID: "executor-123", Name: "test-executor"}}}
	userDB := &MockUserDB{}
	controller := &MockController{period: 1000}
	validator := &MockValidator{}

	server := &MockServer{
		generatorDB: generatorDB,
		executorDB:  executorDB,
		userDB:      userDB,
		controller:  controller,
		validator:   validator,
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

// Tests for HandleAddGenerator
func TestHandleAddGenerator_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	generator := createTestGenerator()
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, jsonString)

	assert.Equal(t, rpc.AddGeneratorPayloadType, server.lastPayloadType)
}

func TestHandleAddGenerator_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddGenerator_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	generator := createTestGenerator()
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddGenerator_NilGenerator(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateAddGeneratorMsg(nil)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddGenerator_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	generator := createTestGenerator()
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleAddGenerator_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.addErr = errors.New("controller error")
	handlers := NewHandlers(server)

	generator := createTestGenerator()
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddGenerator_ReturnsNil(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.returnNil = true
	handlers := NewHandlers(server)

	generator := createTestGenerator()
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddGenerator(ctx, "executor-123", rpc.AddGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetGenerator
func TestHandleGetGenerator_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerator(ctx, "test-user", rpc.GetGeneratorPayloadType, jsonString)

	assert.Equal(t, rpc.GetGeneratorPayloadType, server.lastPayloadType)
}

func TestHandleGetGenerator_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetGenerator(ctx, "test-user", rpc.GetGeneratorPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetGenerator_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerator(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetGenerator_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.generatorDB.returnNilByID = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorMsg("nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerator(ctx, "test-user", rpc.GetGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleGetGenerator_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerator(ctx, "test-user", rpc.GetGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleResolveGenerator
func TestHandleResolveGenerator_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateResolveGeneratorMsg("test-colony", "test-generator")
	jsonString, _ := msg.ToJSON()

	handlers.HandleResolveGenerator(ctx, "test-user", rpc.ResolveGeneratorPayloadType, jsonString)

	assert.Equal(t, rpc.ResolveGeneratorPayloadType, server.lastPayloadType)
}

func TestHandleResolveGenerator_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleResolveGenerator(ctx, "test-user", rpc.ResolveGeneratorPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleResolveGenerator_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateResolveGeneratorMsg("test-colony", "test-generator")
	jsonString, _ := msg.ToJSON()

	handlers.HandleResolveGenerator(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleResolveGenerator_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.generatorDB.returnNilByName = true
	handlers := NewHandlers(server)

	msg := rpc.CreateResolveGeneratorMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleResolveGenerator(ctx, "test-user", rpc.ResolveGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetGenerators
func TestHandleGetGenerators_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorsMsg("test-colony", 100)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerators(ctx, "test-user", rpc.GetGeneratorsPayloadType, jsonString)

	assert.Equal(t, rpc.GetGeneratorsPayloadType, server.lastPayloadType)
}

func TestHandleGetGenerators_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetGenerators(ctx, "test-user", rpc.GetGeneratorsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetGenerators_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorsMsg("test-colony", 100)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerators(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetGenerators_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetGeneratorsMsg("test-colony", 100)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetGenerators(ctx, "test-user", rpc.GetGeneratorsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandlePackGenerator
func TestHandlePackGenerator_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreatePackGeneratorMsg("generator-123", "test-arg")
	jsonString, _ := msg.ToJSON()

	handlers.HandlePackGenerator(ctx, "test-user", rpc.PackGeneratorPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandlePackGenerator_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandlePackGenerator(ctx, "test-user", rpc.PackGeneratorPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandlePackGenerator_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreatePackGeneratorMsg("generator-123", "test-arg")
	jsonString, _ := msg.ToJSON()

	handlers.HandlePackGenerator(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandlePackGenerator_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.generatorDB.returnNilByID = true
	handlers := NewHandlers(server)

	msg := rpc.CreatePackGeneratorMsg("nonexistent", "test-arg")
	jsonString, _ := msg.ToJSON()

	handlers.HandlePackGenerator(ctx, "test-user", rpc.PackGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandlePackGenerator_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreatePackGeneratorMsg("generator-123", "test-arg")
	jsonString, _ := msg.ToJSON()

	handlers.HandlePackGenerator(ctx, "test-user", rpc.PackGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandlePackGenerator_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.packErr = errors.New("pack error")
	handlers := NewHandlers(server)

	msg := rpc.CreatePackGeneratorMsg("generator-123", "test-arg")
	jsonString, _ := msg.ToJSON()

	handlers.HandlePackGenerator(ctx, "test-user", rpc.PackGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleRemoveGenerator
func TestHandleRemoveGenerator_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveGenerator(ctx, "test-user", rpc.RemoveGeneratorPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveGenerator_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveGenerator(ctx, "test-user", rpc.RemoveGeneratorPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveGenerator_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveGenerator(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveGenerator_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.generatorDB.returnNilByID = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveGeneratorMsg("nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveGenerator(ctx, "test-user", rpc.RemoveGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleRemoveGenerator_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveGenerator(ctx, "test-user", rpc.RemoveGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveGenerator_ControllerError(t *testing.T) {
	server, ctx := createMockServer()
	server.controller.removeErr = errors.New("remove error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveGeneratorMsg("generator-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveGenerator(ctx, "test-user", rpc.RemoveGeneratorPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}
