package embedded

import (
	"os"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helpers (using names that do not collide with embedded_test.go)
// ---------------------------------------------------------------------------

func addColonyForPhase3(t *testing.T, db *EmbeddedDatabase, name string) {
	t.Helper()
	colony := core.CreateColony(core.GenerateRandomID(), name)
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}
}

func addExecutorForPhase3(t *testing.T, db *EmbeddedDatabase, colonyName, name, execType string) *core.Executor {
	t.Helper()
	e := core.CreateExecutor(core.GenerateRandomID(), execType, name, colonyName, time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}
	return e
}

// createTestProcessP3 creates a minimal WAITING process. Unlike the shared
// createTestProcess (which uses CreateFunctionSpec and sets resource defaults),
// this one creates a bare-bones process useful for tests that do not need the
// resource fields populated.
func createTestProcessP3(colonyName, executorType string) *core.Process {
	env := make(map[string]string)
	funcSpec := core.CreateFunctionSpec(
		"", "test-func-p3", []interface{}{}, map[string]interface{}{},
		colonyName, []string{}, executorType,
		0, 0, 0, env, []string{}, 0, "",
	)
	return core.CreateProcess(funcSpec)
}

// ---------------------------------------------------------------------------
// 1. FindProcessesByExecutorID
// ---------------------------------------------------------------------------

func TestP3FindProcessesByExecutorID(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	exec := addExecutorForPhase3(t, db, "c1", "exec1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	found, err := db.FindProcessesByExecutorID("c1", exec.ID, 60, core.RUNNING)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p.ID, found[0].ID)
}

// ---------------------------------------------------------------------------
// 2. findProcessesByState with filters
// ---------------------------------------------------------------------------

func TestP3FindProcessesByStateWithExecutorTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "container")
	assert.NoError(t, db.AddProcess(p1))
	assert.NoError(t, db.AddProcess(p2))

	found, err := db.FindWaitingProcesses("c1", "cli", "", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p1.ID, found[0].ID)
}

func TestP3FindProcessesByStateWithLabelFilter(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p1 := createTestProcess("c1", "cli")
	p1.FunctionSpec.Label = "alpha"
	p2 := createTestProcess("c1", "cli")
	p2.FunctionSpec.Label = "beta"
	assert.NoError(t, db.AddProcess(p1))
	assert.NoError(t, db.AddProcess(p2))

	found, err := db.FindWaitingProcesses("c1", "", "alpha", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p1.ID, found[0].ID)
}

func TestP3FindProcessesByStateWithInitiatorFilter(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p1 := createTestProcess("c1", "cli")
	p1.InitiatorName = "user-a"
	p2 := createTestProcess("c1", "cli")
	p2.InitiatorName = "user-b"
	assert.NoError(t, db.AddProcess(p1))
	assert.NoError(t, db.AddProcess(p2))

	found, err := db.FindWaitingProcesses("c1", "", "", "user-a", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p1.ID, found[0].ID)
}

// ---------------------------------------------------------------------------
// 3. MarkSuccessful error paths
// ---------------------------------------------------------------------------

func TestP3MarkSuccessfulNotFound(t *testing.T) {
	db := setupTestDB(t)
	_, _, err := db.MarkSuccessful("nonexistent")
	assert.Error(t, err)
}

func TestP3MarkSuccessfulFailedProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	_, _, err := db.MarkSuccessful(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed process as successful")
}

func TestP3MarkSuccessfulCancelledProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	_, _, err := db.MarkSuccessful(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled process as successful")
}

func TestP3MarkSuccessfulWaitingProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	_, _, err := db.MarkSuccessful(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waiting process as successful")
}

// ---------------------------------------------------------------------------
// 4. MarkFailed error paths
// ---------------------------------------------------------------------------

func TestP3MarkFailedNotFound(t *testing.T) {
	db := setupTestDB(t)
	err := db.MarkFailed("nonexistent", []string{"err"})
	assert.Error(t, err)
}

func TestP3MarkFailedSuccessfulProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	err = db.MarkFailed(p.ID, []string{"err"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "successful process as failed")
}

