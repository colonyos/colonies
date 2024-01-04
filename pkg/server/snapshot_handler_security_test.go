package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshotSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony2Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByID(env.colony1Name, snapshot.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1Name, snapshot.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1Name, snapshot.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony2Name, snapshot.ID, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1Name, snapshot.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByName(env.colony1Name, snapshot.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1Name, snapshot.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1Name, snapshot.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony2Name, snapshot.Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1Name, snapshot.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByColonyNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotsByColonyName(env.colony1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.colony1Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.colony1Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.colony2Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.colony1Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveSnapshotByName(env.colony1Name, snapshot.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.colony1Name, snapshot.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.colony1Name, snapshot.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.colony2Name, snapshot.Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.colony1Name, snapshot.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveSnapshotByID(env.colony1Name, snapshot.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.colony1Name, snapshot.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.colony1Name, snapshot.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.colony2Name, snapshot.ID, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.colony1Name, snapshot.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveAllSnapshotSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.colony1Name, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveAllSnapshots(env.colony1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.colony1Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.colony1Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.colony2Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.colony1Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
