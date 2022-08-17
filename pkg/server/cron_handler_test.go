package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCron(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	cronFromServer, err := client.GetCron(cron.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, addedCron.Equals(cronFromServer))

	cron = utils.FakeCron(t, env.colonyID)
	cron.WorkflowSpec = "error"
	addedCron, err = client.AddCron(cron, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, addedCron)

	cron = utils.FakeCron(t, env.colonyID)
	cron.CronExpression = "error"
	addedCron, err = client.AddCron(cron, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, addedCron)

	server.Shutdown()
	<-done
}

func TestGetCrons(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron1 := utils.FakeCron(t, env.colonyID)
	cron1.Name = "test_cron_1"
	addedCron1, err := client.AddCron(cron1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron1)
	cron2 := utils.FakeCron(t, env.colonyID)
	cron2.Name = "test_cron_2"
	addedCron2, err := client.AddCron(cron2, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron2)

	cronsFromServer, err := client.GetCrons(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)

	assert.Len(t, cronsFromServer, 2)

	counter := 0
	for _, cron := range cronsFromServer {
		if cron.Name == "test_cron_1" {
			counter++
		}
		if cron.Name == "test_cron_2" {
			counter++
		}
	}

	assert.Equal(t, counter, 2)

	server.Shutdown()
	<-done
}

func TestDeleteCron(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	err = client.DeleteCron(cron.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	cronFromServer, err := client.GetCron(cron.ID, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, cronFromServer)

	server.Shutdown()
	<-done
}

func TestRunCron(t *testing.T) {
	// TODO
}
