package blueprint

import (
	"errors"
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/handlers/process"
	"github.com/stretchr/testify/assert"
)

// MockBlueprintDB is a mock implementation of BlueprintDatabase
type MockBlueprintDB struct {
	blueprintDefinitions      []*core.BlueprintDefinition
	blueprints                []*core.Blueprint
	getBlueprintsByKindErr    error
	getBlueprintDefByNameErr  error
	getBlueprintDefByKindErr  error
	getDefsByNamespaceErr     error
}

func (m *MockBlueprintDB) AddBlueprintDefinition(sd *core.BlueprintDefinition) error {
	m.blueprintDefinitions = append(m.blueprintDefinitions, sd)
	return nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitionByID(id string) (*core.BlueprintDefinition, error) {
	for _, sd := range m.blueprintDefinitions {
		if sd.ID == id {
			return sd, nil
		}
	}
	return nil, nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitionByName(namespace, name string) (*core.BlueprintDefinition, error) {
	if m.getBlueprintDefByNameErr != nil {
		return nil, m.getBlueprintDefByNameErr
	}
	for _, sd := range m.blueprintDefinitions {
		if sd.Metadata.ColonyName == namespace && sd.Metadata.Name == name {
			return sd, nil
		}
	}
	return nil, nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitions() ([]*core.BlueprintDefinition, error) {
	return m.blueprintDefinitions, nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitionsByNamespace(namespace string) ([]*core.BlueprintDefinition, error) {
	if m.getDefsByNamespaceErr != nil {
		return nil, m.getDefsByNamespaceErr
	}
	var result []*core.BlueprintDefinition
	for _, sd := range m.blueprintDefinitions {
		if sd.Metadata.ColonyName == namespace {
			result = append(result, sd)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitionsByGroup(group string) ([]*core.BlueprintDefinition, error) {
	var result []*core.BlueprintDefinition
	for _, sd := range m.blueprintDefinitions {
		if sd.Spec.Group == group {
			result = append(result, sd)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) GetBlueprintDefinitionByKind(kind string) (*core.BlueprintDefinition, error) {
	if m.getBlueprintDefByKindErr != nil {
		return nil, m.getBlueprintDefByKindErr
	}
	for _, sd := range m.blueprintDefinitions {
		if sd.Spec.Names.Kind == kind {
			return sd, nil
		}
	}
	return nil, nil
}

func (m *MockBlueprintDB) UpdateBlueprintDefinition(sd *core.BlueprintDefinition) error {
	return nil
}

func (m *MockBlueprintDB) RemoveBlueprintDefinitionByID(id string) error {
	return nil
}

func (m *MockBlueprintDB) RemoveBlueprintDefinitionByName(namespace, name string) error {
	return nil
}

func (m *MockBlueprintDB) CountBlueprintDefinitions() (int, error) {
	return len(m.blueprintDefinitions), nil
}

func (m *MockBlueprintDB) AddBlueprint(blueprint *core.Blueprint) error {
	m.blueprints = append(m.blueprints, blueprint)
	return nil
}

func (m *MockBlueprintDB) GetBlueprintByID(id string) (*core.Blueprint, error) {
	for _, bp := range m.blueprints {
		if bp.ID == id {
			return bp, nil
		}
	}
	return nil, nil
}

func (m *MockBlueprintDB) GetBlueprintByName(namespace, name string) (*core.Blueprint, error) {
	for _, bp := range m.blueprints {
		if bp.Metadata.ColonyName == namespace && bp.Metadata.Name == name {
			return bp, nil
		}
	}
	return nil, nil
}

func (m *MockBlueprintDB) GetBlueprints() ([]*core.Blueprint, error) {
	return m.blueprints, nil
}

func (m *MockBlueprintDB) GetBlueprintsByNamespace(namespace string) ([]*core.Blueprint, error) {
	var result []*core.Blueprint
	for _, bp := range m.blueprints {
		if bp.Metadata.ColonyName == namespace {
			result = append(result, bp)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) GetBlueprintsByKind(kind string) ([]*core.Blueprint, error) {
	var result []*core.Blueprint
	for _, bp := range m.blueprints {
		if bp.Kind == kind {
			result = append(result, bp)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) GetBlueprintsByNamespaceAndKind(namespace, kind string) ([]*core.Blueprint, error) {
	if m.getBlueprintsByKindErr != nil {
		return nil, m.getBlueprintsByKindErr
	}
	var result []*core.Blueprint
	for _, bp := range m.blueprints {
		if bp.Metadata.ColonyName == namespace && bp.Kind == kind {
			result = append(result, bp)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) GetBlueprintsByNamespaceKindAndLocation(namespace, kind, locationName string) ([]*core.Blueprint, error) {
	var result []*core.Blueprint
	for _, bp := range m.blueprints {
		if bp.Metadata.ColonyName == namespace && bp.Kind == kind && bp.Metadata.LocationName == locationName {
			result = append(result, bp)
		}
	}
	return result, nil
}

func (m *MockBlueprintDB) UpdateBlueprint(blueprint *core.Blueprint) error {
	return nil
}

func (m *MockBlueprintDB) UpdateBlueprintStatus(id string, status map[string]interface{}) error {
	return nil
}

func (m *MockBlueprintDB) RemoveBlueprintByID(id string) error {
	return nil
}

func (m *MockBlueprintDB) RemoveBlueprintByName(namespace, name string) error {
	return nil
}

func (m *MockBlueprintDB) RemoveBlueprintsByNamespace(namespace string) error {
	return nil
}

func (m *MockBlueprintDB) CountBlueprints() (int, error) {
	return len(m.blueprints), nil
}

func (m *MockBlueprintDB) CountBlueprintsByNamespace(namespace string) (int, error) {
	count := 0
	for _, bp := range m.blueprints {
		if bp.Metadata.ColonyName == namespace {
			count++
		}
	}
	return count, nil
}

func (m *MockBlueprintDB) AddBlueprintHistory(history *core.BlueprintHistory) error {
	return nil
}

func (m *MockBlueprintDB) GetBlueprintHistory(blueprintID string, limit int) ([]*core.BlueprintHistory, error) {
	return nil, nil
}

func (m *MockBlueprintDB) GetBlueprintHistoryByGeneration(blueprintID string, generation int64) (*core.BlueprintHistory, error) {
	return nil, nil
}

func (m *MockBlueprintDB) RemoveBlueprintHistory(blueprintID string) error {
	return nil
}

// MockExecutorDB is a mock implementation of ExecutorDatabase
type MockExecutorDB struct {
	executors        []*core.Executor
	getByIDErr       error
	getByIDReturnNil bool
}

func (m *MockExecutorDB) AddExecutor(executor *core.Executor) error {
	m.executors = append(m.executors, executor)
	return nil
}

func (m *MockExecutorDB) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	return nil
}

func (m *MockExecutorDB) GetExecutors() ([]*core.Executor, error) {
	return m.executors, nil
}

func (m *MockExecutorDB) GetExecutorByID(executorID string) (*core.Executor, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.getByIDReturnNil {
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
	var result []*core.Executor
	for _, e := range m.executors {
		if e.ColonyName == colonyName {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MockExecutorDB) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	for _, e := range m.executors {
		if e.ColonyName == colonyName && e.Name == executorName {
			return e, nil
		}
	}
	return nil, nil
}

func (m *MockExecutorDB) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	return nil, nil
}

func (m *MockExecutorDB) ApproveExecutor(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) RejectExecutor(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) MarkAlive(executor *core.Executor) error {
	return nil
}

func (m *MockExecutorDB) RemoveExecutorByName(colonyName string, executorName string) error {
	return nil
}

func (m *MockExecutorDB) RemoveExecutorsByColonyName(colonyName string) error {
	return nil
}

func (m *MockExecutorDB) CountExecutors() (int, error) {
	return len(m.executors), nil
}

func (m *MockExecutorDB) CountExecutorsByColonyName(colonyName string) (int, error) {
	count := 0
	for _, e := range m.executors {
		if e.ColonyName == colonyName {
			count++
		}
	}
	return count, nil
}

func (m *MockExecutorDB) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	return 0, nil
}

func (m *MockExecutorDB) UpdateExecutorCapabilities(colonyName string, executorName string, capabilities core.Capabilities) error {
	return nil
}

// MockUserDB is a mock implementation of UserDatabase
type MockUserDB struct {
	users            []*core.User
	getUserByIDErr   error
	getUserReturnNil bool
}

func (m *MockUserDB) AddUser(user *core.User) error {
	m.users = append(m.users, user)
	return nil
}

func (m *MockUserDB) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	var result []*core.User
	for _, u := range m.users {
		if u.ColonyName == colonyName {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *MockUserDB) GetUserByID(colonyName string, userID string) (*core.User, error) {
	if m.getUserByIDErr != nil {
		return nil, m.getUserByIDErr
	}
	if m.getUserReturnNil {
		return nil, nil
	}
	for _, u := range m.users {
		if u.ColonyName == colonyName && u.ID == userID {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserDB) GetUserByName(colonyName string, name string) (*core.User, error) {
	for _, u := range m.users {
		if u.ColonyName == colonyName && u.Name == name {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockUserDB) RemoveUserByID(colonyName string, userID string) error {
	return nil
}

func (m *MockUserDB) RemoveUserByName(colonyName string, name string) error {
	return nil
}

func (m *MockUserDB) RemoveUsersByColonyName(colonyName string) error {
	return nil
}

// MockProcessController is a mock implementation of process.Controller
type MockProcessController struct {
	processes   []*core.Process
	addErr      error
}

func (m *MockProcessController) AddProcessToDB(p *core.Process) (*core.Process, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	m.processes = append(m.processes, p)
	return p, nil
}

func (m *MockProcessController) AddProcess(p *core.Process) (*core.Process, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	m.processes = append(m.processes, p)
	return p, nil
}

func (m *MockProcessController) RemoveProcess(processID string) error {
	return nil
}

func (m *MockProcessController) RemoveAllProcesses(colonyName string, state int) error {
	return nil
}

func (m *MockProcessController) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	return nil
}

func (m *MockProcessController) CloseFailed(processID string, errs []string) error {
	return nil
}

func (m *MockProcessController) Assign(executorID string, colonyName string, cpu int64, memory int64) (*process.AssignResult, error) {
	return nil, nil
}

func (m *MockProcessController) DistributedAssign(executor *core.Executor, colonyName string, cpu int64, memory int64, storage int64) (*process.AssignResult, error) {
	return nil, nil
}

func (m *MockProcessController) UnassignExecutor(processID string) error {
	return nil
}

func (m *MockProcessController) PauseColonyAssignments(colonyName string) error {
	return nil
}

func (m *MockProcessController) ResumeColonyAssignments(colonyName string) error {
	return nil
}

func (m *MockProcessController) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return false, nil
}

func (m *MockProcessController) GetEventHandler() *process.EventHandler {
	return nil
}

func (m *MockProcessController) IsLeader() bool {
	return true
}

func (m *MockProcessController) GetEtcdServer() process.EtcdServer {
	return nil
}

// MockCronController is a mock implementation of CronController
type MockCronController struct {
	crons     []*core.Cron
	addErr    error
	runErr    error
	removeErr error
}

func (m *MockCronController) AddCron(cron *core.Cron) (*core.Cron, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	m.crons = append(m.crons, cron)
	return cron, nil
}

func (m *MockCronController) RunCron(cronID string) (*core.Cron, error) {
	if m.runErr != nil {
		return nil, m.runErr
	}
	for _, c := range m.crons {
		if c.ID == cronID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockCronController) RemoveCron(cronID string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func (m *MockCronController) GetCronPeriod() int {
	return 60
}

// MockServer implements the Server interface for testing
type MockServer struct {
	blueprintDB       *MockBlueprintDB
	executorDB        *MockExecutorDB
	userDB            *MockUserDB
	processController *MockProcessController
	cronController    *MockCronController
	validator         security.Validator
	locationDB        database.LocationDatabase
	cronDB            database.CronDatabase
}

func (m *MockServer) HandleHTTPError(c backends.Context, err error, errorCode int) bool {
	return err != nil
}

func (m *MockServer) SendHTTPReply(c backends.Context, payloadType string, jsonString string) {
}

func (m *MockServer) SendEmptyHTTPReply(c backends.Context, payloadType string) {
}

func (m *MockServer) Validator() security.Validator {
	return m.validator
}

func (m *MockServer) BlueprintDB() database.BlueprintDatabase {
	return m.blueprintDB
}

func (m *MockServer) LocationDB() database.LocationDatabase {
	return m.locationDB
}

func (m *MockServer) ProcessController() process.Controller {
	return m.processController
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

// Helper to create a test BlueprintDefinition
func createTestBlueprintDefinition(kind, executorType, colonyName string) *core.BlueprintDefinition {
	sd := core.CreateBlueprintDefinition(
		"test-def",
		"example.com",
		"v1",
		kind,
		"testresources",
		"Namespaced",
		executorType,
		"reconcile",
	)
	sd.Metadata.ColonyName = colonyName
	return sd
}

// Helper to create a test Blueprint
func createTestBlueprint(kind, name, colonyName, executorType string) *core.Blueprint {
	bp := core.CreateBlueprint(kind, name, colonyName)
	if executorType != "" {
		bp.Handler = &core.BlueprintHandler{
			ExecutorType: executorType,
		}
	}
	return bp
}

// =============================================
// Tests for createConsolidatedReconciliationWorkflowSpec
// =============================================

func TestCreateConsolidatedReconciliationWorkflowSpec_Success(t *testing.T) {
	mockBlueprintDB := &MockBlueprintDB{}
	mockServer := &MockServer{
		blueprintDB: mockBlueprintDB,
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"

	// Create a BlueprintDefinition with handler
	sd := createTestBlueprintDefinition(kind, "docker-reconciler", colonyName)
	mockBlueprintDB.blueprintDefinitions = append(mockBlueprintDB.blueprintDefinitions, sd)

	// Add some blueprints of this kind
	bp1 := createTestBlueprint(kind, "bp1", colonyName, "")
	bp2 := createTestBlueprint(kind, "bp2", colonyName, "")
	mockBlueprintDB.blueprints = append(mockBlueprintDB.blueprints, bp1, bp2)

	// Call the function
	workflowJSON, err := handlers.createConsolidatedReconciliationWorkflowSpec(colonyName, kind, sd)

	assert.Nil(t, err)
	assert.NotEmpty(t, workflowJSON)

	// Parse the JSON to verify structure
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(workflowJSON)
	assert.Nil(t, err)
	assert.Equal(t, colonyName, workflowSpec.ColonyName)
	assert.Equal(t, 1, len(workflowSpec.FunctionSpecs)) // One unique handler type
	assert.Equal(t, "docker-reconciler", workflowSpec.FunctionSpecs[0].Conditions.ExecutorType)
	assert.Equal(t, "reconcile", workflowSpec.FunctionSpecs[0].FuncName)
	assert.Equal(t, kind, workflowSpec.FunctionSpecs[0].KwArgs["kind"])
}

func TestCreateConsolidatedReconciliationWorkflowSpec_NilBlueprintDefinition(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	// Call with nil BlueprintDefinition
	_, err := handlers.createConsolidatedReconciliationWorkflowSpec("test-colony", "TestKind", nil)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no handler defined for blueprint kind")
}

func TestCreateConsolidatedReconciliationWorkflowSpec_NoHandler(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	// Create BlueprintDefinition without handler
	sd := createTestBlueprintDefinition("TestKind", "", "test-colony")
	sd.Spec.Handler.ExecutorType = "" // Empty executor type

	_, err := handlers.createConsolidatedReconciliationWorkflowSpec("test-colony", "TestKind", sd)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no handler defined for blueprint kind")
}

func TestCreateConsolidatedReconciliationWorkflowSpec_DBError(t *testing.T) {
	mockBlueprintDB := &MockBlueprintDB{
		getBlueprintsByKindErr: errors.New("database connection error"),
	}
	mockServer := &MockServer{
		blueprintDB: mockBlueprintDB,
	}
	handlers := NewHandlers(mockServer)

	sd := createTestBlueprintDefinition("TestKind", "docker-reconciler", "test-colony")

	_, err := handlers.createConsolidatedReconciliationWorkflowSpec("test-colony", "TestKind", sd)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to get blueprints for kind")
}

func TestCreateConsolidatedReconciliationWorkflowSpec_MultipleExecutorTypes(t *testing.T) {
	mockBlueprintDB := &MockBlueprintDB{}
	mockServer := &MockServer{
		blueprintDB: mockBlueprintDB,
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"

	// Create a BlueprintDefinition with default handler
	sd := createTestBlueprintDefinition(kind, "default-reconciler", colonyName)
	mockBlueprintDB.blueprintDefinitions = append(mockBlueprintDB.blueprintDefinitions, sd)

	// Add blueprints with different executor types
	bp1 := createTestBlueprint(kind, "bp1", colonyName, "")                     // Uses default
	bp2 := createTestBlueprint(kind, "bp2", colonyName, "custom-reconciler-1") // Override
	bp3 := createTestBlueprint(kind, "bp3", colonyName, "custom-reconciler-2") // Override
	bp4 := createTestBlueprint(kind, "bp4", colonyName, "custom-reconciler-1") // Same as bp2
	mockBlueprintDB.blueprints = append(mockBlueprintDB.blueprints, bp1, bp2, bp3, bp4)

	// Call the function
	workflowJSON, err := handlers.createConsolidatedReconciliationWorkflowSpec(colonyName, kind, sd)

	assert.Nil(t, err)
	assert.NotEmpty(t, workflowJSON)

	// Parse and verify - should have 3 unique executor types
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(workflowJSON)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(workflowSpec.FunctionSpecs))

	// Verify all executor types are present
	executorTypes := make(map[string]bool)
	for _, spec := range workflowSpec.FunctionSpecs {
		executorTypes[spec.Conditions.ExecutorType] = true
	}
	assert.True(t, executorTypes["default-reconciler"])
	assert.True(t, executorTypes["custom-reconciler-1"])
	assert.True(t, executorTypes["custom-reconciler-2"])
}

func TestCreateConsolidatedReconciliationWorkflowSpec_EmptyBlueprints(t *testing.T) {
	mockBlueprintDB := &MockBlueprintDB{}
	mockServer := &MockServer{
		blueprintDB: mockBlueprintDB,
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"

	// Create a BlueprintDefinition but no blueprints
	sd := createTestBlueprintDefinition(kind, "docker-reconciler", colonyName)
	mockBlueprintDB.blueprintDefinitions = append(mockBlueprintDB.blueprintDefinitions, sd)

	// No blueprints added
	workflowJSON, err := handlers.createConsolidatedReconciliationWorkflowSpec(colonyName, kind, sd)

	assert.Nil(t, err)
	assert.NotEmpty(t, workflowJSON)

	// Should have 0 function specs (no blueprints to reconcile)
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(workflowJSON)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(workflowSpec.FunctionSpecs))
}

// =============================================
// Tests for createReconcilerCronWorkflowSpec
// =============================================

func TestCreateReconcilerCronWorkflowSpec_Success(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"
	executorType := "docker-reconciler"
	locationName := "datacenter-east"

	workflowJSON, err := handlers.createReconcilerCronWorkflowSpec(colonyName, kind, executorType, locationName)

	assert.Nil(t, err)
	assert.NotEmpty(t, workflowJSON)

	// Parse and verify
	workflowSpec, err := core.ConvertJSONToWorkflowSpec(workflowJSON)
	assert.Nil(t, err)
	assert.Equal(t, colonyName, workflowSpec.ColonyName)
	assert.Equal(t, 1, len(workflowSpec.FunctionSpecs))
	assert.Equal(t, executorType, workflowSpec.FunctionSpecs[0].Conditions.ExecutorType)
	assert.Equal(t, locationName, workflowSpec.FunctionSpecs[0].Conditions.LocationName)
	assert.Equal(t, "reconcile", workflowSpec.FunctionSpecs[0].FuncName)
	assert.Equal(t, kind, workflowSpec.FunctionSpecs[0].KwArgs["kind"])
}

func TestCreateReconcilerCronWorkflowSpec_EmptyExecutorType(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	_, err := handlers.createReconcilerCronWorkflowSpec("test-colony", "TestKind", "", "datacenter")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no handler executorType defined")
}

func TestCreateReconcilerCronWorkflowSpec_EmptyLocation(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	workflowJSON, err := handlers.createReconcilerCronWorkflowSpec("test-colony", "TestKind", "docker-reconciler", "")

	assert.Nil(t, err)
	assert.NotEmpty(t, workflowJSON)

	workflowSpec, err := core.ConvertJSONToWorkflowSpec(workflowJSON)
	assert.Nil(t, err)
	assert.Equal(t, "", workflowSpec.FunctionSpecs[0].Conditions.LocationName)
}

// =============================================
// Tests for createImmediateReconciliationProcess
// =============================================

func TestCreateImmediateReconciliationProcess_Success(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"

	blueprint := createTestBlueprint(kind, "my-blueprint", colonyName, "")
	blueprint.Metadata.LocationName = "datacenter-east"

	sd := createTestBlueprintDefinition(kind, "docker-reconciler", colonyName)

	process, err := handlers.createImmediateReconciliationProcess(blueprint, sd, "initiator-123", "test-initiator")

	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "reconcile", process.FunctionSpec.FuncName)
	assert.Equal(t, "docker-reconciler", process.FunctionSpec.Conditions.ExecutorType)
	assert.Equal(t, "datacenter-east", process.FunctionSpec.Conditions.LocationName)
	assert.Equal(t, colonyName, process.FunctionSpec.Conditions.ColonyName)
	assert.Equal(t, kind, process.FunctionSpec.KwArgs["kind"])
	assert.Equal(t, "my-blueprint", process.FunctionSpec.KwArgs["blueprintName"])
	assert.Equal(t, "initiator-123", process.InitiatorID)
	assert.Equal(t, "test-initiator", process.InitiatorName)
}

func TestCreateImmediateReconciliationProcess_BlueprintHandlerOverride(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	colonyName := "test-colony"
	kind := "TestKind"

	// Blueprint has its own handler that should override definition
	blueprint := createTestBlueprint(kind, "my-blueprint", colonyName, "custom-reconciler")
	sd := createTestBlueprintDefinition(kind, "docker-reconciler", colonyName)

	process, err := handlers.createImmediateReconciliationProcess(blueprint, sd, "initiator-123", "test-initiator")

	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "custom-reconciler", process.FunctionSpec.Conditions.ExecutorType)
}

func TestCreateImmediateReconciliationProcess_NoHandler(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	blueprint := createTestBlueprint("TestKind", "my-blueprint", "test-colony", "")
	sd := createTestBlueprintDefinition("TestKind", "", "test-colony")
	sd.Spec.Handler.ExecutorType = "" // No handler

	_, err := handlers.createImmediateReconciliationProcess(blueprint, sd, "initiator-123", "test-initiator")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no handler defined for blueprint kind")
}

func TestCreateImmediateReconciliationProcess_NilDefinition(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	// Blueprint has its own handler, so nil definition should work
	blueprint := createTestBlueprint("TestKind", "my-blueprint", "test-colony", "blueprint-handler")

	process, err := handlers.createImmediateReconciliationProcess(blueprint, nil, "initiator-123", "test-initiator")

	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Equal(t, "blueprint-handler", process.FunctionSpec.Conditions.ExecutorType)
}

func TestCreateImmediateReconciliationProcess_NilDefinitionNoHandler(t *testing.T) {
	mockServer := &MockServer{
		blueprintDB: &MockBlueprintDB{},
	}
	handlers := NewHandlers(mockServer)

	// Blueprint without handler and nil definition should fail
	blueprint := createTestBlueprint("TestKind", "my-blueprint", "test-colony", "")

	_, err := handlers.createImmediateReconciliationProcess(blueprint, nil, "initiator-123", "test-initiator")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no handler defined for blueprint kind")
}

// =============================================
// Tests for resolveInitiator
// =============================================

func TestResolveInitiator_ExecutorFound(t *testing.T) {
	mockExecutorDB := &MockExecutorDB{}
	mockServer := &MockServer{
		executorDB: mockExecutorDB,
		userDB:     &MockUserDB{},
	}
	handlers := NewHandlers(mockServer)

	// Add an executor
	executor := &core.Executor{
		ID:         "executor-123",
		Name:       "test-executor",
		ColonyName: "test-colony",
	}
	mockExecutorDB.executors = append(mockExecutorDB.executors, executor)

	name, err := handlers.resolveInitiator("test-colony", "executor-123")

	assert.Nil(t, err)
	assert.Equal(t, "test-executor", name)
}

func TestResolveInitiator_UserFound(t *testing.T) {
	mockExecutorDB := &MockExecutorDB{
		getByIDReturnNil: true, // Executor lookup returns nil (not found)
	}
	mockUserDB := &MockUserDB{}
	mockServer := &MockServer{
		executorDB: mockExecutorDB,
		userDB:     mockUserDB,
	}
	handlers := NewHandlers(mockServer)

	// Add a user
	user := &core.User{
		ID:         "user-123",
		Name:       "test-user",
		ColonyName: "test-colony",
	}
	mockUserDB.users = append(mockUserDB.users, user)

	name, err := handlers.resolveInitiator("test-colony", "user-123")

	assert.Nil(t, err)
	assert.Equal(t, "test-user", name)
}

func TestResolveInitiator_ExecutorDBError(t *testing.T) {
	mockExecutorDB := &MockExecutorDB{
		getByIDErr: errors.New("database error"),
	}
	mockServer := &MockServer{
		executorDB: mockExecutorDB,
		userDB:     &MockUserDB{},
	}
	handlers := NewHandlers(mockServer)

	_, err := handlers.resolveInitiator("test-colony", "executor-123")

	assert.NotNil(t, err)
	assert.Equal(t, "database error", err.Error())
}

func TestResolveInitiator_UserDBError(t *testing.T) {
	mockExecutorDB := &MockExecutorDB{
		getByIDReturnNil: true, // Executor not found
	}
	mockUserDB := &MockUserDB{
		getUserByIDErr: errors.New("user database error"),
	}
	mockServer := &MockServer{
		executorDB: mockExecutorDB,
		userDB:     mockUserDB,
	}
	handlers := NewHandlers(mockServer)

	_, err := handlers.resolveInitiator("test-colony", "user-123")

	assert.NotNil(t, err)
	assert.Equal(t, "user database error", err.Error())
}

func TestResolveInitiator_NotFound(t *testing.T) {
	mockExecutorDB := &MockExecutorDB{
		getByIDReturnNil: true, // Executor not found
	}
	mockUserDB := &MockUserDB{
		getUserReturnNil: true, // User not found
	}
	mockServer := &MockServer{
		executorDB: mockExecutorDB,
		userDB:     mockUserDB,
	}
	handlers := NewHandlers(mockServer)

	_, err := handlers.resolveInitiator("test-colony", "unknown-id")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Could not derive InitiatorName")
}

// =============================================
// Tests for NewHandlers
// =============================================

func TestNewHandlers(t *testing.T) {
	mockServer := &MockServer{}
	handlers := NewHandlers(mockServer)

	assert.NotNil(t, handlers)
}