func TestP3MarkFailedAlreadyFailed(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err1"}))

	err := db.MarkFailed(p.ID, []string{"err2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed process as failed")
}

func TestP3MarkFailedCancelledProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	err := db.MarkFailed(p.ID, []string{"err"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled process as failed")
}

// ---------------------------------------------------------------------------
// 5. MarkCancelled error paths
// ---------------------------------------------------------------------------

func TestP3MarkCancelledNotFound(t *testing.T) {
	db := setupTestDB(t)
	err := db.MarkCancelled("nonexistent")
	assert.Error(t, err)
}

func TestP3MarkCancelledSuccessfulProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	err = db.MarkCancelled(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancel successful")
}

func TestP3MarkCancelledFailedProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	err := db.MarkCancelled(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancel failed")
}

func TestP3MarkCancelledAlreadyCancelled(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	err := db.MarkCancelled(p.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancel already cancelled")
}

// ---------------------------------------------------------------------------
// 6. Assign error paths
// ---------------------------------------------------------------------------

func TestP3AssignNotFound(t *testing.T) {
	db := setupTestDB(t)
	p := createTestProcessP3("c1", "cli")
	err := db.Assign("exec-id", p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestP3AssignAlreadyAssigned(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	err := db.Assign(exec.ID, p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already assigned")
}

func TestP3AssignCancelledProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	err := db.Assign("exec-id", p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled process")
}

// ---------------------------------------------------------------------------
// 7. Unassign
// ---------------------------------------------------------------------------

func TestP3Unassign(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	err := db.Unassign(p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.IsAssigned)
}

// ---------------------------------------------------------------------------
// 8. ResetProcess
// ---------------------------------------------------------------------------

func TestP3ResetProcess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	err := db.ResetProcess(p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.IsAssigned)
	assert.Empty(t, got.AssignedExecutorID)
}

// ---------------------------------------------------------------------------
// 9. SetInput, SetOutput, SetErrors
// ---------------------------------------------------------------------------

func TestP3SetInput(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetInput(p.ID, []interface{}{"hello", 42})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Len(t, got.Input, 2)
}

func TestP3SetOutput(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetOutput(p.ID, []interface{}{"result"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Len(t, got.Output, 1)
}

func TestP3SetErrors(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetErrors(p.ID, []string{"error1", "error2"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Len(t, got.Errors, 2)
}

// ---------------------------------------------------------------------------
// 10. SetProcessState
// ---------------------------------------------------------------------------

func TestP3SetProcessState(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetProcessState(p.ID, core.RUNNING)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.RUNNING, got.State)
}

// ---------------------------------------------------------------------------
// 11. SetParents, SetChildren, SetWaitForParents
// ---------------------------------------------------------------------------

func TestP3SetParents(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetParents(p.ID, []string{"parent1", "parent2"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"parent1", "parent2"}, got.Parents)
}

func TestP3SetChildren(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetChildren(p.ID, []string{"child1"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"child1"}, got.Children)
}

func TestP3SetWaitForParents(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.SetWaitForParents(p.ID, true)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.True(t, got.WaitForParents)
}

// ---------------------------------------------------------------------------
// 12. FindAllRunningProcesses, FindAllWaitingProcesses
// ---------------------------------------------------------------------------

func TestP3FindAllRunningProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	found, err := db.FindAllRunningProcesses()
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p.ID, found[0].ID)
}

func TestP3FindAllWaitingProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	found, err := db.FindAllWaitingProcesses()
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, p.ID, found[0].ID)
}

// ---------------------------------------------------------------------------
// 13. RemoveAllProcessesByProcessGraphID
// ---------------------------------------------------------------------------

func TestP3RemoveAllProcessesByProcessGraphID(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	graphID := core.GenerateRandomID()

	p := createTestProcess("c1", "cli")
	p.ProcessGraphID = graphID
	assert.NoError(t, db.AddProcess(p))

	pg := &core.ProcessGraph{
		ID:         graphID,
		ColonyName: "c1",
		ProcessIDs: []string{p.ID},
	}
	assert.NoError(t, db.AddProcessGraph(pg))

	err := db.RemoveAllProcessesByProcessGraphID(graphID)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

// ---------------------------------------------------------------------------
// 14. RemoveAllProcessesInProcessGraphsByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllProcessesInProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	graphID := core.GenerateRandomID()

	p1 := createTestProcess("c1", "cli")
	p1.ProcessGraphID = graphID
	assert.NoError(t, db.AddProcess(p1))

	// Non-graph process should remain
	p2 := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p2))

	err := db.RemoveAllProcessesInProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got1, _ := db.GetProcessByID(p1.ID)
	assert.Nil(t, got1)

	got2, _ := db.GetProcessByID(p2.ID)
	assert.NotNil(t, got2)
}

// ---------------------------------------------------------------------------
// 15. RemoveAll*ProcessesByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllWaitingProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	err := db.RemoveAllWaitingProcessesByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessByID(p.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllRunningProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	err := db.RemoveAllRunningProcessesByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessByID(p.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllSuccessfulProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	err = db.RemoveAllSuccessfulProcessesByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessByID(p.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllFailedProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	err := db.RemoveAllFailedProcessesByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessByID(p.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllCancelledProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	err := db.RemoveAllCancelledProcessesByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessByID(p.ID)
	assert.Nil(t, got)
}

// ---------------------------------------------------------------------------
// 16. Count processes (global and ByColonyName)
// ---------------------------------------------------------------------------

func TestP3CountProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	count, err := db.CountProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountWaitingProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	count, err := db.CountWaitingProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountRunningProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	count, err := db.CountRunningProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountSuccessfulProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	count, err := db.CountSuccessfulProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountFailedProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	count, err := db.CountFailedProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountCancelledProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	count, err := db.CountCancelledProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountWaitingProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")
	addColonyForPhase3(t, db, "c2")

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c2", "cli")
	assert.NoError(t, db.AddProcess(p1))
	assert.NoError(t, db.AddProcess(p2))

	count, err := db.CountWaitingProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountRunningProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	count, err := db.CountRunningProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountSuccessfulProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	count, err := db.CountSuccessfulProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountFailedProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	count, err := db.CountFailedProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountCancelledProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	count, err := db.CountCancelledProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

// ---------------------------------------------------------------------------
// 17. matchCandidate with resource constraints and location check
// ---------------------------------------------------------------------------

func TestP3MatchCandidateResourceConstraints(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	// createTestProcess sets CPU=1000m, Memory=1Gi, Storage=10Gi, Nodes=1, Processes=1, ProcessesPerNode=1
	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.CPU = "2000m"
	p.FunctionSpec.Conditions.Memory = "1Gi"
	p.FunctionSpec.Conditions.Storage = "10Gi"
	p.FunctionSpec.Conditions.Nodes = 2
	p.FunctionSpec.Conditions.Processes = 4
	p.FunctionSpec.Conditions.ProcessesPerNode = 2
	assert.NoError(t, db.AddProcess(p))

	// Sufficient resources
	found, err := db.FindCandidates("c1", "cli", "", 4000, 2*1024*1024*1024, 20*1024*1024*1024, 4, 8, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 1)

	// Insufficient CPU
	found, err = db.FindCandidates("c1", "cli", "", 1000, 2*1024*1024*1024, 20*1024*1024*1024, 4, 8, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)

	// Insufficient memory
	found, err = db.FindCandidates("c1", "cli", "", 4000, 100, 20*1024*1024*1024, 4, 8, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)

	// Insufficient storage
	found, err = db.FindCandidates("c1", "cli", "", 4000, 2*1024*1024*1024, 100, 4, 8, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)

	// Insufficient nodes
	found, err = db.FindCandidates("c1", "cli", "", 4000, 2*1024*1024*1024, 20*1024*1024*1024, 1, 8, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)

	// Insufficient processes
	found, err = db.FindCandidates("c1", "cli", "", 4000, 2*1024*1024*1024, 20*1024*1024*1024, 4, 2, 4, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)

	// Insufficient processesPerNode
	found, err = db.FindCandidates("c1", "cli", "", 4000, 2*1024*1024*1024, 20*1024*1024*1024, 4, 8, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

func TestP3MatchCandidateLocationCheck(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.LocationName = "datacenter-1"
	assert.NoError(t, db.AddProcess(p))

	// Matching location -- needs enough resources to satisfy default constraints
	found, err := db.FindCandidates("c1", "cli", "datacenter-1", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 1)

	// Non-matching location
	found, err = db.FindCandidates("c1", "cli", "datacenter-2", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// 18. FindCandidatesByName
// ---------------------------------------------------------------------------

func TestP3FindCandidatesByName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.ExecutorNames = []string{"exec-a", "exec-b"}
	assert.NoError(t, db.AddProcess(p))

	// Should find because exec-a is in the target list
	found, err := db.FindCandidatesByName("c1", "exec-a", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 1)

	// Should not find because exec-c is not in the target list
	found, err = db.FindCandidatesByName("c1", "exec-c", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// 19. SelectAndAssign
// ---------------------------------------------------------------------------

func TestP3SelectAndAssignHappy(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))

	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")

	assigned, err := db.SelectAndAssign("c1", exec.ID, "e1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.NotNil(t, assigned)
	assert.Equal(t, p.ID, assigned.ID)
	assert.Equal(t, core.RUNNING, assigned.State)
}

func TestP3SelectAndAssignNoCandidates(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")

	assigned, err := db.SelectAndAssign("c1", exec.ID, "e1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Nil(t, assigned)
}

// ---------------------------------------------------------------------------
// 20. Drop
// ---------------------------------------------------------------------------

func TestP3Drop(t *testing.T) {
	dir, err := os.MkdirTemp("", "embedded-drop-p3-*")
	if err != nil {
		t.Fatal(err)
	}

	db := CreateEmbeddedDatabase(dir)
	if err := db.Initialize(); err != nil {
		t.Fatal(err)
	}

	colony := core.CreateColony(core.GenerateRandomID(), "test-colony")
	assert.NoError(t, db.AddColony(colony))

	err = db.Drop()
	assert.NoError(t, err)

	_, statErr := os.Stat(dir)
	assert.True(t, os.IsNotExist(statErr))
}

// ---------------------------------------------------------------------------
// 21. Close and reopen (WAL persistence)
// ---------------------------------------------------------------------------

func TestP3CloseAndReopen(t *testing.T) {
	dir, err := os.MkdirTemp("", "embedded-close-p3-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	db := CreateEmbeddedDatabase(dir)
	if err := db.Initialize(); err != nil {
		t.Fatal(err)
	}

	colony := core.CreateColony(core.GenerateRandomID(), "persist-colony")
	assert.NoError(t, db.AddColony(colony))

	db.Close()

	// Reopen
	db2 := CreateEmbeddedDatabase(dir)
	if err := db2.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	got, err := db2.GetColonyByName("persist-colony")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "persist-colony", got.Name)
}

// ---------------------------------------------------------------------------
// 22. GetBlueprintsByNamespaceKindAndLocation
// ---------------------------------------------------------------------------

func TestP3GetBlueprintsByNamespaceKindAndLocationNoLocation(t *testing.T) {
	db := setupTestDB(t)

	bp := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:       "bp1",
			ColonyName: "ns1",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprint(bp))

	found, err := db.GetBlueprintsByNamespaceKindAndLocation("ns1", "deployment", "")
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3GetBlueprintsByNamespaceKindAndLocationWithLocation(t *testing.T) {
	db := setupTestDB(t)

	bp1 := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:         "bp1",
			ColonyName:   "ns1",
			LocationName: "loc-a",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	bp2 := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:         "bp2",
			ColonyName:   "ns1",
			LocationName: "loc-b",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprint(bp1))
	assert.NoError(t, db.AddBlueprint(bp2))

	found, err := db.GetBlueprintsByNamespaceKindAndLocation("ns1", "deployment", "loc-a")
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, "bp1", found[0].Metadata.Name)
}

// ---------------------------------------------------------------------------
// 23. RemoveBlueprintHistory
// ---------------------------------------------------------------------------

func TestP3RemoveBlueprintHistory(t *testing.T) {
	db := setupTestDB(t)

	bpID := core.GenerateRandomID()
	h1 := &core.BlueprintHistory{
		ID:          core.GenerateRandomID(),
		BlueprintID: bpID,
		Generation:  1,
		Timestamp:   time.Now(),
		Spec:        map[string]interface{}{"k": "v"},
		Status:      map[string]interface{}{},
	}
	h2 := &core.BlueprintHistory{
		ID:          core.GenerateRandomID(),
		BlueprintID: bpID,
		Generation:  2,
		Timestamp:   time.Now(),
		Spec:        map[string]interface{}{"k": "v2"},
		Status:      map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprintHistory(h1))
	assert.NoError(t, db.AddBlueprintHistory(h2))

	history, err := db.GetBlueprintHistory(bpID, 0)
	assert.NoError(t, err)
	assert.Len(t, history, 2)

	err = db.RemoveBlueprintHistory(bpID)
	assert.NoError(t, err)

	history, err = db.GetBlueprintHistory(bpID, 0)
	assert.NoError(t, err)
	assert.Len(t, history, 0)
}

// ---------------------------------------------------------------------------
// 24. GetBlueprintHistoryByGeneration
// ---------------------------------------------------------------------------

func TestP3GetBlueprintHistoryByGeneration(t *testing.T) {
	db := setupTestDB(t)

	bpID := core.GenerateRandomID()
	h1 := &core.BlueprintHistory{
		ID:          core.GenerateRandomID(),
		BlueprintID: bpID,
		Generation:  1,
		Timestamp:   time.Now(),
		Spec:        map[string]interface{}{"gen": "1"},
		Status:      map[string]interface{}{},
	}
	h2 := &core.BlueprintHistory{
		ID:          core.GenerateRandomID(),
		BlueprintID: bpID,
		Generation:  2,
		Timestamp:   time.Now(),
		Spec:        map[string]interface{}{"gen": "2"},
		Status:      map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprintHistory(h1))
	assert.NoError(t, db.AddBlueprintHistory(h2))

	got, err := db.GetBlueprintHistoryByGeneration(bpID, 2)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, int64(2), got.Generation)

	// Non-existent generation
	got, err = db.GetBlueprintHistoryByGeneration(bpID, 99)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

// ---------------------------------------------------------------------------
// 25. UpdateBlueprintStatus
// ---------------------------------------------------------------------------

func TestP3UpdateBlueprintStatus(t *testing.T) {
	db := setupTestDB(t)

	bp := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:       "bp-status",
			ColonyName: "ns1",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprint(bp))

	newStatus := map[string]interface{}{"phase": "ready", "replicas": float64(3)}
	err := db.UpdateBlueprintStatus(bp.ID, newStatus)
	assert.NoError(t, err)

	got, err := db.GetBlueprintByID(bp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "ready", got.Status["phase"])
	assert.Equal(t, float64(3), got.Status["replicas"])
}

func TestP3UpdateBlueprintStatusNotFound(t *testing.T) {
	db := setupTestDB(t)
	err := db.UpdateBlueprintStatus("nonexistent", map[string]interface{}{})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 26. RemoveBlueprintsByNamespace
// ---------------------------------------------------------------------------

func TestP3RemoveBlueprintsByNamespace(t *testing.T) {
	db := setupTestDB(t)

	bp1 := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:       "bp1",
			ColonyName: "ns1",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	bp2 := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:       "bp2",
			ColonyName: "ns1",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	bp3 := &core.Blueprint{
		ID:   core.GenerateRandomID(),
		Kind: "deployment",
		Metadata: core.BlueprintMetadata{
			Name:       "bp3",
			ColonyName: "ns2",
		},
		Spec:   map[string]interface{}{"key": "value"},
		Status: map[string]interface{}{},
	}
	assert.NoError(t, db.AddBlueprint(bp1))
	assert.NoError(t, db.AddBlueprint(bp2))
	assert.NoError(t, db.AddBlueprint(bp3))

	err := db.RemoveBlueprintsByNamespace("ns1")
	assert.NoError(t, err)

	count, _ := db.CountBlueprintsByNamespace("ns1")
	assert.Equal(t, 0, count)

	count, _ = db.CountBlueprintsByNamespace("ns2")
	assert.Equal(t, 1, count)
}

// ---------------------------------------------------------------------------
// 27. AddExecutor with re-registration
// ---------------------------------------------------------------------------

func TestP3AddExecutorReregistration(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	exec := core.CreateExecutor(core.GenerateRandomID(), "cli", "re-exec", "c1", time.Now(), time.Now())
	assert.NoError(t, db.AddExecutor(exec))

	// RemoveExecutorByName marks as UNREGISTERED
	assert.NoError(t, db.RemoveExecutorByName("c1", "re-exec"))

	got, err := db.GetExecutorByName("c1", "re-exec")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, core.UNREGISTERED, got.State)

	// Re-register with same name
	exec2 := core.CreateExecutor(core.GenerateRandomID(), "cli", "re-exec", "c1", time.Now(), time.Now())
	assert.NoError(t, db.AddExecutor(exec2))

	got2, err := db.GetExecutorByName("c1", "re-exec")
	assert.NoError(t, err)
	assert.NotNil(t, got2)
	assert.Equal(t, core.PENDING, got2.State)
}

// ---------------------------------------------------------------------------
// 28. RemoveExecutorByName with running processes
// ---------------------------------------------------------------------------

func TestP3RemoveExecutorByNameWithRunningProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	exec := addExecutorForPhase3(t, db, "c1", "exec-rm", "cli")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.Assign(exec.ID, p))

	got, _ := db.GetProcessByID(p.ID)
	assert.Equal(t, core.RUNNING, got.State)

	assert.NoError(t, db.RemoveExecutorByName("c1", "exec-rm"))

	got, _ = db.GetProcessByID(p.ID)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.IsAssigned)
}

