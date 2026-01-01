package cron

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// MockCronDB implements database.CronDatabase
type MockCronDB struct {
	crons           []*core.Cron
	addCronErr      error
	getByIDErr      error
	findByColonyErr error
	removeErr       error
	returnNilByID   bool
}

func (m *MockCronDB) AddCron(cron *core.Cron) error {
	if m.addCronErr != nil {
		return m.addCronErr
	}
	m.crons = append(m.crons, cron)
	return nil
}

func (m *MockCronDB) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error {
	return nil
}

func (m *MockCronDB) GetCronByID(cronID string) (*core.Cron, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, c := range m.crons {
		if c.ID == cronID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockCronDB) GetCronByName(colonyName string, cronName string) (*core.Cron, error) {
	for _, c := range m.crons {
		if c.ColonyName == colonyName && c.Name == cronName {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockCronDB) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) {
	if m.findByColonyErr != nil {
		return nil, m.findByColonyErr
	}
	var result []*core.Cron
	for _, c := range m.crons {
		if c.ColonyName == colonyName {
			result = append(result, c)
			if len(result) >= count {
				break
			}
		}
	}
	return result, nil
}

func (m *MockCronDB) FindAllCrons() ([]*core.Cron, error) {
	return m.crons, nil
}

func (m *MockCronDB) RemoveCronByID(cronID string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func (m *MockCronDB) RemoveAllCronsByColonyName(colonyName string) error {
	return nil
}

// MockCronController implements the CronController interface
type MockCronController struct {
	crons          []*core.Cron
	addCronErr     error
	runCronErr     error
	removeCronErr  error
	cronPeriod     int
	returnNilOnAdd bool
	returnNilOnRun bool
}

func (m *MockCronController) AddCron(cron *core.Cron) (*core.Cron, error) {
	if m.addCronErr != nil {
		return nil, m.addCronErr
	}
	if m.returnNilOnAdd {
		return nil, nil
	}
	m.crons = append(m.crons, cron)
	return cron, nil
}

func (m *MockCronController) RunCron(cronID string) (*core.Cron, error) {
	if m.runCronErr != nil {
		return nil, m.runCronErr
	}
	if m.returnNilOnRun {
		return nil, nil
	}
	for _, c := range m.crons {
		if c.ID == cronID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockCronController) RemoveCron(cronID string) error {
	if m.removeCronErr != nil {
		return m.removeCronErr
	}
	return nil
}

func (m *MockCronController) GetCronPeriod() int {
	return m.cronPeriod
}

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

// MockValidator implements security.Validator
type MockValidator struct {
	requireMembershipErr error
}

func (m *MockValidator) RequireServerOwner(recoveredID, serverID string) error   { return nil }
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

// MockServer implements the Server interface for cron handlers
type MockServer struct {
	cronDB         *MockCronDB
	executorDB     *MockExecutorDB
	userDB         *MockUserDB
	validator      *MockValidator
	cronController *MockCronController
	httpErrorCode  int
	replyPayload   string
	replyType      string
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

func (m *MockServer) CronDB() database.CronDatabase {
	return m.cronDB
}

func (m *MockServer) CronController() interface {
	AddCron(cron *core.Cron) (*core.Cron, error)
	RunCron(cronID string) (*core.Cron, error)
	RemoveCron(cronID string) error
	GetCronPeriod() int
} {
	return m.cronController
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) UserDB() database.UserDatabase {
	return m.userDB
}

// Test setup helpers
func createTestCron() *core.Cron {
	return &core.Cron{
		ID:         "cron-123",
		ColonyName: "test-colony",
		Name:       "test-cron",
		Interval:   60,
		Random:     false,
		WorkflowSpec: `{
			"colonyname": "test-colony",
			"funcspecs": [
				{
					"func": "test-func",
					"conditions": {
						"executortype": "test-type"
					}
				}
			]
		}`,
	}
}

func createMockServer() *MockServer {
	executor := &core.Executor{
		ID:         "executor-123",
		Name:       "test-executor",
		ColonyName: "test-colony",
		Type:       "test-type",
	}

	cronDB := &MockCronDB{}
	cronController := &MockCronController{cronPeriod: 10}
	executorDB := &MockExecutorDB{executors: []*core.Executor{executor}}
	userDB := &MockUserDB{}
	validator := &MockValidator{}

	return &MockServer{
		cronDB:         cronDB,
		cronController: cronController,
		executorDB:     executorDB,
		userDB:         userDB,
		validator:      validator,
	}
}

// HandleAddCron tests
func TestHandleAddCron_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.AddCronPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleAddCron_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_NilCron(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateAddCronMsg(nil)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleAddCron_InvalidWorkflowSpec(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	cron.WorkflowSpec = "invalid workflow spec"
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_ZeroInterval(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	cron.Interval = 0
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_CronControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronController.addCronErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddCron_ControllerReturnsNil(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronController.returnNilOnAdd = true
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "executor-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleAddCron_ExecutorNotFoundUsesUser(t *testing.T) {
	mockServer := createMockServer()
	mockServer.executorDB.returnNilByID = true
	mockServer.userDB.users = []*core.User{
		{ID: "user-123", Name: "test-user", ColonyName: "test-colony"},
	}
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "user-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
}

func TestHandleAddCron_InitiatorNotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.executorDB.returnNilByID = true
	mockServer.userDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	cron := createTestCron()
	msg := rpc.CreateAddCronMsg(cron)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddCron(ctx, "unknown-123", rpc.AddCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// HandleGetCron tests
func TestHandleGetCron_Success(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	mockServer.cronController.crons = []*core.Cron{cron}
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", rpc.GetCronPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.GetCronPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleGetCron_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", rpc.GetCronPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCron_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCron_DBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", rpc.GetCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCron_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", rpc.GetCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleGetCron_AuthError(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCron(ctx, "executor-123", rpc.GetCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

// HandleGetCrons tests
func TestHandleGetCrons_Success(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronsMsg("test-colony", 10)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", rpc.GetCronsPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.GetCronsPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleGetCrons_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", rpc.GetCronsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCrons_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronsMsg("test-colony", 10)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCrons_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronsMsg("test-colony", 10)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", rpc.GetCronsPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleGetCrons_DBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronDB.findByColonyErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronsMsg("test-colony", 10)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", rpc.GetCronsPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetCrons_EmptyResult(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetCronsMsg("empty-colony", 10)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetCrons(ctx, "executor-123", rpc.GetCronsPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.GetCronsPayloadType, mockServer.replyType)
	// Should return empty array, not nil
	assert.Contains(t, mockServer.replyPayload, "[]")
}

// HandleRunCron tests
func TestHandleRunCron_Success(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronController.crons = []*core.Cron{cron}
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRunCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", rpc.RunCronPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.RunCronPayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleRunCron_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", rpc.RunCronPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRunCron_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRunCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRunCron_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronController.runCronErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRunCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", rpc.RunCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRunCron_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronController.returnNilOnRun = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRunCronMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", rpc.RunCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleRunCron_AuthError(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronController.crons = []*core.Cron{cron}
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRunCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRunCron(ctx, "executor-123", rpc.RunCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

// HandleRemoveCron tests
func TestHandleRemoveCron_Success(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.RemoveCronPayloadType, mockServer.replyType)
}

func TestHandleRemoveCron_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveCron_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveCron_DBGetError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg("cron-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleRemoveCron_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.cronDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

func TestHandleRemoveCron_AuthError(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleRemoveCron_ControllerError(t *testing.T) {
	mockServer := createMockServer()
	cron := createTestCron()
	mockServer.cronDB.crons = []*core.Cron{cron}
	mockServer.cronController.removeCronErr = errors.New("controller error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateRemoveCronMsg(cron.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleRemoveCron(ctx, "executor-123", rpc.RemoveCronPayloadType, jsonStr)

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
