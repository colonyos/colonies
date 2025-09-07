package cron_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCronSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)

	_, err := client.AddCron(cron, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddCron(cron, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)
	addedCron, err := client.AddCron(cron, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCron(addedCron.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCron(addedCron.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetCronsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)
	_, err := client.AddCron(cron, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetCrons(env.Colony1Name, 100, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.Colony1Name, 100, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.Colony1Name, 100, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetCrons(env.Colony1Name, 100, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRunCronSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)
	addedCron, err := client.AddCron(cron, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.RunCron(addedCron.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.RunCron(addedCron.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveCronSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	cron := utils.FakeCron(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)
	addedCron, err := client.AddCron(cron, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveCron(addedCron.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveCron(addedCron.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
