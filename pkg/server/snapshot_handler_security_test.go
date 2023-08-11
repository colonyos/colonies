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

	_, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony2ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByID(env.colony1ID, snapshot.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1ID, snapshot.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1ID, snapshot.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony2ID, snapshot.ID, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.colony1ID, snapshot.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByName(env.colony1ID, snapshot.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1ID, snapshot.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1ID, snapshot.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony2ID, snapshot.Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.colony1ID, snapshot.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByColonyIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotsByColonyID(env.colony1ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyID(env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyID(env.colony2ID, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyID(env.colony1ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.DeleteSnapshotByName(env.colony1ID, snapshot.Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByName(env.colony1ID, snapshot.Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByName(env.colony1ID, snapshot.Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByName(env.colony2ID, snapshot.Name, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByName(env.colony1ID, snapshot.Name, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.colony1ID, "test_label", "test_snapshot_name", env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.DeleteSnapshotByID(env.colony1ID, snapshot.ID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByID(env.colony1ID, snapshot.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByID(env.colony1ID, snapshot.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByID(env.colony2ID, snapshot.ID, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteSnapshotByID(env.colony1ID, snapshot.ID, env.executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