// ---------------------------------------------------------------------------
// 29. GetExecutorsByBlueprintID
// ---------------------------------------------------------------------------

func TestP3GetExecutorsByBlueprintID(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	bpID := core.GenerateRandomID()
	exec := core.CreateExecutor(core.GenerateRandomID(), "cli", "bp-exec", "c1", time.Now(), time.Now())
	exec.BlueprintID = bpID
	assert.NoError(t, db.AddExecutor(exec))

	found, err := db.GetExecutorsByBlueprintID(bpID)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, "bp-exec", found[0].Name)

	found, err = db.GetExecutorsByBlueprintID("nonexistent")
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// 30. SetAllocations
// ---------------------------------------------------------------------------

func TestP3SetAllocations(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	addExecutorForPhase3(t, db, "c1", "alloc-exec", "cli")

	alloc := core.Allocations{
		Projects: map[string]core.Project{
			"proj-a": {AllocatedCPU: 1000, UsedCPU: 500},
		},
	}
	err := db.SetAllocations("c1", "alloc-exec", alloc)
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "alloc-exec")
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), got.Allocations.Projects["proj-a"].AllocatedCPU)
	assert.Equal(t, int64(500), got.Allocations.Projects["proj-a"].UsedCPU)
}

// ---------------------------------------------------------------------------
// 31. CountExecutorsByColonyNameAndState
// ---------------------------------------------------------------------------

