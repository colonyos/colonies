package embedded

import (
	"os"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *EmbeddedDatabase {
	t.Helper()
	dir, err := os.MkdirTemp("", "embedded-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	db := CreateEmbeddedDatabase(dir)
	if err := db.Initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// Colony tests

func TestAddColony(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("test-id", "test-colony")
	err := db.AddColony(colony)
	assert.NoError(t, err)

	got, err := db.GetColonyByName("test-colony")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "test-id", got.ID)
	assert.Equal(t, "test-colony", got.Name)
}

func TestAddColonyNil(t *testing.T) {
	db := setupTestDB(t)
	err := db.AddColony(nil)
	assert.EqualError(t, err, "Colony is nil")
}

func TestAddColonyDuplicate(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("id1", "test-colony")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	colony2 := core.CreateColony("id2", "test-colony")
	err := db.AddColony(colony2)
	assert.EqualError(t, err, "Colony with name <test-colony> already exists")
}

func TestGetColonyByID(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("test-id", "test-colony")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetColonyByID("test-id")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "test-colony", got.Name)
}

func TestGetColonyByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	got, err := db.GetColonyByID("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetColonies(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddColony(core.CreateColony("id1", "colony1")); err != nil {
		t.Fatal(err)
	}
	if err := db.AddColony(core.CreateColony("id2", "colony2")); err != nil {
		t.Fatal(err)
	}

	colonies, err := db.GetColonies()
	assert.NoError(t, err)
	assert.Len(t, colonies, 2)
}

func TestRenameColony(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("test-id", "old-name")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	err := db.RenameColony("old-name", "new-name")
	assert.NoError(t, err)

	got, err := db.GetColonyByName("new-name")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "new-name", got.Name)

	old, err := db.GetColonyByName("old-name")
	assert.NoError(t, err)
	assert.Nil(t, old)

	// Should still be findable by ID
	byID, err := db.GetColonyByID("test-id")
	assert.NoError(t, err)
	assert.NotNil(t, byID)
	assert.Equal(t, "new-name", byID.Name)
}

func TestRemoveColonyByName(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("cid", "test-colony")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	user := core.CreateUser("test-colony", "uid", "testuser", "test@test.com", "123")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	executor := core.CreateExecutor("eid", "test-type", "test-executor", "test-colony", time.Now(), time.Now())
	if err := db.AddExecutor(executor); err != nil {
		t.Fatal(err)
	}

	location := core.CreateLocation("lid", "test-loc", "test-colony", "desc", 1.0, 2.0)
	if err := db.AddLocation(location); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveColonyByName("test-colony")
	assert.NoError(t, err)

	got, err := db.GetColonyByName("test-colony")
	assert.NoError(t, err)
	assert.Nil(t, got)

	users, err := db.GetUsersByColonyName("test-colony")
	assert.NoError(t, err)
	assert.Len(t, users, 0)

	locations, err := db.GetLocationsByColonyName("test-colony")
	assert.NoError(t, err)
	assert.Len(t, locations, 0)
}

func TestRemoveColonyNotExists(t *testing.T) {
	db := setupTestDB(t)
	err := db.RemoveColonyByName("nonexistent")
	assert.EqualError(t, err, "Colony does not exists")
}

func TestCountColonies(t *testing.T) {
	db := setupTestDB(t)

	count, err := db.CountColonies()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	if err := db.AddColony(core.CreateColony("id1", "colony1")); err != nil {
		t.Fatal(err)
	}
	if err := db.AddColony(core.CreateColony("id2", "colony2")); err != nil {
		t.Fatal(err)
	}

	count, err = db.CountColonies()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// User tests

func TestAddUser(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("colony1", "uid1", "user1", "user@test.com", "123")
	err := db.AddUser(user)
	assert.NoError(t, err)

	got, err := db.GetUserByName("colony1", "user1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "uid1", got.ID)
}

func TestAddUserNil(t *testing.T) {
	db := setupTestDB(t)
	err := db.AddUser(nil)
	assert.EqualError(t, err, "User is nil")
}

func TestAddUserDuplicate(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("colony1", "uid1", "user1", "a@b.com", "1")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	user2 := core.CreateUser("colony1", "uid2", "user1", "c@d.com", "2")
	err := db.AddUser(user2)
	assert.EqualError(t, err, "User with name <user1> already exists in Colony with name <colony1>")
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("colony1", "uid1", "user1", "a@b.com", "1")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetUserByID("colony1", "uid1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "user1", got.Name)
}

func TestGetUsersByColonyName(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddUser(core.CreateUser("c1", "u1", "user1", "", "")); err != nil {
		t.Fatal(err)
	}
	if err := db.AddUser(core.CreateUser("c1", "u2", "user2", "", "")); err != nil {
		t.Fatal(err)
	}
	if err := db.AddUser(core.CreateUser("c2", "u3", "user3", "", "")); err != nil {
		t.Fatal(err)
	}

	users, err := db.GetUsersByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestRemoveUserByID(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("c1", "uid1", "user1", "", "")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveUserByID("c1", "uid1")
	assert.NoError(t, err)

	got, err := db.GetUserByName("c1", "user1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveUserByName(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("c1", "uid1", "user1", "", "")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveUserByName("c1", "user1")
	assert.NoError(t, err)

	got, err := db.GetUserByName("c1", "user1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveUsersByColonyName(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddUser(core.CreateUser("c1", "u1", "user1", "", "")); err != nil {
		t.Fatal(err)
	}
	if err := db.AddUser(core.CreateUser("c1", "u2", "user2", "", "")); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveUsersByColonyName("c1")
	assert.NoError(t, err)

	users, err := db.GetUsersByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

// Security tests

func TestSetGetServerID(t *testing.T) {
	db := setupTestDB(t)

	err := db.SetServerID("", "server1")
	assert.NoError(t, err)

	id, err := db.GetServerID()
	assert.NoError(t, err)
	assert.Equal(t, "server1", id)

	err = db.SetServerID("server1", "server2")
	assert.NoError(t, err)

	id, err = db.GetServerID()
	assert.NoError(t, err)
	assert.Equal(t, "server2", id)
}

func TestChangeColonyID(t *testing.T) {
	db := setupTestDB(t)

	colony := core.CreateColony("old-id", "test-colony")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	err := db.ChangeColonyID("test-colony", "old-id", "new-id")
	assert.NoError(t, err)

	got, err := db.GetColonyByID("new-id")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "new-id", got.ID)

	old, err := db.GetColonyByID("old-id")
	assert.NoError(t, err)
	assert.Nil(t, old)
}

func TestChangeUserID(t *testing.T) {
	db := setupTestDB(t)

	user := core.CreateUser("c1", "old-uid", "user1", "", "")
	if err := db.AddUser(user); err != nil {
		t.Fatal(err)
	}

	err := db.ChangeUserID("c1", "old-uid", "new-uid")
	assert.NoError(t, err)

	got, err := db.GetUserByID("c1", "new-uid")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "new-uid", got.ID)
}

func TestChangeExecutorID(t *testing.T) {
	db := setupTestDB(t)

	executor := core.CreateExecutor("old-eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(executor); err != nil {
		t.Fatal(err)
	}

	err := db.ChangeExecutorID("c1", "old-eid", "new-eid")
	assert.NoError(t, err)

	got, err := db.GetExecutorByID("new-eid")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "new-eid", got.ID)

	old, err := db.GetExecutorByID("old-eid")
	assert.NoError(t, err)
	assert.Nil(t, old)
}

// Location tests

func TestAddLocation(t *testing.T) {
	db := setupTestDB(t)

	loc := core.CreateLocation("lid", "loc1", "c1", "desc", 1.0, 2.0)
	err := db.AddLocation(loc)
	assert.NoError(t, err)

	got, err := db.GetLocationByID("lid")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "loc1", got.Name)
}

func TestAddLocationNil(t *testing.T) {
	db := setupTestDB(t)
	err := db.AddLocation(nil)
	assert.EqualError(t, err, "Location is nil")
}

func TestAddLocationDuplicate(t *testing.T) {
	db := setupTestDB(t)

	loc := core.CreateLocation("lid1", "loc1", "c1", "desc", 1.0, 2.0)
	if err := db.AddLocation(loc); err != nil {
		t.Fatal(err)
	}

	loc2 := core.CreateLocation("lid2", "loc1", "c1", "desc2", 3.0, 4.0)
	err := db.AddLocation(loc2)
	assert.EqualError(t, err, "Location with name <loc1> already exists in Colony with name <c1>")
}

func TestGetLocationByName(t *testing.T) {
	db := setupTestDB(t)

	loc := core.CreateLocation("lid", "loc1", "c1", "desc", 1.0, 2.0)
	if err := db.AddLocation(loc); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetLocationByName("c1", "loc1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "lid", got.ID)
}

func TestGetLocationsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddLocation(core.CreateLocation("l1", "loc1", "c1", "", 0, 0)); err != nil {
		t.Fatal(err)
	}
	if err := db.AddLocation(core.CreateLocation("l2", "loc2", "c1", "", 0, 0)); err != nil {
		t.Fatal(err)
	}
	if err := db.AddLocation(core.CreateLocation("l3", "loc3", "c2", "", 0, 0)); err != nil {
		t.Fatal(err)
	}

	locs, err := db.GetLocationsByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, locs, 2)
}

func TestRemoveLocationByID(t *testing.T) {
	db := setupTestDB(t)

	loc := core.CreateLocation("lid", "loc1", "c1", "", 0, 0)
	if err := db.AddLocation(loc); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveLocationByID("lid")
	assert.NoError(t, err)

	got, err := db.GetLocationByID("lid")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveLocationByName(t *testing.T) {
	db := setupTestDB(t)

	loc := core.CreateLocation("lid", "loc1", "c1", "", 0, 0)
	if err := db.AddLocation(loc); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveLocationByName("c1", "loc1")
	assert.NoError(t, err)

	got, err := db.GetLocationByName("c1", "loc1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveLocationsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	if err := db.AddLocation(core.CreateLocation("l1", "loc1", "c1", "", 0, 0)); err != nil {
		t.Fatal(err)
	}
	if err := db.AddLocation(core.CreateLocation("l2", "loc2", "c1", "", 0, 0)); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveLocationsByColonyName("c1")
	assert.NoError(t, err)

	locs, err := db.GetLocationsByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, locs, 0)
}

// Executor tests

func TestAddExecutor(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	err := db.AddExecutor(e)
	assert.NoError(t, err)

	got, err := db.GetExecutorByID("eid")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "exec1", got.Name)
}

