package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestSubscribeProcessesSecurity(t *testing.T) {
	_, client, server, _, done := setupTestEnv1(t)

	executorType := "test_executor_type"

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcesses(executorType, core.WAITING, 100, invalidPrivateKey)
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
	assert.NotNil(t, err) // Should not work, we should have got an error "executor not found"

	server.Shutdown()
	<-done
}

func TestSubscribeChangeStateProcessSecurity(t *testing.T) {
	_, client, server, _, done := setupTestEnv1(t)

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcess(core.GenerateRandomID(), "test_executor_type", core.WAITING, 100, invalidPrivateKey)
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
	assert.NotNil(t, err) // Should not work, we should have got an error "executor not found"

	server.Shutdown()
	<-done
}
