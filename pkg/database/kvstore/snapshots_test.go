package kvstore

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	// Create test file
	label := "test_label"
	fileID := "test_file_id"
	colonyName := "test_colony"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = fileID
	file.Label = label
	file.Name = "test_file"
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot"
	_, err = db.CreateSnapshot(colonyName, label, snapshotName)
	assert.Nil(t, err)

	_, err = db.GetSnapshotsByColonyName("invalid_colony")
	assert.Nil(t, err) // Returns empty slice

	_, err = db.GetSnapshotByID(colonyName, "invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.GetSnapshotByName(colonyName, "invalid_name")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveSnapshotByID(colonyName, "invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveSnapshotByName(colonyName, "invalid_name")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveSnapshotsByColonyName("invalid_colony")
	assert.Nil(t, err) // No error when nothing to remove
}

func TestCreateSnapshot(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	file2ID := "test_file2_id"
	file3ID := "test_file3_id"
	colonyName := "test_colony"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("test_id", colonyName, now)
	file1.ID = file1ID
	file1.Label = label
	file1.Name = "test_file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("test_id", colonyName, now.Add(time.Minute)) // Later timestamp
	file2.ID = file2ID // Add another revision of test_file1
	file2.Label = label
	file2.Name = "test_file1"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFileWithID("test_id", colonyName, now)
	file3.ID = file3ID
	file3.Label = label
	file3.Name = "test_file3"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 2)

	counter := 0
	for _, fileID := range snapshot.FileIDs {
		if fileID == file2ID { // we want the latest revision, i.e not file1id
			counter++
		}
		if fileID == file3ID {
			counter++
		}
	}
	assert.Equal(t, counter, 2)
	assert.Equal(t, snapshot.ColonyName, colonyName)
	assert.Equal(t, snapshot.Name, snapshotName)
	assert.Equal(t, snapshot.Label, label)

	_, err = db.CreateSnapshot(colonyName, label, snapshotName)
	assert.NotNil(t, err) // name must be unique
}

func TestGetSnapshotByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyName := "test_colony"

	now := time.Now()
	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 1)

	snapshotFromDB, err := db.GetSnapshotByID(colonyName, snapshot.ID)
	assert.Nil(t, err)
	assert.True(t, snapshotFromDB.Equals(snapshot))

	// Test non-existing ID
	_, err = db.GetSnapshotByID(colonyName, "non_existing_id")
	assert.NotNil(t, err)

	// Test with empty colony
	_, err = db.GetSnapshotByID("invalid_colony", snapshot.ID)
	assert.NotNil(t, err)
}

func TestGetSnapshotByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyName := "test_colony"

	now := time.Now()
	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 1)

	snapshotFromDB, err := db.GetSnapshotByName(colonyName, snapshotName)
	assert.Nil(t, err)
	assert.True(t, snapshotFromDB.Equals(snapshot))

	// Test non-existing name
	_, err = db.GetSnapshotByName(colonyName, "non_existing_name")
	assert.NotNil(t, err)

	// Test with empty colony
	_, err = db.GetSnapshotByName("invalid_colony", snapshotName)
	assert.NotNil(t, err)
}

func TestGetSnapshotsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyName := "test_colony"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	_, err = db.CreateSnapshot(colonyName, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	_, err = db.CreateSnapshot(colonyName, label, snapshotName2)
	assert.Nil(t, err)

	snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	// Test non-existing colony
	snapshotsEmpty, err := db.GetSnapshotsByColonyName("non_existing_colony")
	assert.Nil(t, err)
	assert.Empty(t, snapshotsEmpty)
}

func TestRemoveSnapshotByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyName := "test_colony"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	snapshot1, err := db.CreateSnapshot(colonyName, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	snapshot2, err := db.CreateSnapshot(colonyName, label, snapshotName2)
	assert.Nil(t, err)

	err = db.RemoveSnapshotByID(colonyName, snapshot1.ID)
	assert.Nil(t, err)

	_, err = db.GetSnapshotByID(colonyName, snapshot1.ID)
	assert.NotNil(t, err)
	_, err = db.GetSnapshotByID(colonyName, snapshot2.ID)
	assert.Nil(t, err)
	snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	// Test removing non-existing snapshot
	err = db.RemoveSnapshotByID(colonyName, "non_existing_id")
	assert.NotNil(t, err)

	// Test removing from non-existing colony
	err = db.RemoveSnapshotByID("invalid_colony", snapshot2.ID)
	assert.NotNil(t, err)
}

func TestRemoveSnapshotByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyName := "test_colony_id"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	snapshot1, err := db.CreateSnapshot(colonyName, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	snapshot2, err := db.CreateSnapshot(colonyName, label, snapshotName2)
	assert.Nil(t, err)

	err = db.RemoveSnapshotByName(colonyName, snapshotName1)
	assert.Nil(t, err)

	_, err = db.GetSnapshotByID(colonyName, snapshot1.ID)
	assert.NotNil(t, err)
	_, err = db.GetSnapshotByID(colonyName, snapshot2.ID)
	assert.Nil(t, err)
	snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	// Test removing non-existing snapshot
	err = db.RemoveSnapshotByName(colonyName, "non_existing_name")
	assert.NotNil(t, err)

	// Test removing from non-existing colony
	err = db.RemoveSnapshotByName("invalid_colony", snapshotName2)
	assert.NotNil(t, err)
}

func TestRemoveSnapshotsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	file2ID := "test_file2_id"
	file3ID := "test_file3_id"
	colonyName1 := "test_colony_1"
	colonyName2 := "test_colony_2"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName1, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file = utils.CreateTestFileWithID("test_id", colonyName2, now)
	file.ID = file2ID
	file.Label = label
	file.Name = "test_file2"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file = utils.CreateTestFileWithID("test_id", colonyName2, now)
	file.ID = file3ID
	file.Label = label
	file.Name = "test_file3"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name1"
	_, err = db.CreateSnapshot(colonyName1, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	_, err = db.CreateSnapshot(colonyName1, label, snapshotName2)
	assert.Nil(t, err)

	snapshotName3 := "test_snapshot_name3"
	_, err = db.CreateSnapshot(colonyName2, label, snapshotName3)
	assert.Nil(t, err)

	snapshotName4 := "test_snapshot_name4"
	_, err = db.CreateSnapshot(colonyName2, label, snapshotName4)
	assert.Nil(t, err)

	err = db.RemoveSnapshotsByColonyName(colonyName1)
	assert.Nil(t, err)

	snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName1)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 0)

	snapshotsFromDB, err = db.GetSnapshotsByColonyName(colonyName2)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	// Test removing from non-existing colony - should not error
	err = db.RemoveSnapshotsByColonyName("non_existing_colony")
	assert.Nil(t, err)
}

func TestSnapshotComplexScenarios(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"
	now := time.Now()

	// Create multiple files with same label but different names
	file1 := utils.CreateTestFileWithID("file1", colonyName, now)
	file1.ID = "file1_id"
	file1.Label = label
	file1.Name = "file_a"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2", colonyName, now.Add(time.Second))
	file2.ID = "file2_id" 
	file2.Label = label
	file2.Name = "file_b"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Create snapshot
	snapshotName := "complex_snapshot"
	snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 2)

	// Verify snapshot contains correct files
	assert.Contains(t, snapshot.FileIDs, "file1_id")
	assert.Contains(t, snapshot.FileIDs, "file2_id")

	// Test snapshot with empty label (should create empty snapshot)
	emptySnapshot, err := db.CreateSnapshot(colonyName, "empty_label", "empty_snapshot")
	assert.Nil(t, err)
	assert.Empty(t, emptySnapshot.FileIDs)

	// Test snapshot with files from different colonies
	otherColony := "other_colony"
	otherFile := utils.CreateTestFileWithID("other_file", otherColony, now)
	otherFile.ID = "other_file_id"
	otherFile.Label = label
	otherFile.Name = "other_file"
	err = db.AddFile(otherFile)
	assert.Nil(t, err)

	// Snapshot from first colony should not include files from other colony
	snapshot2, err := db.CreateSnapshot(colonyName, label, "snapshot2")
	assert.Nil(t, err)
	assert.Len(t, snapshot2.FileIDs, 2)
	assert.NotContains(t, snapshot2.FileIDs, "other_file_id")
}

func TestSnapshotFileRevisions(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"
	fileName := "versioned_file"

	// Create multiple revisions of the same file
	baseTime := time.Now()
	
	revision1 := utils.CreateTestFileWithID("rev1", colonyName, baseTime)
	revision1.ID = "revision1_id"
	revision1.Label = label
	revision1.Name = fileName
	err = db.AddFile(revision1)
	assert.Nil(t, err)

	revision2 := utils.CreateTestFileWithID("rev2", colonyName, baseTime.Add(time.Minute))
	revision2.ID = "revision2_id"
	revision2.Label = label
	revision2.Name = fileName
	err = db.AddFile(revision2)
	assert.Nil(t, err)

	revision3 := utils.CreateTestFileWithID("rev3", colonyName, baseTime.Add(2*time.Minute))
	revision3.ID = "revision3_id"
	revision3.Label = label
	revision3.Name = fileName
	err = db.AddFile(revision3)
	assert.Nil(t, err)

	// Create snapshot - should only include the latest revision
	snapshot, err := db.CreateSnapshot(colonyName, label, "revision_snapshot")
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 1)
	assert.Contains(t, snapshot.FileIDs, "revision3_id")
	assert.NotContains(t, snapshot.FileIDs, "revision1_id")
	assert.NotContains(t, snapshot.FileIDs, "revision2_id")
}