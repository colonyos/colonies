package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatisticsSecurity(t *testing.T) {
	env, client, server, serverPrvKey, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.Statistics(env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.Statistics(serverPrvKey)
	assert.Nil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestGetClusterInfoSecurity(t *testing.T) {
	env, client, server, serverPrvKey, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	_, err := client.GetClusterInfo(env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetClusterInfo(serverPrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}
