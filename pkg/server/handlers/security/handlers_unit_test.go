package security

import (
	"errors"
	"net/http"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	pkgsecurity "github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	"github.com/stretchr/testify/assert"
)

// MockUserDB implements database.UserDatabase
type MockUserDB struct {
	users         []*core.User
	getByIDErr    error
	returnNilByID bool
}

func (m *MockUserDB) AddUser(user *core.User) error                                         { return nil }
func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error)          { return nil, nil }
func (m *MockUserDB) GetUserByName(colonyName string, name string) (*core.User, error)      { return nil, nil }
func (m *MockUserDB) RemoveUserByID(colonyName string, userID string) error                 { return nil }
func (m *MockUserDB) RemoveUserByName(colonyName string, name string) error                 { return nil }
func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error                       { return nil }
func (m *MockUserDB) CountUsers() (int, error)                                              { return 0, nil }
func (m *MockUserDB) CountUsersByColonyName(colonyName string) (int, error)                 { return 0, nil }

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
	if len(m.users) > 0 {
		return m.users[0], nil
	}
	return nil, nil
}

// MockExecutorDB implements database.ExecutorDatabase
type MockExecutorDB struct {
	executors     []*core.Executor
	getByIDErr    error
	returnNilByID bool
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error                                            { return nil }
func (m *MockExecutorDB) SetAllocations(colonyName, executorName string, allocations core.Allocations) error   { return nil }
func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error)                                              { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error)                 { return nil, nil }
func (m *MockExecutorDB) GetExecutorByName(colonyName, executorName string) (*core.Executor, error)            { return nil, nil }
func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error)               { return nil, nil }
func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error                                        { return nil }
func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error                                         { return nil }
func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error                                              { return nil }
func (m *MockExecutorDB) RemoveExecutorByName(colonyName, executorName string) error                           { return nil }
func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error                                  { return nil }
func (m *MockExecutorDB) CountExecutors() (int, error)                                                         { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error)                            { return 0, nil }
func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error)         { return 0, nil }
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
	if len(m.executors) > 0 {
		return m.executors[0], nil
	}
	return nil, nil
}

// MockColonyDB implements database.ColonyDatabase
type MockColonyDB struct {
	colonies       []*core.Colony
	getByNameErr   error
	returnNilByName bool
}

func (m *MockColonyDB) AddColony(colony *core.Colony) error                       { return nil }
func (m *MockColonyDB) GetColonies() ([]*core.Colony, error)                      { return nil, nil }
func (m *MockColonyDB) GetColonyByID(colonyID string) (*core.Colony, error)       { return nil, nil }
func (m *MockColonyDB) RenameColony(colonyName string, newColonyName string) error { return nil }
func (m *MockColonyDB) RemoveColonyByName(colonyName string) error                { return nil }
func (m *MockColonyDB) CountColonies() (int, error)                               { return 0, nil }

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
	if len(m.colonies) > 0 {
		return m.colonies[0], nil
	}
	return nil, nil
}

// MockSecurityDB implements database.SecurityDatabase
type MockSecurityDB struct {
	changeUserIDErr     error
	changeExecutorIDErr error
	changeColonyIDErr   error
	setServerIDErr      error
	serverID            string
	getServerIDErr      error
}

func (m *MockSecurityDB) ChangeUserID(colonyName string, oldUserID string, newUserID string) error {
	return m.changeUserIDErr
}

func (m *MockSecurityDB) ChangeExecutorID(colonyName string, oldExecutorID string, newExecutorID string) error {
	return m.changeExecutorIDErr
}

func (m *MockSecurityDB) ChangeColonyID(colonyName string, oldColonyID string, newColonyID string) error {
	return m.changeColonyIDErr
}

func (m *MockSecurityDB) SetServerID(oldServerID string, newServerID string) error {
	return m.setServerIDErr
}

func (m *MockSecurityDB) GetServerID() (string, error) {
	return m.serverID, m.getServerIDErr
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
	userDB          *MockUserDB
	executorDB      *MockExecutorDB
	colonyDB        *MockColonyDB
	securityDB      *MockSecurityDB
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

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
	m.lastPayloadType = payloadType
	c.JSON(http.StatusOK, nil)
}

func (m *MockServer) GetServerID() (string, error) {
	return m.serverID, m.serverIDErr
}

func (m *MockServer) Validator() pkgsecurity.Validator {
	return m.validator
}

func (m *MockServer) UserDB() database.UserDatabase {
	return m.userDB
}

func (m *MockServer) ExecutorDB() database.ExecutorDatabase {
	return m.executorDB
}

func (m *MockServer) ColonyDB() database.ColonyDatabase {
	return m.colonyDB
}

func (m *MockServer) SecurityDB() database.SecurityDatabase {
	return m.securityDB
}

