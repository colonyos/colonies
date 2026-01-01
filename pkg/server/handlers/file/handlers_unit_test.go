package file

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

// MockFileDB implements database.FileDatabase
type MockFileDB struct {
	files            []*core.File
	fileData         []*core.FileData
	labels           []*core.Label
	addErr           error
	getByIDErr       error
	getByNameErr     error
	getLatestErr     error
	getDataErr       error
	getLabelsErr     error
	removeByIDErr    error
	removeByNameErr  error
	returnNilByID    bool
	returnEmptyByName bool
}

func (m *MockFileDB) AddFile(file *core.File) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.files = append(m.files, file)
	return nil
}

func (m *MockFileDB) GetFileByID(colonyName string, fileID string) (*core.File, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.returnNilByID {
		return nil, nil
	}
	for _, f := range m.files {
		if f.ID == fileID {
			return f, nil
		}
	}
	// Return the last added file for test convenience
	if len(m.files) > 0 {
		return m.files[len(m.files)-1], nil
	}
	return nil, nil
}

func (m *MockFileDB) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if m.returnEmptyByName {
		return []*core.File{}, nil
	}
	return m.files, nil
}

func (m *MockFileDB) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	if m.getLatestErr != nil {
		return nil, m.getLatestErr
	}
	if len(m.files) > 0 {
		return []*core.File{m.files[len(m.files)-1]}, nil
	}
	return []*core.File{}, nil
}

func (m *MockFileDB) GetFilenamesByLabel(colonyName string, label string) ([]string, error) {
	return nil, nil
}

func (m *MockFileDB) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) {
	if m.getDataErr != nil {
		return nil, m.getDataErr
	}
	return m.fileData, nil
}

func (m *MockFileDB) RemoveFileByID(colonyName string, fileID string) error {
	if m.removeByIDErr != nil {
		return m.removeByIDErr
	}
	return nil
}

func (m *MockFileDB) RemoveFileByName(colonyName string, label string, name string) error {
	if m.removeByNameErr != nil {
		return m.removeByNameErr
	}
	return nil
}

func (m *MockFileDB) GetFileLabels(colonyName string) ([]*core.Label, error) {
	if m.getLabelsErr != nil {
		return nil, m.getLabelsErr
	}
	return m.labels, nil
}

func (m *MockFileDB) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) {
	if m.getLabelsErr != nil {
		return nil, m.getLabelsErr
	}
	return m.labels, nil
}

func (m *MockFileDB) CountFilesWithLabel(colonyName string, label string) (int, error) {
	return len(m.files), nil
}

func (m *MockFileDB) CountFiles(colonyName string) (int, error) {
	return len(m.files), nil
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
	fileDB          *MockFileDB
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

func (m *MockServer) FileDB() database.FileDatabase {
	return m.fileDB
}

// Helper to create test file
func createTestFile() *core.File {
	return &core.File{
		ID:         "file-123",
		ColonyName: "test-colony",
		Label:      "/test/label",
		Name:       "test-file.txt",
		Size:       1024,
	}
}

// Helper to create mock server
func createMockServer() (*MockServer, *MockContext) {
	file := createTestFile()
	fileDB := &MockFileDB{files: []*core.File{file}}
	validator := &MockValidator{}

	server := &MockServer{
		fileDB:    fileDB,
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

// Tests for HandleAddFile
func TestHandleAddFile_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	file := createTestFile()
	msg := rpc.CreateAddFileMsg(file)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, jsonString)

	assert.Equal(t, rpc.AddFilePayloadType, server.lastPayloadType)
	assert.Nil(t, server.lastError)
}

func TestHandleAddFile_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddFile_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	file := createTestFile()
	msg := rpc.CreateAddFileMsg(file)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddFile(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddFile_NilFile(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateAddFileMsg(nil)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleAddFile_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	file := createTestFile()
	msg := rpc.CreateAddFileMsg(file)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleAddFile_GetFileError(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.getByIDErr = errors.New("db error")
	handlers := NewHandlers(server)

	file := createTestFile()
	msg := rpc.CreateAddFileMsg(file)
	jsonString, _ := msg.ToJSON()

	handlers.HandleAddFile(ctx, "test-user", rpc.AddFilePayloadType, jsonString)

	assert.Equal(t, http.StatusInternalServerError, server.lastStatusCode)
}

// Tests for HandleGetFile
func TestHandleGetFile_ByID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "file-123", "", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, jsonString)

	assert.Equal(t, rpc.GetFilePayloadType, server.lastPayloadType)
}

func TestHandleGetFile_ByName_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "", "/test/label", "test-file.txt", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, jsonString)

	assert.Equal(t, rpc.GetFilePayloadType, server.lastPayloadType)
}