func TestP3CountExecutorsByColonyNameAndState(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	exec := addExecutorForPhase3(t, db, "c1", "state-exec", "cli")

	count, err := db.CountExecutorsByColonyNameAndState("c1", core.PENDING)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	assert.NoError(t, db.ApproveExecutor(exec))

	count, err = db.CountExecutorsByColonyNameAndState("c1", core.APPROVED)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = db.CountExecutorsByColonyNameAndState("c1", core.PENDING)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// ---------------------------------------------------------------------------
// 32. UpdateExecutorCapabilities
// ---------------------------------------------------------------------------

func TestP3UpdateExecutorCapabilities(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	addExecutorForPhase3(t, db, "c1", "cap-exec", "cli")

	caps := core.Capabilities{
		Hardware: []core.Hardware{
			{Model: "gpu-a100", CPU: "4000m", Memory: "16Gi", Network: []string{"10.0.0.1"}},
		},
		Software: []core.Software{
			{Name: "cuda", Version: "12.0"},
		},
	}
	err := db.UpdateExecutorCapabilities("c1", "cap-exec", caps)
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "cap-exec")
	assert.NoError(t, err)
	assert.Len(t, got.Capabilities.Hardware, 1)
	assert.Equal(t, "gpu-a100", got.Capabilities.Hardware[0].Model)
	assert.Len(t, got.Capabilities.Software, 1)
	assert.Equal(t, "cuda", got.Capabilities.Software[0].Name)
}

