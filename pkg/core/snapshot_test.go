package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsSnapshotEquals(t *testing.T) {
	now := time.Now()
	fileID1s := []string{"test_file_id1", "test_file_id2"}
	snapshot1 := &Snapshot{ID: "test_snapshot_id1", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname", FileIDs: fileID1s, Added: now}

	fileID2s := []string{"test_file_id1", "test_file_id2"}
	snapshot2 := &Snapshot{ID: "test_snapshot_id1", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname", FileIDs: fileID2s, Added: now}

	assert.True(t, snapshot1.Equals(snapshot2))
	snapshot1.Name = "changed_name"
	assert.False(t, snapshot1.Equals(snapshot2))
}

func TestSnapshotToJSON(t *testing.T) {
	now := time.Now()
	fileID1s := []string{"test_file_id1", "test_file_id2"}
	snapshot1 := &Snapshot{ID: "test_snapshot_id1", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname", FileIDs: fileID1s, Added: now}
	jsonStr, err := snapshot1.ToJSON()
	assert.Nil(t, err)

	snapshot2, err := ConvertJSONToSnapshot(jsonStr)
	assert.Nil(t, err)
	assert.True(t, snapshot1.Equals(snapshot2))
}

func TestIsSnapshotArraysEquals(t *testing.T) {
	now := time.Now()
	fileIDs := []string{"test_file_id1", "test_file_id2"}
	snapshot1 := &Snapshot{ID: "test_snapshot_id1", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname1", FileIDs: fileIDs, Added: now}
	snapshot2 := &Snapshot{ID: "test_snapshot_id2", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname2", FileIDs: fileIDs, Added: now}
	snapshot3 := &Snapshot{ID: "test_snapshot_id3", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname3", FileIDs: fileIDs, Added: now}
	snapshot4 := &Snapshot{ID: "test_snapshot_id4", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname4", FileIDs: fileIDs, Added: now}

	snapshots1 := []*Snapshot{snapshot1, snapshot2}
	snapshots2 := []*Snapshot{snapshot3, snapshot4}
	assert.True(t, IsSnapshotArraysEqual(snapshots1, snapshots1))
	assert.False(t, IsSnapshotArraysEqual(snapshots1, snapshots2))
}

func TestSnapshotArrayToJSON(t *testing.T) {
	now := time.Now()
	fileIDs := []string{"test_file_id1", "test_file_id2"}
	snapshot1 := &Snapshot{ID: "test_snapshot_id1", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname1", FileIDs: fileIDs, Added: now}
	snapshot2 := &Snapshot{ID: "test_snapshot_id2", ColonyName: "test_colony_1", Label: "test_label", Name: "test_snapshotname2", FileIDs: fileIDs, Added: now}
	snapshots1 := []*Snapshot{snapshot1, snapshot2}

	jsonStr, err := ConvertSnapshotArrayToJSON(snapshots1)
	assert.Nil(t, err)

	snapshots2, err := ConvertJSONToSnapshotsArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsSnapshotArraysEqual(snapshots1, snapshots2))
}
