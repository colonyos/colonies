package server_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestGetStatisticsSecurity(t *testing.T) {
	env, client, coloniesServer, serverPrvKey, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.Statistics(env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(serverPrvKey)
	assert.Nil(t, err) // Should not work

	coloniesServer.Shutdown()
	<-done
}

func TestGetClusterInfoSecurity(t *testing.T) {
	env, client, coloniesServer, serverPrvKey, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetClusterInfo(env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(serverPrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}
