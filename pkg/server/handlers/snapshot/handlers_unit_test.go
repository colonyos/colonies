package snapshot

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

// MockSnapshotDB implements database.SnapshotDatabase
type MockSnapshotDB struct {
	snapshots       []*core.Snapshot
	createErr       error
	getByIDErr      error
	getByNameErr    error
	getByColonyErr  error
	removeByIDErr   error
	removeByNameErr error
	removeAllErr    error
	returnNilByID   bool
	returnNilByName bool
}

func (m *MockSnapshotDB) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	snapshot := &core.Snapshot{
		ID:         "snapshot-123",
		ColonyName: colonyName,
		Label:      label,
		Name:       name,
	}
	m.snapshots = append(m.snapshots, snapshot)
	return snapshot, nil
}

func (m *MockSnapshotDB) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, s := range m.snapshots {
		if s.ID == snapshotID {
			return s, nil
		}
	}
	if len(m.snapshots) > 0 {
		return m.snapshots[0], nil
	}
	return nil, nil
}

func (m *MockSnapshotDB) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnNilByName {
		return nil, nil
	}
	for _, s := range m.snapshots {
		if s.Name == name {
			return s, nil
		}
	}
	if len(m.snapshots) > 0 {
		return m.snapshots[0], nil
	}
	return nil, nil
}

func (m *MockSnapshotDB) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) {
	if m.getByColonyErr != nil {
		return nil, m.getByColonyErr
	}
	return m.snapshots, nil
}

func (m *MockSnapshotDB) RemoveSnapshotByID(colonyName string, snapshotID string) error {
	if m.removeByIDErr != nil {
		return m.removeByIDErr
	}
	return nil
}

func (m *MockSnapshotDB) RemoveSnapshotByName(colonyName string, name string) error {
	if m.removeByNameErr != nil {
		return m.removeByNameErr
	}
	return nil
}

func (m *MockSnapshotDB) RemoveSnapshotsByColonyName(colonyName string) error {
	if m.removeAllErr != nil {
		return m.removeAllErr
	}
	return nil
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
	snapshotDB      *MockSnapshotDB
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

func (m *MockServer) SnapshotDB() database.SnapshotDatabase {
	return m.snapshotDB
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	snapshot := &core.Snapshot{
		ID:         "snapshot-123",
		ColonyName: "test-colony",
		Label:      "/test/label",
		Name:       "test-snapshot",
	}
	snapshotDB := &MockSnapshotDB{snapshots: []*core.Snapshot{snapshot}}
	validator := &MockValidator{}

	server := &MockServer{
		snapshotDB: snapshotDB,
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

// Tests for HandleCreateSnapshot
func TestHandleCreateSnapshot_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateCreateSnapshotMsg("test-colony", "/test/label", "new-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleCreateSnapshot(ctx, "test-user", rpc.CreateSnapshotPayloadType, jsonString)

	assert.Equal(t, rpc.CreateSnapshotPayloadType, server.lastPayloadType)
	assert.Nil(t, server.lastError)
}

func TestHandleCreateSnapshot_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleCreateSnapshot(ctx, "test-user", rpc.CreateSnapshotPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleCreateSnapshot_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateCreateSnapshotMsg("test-colony", "/test/label", "new-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleCreateSnapshot(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleCreateSnapshot_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateCreateSnapshotMsg("test-colony", "/test/label", "new-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleCreateSnapshot(ctx, "test-user", rpc.CreateSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleCreateSnapshot_DBError(t *testing.T) {
	server, ctx := createMockServer()
	server.snapshotDB.createErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateCreateSnapshotMsg("test-colony", "/test/label", "new-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleCreateSnapshot(ctx, "test-user", rpc.CreateSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetSnapshot
func TestHandleGetSnapshot_ByID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshot(ctx, "test-user", rpc.GetSnapshotPayloadType, jsonString)

	assert.Equal(t, rpc.GetSnapshotPayloadType, server.lastPayloadType)
}

func TestHandleGetSnapshot_ByName_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotMsg("test-colony", "", "test-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshot(ctx, "test-user", rpc.GetSnapshotPayloadType, jsonString)

	assert.Equal(t, rpc.GetSnapshotPayloadType, server.lastPayloadType)
}

func TestHandleGetSnapshot_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetSnapshot(ctx, "test-user", rpc.GetSnapshotPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetSnapshot_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshot(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetSnapshot_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshot(ctx, "test-user", rpc.GetSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetSnapshot_MalformedMsg(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	// Neither snapshotID nor name provided
	msg := rpc.CreateGetSnapshotMsg("test-colony", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshot(ctx, "test-user", rpc.GetSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetSnapshots
func TestHandleGetSnapshots_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshots(ctx, "test-user", rpc.GetSnapshotsPayloadType, jsonString)

	assert.Equal(t, rpc.GetSnapshotsPayloadType, server.lastPayloadType)
}

func TestHandleGetSnapshots_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetSnapshots(ctx, "test-user", rpc.GetSnapshotsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetSnapshots_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshots(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetSnapshots_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshots(ctx, "test-user", rpc.GetSnapshotsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetSnapshots_DBError(t *testing.T) {
	server, ctx := createMockServer()
	server.snapshotDB.getByColonyErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetSnapshots(ctx, "test-user", rpc.GetSnapshotsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleRemoveSnapshot
func TestHandleRemoveSnapshot_ByID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveSnapshot(ctx, "test-user", rpc.RemoveSnapshotPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveSnapshot_ByName_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveSnapshotMsg("test-colony", "", "test-snapshot")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveSnapshot(ctx, "test-user", rpc.RemoveSnapshotPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveSnapshot_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveSnapshot(ctx, "test-user", rpc.RemoveSnapshotPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveSnapshot_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveSnapshot(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveSnapshot_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveSnapshotMsg("test-colony", "snapshot-123", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveSnapshot(ctx, "test-user", rpc.RemoveSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveSnapshot_MalformedMsg(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	// Neither snapshotID nor name provided
	msg := rpc.CreateRemoveSnapshotMsg("test-colony", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveSnapshot(ctx, "test-user", rpc.RemoveSnapshotPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleRemoveAllSnapshots
func TestHandleRemoveAllSnapshots_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllSnapshots(ctx, "test-user", rpc.RemoveAllSnapshotsPayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveAllSnapshots_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveAllSnapshots(ctx, "test-user", rpc.RemoveAllSnapshotsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveAllSnapshots_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllSnapshots(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveAllSnapshots_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllSnapshots(ctx, "test-user", rpc.RemoveAllSnapshotsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveAllSnapshots_DBError(t *testing.T) {
	server, ctx := createMockServer()
	server.snapshotDB.removeAllErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveAllSnapshotsMsg("test-colony")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveAllSnapshots(ctx, "test-user", rpc.RemoveAllSnapshotsPayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}
