package realtime_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// Executor 2 subscribes on process events and expects to receive an event when a new process is submitted
// Executor 1 submitts a new process
// Executor 2 receives an event
func TestSubscribeProcesses(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	executorType := "test_executor_type"

	subscription, err := client.SubscribeProcesses(env.Colony1Name, executorType, core.WAITING, 100, env.Executor1PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		}
	}()

	time.Sleep(1 * time.Second)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err = client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Executor 1 submits a process
// Executor 2 subscribes on process events and expects to receive an event when the process finishes.
// Executor 1 gets assign the process
// Executor 1 finish the process
// Executor 2 receives an event
func TestSubscribeChangeStateProcess(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(env.Colony1Name,
		addedProcess.ID,
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.Executor1PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			fmt.Println(err)
			waitForProcess <- err
		}
	}()

	assignedProcess, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Test subscribing to RUNNING state - subscriber should receive event when process is assigned
// Executor 1 submits a process
// Executor 2 subscribes on RUNNING state events
// Executor 1 gets assigned the process
// Executor 2 receives an event when process transitions to RUNNING
func TestSubscribeChangeStateProcessRunning(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(env.Colony1Name,
		addedProcess.ID,
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.RUNNING,
		100,
		env.Executor2PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	var receivedProcess *core.Process
	go func() {
		select {
		case p := <-subscription.ProcessChan:
			receivedProcess = p
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			fmt.Println(err)
			waitForProcess <- err
		}
	}()

	// Assign the process - this should trigger the RUNNING state notification
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	assert.NotNil(t, receivedProcess)
	assert.Equal(t, core.RUNNING, receivedProcess.State)
	server.Shutdown()
	<-done
}

func TestSubscribeChangeStateProcessInvalidID(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(env.Colony1Name,
		"invalid_id",
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.Executor2PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			fmt.Println(err)
			waitForProcess <- err
		}
	}()

	err = <-waitForProcess
	assert.NotNil(t, err)
	server.Shutdown()
	<-done
}

// Let change the order of the operations a bit, what about if the subscriber subscribes on an
// process state change event, but that event has already occurred? Then, the subscriber would what forever.
// The solution is to let the server send an event anyway if the wanted state is true already.
//
// Executor 1 submits a process
// Executor 1 gets assign the process
// Executor 1 finish the process
// Executor 2 subscribes on process events and expects to receive an event when the process finishes.
// Executor 2 receives an event
func TestSubscribeChangeStateProcess2(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.WAITING, addedProcess.State)

	assignedProcess, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	subscription, err := client.SubscribeProcess(env.Colony1Name,
		addedProcess.ID,
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.Executor1PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		}
	}()

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}
