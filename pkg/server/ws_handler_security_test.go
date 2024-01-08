package server

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	executorType := "test_executor_type"

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcesses(env.colony1Name, executorType, core.WAITING, 100, invalidPrivateKey)
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
	fmt.Println(err)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestSubscribeChangeStateProcessSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcess(env.colony1Name, core.GenerateRandomID(), "test_executor_type", core.WAITING, 100, invalidPrivateKey)
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
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestSubscribeProcessSecurityInvalidProcessID(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colony1Name)
	_, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)

	funcSpec = utils.CreateTestFunctionSpec(env.colony2Name)
	addedProcess2, err := client.Submit(funcSpec, env.executor2PrvKey)
	assert.Nil(t, err)

	// Executor1 is member of colony1 and executor2 is member of colony2
	// Both executors are valid members of their respective colonies,
	// it should only be possible to subscribe to process of the same colony as the executor
	subscription, err := client.SubscribeProcess(env.colony1Name,
		addedProcess2.ID,
		addedProcess2.FunctionSpec.Conditions.ExecutorType,
		core.SUCCESS,
		100,
		env.executor1PrvKey)
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

	assignedProcess, err := client.Assign(env.colony1Name, -1, "", "", env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(assignedProcess.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.NotNil(t, err) // Should not work
	server.Shutdown()
	<-done
}
