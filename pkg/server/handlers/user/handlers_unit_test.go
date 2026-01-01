package user

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

// MockUserDB implements database.UserDatabase
type MockUserDB struct {
	users           []*core.User
	addErr          error
	getByColonyErr  error
	getByNameErr    error
	getByIDErr      error
	removeByNameErr error
	returnNilByName bool
	returnNilByID   bool
}

func (m *MockUserDB) AddUser(user *core.User) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.users = append(m.users, user)
	return nil
}

func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	if m.getByColonyErr != nil {
		return nil, m.getByColonyErr
	}
	return m.users, nil
}

func (m *MockUserDB) GetUserByName(colonyName string, name string) (*core.User, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, u := range m.users {
		if u.Name == name && u.ColonyName == colonyName {
			return u, nil
		}
	}
	return nil, nil
}

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

func (m *MockUserDB) RemoveUserByID(colonyName string, userID string) error {
	return nil
}

func (m *MockUserDB) RemoveUserByName(colonyName string, name string) error {
	if m.removeByNameErr != nil {
		return m.removeByNameErr
	}
	return nil
}

func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error {
	return nil
}

func (m *MockUserDB) CountUsers() (int, error) {
	return len(m.users), nil
}

func (m *MockUserDB) CountUsersByColonyName(colonyName string) (int, error) {
	return len(m.users), nil
}

// MockColonyDB implements database.ColonyDatabase
type MockColonyDB struct {
	colonies      []*core.Colony
	getByNameErr  error
	returnNil     bool
}

func (m *MockColonyDB) AddColony(colony *core.Colony) error                     { return nil }
func (m *MockColonyDB) GetColonies() ([]*core.Colony, error)                    { return nil, nil }
func (m *MockColonyDB) GetColonyByID(colonyID string) (*core.Colony, error)     { return nil, nil }
func (m *MockColonyDB) RemoveColonyByName(colonyName string) error              { return nil }
func (m *MockColonyDB) RemoveColonies() error                                   { return nil }
func (m *MockColonyDB) CountColonies() (int, error)                             { return 0, nil }
func (m *MockColonyDB) RenameColony(colonyName, newName string) error           { return nil }

func (m *MockColonyDB) GetColonyByName(colonyName string) (*core.Colony, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNil {
		return nil, nil
	}
	for _, c := range m.colonies {
		if c.Name == colonyName {
			return c, nil
		}
	}
	return &core.Colony{ID: "colony-123", Name: colonyName}, nil
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
	colonyDB        *MockColonyDB
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

func (m *MockServer) GetValidator() security.Validator {
	return m.validator
}

func (m *MockServer) GetUserDB() database.UserDatabase {
	return m.userDB
}

func (m *MockServer) GetColonyDB() database.ColonyDatabase {
	return m.colonyDB
}

// Helper to create test user
func createTestUser() *core.User {
	return &core.User{
		ID:         "user-123",
		ColonyName: "test-colony",
		Name:       "test-user",
	}
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	user := createTestUser()
	userDB := &MockUserDB{users: []*core.User{user}}
	colonyDB := &MockColonyDB{}
	validator := &MockValidator{}

	server := &MockServer{
		userDB:    userDB,
		colonyDB:  colonyDB,
		validator: validator,
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

// Tests for HandleAddUser
func TestHandleAddUser_Success(t *testing.T) {
	// Create a fresh server without any existing users
	userDB := &MockUserDB{users: []*core.User{}}
	colonyDB := &MockColonyDB{}
	validator := &MockValidator{}

	server := &MockServer{
		userDB:    userDB,
		colonyDB:  colonyDB,
		validator: validator,
	}
	ctx := &MockContext{}
	handlers := NewHandlers(server)

	user := &core.User{
		ID:         "new-user-id",
		ColonyName: "test-colony",
		Name:       "new-user",
	}
	msg := rpc.CreateAddUserMsg(user)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-owner", rpc.AddUserPayloadType, jsonString)

	assert.Equal(t, rpc.AddUserPayloadType, server.lastPayloadType)
}

func TestHandleAddUser_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleAddUser(ctx, "test-user", rpc.AddUserPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddUser_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	user := createTestUser()
	msg := rpc.CreateAddUserMsg(user)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddUser_NilUser(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateAddUserMsg(nil)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-user", rpc.AddUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddUser_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	user := createTestUser()
	msg := rpc.CreateAddUserMsg(user)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-user", rpc.AddUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddUser_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("not owner")
	handlers := NewHandlers(server)

	user := createTestUser()
	msg := rpc.CreateAddUserMsg(user)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-user", rpc.AddUserPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleAddUser_NameAlreadyExists(t *testing.T) {
	server, ctx := createMockServer()
	// User with same name already exists
	server.userDB.returnNilByID = true // But ID doesn't exist
	handlers := NewHandlers(server)

	user := createTestUser()
	msg := rpc.CreateAddUserMsg(user)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddUser(ctx, "test-user", rpc.AddUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleGetUsers
func TestHandleGetUsers_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUsersMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUsers(ctx, "test-user", rpc.GetUsersPayloadType, jsonString)

	assert.Equal(t, rpc.GetUsersPayloadType, server.lastPayloadType)
}

func TestHandleGetUsers_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetUsers(ctx, "test-user", rpc.GetUsersPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUsers_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUsersMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUsers(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUsers_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUsersMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUsers(ctx, "test-user", rpc.GetUsersPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUsers_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("not member")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUsersMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUsers(ctx, "test-user", rpc.GetUsersPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleGetUser
func TestHandleGetUser_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUser(ctx, "test-user", rpc.GetUserPayloadType, jsonString)

	assert.Equal(t, rpc.GetUserPayloadType, server.lastPayloadType)
}

func TestHandleGetUser_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetUser(ctx, "test-user", rpc.GetUserPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUser_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUser(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUser_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUser(ctx, "test-user", rpc.GetUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUser_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("not member")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUser(ctx, "test-user", rpc.GetUserPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetUser_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.userDB.returnNilByName = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUser(ctx, "test-user", rpc.GetUserPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetUserByID
func TestHandleGetUserByID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserByIDMsg("test-colony", "user-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUserByID(ctx, "test-user", rpc.GetUserByIDPayloadType, jsonString)

	assert.Equal(t, rpc.GetUserByIDPayloadType, server.lastPayloadType)
}

func TestHandleGetUserByID_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetUserByID(ctx, "test-user", rpc.GetUserByIDPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUserByID_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserByIDMsg("test-colony", "user-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUserByID(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUserByID_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserByIDMsg("test-colony", "user-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUserByID(ctx, "test-user", rpc.GetUserByIDPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetUserByID_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("not member")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserByIDMsg("test-colony", "user-123")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUserByID(ctx, "test-user", rpc.GetUserByIDPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetUserByID_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.userDB.returnNilByID = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetUserByIDMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetUserByID(ctx, "test-user", rpc.GetUserByIDPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

// Tests for HandleRemoveUser
func TestHandleRemoveUser_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveUser(ctx, "test-owner", rpc.RemoveUserPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveUser_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveUser(ctx, "test-user", rpc.RemoveUserPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveUser_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveUser(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveUser_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveUser(ctx, "test-user", rpc.RemoveUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveUser_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("not owner")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveUserMsg("test-colony", "test-user")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveUser(ctx, "test-user", rpc.RemoveUserPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveUser_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.userDB.returnNilByName = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveUserMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveUser(ctx, "test-user", rpc.RemoveUserPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}