// ---------------------------------------------------------------------------
// 33. RejectExecutor
// ---------------------------------------------------------------------------

func TestP3RejectExecutor(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	exec := addExecutorForPhase3(t, db, "c1", "rej-exec", "cli")
	assert.Equal(t, core.PENDING, exec.State)

	err := db.RejectExecutor(exec)
	assert.NoError(t, err)

	got, err := db.GetExecutorByID(exec.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.REJECTED, got.State)
}

// ---------------------------------------------------------------------------
// 34. SetProcessGraphState transitions
// ---------------------------------------------------------------------------

func TestP3SetProcessGraphStateWaitingToRunning(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))

	err := db.SetProcessGraphState(pg.ID, core.RUNNING)
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Equal(t, core.RUNNING, got.State)
	assert.False(t, got.StartTime.IsZero())
}

func TestP3SetProcessGraphStateRunningToSuccess(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	err := db.SetProcessGraphState(pg.ID, core.SUCCESS)
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Equal(t, core.SUCCESS, got.State)
	assert.False(t, got.EndTime.IsZero())
}

func TestP3SetProcessGraphStateRunningToFailed(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	err := db.SetProcessGraphState(pg.ID, core.FAILED)
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Equal(t, core.FAILED, got.State)
}

func TestP3SetProcessGraphStateRunningToCancelled(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	err := db.SetProcessGraphState(pg.ID, core.CANCELLED)
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Equal(t, core.CANCELLED, got.State)
}

// ---------------------------------------------------------------------------
// 35. RemoveAll*ProcessGraphsByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllWaitingProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))

	err := db.RemoveAllWaitingProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllRunningProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	err := db.RemoveAllRunningProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllSuccessfulProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.SUCCESS))

	err := db.RemoveAllSuccessfulProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllFailedProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.FAILED))

	err := db.RemoveAllFailedProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Nil(t, got)
}

func TestP3RemoveAllCancelledProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.CANCELLED))

	err := db.RemoveAllCancelledProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	got, _ := db.GetProcessGraphByID(pg.ID)
	assert.Nil(t, got)
}

// ---------------------------------------------------------------------------
// 36. Find*ProcessGraphs
// ---------------------------------------------------------------------------

func TestP3FindRunningProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	found, err := db.FindRunningProcessGraphs("c1", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindSuccessfulProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.SUCCESS))

	found, err := db.FindSuccessfulProcessGraphs("c1", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindFailedProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.FAILED))

	found, err := db.FindFailedProcessGraphs("c1", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindCancelledProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{
		ID:         core.GenerateRandomID(),
		ColonyName: "c1",
		ProcessIDs: []string{},
	}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.CANCELLED))

	found, err := db.FindCancelledProcessGraphs("c1", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

// ---------------------------------------------------------------------------
// 37. Count*ProcessGraphs and ByColonyName variants
// ---------------------------------------------------------------------------

func TestP3CountWaitingProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))

	count, err := db.CountWaitingProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountRunningProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	count, err := db.CountRunningProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountSuccessfulProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.SUCCESS))

	count, err := db.CountSuccessfulProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountFailedProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.FAILED))

	count, err := db.CountFailedProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountCancelledProcessGraphs(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.CANCELLED))

	count, err := db.CountCancelledProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountWaitingProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))

	count, err := db.CountWaitingProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountRunningProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))

	count, err := db.CountRunningProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountSuccessfulProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.SUCCESS))

	count, err := db.CountSuccessfulProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountFailedProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.FAILED))

	count, err := db.CountFailedProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestP3CountCancelledProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	pg := &core.ProcessGraph{ID: core.GenerateRandomID(), ColonyName: "c1", ProcessIDs: []string{}}
	assert.NoError(t, db.AddProcessGraph(pg))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.RUNNING))
	assert.NoError(t, db.SetProcessGraphState(pg.ID, core.CANCELLED))

	count, err := db.CountCancelledProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

// ---------------------------------------------------------------------------
// 38. GetAttributesByColonyName
// ---------------------------------------------------------------------------

func TestP3GetAttributesByColonyName(t *testing.T) {
	db := setupTestDB(t)

	attr1 := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "val1")
	attr2 := core.CreateAttribute("target2", "c1", "", core.OUT, "key2", "val2")
	attr3 := core.CreateAttribute("target3", "c2", "", core.IN, "key3", "val3")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))
	assert.NoError(t, db.AddAttribute(attr3))

	found, err := db.GetAttributesByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, found, 2)
}

