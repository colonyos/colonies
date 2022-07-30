package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// Runtime 2 subscribes on process events and expects to receive an event when a new process is submitted
// Runtime 1 submitts a new process
// Runtime 2 receives an event
func TestSubscribeProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	runtimeType := "test_runtime_type"

	subscription, err := client.SubscribeProcesses(runtimeType, core.WAITING, 100, env.runtime2PrvKey)
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

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err = client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Runtime 1 submits a process
// Runtime 2 subscribes on process events and expects to receive an event when the process finishes.
// Runtime 1 gets assign the process
// Runtime 1 finish the process
// Runtime 2 receives an event
func TestSubscribeChangeStateProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(addedProcess.ID,
		addedProcess.ProcessSpec.Conditions.RuntimeType,
		core.SUCCESS,
		100,
		env.runtime2PrvKey)
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

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(assignedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Let change the order of the operations a bit, what about if the subscriber subscribes on an
// process state change event, but that event has already occurred. Then, the subscriber would what forever.
// The solution is to let the server send an event anyway if the wanted state is true already.
//
// Runtime 1 submits a process
// Runtime 1 gets assign the process
// Runtime 1 finish the process
// Runtime 2 subscribes on process events and expects to receive an event when the process finishes.
// Runtime 2 receives an event
func TestSubscribeChangeStateProcess2(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(assignedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	subscription, err := client.SubscribeProcess(addedProcess.ID,
		addedProcess.ProcessSpec.Conditions.RuntimeType,
		core.SUCCESS,
		100,
		env.runtime2PrvKey)
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