// Helper to generate a 64-character ID
func generateValidID() string {
	return "1234567890123456789012345678901234567890123456789012345678901234"
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	user := &core.User{
		ID:         "user-123",
		Name:       "test-user",
		ColonyName: "test-colony",
	}
	executor := &core.Executor{
		ID:         "executor-123",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	colony := &core.Colony{
		ID:   "colony-123",
		Name: "test-colony",
	}

	userDB := &MockUserDB{users: []*core.User{user}}
	executorDB := &MockExecutorDB{executors: []*core.Executor{executor}}
	colonyDB := &MockColonyDB{colonies: []*core.Colony{colony}}
	securityDB := &MockSecurityDB{}
	validator := &MockValidator{}

	server := &MockServer{
		userDB:     userDB,
		executorDB: executorDB,
		colonyDB:   colonyDB,
		securityDB: securityDB,
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

// Tests for HandleChangeUserID
func TestHandleChangeUserID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newUserID := generateValidID()
	msg := rpc.CreateChangeUserIDMsg("test-colony", newUserID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, rpc.ChangeUserIDPayloadType, server.lastPayloadType)
}

func TestHandleChangeUserID_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeUserID_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newUserID := generateValidID()
	msg := rpc.CreateChangeUserIDMsg("test-colony", newUserID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeUserID_EmptyUserID(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeUserIDMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeUserID_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	newUserID := generateValidID()
	msg := rpc.CreateChangeUserIDMsg("test-colony", newUserID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChangeUserID_UserNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.userDB.returnNilByID = true
	handlers := NewHandlers(server)

	newUserID := generateValidID()
	msg := rpc.CreateChangeUserIDMsg("test-colony", newUserID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeUserID_InvalidIDLength(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeUserIDMsg("test-colony", "short-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeUserID_SecurityDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.securityDB.changeUserIDErr = errors.New("security error")
	handlers := NewHandlers(server)

	newUserID := generateValidID()
	msg := rpc.CreateChangeUserIDMsg("test-colony", newUserID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeUserID(ctx, "user-123", rpc.ChangeUserIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleChangeExecutorID
func TestHandleChangeExecutorID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newExecutorID := generateValidID()
	msg := rpc.CreateChangeExecutorIDMsg("test-colony", newExecutorID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, rpc.ChangeExecutorIDPayloadType, server.lastPayloadType)
}

func TestHandleChangeExecutorID_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeExecutorID_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newExecutorID := generateValidID()
	msg := rpc.CreateChangeExecutorIDMsg("test-colony", newExecutorID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeExecutorID_EmptyExecutorID(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeExecutorIDMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeExecutorID_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	newExecutorID := generateValidID()
	msg := rpc.CreateChangeExecutorIDMsg("test-colony", newExecutorID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChangeExecutorID_ExecutorNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.executorDB.returnNilByID = true
	handlers := NewHandlers(server)

	newExecutorID := generateValidID()
	msg := rpc.CreateChangeExecutorIDMsg("test-colony", newExecutorID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeExecutorID_InvalidIDLength(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeExecutorIDMsg("test-colony", "short-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeExecutorID_SecurityDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.securityDB.changeExecutorIDErr = errors.New("security error")
	handlers := NewHandlers(server)

	newExecutorID := generateValidID()
	msg := rpc.CreateChangeExecutorIDMsg("test-colony", newExecutorID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeExecutorID(ctx, "executor-123", rpc.ChangeExecutorIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleChangeColonyID
func TestHandleChangeColonyID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newColonyID := generateValidID()
	msg := rpc.CreateChangeColonyIDMsg("test-colony", newColonyID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, jsonString)

	assert.Equal(t, rpc.ChangeColonyIDPayloadType, server.lastPayloadType)
}

func TestHandleChangeColonyID_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeColonyID_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	newColonyID := generateValidID()
	msg := rpc.CreateChangeColonyIDMsg("test-colony", newColonyID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeColonyID_EmptyColonyID(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeColonyIDMsg("test-colony", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeColonyID_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("colony owner error")
	handlers := NewHandlers(server)

	newColonyID := generateValidID()
	msg := rpc.CreateChangeColonyIDMsg("test-colony", newColonyID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChangeColonyID_InvalidIDLength(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeColonyIDMsg("test-colony", "short-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeColonyID_SecurityDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.securityDB.changeColonyIDErr = errors.New("security error")
	handlers := NewHandlers(server)

	newColonyID := generateValidID()
	msg := rpc.CreateChangeColonyIDMsg("test-colony", newColonyID)
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeColonyID(ctx, "owner-123", rpc.ChangeColonyIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleChangeServerID
func TestHandleChangeServerID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("new-server-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, jsonString)

	assert.Equal(t, rpc.ChangeServerIDPayloadType, server.lastPayloadType)
}

func TestHandleChangeServerID_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeServerID_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("new-server-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeServerID_EmptyServerID(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleChangeServerID_GetServerIDError(t *testing.T) {
	server, ctx := createMockServer()
	server.serverIDErr = errors.New("server id error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("new-server-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

func TestHandleChangeServerID_ServerOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.serverOwnerErr = errors.New("server owner error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("new-server-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleChangeServerID_SecurityDBError(t *testing.T) {
	server, ctx := createMockServer()
	server.securityDB.setServerIDErr = errors.New("security error")
	handlers := NewHandlers(server)

	msg := rpc.CreateChangeServerIDMsg("new-server-id")
	jsonString, _ := msg.ToJSON()

	handlers.HandleChangeServerID(ctx, "server-owner-123", rpc.ChangeServerIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}