// ---------------------------------------------------------------------------
// 39. GetAttribute by targetID, key, and type
// ---------------------------------------------------------------------------

func TestP3GetAttribute(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "my-key", "my-val")
	assert.NoError(t, db.AddAttribute(attr))

	got, err := db.GetAttribute("target1", "my-key", core.IN)
	assert.NoError(t, err)
	assert.Equal(t, "my-val", got.Value)

	// Wrong type
	_, err = db.GetAttribute("target1", "my-key", core.OUT)
	assert.Error(t, err)

	// Wrong key
	_, err = db.GetAttribute("target1", "wrong-key", core.IN)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 40. GetAttributesByType
// ---------------------------------------------------------------------------

func TestP3GetAttributesByType(t *testing.T) {
	db := setupTestDB(t)

	attr1 := core.CreateAttribute("target1", "c1", "", core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("target1", "c1", "", core.OUT, "k2", "v2")
	attr3 := core.CreateAttribute("target1", "c1", "", core.IN, "k3", "v3")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))
	assert.NoError(t, db.AddAttribute(attr3))

	found, err := db.GetAttributesByType("target1", core.IN)
	assert.NoError(t, err)
	assert.Len(t, found, 2)

	found, err = db.GetAttributesByType("target1", core.OUT)
	assert.NoError(t, err)
	assert.Len(t, found, 1)

	// No matches returns empty slice (not nil)
	found, err = db.GetAttributesByType("target1", core.ERR)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// 41. UpdateAttribute and UpdateAttribute not found
// ---------------------------------------------------------------------------

func TestP3UpdateAttribute(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "old-val")
	assert.NoError(t, db.AddAttribute(attr))

	attr.Value = "new-val"
	err := db.UpdateAttribute(attr)
	assert.NoError(t, err)

	got, err := db.GetAttributeByID(attr.ID)
	assert.NoError(t, err)
	assert.Equal(t, "new-val", got.Value)
}

func TestP3UpdateAttributeNotFound(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "val")
	attr.ID = "nonexistent-id"
	err := db.UpdateAttribute(attr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// ---------------------------------------------------------------------------
// 42. RemoveAttributeByID
// ---------------------------------------------------------------------------

func TestP3RemoveAttributeByID(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "val")
	assert.NoError(t, db.AddAttribute(attr))

	err := db.RemoveAttributeByID(attr.ID)
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attr.ID)
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 43. RemoveAllAttributesByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributesByColonyName(t *testing.T) {
	db := setupTestDB(t)

	attr1 := core.CreateAttribute("t1", "c1", "", core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("t2", "c1", "", core.OUT, "k2", "v2")
	attr3 := core.CreateAttribute("t3", "c2", "", core.IN, "k3", "v3")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))
	assert.NoError(t, db.AddAttribute(attr3))

	err := db.RemoveAllAttributesByColonyName("c1")
	assert.NoError(t, err)

	found, _ := db.GetAttributesByColonyName("c1")
	assert.Len(t, found, 0)

	found, _ = db.GetAttributesByColonyName("c2")
	assert.Len(t, found, 1)
}

// ---------------------------------------------------------------------------
// 44. RemoveAllAttributesByColonyNameWithState
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributesByColonyNameWithState(t *testing.T) {
	db := setupTestDB(t)

	// Create attributes with different states (non-process-graph attributes)
	attr1 := core.CreateAttribute("t1", "c1", "", core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("t2", "c1", "", core.IN, "k2", "v2")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))

	// Change state of attr2 to SUCCESS via direct store mutation
	a2, _ := db.GetAttributeByID(attr2.ID)
	cp := a2
	cp.State = core.SUCCESS
	db.attributes.Put(cp.ID, &cp)

	// Only remove PENDING state attributes (attr1 has PENDING = 0 by default)
	err := db.RemoveAllAttributesByColonyNameWithState("c1", core.PENDING)
	assert.NoError(t, err)

	// attr1 should be removed, attr2 should remain
	_, err = db.GetAttributeByID(attr1.ID)
	assert.Error(t, err) // removed

	_, err = db.GetAttributeByID(attr2.ID)
	assert.NoError(t, err) // still present
}

// ---------------------------------------------------------------------------
// 45. RemoveAllAttributesByProcessGraphID
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributesByProcessGraphID(t *testing.T) {
	db := setupTestDB(t)

	graphID := core.GenerateRandomID()
	attr1 := core.CreateAttribute("t1", "c1", graphID, core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("t2", "c1", graphID, core.OUT, "k2", "v2")
	attr3 := core.CreateAttribute("t3", "c1", "", core.IN, "k3", "v3")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))
	assert.NoError(t, db.AddAttribute(attr3))

	err := db.RemoveAllAttributesByProcessGraphID(graphID)
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attr1.ID)
	assert.Error(t, err) // removed

	_, err = db.GetAttributeByID(attr2.ID)
	assert.Error(t, err) // removed

	_, err = db.GetAttributeByID(attr3.ID)
	assert.NoError(t, err) // still present
}

// ---------------------------------------------------------------------------
// 46. RemoveAllAttributesInProcessGraphsByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributesInProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	graphID := core.GenerateRandomID()
	attrInGraph := core.CreateAttribute("t1", "c1", graphID, core.IN, "k1", "v1")
	attrNotInGraph := core.CreateAttribute("t2", "c1", "", core.IN, "k2", "v2")
	assert.NoError(t, db.AddAttribute(attrInGraph))
	assert.NoError(t, db.AddAttribute(attrNotInGraph))

	err := db.RemoveAllAttributesInProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attrInGraph.ID)
	assert.Error(t, err) // removed

	_, err = db.GetAttributeByID(attrNotInGraph.ID)
	assert.NoError(t, err) // still present
}

