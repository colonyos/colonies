package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// Executor 2 subscribes on process events and expects to receive an event when a new process is submitted
// Executor 1 submitts a new process
// Executor 2 receives an event
func TestSubscribeProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	executorType := "test_executor_type"

	subscription, err := client.SubscribeProcesses(executorType, core.WAITING, 100, env.executor2PrvKey)
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

	funcSpec := utils.CreateTestFunctionSpec(env.colony1ID)
	_, err = client.Submit(funcSpec, env.executor1PrvKey)
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
	env, client, server, _, done := setupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colony1ID)
	addedProcess, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(addedProcess.ID,
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.executor2PrvKey)
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

	assignedProcess, err := client.Assign(env.colony1ID, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
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
	env, client, server, _, done := setupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colony1ID)
	addedProcess, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.WAITING, addedProcess.State)

	assignedProcess, err := client.Assign(env.colony1ID, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	subscription, err := client.SubscribeProcess(addedProcess.ID,
		addedProcess.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.executor2PrvKey)
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
