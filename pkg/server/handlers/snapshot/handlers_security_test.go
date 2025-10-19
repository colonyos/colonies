package snapshot_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestCreateSnapshotSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.Colony2Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByID(env.Colony1Name, snapshot.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.Colony1Name, snapshot.ID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.Colony1Name, snapshot.ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.Colony2Name, snapshot.ID, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByID(env.Colony1Name, snapshot.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotByName(env.Colony1Name, snapshot.Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.Colony1Name, snapshot.Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.Colony1Name, snapshot.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.Colony2Name, snapshot.Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotByName(env.Colony1Name, snapshot.Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSnapshotByColonyNameSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetSnapshotsByColonyName(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.Colony1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.Colony2Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSnapshotsByColonyName(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByNameSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveSnapshotByName(env.Colony1Name, snapshot.Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.Colony1Name, snapshot.Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.Colony1Name, snapshot.Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.Colony2Name, snapshot.Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByName(env.Colony1Name, snapshot.Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveSnapshotByIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	snapshot, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveSnapshotByID(env.Colony1Name, snapshot.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.Colony1Name, snapshot.ID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.Colony1Name, snapshot.ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.Colony2Name, snapshot.ID, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveSnapshotByID(env.Colony1Name, snapshot.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRemoveAllSnapshotSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.CreateSnapshot(env.Colony1Name, "test_label", "test_snapshot_name", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	err = client.RemoveAllSnapshots(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.Colony1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.Colony2Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllSnapshots(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
