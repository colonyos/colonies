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
	colonyName := "test_colony"
	now := time.Now()

	file := utils.CreateTestFileWithID("test_id", colonyName, now)
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
	db, err := PrepareTests()
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
}

func TestGetSnapshotByName(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestGetSnapshotsByColonyName(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestRemoveSnapshotByID(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestRemoveSnapshotByName(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestRemoveSnapshotsByColonyName(t *testing.T) {
	db, err := PrepareTests()
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
}
