package location

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

// MockLocationDB implements database.LocationDatabase
type MockLocationDB struct {
	locations       []*core.Location
	addErr          error
	getByColonyErr  error
	getByIDErr      error
	getByNameErr    error
	removeByIDErr   error
	removeByNameErr error
	returnNilByName bool
}

func (m *MockLocationDB) AddLocation(location *core.Location) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.locations = append(m.locations, location)
	return nil
}

func (m *MockLocationDB) GetLocationsByColonyName(colonyName string) ([]*core.Location, error) {
	if m.getByColonyErr != nil {
		return nil, m.getByColonyErr
	}
	return m.locations, nil
}

func (m *MockLocationDB) GetLocationByID(locationID string) (*core.Location, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, l := range m.locations {
		if l.ID == locationID {
			return l, nil
		}
	}
	return nil, nil
}

func (m *MockLocationDB) GetLocationByName(colonyName string, name string) (*core.Location, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, l := range m.locations {
		if l.Name == name && l.ColonyName == colonyName {
			return l, nil
		}
	}
	return nil, nil
}

func (m *MockLocationDB) RemoveLocationByID(locationID string) error {
	if m.removeByIDErr != nil {
		return m.removeByIDErr
	}
	return nil
}

func (m *MockLocationDB) RemoveLocationByName(colonyName string, name string) error {
	if m.removeByNameErr != nil {
		return m.removeByNameErr
	}
	return nil
}

func (m *MockLocationDB) RemoveLocationsByColonyName(colonyName string) error {
	return nil
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
	locationDB      *MockLocationDB
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

func (m *MockServer) GetLocationDB() database.LocationDatabase {
	return m.locationDB
}

func (m *MockServer) GetColonyDB() database.ColonyDatabase {
	return m.colonyDB
}

// Helper to create test location
func createTestLocation() *core.Location {
	return &core.Location{
		ID:         "location-123",
		ColonyName: "test-colony",
		Name:       "test-location",
	}
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	location := createTestLocation()
	locationDB := &MockLocationDB{locations: []*core.Location{location}}
	colonyDB := &MockColonyDB{}
	validator := &MockValidator{}

	server := &MockServer{
		locationDB: locationDB,
		colonyDB:   colonyDB,
		validator:  validator,
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

// Tests for HandleAddLocation
func TestHandleAddLocation_Success(t *testing.T) {
	// Create a fresh server without any existing locations
	locationDB := &MockLocationDB{locations: []*core.Location{}}
	colonyDB := &MockColonyDB{}
	validator := &MockValidator{}

	server := &MockServer{
		locationDB: locationDB,
		colonyDB:   colonyDB,
		validator:  validator,
	}
	ctx := &MockContext{}
	handlers := NewHandlers(server)

	location := &core.Location{
		ColonyName: "test-colony",
		Name:       "new-location",
	}
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, jsonString)

	assert.Equal(t, rpc.AddLocationPayloadType, server.lastPayloadType)
}

func TestHandleAddLocation_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddLocation_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	location := createTestLocation()
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddLocation_NilLocation(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateAddLocationMsg(nil)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddLocation_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	location := createTestLocation()
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddLocation_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("not owner")
	handlers := NewHandlers(server)

	location := createTestLocation()
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleAddLocation_AlreadyExists(t *testing.T) {
	server, ctx := createMockServer()
	// Location already exists (returnNilByName = false by default)
	handlers := NewHandlers(server)

	location := createTestLocation()
	msg := rpc.CreateAddLocationMsg(location)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddLocation(ctx, "test-user", rpc.AddLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleGetLocations
func TestHandleGetLocations_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocations(ctx, "test-user", rpc.GetLocationsPayloadType, jsonString)

	assert.Equal(t, rpc.GetLocationsPayloadType, server.lastPayloadType)
}

func TestHandleGetLocations_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetLocations(ctx, "test-user", rpc.GetLocationsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocations_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocations(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocations_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocations(ctx, "test-user", rpc.GetLocationsPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocations_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("not member")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocations(ctx, "test-user", rpc.GetLocationsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleGetLocation
func TestHandleGetLocation_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocation(ctx, "test-user", rpc.GetLocationPayloadType, jsonString)

	assert.Equal(t, rpc.GetLocationPayloadType, server.lastPayloadType)
}

func TestHandleGetLocation_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetLocation(ctx, "test-user", rpc.GetLocationPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocation_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocation(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocation_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocation(ctx, "test-user", rpc.GetLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetLocation_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("not member")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocation(ctx, "test-user", rpc.GetLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetLocation_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.locationDB.returnNilByName = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetLocationMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetLocation(ctx, "test-user", rpc.GetLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusNotFound, server.lastStatusCode)
}

// Tests for HandleRemoveLocation
func TestHandleRemoveLocation_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveLocation(ctx, "test-user", rpc.RemoveLocationPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveLocation_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveLocation(ctx, "test-user", rpc.RemoveLocationPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveLocation_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveLocation(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveLocation_ColonyNotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.colonyDB.returnNil = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveLocation(ctx, "test-user", rpc.RemoveLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveLocation_ColonyOwnerError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.colonyOwnerErr = errors.New("not owner")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveLocationMsg("test-colony", "test-location")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveLocation(ctx, "test-user", rpc.RemoveLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveLocation_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.locationDB.returnNilByName = true
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveLocationMsg("test-colony", "nonexistent")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveLocation(ctx, "test-user", rpc.RemoveLocationPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}