// ---------------------------------------------------------------------------
// 47. RemoveAllAttributesInProcessGraphsByColonyNameWithState
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributesInProcessGraphsByColonyNameWithState(t *testing.T) {
	db := setupTestDB(t)

	graphID := core.GenerateRandomID()
	attr1 := core.CreateAttribute("t1", "c1", graphID, core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("t2", "c1", graphID, core.IN, "k2", "v2")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))

	// Change attr2 state to SUCCESS
	a2, _ := db.GetAttributeByID(attr2.ID)
	cp := a2
	cp.State = core.SUCCESS
	db.attributes.Put(cp.ID, &cp)

	// Remove only PENDING state in process graphs
	err := db.RemoveAllAttributesInProcessGraphsByColonyNameWithState("c1", core.PENDING)
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attr1.ID)
	assert.Error(t, err) // removed

	_, err = db.GetAttributeByID(attr2.ID)
	assert.NoError(t, err) // still present
}

// ---------------------------------------------------------------------------
// 48. RemoveAttributesByTargetID by attribute type
// ---------------------------------------------------------------------------

func TestP3RemoveAttributesByTargetID(t *testing.T) {
	db := setupTestDB(t)

	attr1 := core.CreateAttribute("target1", "c1", "", core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("target1", "c1", "", core.OUT, "k2", "v2")
	attr3 := core.CreateAttribute("target1", "c1", "", core.IN, "k3", "v3")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))
	assert.NoError(t, db.AddAttribute(attr3))

	err := db.RemoveAttributesByTargetID("target1", core.IN)
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attr1.ID)
	assert.Error(t, err) // removed
	_, err = db.GetAttributeByID(attr3.ID)
	assert.Error(t, err) // removed
	_, err = db.GetAttributeByID(attr2.ID)
	assert.NoError(t, err) // OUT remains
}

// ---------------------------------------------------------------------------
// 49. RemoveAllAttributes
// ---------------------------------------------------------------------------

func TestP3RemoveAllAttributes(t *testing.T) {
	db := setupTestDB(t)

	attr1 := core.CreateAttribute("t1", "c1", "", core.IN, "k1", "v1")
	attr2 := core.CreateAttribute("t2", "c2", "g1", core.OUT, "k2", "v2")
	assert.NoError(t, db.AddAttribute(attr1))
	assert.NoError(t, db.AddAttribute(attr2))

	err := db.RemoveAllAttributes()
	assert.NoError(t, err)

	found, _ := db.GetAttributesByColonyName("c1")
	assert.Len(t, found, 0)
	found, _ = db.GetAttributesByColonyName("c2")
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// 50. SetGeneratorLastRun, SetGeneratorFirstPack
// ---------------------------------------------------------------------------

func TestP3SetGeneratorLastRun(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	gen := core.CreateGenerator("c1", "gen1", "workflow", 10, 60)
	gen.ID = core.GenerateRandomID()
	assert.NoError(t, db.AddGenerator(gen))

	err := db.SetGeneratorLastRun(gen.ID)
	assert.NoError(t, err)

	got, err := db.GetGeneratorByID(gen.ID)
	assert.NoError(t, err)
	assert.False(t, got.LastRun.IsZero())
}

func TestP3SetGeneratorFirstPack(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	gen := core.CreateGenerator("c1", "gen1", "workflow", 10, 60)
	gen.ID = core.GenerateRandomID()
	assert.NoError(t, db.AddGenerator(gen))

	err := db.SetGeneratorFirstPack(gen.ID)
	assert.NoError(t, err)

	got, err := db.GetGeneratorByID(gen.ID)
	assert.NoError(t, err)
	assert.False(t, got.FirstPack.IsZero())
}

// ---------------------------------------------------------------------------
// 51. FindAllGenerators
// ---------------------------------------------------------------------------

func TestP3FindAllGenerators(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")
	addColonyForPhase3(t, db, "c2")

	gen1 := core.CreateGenerator("c1", "gen1", "wf1", 10, 60)
	gen1.ID = core.GenerateRandomID()
	gen2 := core.CreateGenerator("c2", "gen2", "wf2", 5, 30)
	gen2.ID = core.GenerateRandomID()
	assert.NoError(t, db.AddGenerator(gen1))
	assert.NoError(t, db.AddGenerator(gen2))

	all, err := db.FindAllGenerators()
	assert.NoError(t, err)
	assert.Len(t, all, 2)
}

// ---------------------------------------------------------------------------
// 52. RemoveAllGeneratorsByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveAllGeneratorsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")
	addColonyForPhase3(t, db, "c2")

	gen1 := core.CreateGenerator("c1", "gen1", "wf1", 10, 60)
	gen1.ID = core.GenerateRandomID()
	gen2 := core.CreateGenerator("c1", "gen2", "wf2", 5, 30)
	gen2.ID = core.GenerateRandomID()
	gen3 := core.CreateGenerator("c2", "gen3", "wf3", 5, 30)
	gen3.ID = core.GenerateRandomID()
	assert.NoError(t, db.AddGenerator(gen1))
	assert.NoError(t, db.AddGenerator(gen2))
	assert.NoError(t, db.AddGenerator(gen3))

	err := db.RemoveAllGeneratorsByColonyName("c1")
	assert.NoError(t, err)

	all, _ := db.FindAllGenerators()
	assert.Len(t, all, 1)
	assert.Equal(t, "c2", all[0].ColonyName)
}

// ---------------------------------------------------------------------------
// 53. AddGeneratorArg, GetGeneratorArgs, CountGeneratorArgs
// ---------------------------------------------------------------------------

