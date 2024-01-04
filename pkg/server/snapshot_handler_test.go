package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshot(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name")

	server.Shutdown()
	<-done
}

func TestGetSnapshot(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name")

	snapshotFromDB, err := client.GetSnapshotByID(env.colonyName, snapshot.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, snapshot.Equals(snapshotFromDB))

	snapshotFromDB, err = client.GetSnapshotByName(env.colonyName, "test_snapshot_name", env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, snapshot.Equals(snapshotFromDB))

	server.Shutdown()
	<-done
}

func TestGetSnapshotsByColonyName(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name1", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name1")

	snapshot, err = client.CreateSnapshot(env.colonyName, label, "test_snapshot_name2", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name1", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name2", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	err = client.RemoveSnapshotByID(env.colonyName, snapshot2.ID, env.executorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err = client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	_, err = client.GetSnapshotByID(env.colonyName, snapshot1.ID, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetSnapshotByID(env.colonyName, snapshot2.ID, env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByName(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name1", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name2", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	err = client.RemoveSnapshotByName(env.colonyName, "test_snapshot_name2", env.executorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err = client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	_, err = client.GetSnapshotByID(env.colonyName, snapshot1.ID, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetSnapshotByID(env.colonyName, snapshot2.ID, env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveAllSnapshots(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.colonyName)
	addedFile, err := client.AddFile(file, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name1", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.colonyName, label, "test_snapshot_name2", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	err = client.RemoveAllSnapshots(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 0)

	server.Shutdown()
	<-done
}
