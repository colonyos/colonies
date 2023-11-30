package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.colony1Name)

	_, err := client.AddCron(cron, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.colony1Name)
	addedCron, err := client.AddCron(cron, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCron(addedCron.ID, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.colony1Name)
	_, err := client.AddCron(cron, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCrons(env.colony1Name, 100, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1Name, 100, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1Name, 100, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1Name, 100, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRunCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.colony1Name)
	addedCron, err := client.AddCron(cron, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.RunCron(addedCron.ID, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.colony1Name)
	addedCron, err := client.AddCron(cron, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveCron(addedCron.ID, env.executor2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
