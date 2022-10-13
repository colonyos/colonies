package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCron(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// If the cron is successful, there should be a process we can assign
	process, err := client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronWaitForPrevProcessGraph(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.colonyID)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// Wait for 5 seconds, we should only have 1 cron workflow since WaitForPrevProcessGraph is true
	time.Sleep(5 * time.Second)

	processes, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	firstProcessID := processes[0]

	processgraphs, err := client.GetWaitingProcessGraphs(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processgraphs, 1)

	// Now assign a the cron process, then a new cron should be triggered
	process, err := client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Close(process.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	processes, err = client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	secondProcessID := processes[0]

	assert.NotEqual(t, firstProcessID, secondProcessID)

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)
	assert.Equal(t, stat.WaitingWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronInputOutput(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.colonyID)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	process, err := client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.CloseWithOutput(process.ID, []string{"result_cron1"}, env.runtimePrvKey)
	fmt.Println(process.ID)

	process, err = client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Len(t, process.Input, 1)
	assert.Equal(t, process.Input[0], "result_cron1")

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 0)
	assert.Equal(t, stat.RunningWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronInputOutput2(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	process, err := client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Close(process.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	process, err = client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.CloseWithOutput(process.ID, []string{"result_cron1"}, env.runtimePrvKey)
	fmt.Println(process.ID)

	process, err = client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Len(t, process.Input, 1)
	assert.Equal(t, process.Input[0], "result_cron1")

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 0)
	assert.Equal(t, stat.RunningWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronWithCronExpr(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.CronExpression = "0/1 * * * * *" // every second

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// If the cron is successful, there should be a process we can assign
	process, err := client.AssignProcess(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronFail(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.WorkflowSpec = "error"
	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
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

func TestGetCron(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	cronFromServer, err := client.GetCron(addedCron.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, addedCron.Equals(cronFromServer))

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

	err = client.DeleteCron(addedCron.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	cronFromServer, err := client.GetCron(addedCron.ID, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, cronFromServer)

	server.Shutdown()
	<-done
}

func TestRunCron(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	cron := utils.FakeCron(t, env.colonyID)
	cron.Interval = 1000 // Will be triggered in 1000 seconds

	addedCron, err := client.AddCron(cron, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	_, err = client.RunCron(addedCron.ID, env.runtimePrvKey)

	// If the cron is successful, there should be a process we can assign
	process, err := client.AssignProcess(env.colonyID, 10, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}
