package attribute

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

// MockAttributeDB implements database.AttributeDatabase
type MockAttributeDB struct {
	attributes      []core.Attribute
	addErr          error
	getByIDErr      error
	returnEmptyByID bool
}

func (m *MockAttributeDB) AddAttribute(attribute core.Attribute) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.attributes = append(m.attributes, attribute)
	return nil
}

func (m *MockAttributeDB) AddAttributes(attributes []core.Attribute) error {
	return nil
}

func (m *MockAttributeDB) GetAttributeByID(attributeID string) (core.Attribute, error) {
	if m.getByIDErr != nil {
		return core.Attribute{}, m.getByIDErr
	}
	if m.returnEmptyByID {
		return core.Attribute{}, nil
	}
	for _, a := range m.attributes {
		if a.ID == attributeID {
			return a, nil
		}
	}
	// Return the last added attribute (for AddAttribute tests)
	if len(m.attributes) > 0 {
		return m.attributes[len(m.attributes)-1], nil
	}
	return core.Attribute{}, nil
}

func (m *MockAttributeDB) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	return nil, nil
}
func (m *MockAttributeDB) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	return core.Attribute{}, nil
}
func (m *MockAttributeDB) GetAttributes(targetID string) ([]core.Attribute, error) {
	return nil, nil
}
func (m *MockAttributeDB) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	return nil, nil
}
func (m *MockAttributeDB) UpdateAttribute(attribute core.Attribute) error { return nil }
func (m *MockAttributeDB) RemoveAttributeByID(attributeID string) error   { return nil }
func (m *MockAttributeDB) RemoveAllAttributesByColonyName(colonyName string) error {
	return nil
}
func (m *MockAttributeDB) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	return nil
}
func (m *MockAttributeDB) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	return nil
}
func (m *MockAttributeDB) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	return nil
}
func (m *MockAttributeDB) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	return nil
}
func (m *MockAttributeDB) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	return nil
}
func (m *MockAttributeDB) RemoveAllAttributesByTargetID(targetID string) error { return nil }
func (m *MockAttributeDB) RemoveAllAttributes() error                          { return nil }

// MockProcessDB implements database.ProcessDatabase (minimal for attribute tests)
type MockProcessDB struct {
	processes     []*core.Process
	getByIDErr    error
	returnNilByID bool
}

func (m *MockProcessDB) AddProcess(process *core.Process) error { return nil }
func (m *MockProcessDB) GetProcesses() ([]*core.Process, error) { return nil, nil }
func (m *MockProcessDB) GetProcessByID(processID string) (*core.Process, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, p := range m.processes {
		if p.ID == processID {
			return p, nil
		}
	}
	return nil, nil
}
func (m *MockProcessDB) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindAllRunningProcesses() ([]*core.Process, error)  { return nil, nil }
func (m *MockProcessDB) FindAllWaitingProcesses() ([]*core.Process, error)  { return nil, nil }
func (m *MockProcessDB) RemoveProcessByID(processID string) error           { return nil }
func (m *MockProcessDB) RemoveAllProcesses() error                          { return nil }
func (m *MockProcessDB) RemoveAllWaitingProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllRunningProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllSuccessfulProcessesByColonyName(string) error {
	return nil
}
func (m *MockProcessDB) RemoveAllFailedProcessesByColonyName(string) error { return nil }
func (m *MockProcessDB) RemoveAllProcessesByColonyName(string) error       { return nil }
func (m *MockProcessDB) RemoveAllProcessesByProcessGraphID(string) error   { return nil }
func (m *MockProcessDB) RemoveAllProcessesInProcessGraphsByColonyName(string) error {
	return nil
}
func (m *MockProcessDB) ResetProcess(process *core.Process) error { return nil }
func (m *MockProcessDB) SetInput(processID string, output []interface{}) error {
	return nil
}
func (m *MockProcessDB) SetOutput(processID string, output []interface{}) error {
	return nil
}
func (m *MockProcessDB) SetErrors(processID string, errs []string) error { return nil }
func (m *MockProcessDB) SetProcessState(processID string, state int) error {
	return nil
}
func (m *MockProcessDB) SetParents(processID string, parents []string) error {
	return nil
}
func (m *MockProcessDB) SetChildren(processID string, children []string) error {
	return nil
}
func (m *MockProcessDB) SetWaitForParents(processID string, waitingForParent bool) error {
	return nil
}
func (m *MockProcessDB) Assign(executorID string, process *core.Process) error {
	return nil
}
func (m *MockProcessDB) SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) Unassign(process *core.Process) error { return nil }
func (m *MockProcessDB) MarkSuccessful(processID string) (float64, float64, error) {
	return 0, 0, nil
}
func (m *MockProcessDB) MarkFailed(processID string, errs []string) error { return nil }
func (m *MockProcessDB) CountProcesses() (int, error)                     { return 0, nil }
func (m *MockProcessDB) CountWaitingProcesses() (int, error)              { return 0, nil }
func (m *MockProcessDB) CountRunningProcesses() (int, error)              { return 0, nil }
func (m *MockProcessDB) CountSuccessfulProcesses() (int, error)           { return 0, nil }
func (m *MockProcessDB) CountFailedProcesses() (int, error)               { return 0, nil }
func (m *MockProcessDB) CountWaitingProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountRunningProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountSuccessfulProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) CountFailedProcessesByColonyName(string) (int, error) {
	return 0, nil
}
func (m *MockProcessDB) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}
func (m *MockProcessDB) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

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

