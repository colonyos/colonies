package snapshot_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshot(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name")

	server.Shutdown()
	<-done
}

func TestGetSnapshot(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name")

	snapshotFromDB, err := client.GetSnapshotByID(env.ColonyName, snapshot.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, snapshot.Equals(snapshotFromDB))

	snapshotFromDB, err = client.GetSnapshotByName(env.ColonyName, "test_snapshot_name", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, snapshot.Equals(snapshotFromDB))

	server.Shutdown()
	<-done
}

func TestGetSnapshotsByColonyName(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name1", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name1")

	snapshot, err = client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name2", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByID(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name1", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name2", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	err = client.RemoveSnapshotByID(env.ColonyName, snapshot2.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err = client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	_, err = client.GetSnapshotByID(env.ColonyName, snapshot1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetSnapshotByID(env.ColonyName, snapshot2.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByName(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name1", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name2", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 2)

	err = client.RemoveSnapshotByName(env.ColonyName, "test_snapshot_name2", env.ExecutorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err = client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 1)

	_, err = client.GetSnapshotByID(env.ColonyName, snapshot1.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.GetSnapshotByID(env.ColonyName, snapshot2.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveAllSnapshots(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	label := "test_label"

	file := utils.CreateTestFile(env.ColonyName)
	addedFile, err := client.AddFile(file, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedFile)
	file.ID = addedFile.ID
	file.Added = addedFile.Added
	file.SequenceNumber = addedFile.SequenceNumber
	assert.True(t, file.Equals(addedFile))

	snapshot1, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name1", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot1.Name, "test_snapshot_name1")

	snapshot2, err := client.CreateSnapshot(env.ColonyName, label, "test_snapshot_name2", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, snapshot2.Name, "test_snapshot_name2")

	err = client.RemoveAllSnapshots(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)

	snapshotsFromDB, err := client.GetSnapshotsByColonyName(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, snapshotsFromDB, 0)

	server.Shutdown()
	<-done
}

// TestCreateSnapshotUnauthorized tests creating snapshot from different colony
func TestCreateSnapshotUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to create snapshot in colony1 with executor2's key
	_, err = client.CreateSnapshot(colony1.Name, "label", "test_snapshot", executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetSnapshotUnauthorized tests getting snapshot from different colony
func TestGetSnapshotUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Create snapshot in colony1
	file := utils.CreateTestFile(colony1.Name)
	_, err = client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	snapshot, err := client.CreateSnapshot(colony1.Name, "label", "test_snapshot", executor1PrvKey)
	assert.Nil(t, err)

	// Try to get snapshot from colony1 with executor2's key
	_, err = client.GetSnapshotByID(colony1.Name, snapshot.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetSnapshotsUnauthorized tests getting snapshots list from different colony
func TestGetSnapshotsUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Create snapshot in colony1
	file := utils.CreateTestFile(colony1.Name)
	_, err = client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.CreateSnapshot(colony1.Name, "label", "test_snapshot", executor1PrvKey)
	assert.Nil(t, err)

	// Try to get snapshots from colony1 with executor2's key
	_, err = client.GetSnapshotsByColonyName(colony1.Name, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveSnapshotUnauthorized tests removing snapshot from different colony
func TestRemoveSnapshotUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Create snapshot in colony1
	file := utils.CreateTestFile(colony1.Name)
	_, err = client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	snapshot, err := client.CreateSnapshot(colony1.Name, "label", "test_snapshot", executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove snapshot from colony1 with executor2's key
	err = client.RemoveSnapshotByID(colony1.Name, snapshot.ID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveAllSnapshotsUnauthorized tests removing all snapshots from different colony
func TestRemoveAllSnapshotsUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Create snapshot in colony1
	file := utils.CreateTestFile(colony1.Name)
	_, err = client.AddFile(file, executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.CreateSnapshot(colony1.Name, "label", "test_snapshot", executor1PrvKey)
	assert.Nil(t, err)

	// Try to remove all snapshots from colony1 with executor2's key
	err = client.RemoveAllSnapshots(colony1.Name, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetSnapshotByIDNotFound tests getting non-existent snapshot by ID
func TestGetSnapshotByIDNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.GetSnapshotByID(env.ColonyName, "non_existent_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetSnapshotByNameNotFound tests getting non-existent snapshot by name
func TestGetSnapshotByNameNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.GetSnapshotByName(env.ColonyName, "non_existent_name", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
