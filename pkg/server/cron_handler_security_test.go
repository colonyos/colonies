package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	cron := utils.FakeCron(t, env.colony1ID)

	_, err := client.AddCron(cron, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	cron := utils.FakeCron(t, env.colony1ID)
	_, err := client.AddCron(cron, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCron(cron.ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(cron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(cron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(cron.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	cron := utils.FakeCron(t, env.colony1ID)
	_, err := client.AddCron(cron, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCrons(env.colony1ID, 100, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1ID, 100, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1ID, 100, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.colony1ID, 100, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRunCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	cron := utils.FakeCron(t, env.colony1ID)
	_, err := client.AddCron(cron, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.RunCron(cron.ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(cron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(cron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(cron.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteCronSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	cron := utils.FakeCron(t, env.colony1ID)
	_, err := client.AddCron(cron, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteCron(cron.ID, env.runtime2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteCron(cron.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteCron(cron.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteCron(cron.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
