package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshot(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	file2ID := "test_file2_id"
	file3ID := "test_file3_id"
	colonyID := "test_colony_id"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file.ID = file2ID // Add another revision of test_file1
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file.ID = file3ID
	file.Label = label
	file.Name = "test_file3"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyID, label, snapshotName)
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
	assert.Equal(t, snapshot.ColonyID, colonyID)
	assert.Equal(t, snapshot.Name, snapshotName)
	assert.Equal(t, snapshot.Label, label)

	_, err = db.CreateSnapshot(colonyID, label, snapshotName)
	assert.NotNil(t, err) // name must be unique
}

func TestGetSnapshotByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyID := "test_colony_id"

	now := time.Now()
	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyID, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 1)

	snapshotFromDB, err := db.GetSnapshotByID(colonyID, snapshot.ID)
	assert.Nil(t, err)
	assert.True(t, snapshotFromDB.Equals(snapshot))
}

func TestGetSnapshotByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyID := "test_colony_id"

	now := time.Now()
	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName := "test_snapshot_name"
	snapshot, err := db.CreateSnapshot(colonyID, label, snapshotName)
	assert.Nil(t, err)
	assert.Len(t, snapshot.FileIDs, 1)

	snapshotFromDB, err := db.GetSnapshotByName(colonyID, snapshotName)
	assert.Nil(t, err)
	assert.True(t, snapshotFromDB.Equals(snapshot))
}

func TestGetSnapshotsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyID := "test_colony_id"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	_, err = db.CreateSnapshot(colonyID, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	_, err = db.CreateSnapshot(colonyID, label, snapshotName2)
	assert.Nil(t, err)

	snapshotsFromDB, err := db.GetSnapshotsByColonyID(colonyID)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)
}

func TestDeleteSnapshotByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyID := "test_colony_id"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	snapshot1, err := db.CreateSnapshot(colonyID, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	snapshot2, err := db.CreateSnapshot(colonyID, label, snapshotName2)
	assert.Nil(t, err)

	err = db.DeleteSnapshotByID(colonyID, snapshot1.ID)
	assert.Nil(t, err)

	_, err = db.GetSnapshotByID(colonyID, snapshot1.ID)
	assert.NotNil(t, err)
	_, err = db.GetSnapshotByID(colonyID, snapshot2.ID)
	assert.Nil(t, err)
	snapshotsFromDB, err := db.GetSnapshotsByColonyID(colonyID)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)
}

func TestDeleteSnapshotByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	colonyID := "test_colony_id"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyID, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name"
	snapshot1, err := db.CreateSnapshot(colonyID, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	snapshot2, err := db.CreateSnapshot(colonyID, label, snapshotName2)
	assert.Nil(t, err)

	err = db.DeleteSnapshotByName(colonyID, snapshotName1)
	assert.Nil(t, err)

	_, err = db.GetSnapshotByID(colonyID, snapshot1.ID)
	assert.NotNil(t, err)
	_, err = db.GetSnapshotByID(colonyID, snapshot2.ID)
	assert.Nil(t, err)
	snapshotsFromDB, err := db.GetSnapshotsByColonyID(colonyID)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)
}

func TestDeleteSnapshotsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	label := "test_label"
	file1ID := "test_file1_id"
	file2ID := "test_file2_id"
	file3ID := "test_file3_id"
	colonyID1 := "test_colony_id1"
	colonyID2 := "test_colony_id2"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyID1, now)
	file.ID = file1ID
	file.Label = label
	file.Name = "test_file1"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file = utils.CreateTestFileWithID("test_id", colonyID2, now)
	file.ID = file2ID
	file.Label = label
	file.Name = "test_file2"
	err = db.AddFile(file)
	assert.Nil(t, err)

	file = utils.CreateTestFileWithID("test_id", colonyID2, now)
	file.ID = file3ID
	file.Label = label
	file.Name = "test_file3"
	err = db.AddFile(file)
	assert.Nil(t, err)

	snapshotName1 := "test_snapshot_name1"
	_, err = db.CreateSnapshot(colonyID1, label, snapshotName1)
	assert.Nil(t, err)

	snapshotName2 := "test_snapshot_name2"
	_, err = db.CreateSnapshot(colonyID1, label, snapshotName2)
	assert.Nil(t, err)

	snapshotName3 := "test_snapshot_name3"
	_, err = db.CreateSnapshot(colonyID2, label, snapshotName3)
	assert.Nil(t, err)

	snapshotName4 := "test_snapshot_name4"
	_, err = db.CreateSnapshot(colonyID2, label, snapshotName4)
	assert.Nil(t, err)

	err = db.DeleteSnapshotsByColonyID(colonyID1)
	assert.Nil(t, err)

	snapshotsFromDB, err := db.GetSnapshotsByColonyID(colonyID1)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 0)

	snapshotsFromDB, err = db.GetSnapshotsByColonyID(colonyID2)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)
}