func TestAddExecutorNil(t *testing.T) {
	db := setupTestDB(t)
	err := db.AddExecutor(nil)
	assert.EqualError(t, err, "Executor is nil")
}

func TestAddExecutorDuplicate(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	e2 := core.CreateExecutor("eid2", "type1", "exec1", "c1", time.Now(), time.Now())
	err := db.AddExecutor(e2)
	assert.EqualError(t, err, "Executor with name <exec1> already exists in Colony with name <c1>")
}

func TestAddExecutorReactivateUnregistered(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	// Mark as unregistered
	if err := db.RemoveExecutorByName("c1", "exec1"); err != nil {
		t.Fatal(err)
	}

	// Reactivate
	e2 := core.CreateExecutor("eid2", "type1", "exec1", "c1", time.Now(), time.Now())
	err := db.AddExecutor(e2)
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "exec1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, core.PENDING, got.State)
	assert.Equal(t, "eid2", got.ID)
}

func TestGetExecutors(t *testing.T) {
	db := setupTestDB(t)

	e1 := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	e2 := core.CreateExecutor("eid2", "type1", "exec2", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddExecutor(e2); err != nil {
		t.Fatal(err)
	}

	executors, err := db.GetExecutors()
	assert.NoError(t, err)
	assert.Len(t, executors, 2)
}

func TestGetExecutorsExcludesUnregistered(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	if err := db.RemoveExecutorByName("c1", "exec1"); err != nil {
		t.Fatal(err)
	}

	executors, err := db.GetExecutors()
	assert.NoError(t, err)
	assert.Len(t, executors, 0)
}

func TestGetExecutorByName(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetExecutorByName("c1", "exec1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "eid", got.ID)
}

func TestApproveRejectExecutor(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, core.PENDING, e.State)

	err := db.ApproveExecutor(e)
	assert.NoError(t, err)

	got, err := db.GetExecutorByID("eid")
	assert.NoError(t, err)
	assert.Equal(t, core.APPROVED, got.State)

	e2 := core.CreateExecutor("eid2", "type1", "exec2", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e2); err != nil {
		t.Fatal(err)
	}

	err = db.RejectExecutor(e2)
	assert.NoError(t, err)

	got2, err := db.GetExecutorByID("eid2")
	assert.NoError(t, err)
	assert.Equal(t, core.REJECTED, got2.State)
}

func TestSetAllocations(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	allocs := core.Allocations{
		Projects: map[string]core.Project{
			"proj1": {AllocatedCPU: 100},
		},
	}

	err := db.SetAllocations("c1", "exec1", allocs)
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "exec1")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), got.Allocations.Projects["proj1"].AllocatedCPU)
}