// MockServer implements the Server interface for attribute handlers
type MockServer struct {
	attributeDB   *MockAttributeDB
	processDB     *MockProcessDB
	validator     *MockValidator
	httpErrorCode int
	replyPayload  string
	replyType     string
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

func (m *MockServer) Validator() security.Validator {
	return m.validator
}

func (m *MockServer) ProcessDB() database.ProcessDatabase {
	return m.processDB
}

func (m *MockServer) AttributeDB() database.AttributeDatabase {
	return m.attributeDB
}

// Test setup helpers
func createTestProcess() *core.Process {
	return &core.Process{
		ID:                 "process-123",
		State:              core.RUNNING,
		AssignedExecutorID: "executor-123",
		ProcessGraphID:     "graph-123",
		FunctionSpec: core.FunctionSpec{
			Conditions: core.Conditions{
				ColonyName: "test-colony",
			},
		},
	}
}

func createTestAttribute() core.Attribute {
	return core.Attribute{
		ID:            "attr-123",
		Key:           "test-key",
		Value:         "test-value",
		TargetID:      "process-123",
		AttributeType: core.OUT,
	}
}

func createMockServer() *MockServer {
	process := createTestProcess()

	attributeDB := &MockAttributeDB{}
	processDB := &MockProcessDB{processes: []*core.Process{process}}
	validator := &MockValidator{}

	return &MockServer{
		attributeDB: attributeDB,
		processDB:   processDB,
		validator:   validator,
	}
}

// HandleAddAttribute tests
func TestHandleAddAttribute_Success(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.AddAttributePayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleAddAttribute_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_ProcessDBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_ProcessNotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusNotFound, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_AuthError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_ProcessNotRunning(t *testing.T) {
	mockServer := createMockServer()
	mockServer.processDB.processes[0].State = core.WAITING
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_WrongExecutor(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "wrong-executor", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_AddDBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.attributeDB.addErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleAddAttribute_GetAddedAttributeError(t *testing.T) {
	mockServer := createMockServer()
	// First add will succeed, but get by ID will fail
	mockServer.attributeDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	attr := createTestAttribute()
	msg := rpc.CreateAddAttributeMsg(attr)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleAddAttribute(ctx, "executor-123", rpc.AddAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusInternalServerError, mockServer.httpErrorCode)
}

// HandleGetAttribute tests
func TestHandleGetAttribute_Success(t *testing.T) {
	mockServer := createMockServer()
	attr := createTestAttribute()
	mockServer.attributeDB.attributes = []core.Attribute{attr}
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg(attr.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, 0, mockServer.httpErrorCode)
	assert.Equal(t, rpc.GetAttributePayloadType, mockServer.replyType)
	assert.NotEmpty(t, mockServer.replyPayload)
}

func TestHandleGetAttribute_InvalidJSON(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_MsgTypeMismatch(t *testing.T) {
	mockServer := createMockServer()
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg("attr-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", "wrong-type", jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_DBError(t *testing.T) {
	mockServer := createMockServer()
	mockServer.attributeDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg("attr-123")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_NotFound(t *testing.T) {
	mockServer := createMockServer()
	mockServer.attributeDB.returnEmptyByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg("non-existent")
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusNotFound, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_ProcessDBError(t *testing.T) {
	mockServer := createMockServer()
	attr := createTestAttribute()
	mockServer.attributeDB.attributes = []core.Attribute{attr}
	mockServer.processDB.getByIDErr = errors.New("database error")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg(attr.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusBadRequest, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_ProcessNotFound(t *testing.T) {
	mockServer := createMockServer()
	attr := createTestAttribute()
	mockServer.attributeDB.attributes = []core.Attribute{attr}
	mockServer.processDB.returnNilByID = true
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg(attr.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusNotFound, mockServer.httpErrorCode)
}

func TestHandleGetAttribute_AuthError(t *testing.T) {
	mockServer := createMockServer()
	attr := createTestAttribute()
	mockServer.attributeDB.attributes = []core.Attribute{attr}
	mockServer.validator.requireMembershipErr = errors.New("not a member")
	handlers := NewHandlers(mockServer)

	msg := rpc.CreateGetAttributeMsg(attr.ID)
	jsonStr, _ := msg.ToJSON()

	ctx := &MockContext{}
	handlers.HandleGetAttribute(ctx, "executor-123", rpc.GetAttributePayloadType, jsonStr)

	assert.Equal(t, http.StatusForbidden, mockServer.httpErrorCode)
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