func TestP3AddGeneratorArgAndGetAndCount(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	genID := core.GenerateRandomID()

	arg1 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID,
		ColonyName:  "c1",
		Arg:         "arg-value-1",
	}
	arg2 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID,
		ColonyName:  "c1",
		Arg:         "arg-value-2",
	}
	assert.NoError(t, db.AddGeneratorArg(arg1))
	assert.NoError(t, db.AddGeneratorArg(arg2))

	args, err := db.GetGeneratorArgs(genID, 100)
	assert.NoError(t, err)
	assert.Len(t, args, 2)

	count, err := db.CountGeneratorArgs(genID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// ---------------------------------------------------------------------------
// 54. RemoveGeneratorArgByID, RemoveAllGeneratorArgsByGeneratorID,
//     RemoveAllGeneratorArgsByColonyName
// ---------------------------------------------------------------------------

func TestP3RemoveGeneratorArgByID(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	genID := core.GenerateRandomID()
	arg := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID,
		ColonyName:  "c1",
		Arg:         "test-arg",
	}
	assert.NoError(t, db.AddGeneratorArg(arg))

	err := db.RemoveGeneratorArgByID(arg.ID)
	assert.NoError(t, err)

	count, _ := db.CountGeneratorArgs(genID)
	assert.Equal(t, 0, count)
}

func TestP3RemoveAllGeneratorArgsByGeneratorID(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	genID := core.GenerateRandomID()
	arg1 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID,
		ColonyName:  "c1",
		Arg:         "arg1",
	}
	arg2 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID,
		ColonyName:  "c1",
		Arg:         "arg2",
	}
	assert.NoError(t, db.AddGeneratorArg(arg1))
	assert.NoError(t, db.AddGeneratorArg(arg2))

	err := db.RemoveAllGeneratorArgsByGeneratorID(genID)
	assert.NoError(t, err)

	count, _ := db.CountGeneratorArgs(genID)
	assert.Equal(t, 0, count)
}

func TestP3RemoveAllGeneratorArgsByColonyName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")
	addColonyForPhase3(t, db, "c2")

	genID1 := core.GenerateRandomID()
	genID2 := core.GenerateRandomID()
	arg1 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID1,
		ColonyName:  "c1",
		Arg:         "arg1",
	}
	arg2 := &core.GeneratorArg{
		ID:          core.GenerateRandomID(),
		GeneratorID: genID2,
		ColonyName:  "c2",
		Arg:         "arg2",
	}
	assert.NoError(t, db.AddGeneratorArg(arg1))
	assert.NoError(t, db.AddGeneratorArg(arg2))

	err := db.RemoveAllGeneratorArgsByColonyName("c1")
	assert.NoError(t, err)

	count, _ := db.CountGeneratorArgs(genID1)
	assert.Equal(t, 0, count)

	count, _ = db.CountGeneratorArgs(genID2)
	assert.Equal(t, 1, count)
}

// ---------------------------------------------------------------------------
// Additional: FindRunningProcesses, FindSuccessfulProcesses, FindFailedProcesses,
// FindCancelledProcesses via findProcessesByState (exercises non-WAITING state ordering)
// ---------------------------------------------------------------------------

func TestP3FindRunningProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	found, err := db.FindRunningProcesses("c1", "", "", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindSuccessfulProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	found, err := db.FindSuccessfulProcesses("c1", "", "", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindFailedProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	assert.NoError(t, db.MarkFailed(p.ID, []string{"err"}))

	found, err := db.FindFailedProcesses("c1", "", "", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindCancelledProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	assert.NoError(t, db.MarkCancelled(p.ID))

	found, err := db.FindCancelledProcesses("c1", "", "", "", 100)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

// ---------------------------------------------------------------------------
// Additional: SelectAndAssign with named executor target
// ---------------------------------------------------------------------------

func TestP3SelectAndAssignByName(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.ExecutorNames = []string{"e1"}
	assert.NoError(t, db.AddProcess(p))

	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")

	assigned, err := db.SelectAndAssign("c1", exec.ID, "e1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.NotNil(t, assigned)
	assert.Equal(t, p.ID, assigned.ID)
}

func TestP3SelectAndAssignByNameNoMatch(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.ExecutorNames = []string{"e2"}
	assert.NoError(t, db.AddProcess(p))

	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")

	assigned, err := db.SelectAndAssign("c1", exec.ID, "e1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Nil(t, assigned)
}

// ---------------------------------------------------------------------------
// Additional: matchCandidate with WaitForParents true
// ---------------------------------------------------------------------------

func TestP3MatchCandidateWaitForParents(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.WaitForParents = true
	assert.NoError(t, db.AddProcess(p))

	found, err := db.FindCandidates("c1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// Additional: FindCandidates excludes processes with target executor names
// ---------------------------------------------------------------------------

func TestP3FindCandidatesExcludesNamedProcesses(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	p.FunctionSpec.Conditions.ExecutorNames = []string{"some-exec"}
	assert.NoError(t, db.AddProcess(p))

	found, err := db.FindCandidates("c1", "cli", "", 10000, 100*1024*1024*1024, 100*1024*1024*1024, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, found, 0)
}

// ---------------------------------------------------------------------------
// Additional: MarkSuccessful returns timing values
// ---------------------------------------------------------------------------

func TestP3MarkSuccessfulReturnsTiming(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	waitTime, procTime, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, waitTime, float64(0))
	assert.GreaterOrEqual(t, procTime, float64(0))
}

// ---------------------------------------------------------------------------
// Additional: FindProcessesByColonyName for each state
// ---------------------------------------------------------------------------

func TestP3FindProcessesByColonyNameRunning(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))

	found, err := db.FindProcessesByColonyName("c1", 60, core.RUNNING)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestP3FindProcessesByColonyNameSuccessful(t *testing.T) {
	db := setupTestDB(t)
	addColonyForPhase3(t, db, "c1")

	p := createTestProcess("c1", "cli")
	assert.NoError(t, db.AddProcess(p))
	exec := addExecutorForPhase3(t, db, "c1", "e1", "cli")
	assert.NoError(t, db.Assign(exec.ID, p))
	_, _, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)

	found, err := db.FindProcessesByColonyName("c1", 60, core.SUCCESS)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
}