func TestCountExecutors(t *testing.T) {
	db := setupTestDB(t)

	e1 := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	e2 := core.CreateExecutor("eid2", "type1", "exec2", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddExecutor(e2); err != nil {
		t.Fatal(err)
	}

	count, err := db.CountExecutors()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestCountExecutorsByColonyNameAndState(t *testing.T) {
	db := setupTestDB(t)

	e1 := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	e2 := core.CreateExecutor("eid2", "type1", "exec2", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddExecutor(e2); err != nil {
		t.Fatal(err)
	}
	if err := db.ApproveExecutor(e1); err != nil {
		t.Fatal(err)
	}

	count, err := db.CountExecutorsByColonyNameAndState("c1", core.APPROVED)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = db.CountExecutorsByColonyNameAndState("c1", core.PENDING)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRemoveExecutorByName(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveExecutorByName("c1", "exec1")
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "exec1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, core.UNREGISTERED, got.State)
}

func TestRemoveExecutorNotExists(t *testing.T) {
	db := setupTestDB(t)
	err := db.RemoveExecutorByName("c1", "nonexistent")
	assert.EqualError(t, err, "Executor <nonexistent> does not exists")
}

func TestRemoveExecutorsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	e1 := core.CreateExecutor("eid1", "type1", "exec1", "c1", time.Now(), time.Now())
	e2 := core.CreateExecutor("eid2", "type1", "exec2", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddExecutor(e2); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveExecutorsByColonyName("c1")
	assert.NoError(t, err)

	executors, err := db.GetExecutorsByColonyName("c1")
	assert.NoError(t, err)
	assert.Len(t, executors, 0)
}

func TestUpdateExecutorCapabilities(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	caps := core.Capabilities{
		Hardware: []core.Hardware{{Model: "gpu-model", Cores: 8}},
		Software: []core.Software{{Name: "cuda", Version: "12.0"}},
	}

	err := db.UpdateExecutorCapabilities("c1", "exec1", caps)
	assert.NoError(t, err)

	got, err := db.GetExecutorByName("c1", "exec1")
	assert.NoError(t, err)
	assert.Len(t, got.Capabilities.Hardware, 1)
	assert.Equal(t, "gpu-model", got.Capabilities.Hardware[0].Model)
}

func TestGetExecutorsByBlueprintID(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Now())
	e.BlueprintID = "bp1"
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	executors, err := db.GetExecutorsByBlueprintID("bp1")
	assert.NoError(t, err)
	assert.Len(t, executors, 1)
	assert.Equal(t, "exec1", executors[0].Name)
}

// Function tests

func TestAddFunction(t *testing.T) {
	db := setupTestDB(t)

	f := core.CreateFunction("fid", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	err := db.AddFunction(f)
	assert.NoError(t, err)

	got, err := db.GetFunctionByID("fid")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "func1", got.FuncName)
}

func TestGetFunctionsByExecutorName(t *testing.T) {
	db := setupTestDB(t)

	f1 := core.CreateFunction("fid1", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	f2 := core.CreateFunction("fid2", "exec1", "type1", "c1", "func2", 0, 0, 0, 0, 0, 0, 0)
	f3 := core.CreateFunction("fid3", "exec2", "type1", "c1", "func3", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddFunction(f2); err != nil {
		t.Fatal(err)
	}
	if err := db.AddFunction(f3); err != nil {
		t.Fatal(err)
	}

	funcs, err := db.GetFunctionsByExecutorName("c1", "exec1")
	assert.NoError(t, err)
	assert.Len(t, funcs, 2)
}

func TestGetFunctionsByExecutorAndName(t *testing.T) {
	db := setupTestDB(t)

	f := core.CreateFunction("fid1", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetFunctionsByExecutorAndName("c1", "exec1", "func1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "fid1", got.FunctionID)

	notFound, err := db.GetFunctionsByExecutorAndName("c1", "exec1", "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestUpdateFunctionStats(t *testing.T) {
	db := setupTestDB(t)

	f := core.CreateFunction("fid1", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f); err != nil {
		t.Fatal(err)
	}

	err := db.UpdateFunctionStats("c1", "exec1", "func1", 10, 1.0, 5.0, 0.5, 3.0, 2.0, 1.5)
	assert.NoError(t, err)

	got, err := db.GetFunctionByID("fid1")
	assert.NoError(t, err)
	assert.Equal(t, 10, got.Counter)
	assert.Equal(t, 1.0, got.MinWaitTime)
	assert.Equal(t, 5.0, got.MaxWaitTime)
}

func TestRemoveFunctionByID(t *testing.T) {
	db := setupTestDB(t)

	f := core.CreateFunction("fid", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveFunctionByID("fid")
	assert.NoError(t, err)

	got, err := db.GetFunctionByID("fid")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveFunctionsByExecutorName(t *testing.T) {
	db := setupTestDB(t)

	f1 := core.CreateFunction("fid1", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	f2 := core.CreateFunction("fid2", "exec1", "type1", "c1", "func2", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddFunction(f2); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveFunctionsByExecutorName("c1", "exec1")
	assert.NoError(t, err)

	funcs, err := db.GetFunctionsByExecutorName("c1", "exec1")
	assert.NoError(t, err)
	assert.Len(t, funcs, 0)
}

func TestRemoveFunctions(t *testing.T) {
	db := setupTestDB(t)

	f1 := core.CreateFunction("fid1", "exec1", "type1", "c1", "func1", 0, 0, 0, 0, 0, 0, 0)
	f2 := core.CreateFunction("fid2", "exec2", "type1", "c2", "func2", 0, 0, 0, 0, 0, 0, 0)
	if err := db.AddFunction(f1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddFunction(f2); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveFunctions()
	assert.NoError(t, err)

	got, err := db.GetFunctionByID("fid1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

// Generator tests

func TestAddGenerator(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	err := db.AddGenerator(g)
	assert.NoError(t, err)

	got, err := db.GetGeneratorByID("gid1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "gen1", got.Name)
}

func TestAddGeneratorDuplicate(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	if err := db.AddGenerator(g); err != nil {
		t.Fatal(err)
	}

	g2 := core.CreateGenerator("c1", "gen1", "{}", 5, 30)
	g2.ID = "gid2"
	err := db.AddGenerator(g2)
	assert.EqualError(t, err, "Generator with name <gen1> in Colony <c1> already exists")
}

func TestGetGeneratorByName(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	if err := db.AddGenerator(g); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetGeneratorByName("c1", "gen1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "gid1", got.ID)
}

func TestRemoveGeneratorByID(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	if err := db.AddGenerator(g); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveGeneratorByID("gid1")
	assert.NoError(t, err)

	got, err := db.GetGeneratorByID("gid1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestGeneratorArgs(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	if err := db.AddGenerator(g); err != nil {
		t.Fatal(err)
	}

	arg1 := &core.GeneratorArg{ID: "aid1", GeneratorID: "gid1", ColonyName: "c1", Arg: "arg1"}
	arg2 := &core.GeneratorArg{ID: "aid2", GeneratorID: "gid1", ColonyName: "c1", Arg: "arg2"}
	if err := db.AddGeneratorArg(arg1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddGeneratorArg(arg2); err != nil {
		t.Fatal(err)
	}

	count, err := db.CountGeneratorArgs("gid1")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	args, err := db.GetGeneratorArgs("gid1", 10)
	assert.NoError(t, err)
	assert.Len(t, args, 2)

	err = db.RemoveGeneratorArgByID("aid1")
	assert.NoError(t, err)

	count, err = db.CountGeneratorArgs("gid1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRemoveGeneratorCascadesArgs(t *testing.T) {
	db := setupTestDB(t)

	g := core.CreateGenerator("c1", "gen1", "{}", 10, 60)
	g.ID = "gid1"
	if err := db.AddGenerator(g); err != nil {
		t.Fatal(err)
	}

	arg := &core.GeneratorArg{ID: "aid1", GeneratorID: "gid1", ColonyName: "c1", Arg: "arg1"}
	if err := db.AddGeneratorArg(arg); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveGeneratorByID("gid1")
	assert.NoError(t, err)

	count, err := db.CountGeneratorArgs("gid1")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// Cron tests

func TestAddCron(t *testing.T) {
	db := setupTestDB(t)

	cron := &core.Cron{
		ID:             "cid1",
		ColonyName:     "c1",
		Name:           "cron1",
		CronExpression: "* * * * *",
		WorkflowSpec:   "{}",
	}
	err := db.AddCron(cron)
	assert.NoError(t, err)

	got, err := db.GetCronByID("cid1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "cron1", got.Name)
}

func TestAddCronDuplicate(t *testing.T) {
	db := setupTestDB(t)

	cron := &core.Cron{ID: "cid1", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	if err := db.AddCron(cron); err != nil {
		t.Fatal(err)
	}

	cron2 := &core.Cron{ID: "cid2", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	err := db.AddCron(cron2)
	assert.EqualError(t, err, "Cron with name <cron1> in Colony <c1> already exists")
}

func TestGetCronByName(t *testing.T) {
	db := setupTestDB(t)

	cron := &core.Cron{ID: "cid1", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	if err := db.AddCron(cron); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetCronByName("c1", "cron1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "cid1", got.ID)
}

func TestUpdateCron(t *testing.T) {
	db := setupTestDB(t)

	cron := &core.Cron{ID: "cid1", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	if err := db.AddCron(cron); err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	err := db.UpdateCron("cid1", now.Add(time.Hour), now, "pg-id")
	assert.NoError(t, err)

	got, err := db.GetCronByID("cid1")
	assert.NoError(t, err)
	assert.Equal(t, "pg-id", got.PrevProcessGraphID)
}

func TestRemoveCronByID(t *testing.T) {
	db := setupTestDB(t)

	cron := &core.Cron{ID: "cid1", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	if err := db.AddCron(cron); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveCronByID("cid1")
	assert.NoError(t, err)

	got, err := db.GetCronByID("cid1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveAllCronsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	cron1 := &core.Cron{ID: "cid1", ColonyName: "c1", Name: "cron1", WorkflowSpec: "{}"}
	cron2 := &core.Cron{ID: "cid2", ColonyName: "c1", Name: "cron2", WorkflowSpec: "{}"}
	if err := db.AddCron(cron1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddCron(cron2); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveAllCronsByColonyName("c1")
	assert.NoError(t, err)

	crons, err := db.FindCronsByColonyName("c1", 100)
	assert.NoError(t, err)
	assert.Len(t, crons, 0)
}

// Blueprint tests

func TestAddBlueprint(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	err := db.AddBlueprint(bp)
	assert.NoError(t, err)

	got, err := db.GetBlueprintByID(bp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "bp1", got.Metadata.Name)
}

func TestAddBlueprintDuplicate(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	bp2 := core.CreateBlueprint("TestKind", "bp1", "ns1")
	err := db.AddBlueprint(bp2)
	assert.EqualError(t, err, "Blueprint with name <bp1> in namespace <ns1> already exists")
}

func TestGetBlueprintByName(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetBlueprintByName("ns1", "bp1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, bp.ID, got.ID)
}

func TestGetBlueprintsByKind(t *testing.T) {
	db := setupTestDB(t)

	bp1 := core.CreateBlueprint("KindA", "bp1", "ns1")
	bp2 := core.CreateBlueprint("KindA", "bp2", "ns1")
	bp3 := core.CreateBlueprint("KindB", "bp3", "ns1")
	if err := db.AddBlueprint(bp1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddBlueprint(bp2); err != nil {
		t.Fatal(err)
	}
	if err := db.AddBlueprint(bp3); err != nil {
		t.Fatal(err)
	}

	bps, err := db.GetBlueprintsByKind("KindA")
	assert.NoError(t, err)
	assert.Len(t, bps, 2)
}

func TestUpdateBlueprint(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	bp.SetSpec("replicas", 3)
	err := db.UpdateBlueprint(bp)
	assert.NoError(t, err)

	got, err := db.GetBlueprintByID(bp.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, got.Spec["replicas"])
}

func TestUpdateBlueprintStatus(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	status := map[string]interface{}{"ready": true, "replicas": float64(3)}
	err := db.UpdateBlueprintStatus(bp.ID, status)
	assert.NoError(t, err)

	got, err := db.GetBlueprintByID(bp.ID)
	assert.NoError(t, err)
	assert.Equal(t, true, got.Status["ready"])
}

func TestRemoveBlueprintByID(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	err := db.RemoveBlueprintByID(bp.ID)
	assert.NoError(t, err)

	got, err := db.GetBlueprintByID(bp.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestCountBlueprints(t *testing.T) {
	db := setupTestDB(t)

	bp1 := core.CreateBlueprint("TestKind", "bp1", "ns1")
	bp2 := core.CreateBlueprint("TestKind", "bp2", "ns1")
	if err := db.AddBlueprint(bp1); err != nil {
		t.Fatal(err)
	}
	if err := db.AddBlueprint(bp2); err != nil {
		t.Fatal(err)
	}

	count, err := db.CountBlueprints()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = db.CountBlueprintsByNamespace("ns1")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// BlueprintDefinition tests

func TestAddBlueprintDefinition(t *testing.T) {
	db := setupTestDB(t)

	sd := core.CreateBlueprintDefinition("testsd", "group1", "v1", "TestKind", "testkinds", "Namespaced", "reconciler", "reconcile")
	sd.Metadata.ColonyName = "ns1"
	err := db.AddBlueprintDefinition(sd)
	assert.NoError(t, err)

	got, err := db.GetBlueprintDefinitionByID(sd.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "testsd", got.Metadata.Name)
}

func TestGetBlueprintDefinitionByKind(t *testing.T) {
	db := setupTestDB(t)

	sd := core.CreateBlueprintDefinition("testsd", "group1", "v1", "TestKind", "testkinds", "Namespaced", "reconciler", "reconcile")
	sd.Metadata.ColonyName = "ns1"
	if err := db.AddBlueprintDefinition(sd); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetBlueprintDefinitionByKind("TestKind")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "testsd", got.Metadata.Name)
}

// BlueprintHistory tests

func TestBlueprintHistory(t *testing.T) {
	db := setupTestDB(t)

	bp := core.CreateBlueprint("TestKind", "bp1", "ns1")
	if err := db.AddBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	history := core.CreateBlueprintHistory(bp, "test-user", "create")
	err := db.AddBlueprintHistory(history)
	assert.NoError(t, err)

	histories, err := db.GetBlueprintHistory(bp.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.Equal(t, "create", histories[0].ChangeType)

	got, err := db.GetBlueprintHistoryByGeneration(bp.ID, history.Generation)
	assert.NoError(t, err)
	assert.NotNil(t, got)

	err = db.RemoveBlueprintHistory(bp.ID)
	assert.NoError(t, err)

	histories, err = db.GetBlueprintHistory(bp.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, histories, 0)
}

// Drop/Initialize test

func TestDropAndReinitialize(t *testing.T) {
	dir, err := os.MkdirTemp("", "embedded-drop-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db := CreateEmbeddedDatabase(dir)
	if err := db.Initialize(); err != nil {
		t.Fatal(err)
	}

	colony := core.CreateColony("cid", "test-colony")
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}

	err = db.Drop()
	assert.NoError(t, err)

	// Reinitialize
	db2 := CreateEmbeddedDatabase(dir)
	if err := db2.Initialize(); err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	count, err := db2.CountColonies()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// MarkAlive test

func TestMarkAlive(t *testing.T) {
	db := setupTestDB(t)

	e := core.CreateExecutor("eid", "type1", "exec1", "c1", time.Now(), time.Time{})
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}

	before, _ := db.GetExecutorByID("eid")
	assert.True(t, before.LastHeardFromTime.IsZero() || before.LastHeardFromTime.Before(time.Now()))

	err := db.MarkAlive(e)
	assert.NoError(t, err)

	after, _ := db.GetExecutorByID("eid")
	assert.False(t, after.LastHeardFromTime.IsZero())
}

// Interface compliance test
func TestInterfaceCompliance(t *testing.T) {
	db := setupTestDB(t)
	// This line verifies at runtime that EmbeddedDatabase satisfies the interface
	var _ interface{ Close() } = db
}

// Process tests

func createTestProcess(colonyName string, executorType string) *core.Process {
	env := make(map[string]string)
	funcSpec := core.CreateFunctionSpec(
		"", "testfunc", []interface{}{"arg1"}, map[string]interface{}{"key": "val"},
		colonyName, []string{}, executorType,
		0, 0, 0, env, []string{}, 0, "",
	)
	funcSpec.Conditions.CPU = "1000m"
	funcSpec.Conditions.Memory = "1Gi"
	funcSpec.Conditions.Storage = "10Gi"
	funcSpec.Conditions.Nodes = 1
	funcSpec.Conditions.Processes = 1
	funcSpec.Conditions.ProcessesPerNode = 1
	return core.CreateProcess(funcSpec)
}

func TestAddProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	err := db.AddProcess(p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.SubmissionTime.IsZero())
}

func TestGetProcessByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	got, err := db.GetProcessByID("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)

	all, err := db.GetProcesses()
	assert.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestAssignProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.Assign("executor1", p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.RUNNING, got.State)
	assert.True(t, got.IsAssigned)
	assert.Equal(t, "executor1", got.AssignedExecutorID)
}

func TestAssignAlreadyAssigned(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	err := db.Assign("executor2", p)
	assert.EqualError(t, err, "Process already assigned")
}

func TestAssignCancelledProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.MarkCancelled(p.ID)

	err := db.Assign("executor1", p)
	assert.EqualError(t, err, "Cannot assign cancelled process")
}

func TestUnassignProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	err := db.Unassign(p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.IsAssigned)
	assert.Equal(t, 1, got.Retries)
}

func TestMarkSuccessful(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	waitTime, procTime, err := db.MarkSuccessful(p.ID)
	assert.NoError(t, err)
	assert.True(t, waitTime >= 0)
	assert.True(t, procTime >= 0)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.SUCCESS, got.State)
}

func TestMarkSuccessfulFailedProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)
	db.MarkFailed(p.ID, []string{"error"})

	_, _, err := db.MarkSuccessful(p.ID)
	assert.EqualError(t, err, "Tried to set failed process as successful")
}

func TestMarkSuccessfulCancelledProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.MarkCancelled(p.ID)

	_, _, err := db.MarkSuccessful(p.ID)
	assert.EqualError(t, err, "Tried to set cancelled process as successful")
}

func TestMarkSuccessfulWaitingProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	_, _, err := db.MarkSuccessful(p.ID)
	assert.EqualError(t, err, "Tried to set waiting process as successful without being running")
}

func TestMarkFailed(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	err := db.MarkFailed(p.ID, []string{"something went wrong"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.FAILED, got.State)
	assert.Contains(t, got.Errors, "something went wrong")
}

func TestMarkFailedErrors(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	db.MarkFailed(p.ID, []string{"err"})
	err := db.MarkFailed(p.ID, []string{"err2"})
	assert.EqualError(t, err, "Tried to set failed process as failed")
}

func TestMarkCancelled(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.MarkCancelled(p.ID)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.CANCELLED, got.State)
}

func TestMarkCancelledAlreadyCancelled(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.MarkCancelled(p.ID)

	err := db.MarkCancelled(p.ID)
	assert.EqualError(t, err, "Tried to cancel already cancelled process")
}

func TestMarkCancelledSuccessful(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("e1", p)
	db.MarkSuccessful(p.ID)

	err := db.MarkCancelled(p.ID)
	assert.EqualError(t, err, "Tried to cancel successful process")
}

func TestResetProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("executor1", p)

	err := db.ResetProcess(p)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.WAITING, got.State)
	assert.False(t, got.IsAssigned)
	assert.Equal(t, "", got.AssignedExecutorID)
}

func TestSetInputOutput(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.SetInput(p.ID, []interface{}{"in1", "in2"})
	assert.NoError(t, err)
	err = db.SetOutput(p.ID, []interface{}{"out1"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Len(t, got.Input, 2)
	assert.Len(t, got.Output, 1)
}

func TestSetErrors(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.SetErrors(p.ID, []string{"err1", "err2"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Len(t, got.Errors, 2)
}

func TestSetParentsChildren(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.SetParents(p.ID, []string{"parent1"})
	assert.NoError(t, err)
	err = db.SetChildren(p.ID, []string{"child1", "child2"})
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"parent1"}, got.Parents)
	assert.Equal(t, []string{"child1", "child2"}, got.Children)
}

func TestSetWaitForParents(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.SetWaitForParents(p.ID, true)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.True(t, got.WaitForParents)
}

func TestFindWaitingProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "cli")
	p3 := createTestProcess("c2", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)
	db.AddProcess(p3)

	waiting, err := db.FindWaitingProcesses("c1", "", "", "", 10)
	assert.NoError(t, err)
	assert.Len(t, waiting, 2)
}

func TestFindRunningProcesses(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)
	db.Assign("e1", p)

	running, err := db.FindRunningProcesses("c1", "", "", "", 10)
	assert.NoError(t, err)
	assert.Len(t, running, 1)
}

func TestFindCandidates(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	candidates, err := db.FindCandidates("c1", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, candidates, 1)
}

func TestFindCandidatesNoMatch(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	// CPU too low
	candidates, err := db.FindCandidates("c1", "cli", "", 0, 10737418240, 107374182400, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, candidates, 0)
}

func TestFindCandidatesByName(t *testing.T) {
	db := setupTestDB(t)

	env := make(map[string]string)
	funcSpec := core.CreateFunctionSpec(
		"", "testfunc", nil, nil, "c1",
		[]string{"myexecutor"}, "cli",
		0, 0, 0, env, nil, 0, "",
	)
	funcSpec.Conditions.CPU = "1000m"
	funcSpec.Conditions.Memory = "1Gi"
	funcSpec.Conditions.Storage = "10Gi"
	funcSpec.Conditions.Nodes = 1
	funcSpec.Conditions.Processes = 1
	funcSpec.Conditions.ProcessesPerNode = 1
	p := core.CreateProcess(funcSpec)
	db.AddProcess(p)

	candidates, err := db.FindCandidatesByName("c1", "myexecutor", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, candidates, 1)

	// Wrong executor name
	candidates, err = db.FindCandidatesByName("c1", "wrongexecutor", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 10)
	assert.NoError(t, err)
	assert.Len(t, candidates, 0)
}

func TestSelectAndAssign(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	assigned, err := db.SelectAndAssign("c1", "eid1", "exec1", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 1)
	assert.NoError(t, err)
	assert.NotNil(t, assigned)
	assert.Equal(t, core.RUNNING, assigned.State)
	assert.Equal(t, "eid1", assigned.AssignedExecutorID)
}

func TestSelectAndAssignNoCandidate(t *testing.T) {
	db := setupTestDB(t)

	assigned, err := db.SelectAndAssign("c1", "eid1", "exec1", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 1)
	assert.NoError(t, err)
	assert.Nil(t, assigned)
}

func TestSelectAndAssignByName(t *testing.T) {
	db := setupTestDB(t)

	env := make(map[string]string)
	funcSpec := core.CreateFunctionSpec(
		"", "testfunc", nil, nil, "c1",
		[]string{"exec1"}, "cli",
		0, 0, 0, env, nil, 0, "",
	)
	funcSpec.Conditions.CPU = "1000m"
	funcSpec.Conditions.Memory = "1Gi"
	funcSpec.Conditions.Storage = "10Gi"
	funcSpec.Conditions.Nodes = 1
	funcSpec.Conditions.Processes = 1
	funcSpec.Conditions.ProcessesPerNode = 1
	p := core.CreateProcess(funcSpec)
	db.AddProcess(p)

	assigned, err := db.SelectAndAssign("c1", "eid1", "exec1", "cli", "", 10000, 10737418240, 107374182400, 10, 10, 10, 1)
	assert.NoError(t, err)
	assert.NotNil(t, assigned)
}

func TestRemoveProcessByID(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.RemoveProcessByID(p.ID)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveAllProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)

	err := db.RemoveAllProcesses()
	assert.NoError(t, err)

	count, err := db.CountProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCountProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)

	count, err := db.CountProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	wc, err := db.CountWaitingProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 2, wc)

	db.Assign("e1", p1)

	rc, err := db.CountRunningProcesses()
	assert.NoError(t, err)
	assert.Equal(t, 1, rc)
}

func TestCountProcessesByColonyName(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c2", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)

	c, err := db.CountWaitingProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func TestRemoveWaitingProcessesByColony(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c1", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)
	db.Assign("e1", p2)

	err := db.RemoveAllWaitingProcessesByColonyName("c1")
	assert.NoError(t, err)

	wc, err := db.CountWaitingProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 0, wc)

	// Running process should still exist
	rc, err := db.CountRunningProcessesByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, rc)
}

func TestSetProcessState(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	err := db.SetProcessState(p.ID, core.RUNNING)
	assert.NoError(t, err)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.RUNNING, got.State)
}

func TestFindAllRunningProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c2", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)
	db.Assign("e1", p1)
	db.Assign("e2", p2)

	running, err := db.FindAllRunningProcesses()
	assert.NoError(t, err)
	assert.Len(t, running, 2)
}

func TestFindAllWaitingProcesses(t *testing.T) {
	db := setupTestDB(t)

	p1 := createTestProcess("c1", "cli")
	p2 := createTestProcess("c2", "cli")
	db.AddProcess(p1)
	db.AddProcess(p2)

	waiting, err := db.FindAllWaitingProcesses()
	assert.NoError(t, err)
	assert.Len(t, waiting, 2)
}

func TestProcessWithEnvAttributes(t *testing.T) {
	db := setupTestDB(t)

	env := map[string]string{"KEY1": "val1", "KEY2": "val2"}
	funcSpec := core.CreateFunctionSpec(
		"", "testfunc", nil, nil, "c1", nil, "cli",
		0, 0, 0, env, nil, 0, "",
	)
	p := core.CreateProcess(funcSpec)
	db.AddProcess(p)

	got, err := db.GetProcessByID(p.ID)
	assert.NoError(t, err)
	assert.Equal(t, "val1", got.FunctionSpec.Env["KEY1"])
	assert.Equal(t, "val2", got.FunctionSpec.Env["KEY2"])
}

// Attribute tests

func TestAddAttribute(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "mykey", "myval")
	err := db.AddAttribute(attr)
	assert.NoError(t, err)

	got, err := db.GetAttributeByID(attr.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mykey", got.Key)
	assert.Equal(t, "myval", got.Value)
}

func TestGetAttributeByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.GetAttributeByID("nonexistent")
	assert.EqualError(t, err, "Attribute does not exists")
}

func TestGetAttribute(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "mykey", "myval")
	db.AddAttribute(attr)

	got, err := db.GetAttribute("target1", "mykey", core.IN)
	assert.NoError(t, err)
	assert.Equal(t, "myval", got.Value)

	_, err = db.GetAttribute("target1", "nonexistent", core.IN)
	assert.EqualError(t, err, "Attribute does not exists")
}

func TestGetAttributes(t *testing.T) {
	db := setupTestDB(t)

	a1 := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "val1")
	a2 := core.CreateAttribute("target1", "c1", "", core.OUT, "key2", "val2")
	a3 := core.CreateAttribute("target2", "c1", "", core.IN, "key3", "val3")
	db.AddAttribute(a1)
	db.AddAttribute(a2)
	db.AddAttribute(a3)

	attrs, err := db.GetAttributes("target1")
	assert.NoError(t, err)
	assert.Len(t, attrs, 2)
}

func TestGetAttributesByType(t *testing.T) {
	db := setupTestDB(t)

	a1 := core.CreateAttribute("target1", "c1", "", core.IN, "key1", "val1")
	a2 := core.CreateAttribute("target1", "c1", "", core.OUT, "key2", "val2")
	db.AddAttribute(a1)
	db.AddAttribute(a2)

	attrs, err := db.GetAttributesByType("target1", core.IN)
	assert.NoError(t, err)
	assert.Len(t, attrs, 1)
	assert.Equal(t, "key1", attrs[0].Key)
}

func TestUpdateAttribute(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "mykey", "oldval")
	db.AddAttribute(attr)

	attr.Value = "newval"
	err := db.UpdateAttribute(attr)
	assert.NoError(t, err)

	got, err := db.GetAttributeByID(attr.ID)
	assert.NoError(t, err)
	assert.Equal(t, "newval", got.Value)
}

func TestRemoveAttributeByID(t *testing.T) {
	db := setupTestDB(t)

	attr := core.CreateAttribute("target1", "c1", "", core.IN, "mykey", "myval")
	db.AddAttribute(attr)

	err := db.RemoveAttributeByID(attr.ID)
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(attr.ID)
	assert.Error(t, err)
}

func TestAttributeStateFollowsProcess(t *testing.T) {
	db := setupTestDB(t)

	p := createTestProcess("c1", "cli")
	db.AddProcess(p)

	// Add an attribute for this process
	attr := core.CreateAttribute(p.ID, "c1", "", core.IN, "key1", "val1")
	db.AddAttribute(attr)

	// Assign the process (changes state to RUNNING)
	db.Assign("e1", p)

	got, err := db.GetAttributeByID(attr.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.RUNNING, got.State)
}

func TestRemoveAllAttributes(t *testing.T) {
	db := setupTestDB(t)

	a1 := core.CreateAttribute("t1", "c1", "", core.IN, "k1", "v1")
	a2 := core.CreateAttribute("t2", "c1", "", core.IN, "k2", "v2")
	db.AddAttribute(a1)
	db.AddAttribute(a2)

	err := db.RemoveAllAttributes()
	assert.NoError(t, err)

	_, err = db.GetAttributeByID(a1.ID)
	assert.Error(t, err)
}

// ProcessGraph tests

func TestAddProcessGraph(t *testing.T) {
	db := setupTestDB(t)

	graph, err := core.CreateProcessGraph("c1")
	assert.NoError(t, err)

	err = db.AddProcessGraph(graph)
	assert.NoError(t, err)

	got, err := db.GetProcessGraphByID(graph.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "c1", got.ColonyName)
	assert.Equal(t, core.WAITING, got.State)
}

func TestGetProcessGraphByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	got, err := db.GetProcessGraphByID("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestSetProcessGraphState(t *testing.T) {
	db := setupTestDB(t)

	graph, _ := core.CreateProcessGraph("c1")
	db.AddProcessGraph(graph)

	// Transition to RUNNING (sets StartTime)
	err := db.SetProcessGraphState(graph.ID, core.RUNNING)
	assert.NoError(t, err)

	got, err := db.GetProcessGraphByID(graph.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.RUNNING, got.State)
	assert.False(t, got.StartTime.IsZero())

	// Transition to SUCCESS (sets EndTime)
	err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
	assert.NoError(t, err)

	got, err = db.GetProcessGraphByID(graph.ID)
	assert.NoError(t, err)
	assert.Equal(t, core.SUCCESS, got.State)
	assert.False(t, got.EndTime.IsZero())
}

func TestFindProcessGraphsByState(t *testing.T) {
	db := setupTestDB(t)

	g1, _ := core.CreateProcessGraph("c1")
	g2, _ := core.CreateProcessGraph("c1")
	db.AddProcessGraph(g1)
	db.AddProcessGraph(g2)

	waiting, err := db.FindWaitingProcessGraphs("c1", 10)
	assert.NoError(t, err)
	assert.Len(t, waiting, 2)

	db.SetProcessGraphState(g1.ID, core.RUNNING)

	waiting, err = db.FindWaitingProcessGraphs("c1", 10)
	assert.NoError(t, err)
	assert.Len(t, waiting, 1)

	running, err := db.FindRunningProcessGraphs("c1", 10)
	assert.NoError(t, err)
	assert.Len(t, running, 1)
}

func TestRemoveProcessGraphByID(t *testing.T) {
	db := setupTestDB(t)

	graph, _ := core.CreateProcessGraph("c1")
	db.AddProcessGraph(graph)

	err := db.RemoveProcessGraphByID(graph.ID)
	assert.NoError(t, err)

	got, err := db.GetProcessGraphByID(graph.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestCountProcessGraphs(t *testing.T) {
	db := setupTestDB(t)

	g1, _ := core.CreateProcessGraph("c1")
	g2, _ := core.CreateProcessGraph("c1")
	db.AddProcessGraph(g1)
	db.AddProcessGraph(g2)

	c, err := db.CountWaitingProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 2, c)

	db.SetProcessGraphState(g1.ID, core.RUNNING)

	c, err = db.CountRunningProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, c)

	c, err = db.CountWaitingProcessGraphsByColonyName("c1")
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func TestRemoveAllProcessGraphsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	g1, _ := core.CreateProcessGraph("c1")
	g2, _ := core.CreateProcessGraph("c2")
	db.AddProcessGraph(g1)
	db.AddProcessGraph(g2)

	err := db.RemoveAllProcessGraphsByColonyName("c1")
	assert.NoError(t, err)

	c, err := db.CountWaitingProcessGraphs()
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

// Log tests

func TestAddLog(t *testing.T) {
	db := setupTestDB(t)

	err := db.AddLog("pid1", "c1", "exec1", 1000, "test message")
	assert.NoError(t, err)

	logs, err := db.GetLogsByProcessID("pid1", 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, "test message", logs[0].Message)
	assert.Equal(t, int64(1000), logs[0].Timestamp)
}

func TestGetLogsByProcessID(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid1", "c1", "exec1", 200, "msg2")
	db.AddLog("pid1", "c1", "exec1", 300, "msg3")

	logs, err := db.GetLogsByProcessID("pid1", 2)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, int64(100), logs[0].Timestamp)
	assert.Equal(t, int64(200), logs[1].Timestamp)
}

func TestGetLogsByProcessIDSince(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid1", "c1", "exec1", 200, "msg2")
	db.AddLog("pid1", "c1", "exec1", 300, "msg3")

	logs, err := db.GetLogsByProcessIDSince("pid1", 10, 150)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, int64(200), logs[0].Timestamp)
}

func TestGetLogsByProcessIDLatest(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid1", "c1", "exec1", 200, "msg2")
	db.AddLog("pid1", "c1", "exec1", 300, "msg3")

	logs, err := db.GetLogsByProcessIDLatest("pid1", 2)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	// Should be in chronological order (oldest of latest first)
	assert.Equal(t, int64(200), logs[0].Timestamp)
	assert.Equal(t, int64(300), logs[1].Timestamp)
}

func TestGetLogsByExecutor(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid2", "c1", "exec1", 200, "msg2")
	db.AddLog("pid3", "c1", "exec2", 300, "msg3")

	logs, err := db.GetLogsByExecutor("exec1", 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

func TestGetLogsByExecutorSince(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid2", "c1", "exec1", 200, "msg2")

	logs, err := db.GetLogsByExecutorSince("exec1", 10, 150)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)
}

func TestGetLogsByExecutorLatest(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid2", "c1", "exec1", 200, "msg2")
	db.AddLog("pid3", "c1", "exec1", 300, "msg3")

	logs, err := db.GetLogsByExecutorLatest("exec1", 2)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, int64(200), logs[0].Timestamp)
	assert.Equal(t, int64(300), logs[1].Timestamp)
}

func TestRemoveLogsByColonyName(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid2", "c1", "exec1", 200, "msg2")
	db.AddLog("pid3", "c2", "exec1", 300, "msg3")

	err := db.RemoveLogsByColonyName("c1")
	assert.NoError(t, err)

	count, err := db.CountLogs("c1")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	count, err = db.CountLogs("c2")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCountLogs(t *testing.T) {
	db := setupTestDB(t)

	db.AddLog("pid1", "c1", "exec1", 100, "msg1")
	db.AddLog("pid2", "c1", "exec1", 200, "msg2")

	count, err := db.CountLogs("c1")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestSearchLogs(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now().UnixNano()
	db.AddLog("pid1", "c1", "exec1", now, "error occurred in module")
	db.AddLog("pid2", "c1", "exec1", now, "all good")
	db.AddLog("pid3", "c1", "exec1", now, "another error here")

	logs, err := db.SearchLogs("c1", "error", 30, 10)
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
}

// File tests

func TestAddFile(t *testing.T) {
	db := setupTestDB(t)

	file := &core.File{
		ID:         "fid1",
		ColonyName: "c1",
		Label:      "data",
		Name:       "test.txt",
		Size:       100,
		Checksum:   "abc123",
	}

	err := db.AddFile(file)
	assert.NoError(t, err)
	assert.True(t, file.SequenceNumber > 0)

	got, err := db.GetFileByID("c1", "fid1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "test.txt", got.Name)
}

func TestGetFileByIDNotFound(t *testing.T) {
	db := setupTestDB(t)

	got, err := db.GetFileByID("c1", "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetFileByIDWrongColony(t *testing.T) {
	db := setupTestDB(t)

	file := &core.File{ID: "fid1", ColonyName: "c1", Label: "data", Name: "test.txt"}
	db.AddFile(file)

	got, err := db.GetFileByID("c2", "fid1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestGetLatestFileByName(t *testing.T) {
	db := setupTestDB(t)

	f1 := &core.File{ID: "fid1", ColonyName: "c1", Label: "data", Name: "test.txt", Size: 100}
	f2 := &core.File{ID: "fid2", ColonyName: "c1", Label: "data", Name: "test.txt", Size: 200}
	db.AddFile(f1)
	db.AddFile(f2)

	files, err := db.GetLatestFileByName("c1", "data", "test.txt")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "fid2", files[0].ID) // Latest by sequence number
}

func TestGetFileByName(t *testing.T) {
	db := setupTestDB(t)

	f1 := &core.File{ID: "fid1", ColonyName: "c1", Label: "data", Name: "test.txt"}
	f2 := &core.File{ID: "fid2", ColonyName: "c1", Label: "data", Name: "test.txt"}
	db.AddFile(f1)
	db.AddFile(f2)

	files, err := db.GetFileByName("c1", "data", "test.txt")
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	// Should be sorted by sequence number descending
	assert.True(t, files[0].SequenceNumber > files[1].SequenceNumber)
}

func TestGetFilenamesByLabel(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "data", Name: "b.txt"})
	db.AddFile(&core.File{ID: "f3", ColonyName: "c1", Label: "data", Name: "a.txt"}) // duplicate name

	filenames, err := db.GetFilenamesByLabel("c1", "data")
	assert.NoError(t, err)
	assert.Len(t, filenames, 2) // "a.txt" and "b.txt"
}

func TestGetFileDataByLabel(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt", Size: 100, Checksum: "aaa"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "data", Name: "a.txt", Size: 200, Checksum: "bbb"})
	db.AddFile(&core.File{ID: "f3", ColonyName: "c1", Label: "data", Name: "b.txt", Size: 300, Checksum: "ccc"})

	data, err := db.GetFileDataByLabel("c1", "data")
	assert.NoError(t, err)
	assert.Len(t, data, 2) // One per unique filename, latest version

	// Find the "a.txt" entry - should have the latest version
	for _, d := range data {
		if d.Name == "a.txt" {
			assert.Equal(t, int64(200), d.Size)
		}
	}
}

func TestRemoveFileByID(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"})

	err := db.RemoveFileByID("c1", "f1")
	assert.NoError(t, err)

	got, err := db.GetFileByID("c1", "f1")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestRemoveFileByName(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "data", Name: "a.txt"})

	err := db.RemoveFileByName("c1", "data", "a.txt")
	assert.NoError(t, err)

	files, err := db.GetFileByName("c1", "data", "a.txt")
	assert.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestGetFileLabels(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "models", Name: "b.txt"})
	db.AddFile(&core.File{ID: "f3", ColonyName: "c1", Label: "data", Name: "c.txt"})

	labels, err := db.GetFileLabels("c1")
	assert.NoError(t, err)
	assert.Len(t, labels, 2)
}

func TestGetFileLabelsByName(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data/train", Name: "a.txt"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "data/test", Name: "b.txt"})
	db.AddFile(&core.File{ID: "f3", ColonyName: "c1", Label: "models", Name: "c.txt"})

	// Exact: match "data" itself or "data/*"
	labels, err := db.GetFileLabelsByName("c1", "data", true)
	assert.NoError(t, err)
	assert.Len(t, labels, 2) // data/train, data/test

	// Prefix: match "data*"
	labels, err = db.GetFileLabelsByName("c1", "data", false)
	assert.NoError(t, err)
	assert.Len(t, labels, 2) // data/train, data/test
}

func TestCountFiles(t *testing.T) {
	db := setupTestDB(t)

	db.AddFile(&core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"})
	db.AddFile(&core.File{ID: "f2", ColonyName: "c1", Label: "data", Name: "b.txt"})
	db.AddFile(&core.File{ID: "f3", ColonyName: "c1", Label: "models", Name: "c.txt"})

	count, err := db.CountFiles("c1")
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	count, err = db.CountFilesWithLabel("c1", "data")
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestFileSequenceNumbers(t *testing.T) {
	db := setupTestDB(t)

	f1 := &core.File{ID: "f1", ColonyName: "c1", Label: "data", Name: "a.txt"}
	f2 := &core.File{ID: "f2", ColonyName: "c1", Label: "data", Name: "b.txt"}
	f3 := &core.File{ID: "f3", ColonyName: "c1", Label: "data", Name: "c.txt"}
	db.AddFile(f1)
	db.AddFile(f2)
	db.AddFile(f3)

	assert.Equal(t, int64(1), f1.SequenceNumber)
	assert.Equal(t, int64(2), f2.SequenceNumber)
	assert.Equal(t, int64(3), f3.SequenceNumber)
}
