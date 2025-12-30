package cron_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddCronDebug(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronRemoveAllProcesses(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	time.Sleep(2 * time.Second)

	err = client.RemoveAllProcesses(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronRemoveAllProcessGraphs(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	time.Sleep(2 * time.Second)

	err = client.RemoveAllProcessGraphs(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestFailCron(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	err = client.Fail(process.ID, []string{}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Cron should still generate a cron workflow even if the last process fails
	process, err = client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronWaitForPrevProcessGraph(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// Wait for 5 seconds, we should only have 1 cron workflow since WaitForPrevProcessGraph is true
	time.Sleep(5 * time.Second)

	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	firstProcessID := processes[0]

	processgraphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processgraphs, 1)

	// Now assign a the cron process, then a new cron should be triggered
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	processes, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	secondProcessID := processes[0]

	assert.NotEqual(t, firstProcessID, secondProcessID)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)
	assert.Equal(t, stat.WaitingWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronWaitForPrevProcessGraphFail(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// Wait for 5 seconds, we should only have 1 cron workflow since WaitForPrevProcessGraph is true
	time.Sleep(5 * time.Second)

	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	firstProcessID := processes[0]

	processgraphs, err := client.GetWaitingProcessGraphs(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processgraphs, 1)

	// Now assign a the cron process, then a new cron should be triggered
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Fail(process.ID, []string{""}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	processes, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	secondProcessID := processes[0]

	assert.NotEqual(t, firstProcessID, secondProcessID)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 0)
	assert.Equal(t, stat.FailedWorkflows, 1)
	assert.Equal(t, stat.WaitingWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronInputOutput(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	output := make([]interface{}, 1)
	output[0] = "result_cron1"
	err = client.CloseWithOutput(process.ID, output, env.ExecutorPrvKey)

	process, err = client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Len(t, process.Input, 1)
	assert.Equal(t, process.Input[0], "result_cron1")

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 0)
	assert.Equal(t, stat.RunningWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronInputOutput2(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	process, err = client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	output := make([]interface{}, 1)
	output[0] = "result_cron1"
	err = client.CloseWithOutput(process.ID, output, env.ExecutorPrvKey)

	process, err = client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	assert.Len(t, process.Input, 1)
	assert.Equal(t, process.Input[0], "result_cron1")

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingWorkflows, 0)
	assert.Equal(t, stat.RunningWorkflows, 1)
	assert.Equal(t, stat.SuccessfulWorkflows, 1)

	server.Shutdown()
	<-done
}

func TestAddCronWithCronExpr(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.CronExpression = "0/1 * * * * *" // every second

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestAddCronFail(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.WorkflowSpec = "error"
	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, addedCron)

	cron = utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.CronExpression = "error"
	addedCron, err = client.AddCron(cron, env.ExecutorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, addedCron)

	server.Shutdown()
	<-done
}

func TestGetCron(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	cronFromServer, err := client.GetCron(addedCron.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedCron.ID, cronFromServer.ID)

	server.Shutdown()
	<-done
}

func TestCronArgs(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 2

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	process, err := client.Assign(env.ColonyName, 100, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	// TODO:
	fmt.Println(process)

	server.Shutdown()
	<-done
}

func TestGetCrons(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron1 := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron1.Name = "test_cron_1"
	addedCron1, err := client.AddCron(cron1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron1)
	cron2 := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron2.Name = "test_cron_2"
	addedCron2, err := client.AddCron(cron2, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron2)

	cronsFromServer, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
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

func TestRemoveCron(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	err = client.RemoveCron(addedCron.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	cronFromServer, err := client.GetCron(addedCron.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, cronFromServer)

	server.Shutdown()
	<-done
}

func TestRunCron(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000 // Will be triggered in 1000 seconds

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	_, err = client.RunCron(addedCron.ID, env.ExecutorPrvKey)

	// If the cron is successful, there should be a process we can assign
	process, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)

	server.Shutdown()
	<-done
}

func TestRunCronWaitForPrevProcessGraph(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeSingleCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000 // Long interval so it won't trigger automatically
	cron.WaitForPrevProcessGraph = true

	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)

	// First RunCron should succeed and create a process
	_, err = client.RunCron(addedCron.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify a process was created
	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	// Second RunCron should be skipped because previous process is still running
	_, err = client.RunCron(addedCron.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Should still only have 1 process (second RunCron was skipped)
	processes, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	// Now complete the process
	process, err := client.Assign(env.ColonyName, 10, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, process)
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Now RunCron should work again
	_, err = client.RunCron(addedCron.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Should now have a new waiting process
	processes, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processes, 1)

	server.Shutdown()
	<-done
}

// TestAddCronWithUserAsInitiator tests that a user can create a cron (covers resolveInitiator user path)
func TestAddCronWithUserAsInitiator(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Create a user
	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "cron-user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)

	// User creates a cron
	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000 // Long interval

	addedCron, err := client.AddCron(cron, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedCron)
	assert.Equal(t, "cron-user", addedCron.InitiatorName)

	server.Shutdown()
	<-done
}

// TestGetCronNotFound tests getting a non-existent cron
func TestGetCronNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.GetCron("nonexistent-cron-id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRemoveCronNotFound tests removing a non-existent cron
func TestRemoveCronNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	err := client.RemoveCron("nonexistent-cron-id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRunCronNotFound tests running a non-existent cron
func TestRunCronNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	_, err := client.RunCron("nonexistent-cron-id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddCronIntervalZero tests that interval=0 is rejected
func TestAddCronIntervalZero(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 0 // Invalid - must be -1 or > 0
	cron.CronExpression = ""

	_, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddCronRandomWithCronExpression tests that random=true with cron expression is rejected
func TestAddCronRandomWithCronExpression(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = -1                     // Use cron expression
	cron.CronExpression = "0/1 * * * * *"  // Every second
	cron.Random = true                     // Invalid with cron expression

	_, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetCronsEmpty tests getting crons when none exist
func TestGetCronsEmpty(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	crons, err := client.GetCrons(env.ColonyName, 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, crons, 0)

	server.Shutdown()
	<-done
}

// TestGetCronsUnauthorized tests that non-members cannot get crons
func TestGetCronsUnauthorized(t *testing.T) {
	env, client, server, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to get crons from a different colony
	_, err = client.GetCrons(env.ColonyName, 100, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddCronUnauthorized tests that non-members cannot add crons
func TestAddCronUnauthorized(t *testing.T) {
	env, client, server, serverPrvKey, done := server.SetupTestEnv2(t)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to add cron to a different colony
	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	_, err = client.AddCron(cron, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRemoveCronUnauthorized tests that non-members cannot remove crons
func TestRemoveCronUnauthorized(t *testing.T) {
	env, client, server, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a cron
	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000
	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to remove cron from a different colony
	err = client.RemoveCron(addedCron.ID, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetCronUnauthorized tests that non-members cannot get cron details
func TestGetCronUnauthorized(t *testing.T) {
	env, client, server, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a cron
	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000
	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to get cron from a different colony
	_, err = client.GetCron(addedCron.ID, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRunCronUnauthorized tests that non-members cannot run crons
func TestRunCronUnauthorized(t *testing.T) {
	env, client, server, serverPrvKey, done := server.SetupTestEnv2(t)

	// Add a cron
	cron := utils.FakeCron(t, env.ColonyName, env.ExecutorID, env.ExecutorName)
	cron.Interval = 1000
	addedCron, err := client.AddCron(cron, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Create another colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Try to run cron from a different colony
	_, err = client.RunCron(addedCron.ID, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