func TestHandleGetFile_Latest_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "", "/test/label", "test-file.txt", true)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, jsonString)

	assert.Equal(t, rpc.GetFilePayloadType, server.lastPayloadType)
}

func TestHandleGetFile_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFile_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "file-123", "", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFile_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "file-123", "", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetFile_NotFound(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.returnNilByID = true
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileMsg("test-colony", "nonexistent", "", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFile(ctx, "test-user", rpc.GetFilePayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleGetFiles
func TestHandleGetFiles_Success(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.fileData = []*core.FileData{{Name: "test-file.txt"}}
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFilesMsg("test-colony", "/test/label")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFiles(ctx, "test-user", rpc.GetFilesPayloadType, jsonString)

	assert.Equal(t, rpc.GetFilesPayloadType, server.lastPayloadType)
}

func TestHandleGetFiles_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetFiles(ctx, "test-user", rpc.GetFilesPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFiles_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFilesMsg("test-colony", "/test/label")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFiles(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFiles_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFilesMsg("test-colony", "/test/label")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFiles(ctx, "test-user", rpc.GetFilesPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleGetFiles_DBError(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.getDataErr = errors.New("db error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFilesMsg("test-colony", "/test/label")
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFiles(ctx, "test-user", rpc.GetFilesPayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

// Tests for HandleGetFileLabels
func TestHandleGetFileLabels_Success(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.labels = []*core.Label{{Name: "/test/label"}}
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileLabelsMsg("test-colony", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFileLabels(ctx, "test-user", rpc.GetFileLabelsPayloadType, jsonString)

	assert.Equal(t, rpc.GetFileLabelsPayloadType, server.lastPayloadType)
}

func TestHandleGetFileLabels_ByName_Success(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.labels = []*core.Label{{Name: "/test/label"}}
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileLabelsMsg("test-colony", "test", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFileLabels(ctx, "test-user", rpc.GetFileLabelsPayloadType, jsonString)

	assert.Equal(t, rpc.GetFileLabelsPayloadType, server.lastPayloadType)
}

func TestHandleGetFileLabels_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleGetFileLabels(ctx, "test-user", rpc.GetFileLabelsPayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFileLabels_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileLabelsMsg("test-colony", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFileLabels(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleGetFileLabels_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateGetFileLabelsMsg("test-colony", "", false)
	jsonString, _ := msg.ToJSON()

	handlers.HandleGetFileLabels(ctx, "test-user", rpc.GetFileLabelsPayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

// Tests for HandleRemoveFile
func TestHandleRemoveFile_ByID_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveFileMsg("test-colony", "file-123", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveFile_ByName_Success(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveFileMsg("test-colony", "", "/test/label", "test-file.txt")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, jsonString)

	assert.True(t, server.emptyReplySent)
}

func TestHandleRemoveFile_InvalidJSON(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, "invalid json")

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveFile_MsgTypeMismatch(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveFileMsg("test-colony", "file-123", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", "wrong-type", jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveFile_MembershipError(t *testing.T) {
	server, ctx := createMockServer()
	server.validator.membershipErr = errors.New("membership error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveFileMsg("test-colony", "file-123", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, jsonString)

	assert.Equal(t, http.StatusForbidden, server.lastStatusCode)
}

func TestHandleRemoveFile_MalformedMsg(t *testing.T) {
	server, ctx := createMockServer()
	handlers := NewHandlers(server)

	// Neither fileID nor label+name provided
	msg := rpc.CreateRemoveFileMsg("test-colony", "", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}

func TestHandleRemoveFile_RemoveByIDError(t *testing.T) {
	server, ctx := createMockServer()
	server.fileDB.removeByIDErr = errors.New("remove error")
	handlers := NewHandlers(server)

	msg := rpc.CreateRemoveFileMsg("test-colony", "file-123", "", "")
	jsonString, _ := msg.ToJSON()

	handlers.HandleRemoveFile(ctx, "test-user", rpc.RemoveFilePayloadType, jsonString)

	assert.Equal(t, http.StatusBadRequest, server.lastStatusCode)
}
