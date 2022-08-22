package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGeneratorTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.WorkflowSpec = "error"
	addedGenerator, err := client.AddGenerator(generator, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, addedGenerator)

	generator = utils.FakeGenerator(t, colonyID)
	addedGenerator, err = client.AddGenerator(generator, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	doneInc := make(chan bool)
	go func() {
		for i := 0; i < 15; i++ {
			err = client.AddArgToGenerator(addedGenerator.ID, "arg", env.runtimePrvKey)
			assert.Nil(t, err)
			time.Sleep(500 * time.Millisecond)

		}
		doneInc <- true
	}()
	<-doneInc

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, len(graphs) > 2)

	server.Shutdown()
	<-done
}

func TestAddGeneratorCounter(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	// trigger = 10 and timeout = -1
	generator := utils.FakeGenerator(t, colonyID)
	generator.Timeout = -1
	addedGenerator, err := client.AddGenerator(generator, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	doneInc := make(chan bool)
	go func() {
		for i := 0; i < 70; i++ {
			err = client.AddArgToGenerator(addedGenerator.ID, "arg", env.runtimePrvKey)
			assert.Nil(t, err)
		}
		doneInc <- true
	}()
	<-doneInc

	time.Sleep(20 * time.Second)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 3)

	server.Shutdown()
	<-done
}

func TestGetGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	addedGenerator, err := client.AddGenerator(generator, env.runtimePrvKey)
	assert.Nil(t, err)

	generatorFromServer, err := client.GetGenerator(addedGenerator.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, generatorFromServer.ID, addedGenerator.ID)

	server.Shutdown()
	<-done
}

func TestGetGenerators(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator1 := utils.FakeGenerator(t, colonyID)
	addedGenerator1, err := client.AddGenerator(generator1, env.runtimePrvKey)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyID)
	addedGenerator2, err := client.AddGenerator(generator2, env.runtimePrvKey)
	assert.Nil(t, err)

	generator3 := utils.FakeGenerator(t, colonyID)
	addedGenerator3, err := client.AddGenerator(generator3, env.runtimePrvKey)
	assert.Nil(t, err)

	generatorsFromServer, err := client.GetGenerators(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromServer, 3)

	generator1Found := false
	generator2Found := false
	generator3Found := false
	for _, generator := range generatorsFromServer {
		if generator.ID == addedGenerator1.ID {
			generator1Found = true
		}
		if generator.ID == addedGenerator2.ID {
			generator2Found = true
		}
		if generator.ID == addedGenerator3.ID {
			generator3Found = true
		}
	}

	assert.True(t, generator1Found)
	assert.True(t, generator2Found)
	assert.True(t, generator3Found)

	generatorsFromServer, err = client.GetGenerators(colonyID, 1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromServer, 1)

	server.Shutdown()
	<-done
}

func TestDeleteGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	addedGenerator, err := client.AddGenerator(generator, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.AddArgToGenerator(addedGenerator.ID, "arg", env.runtimePrvKey)
	assert.Nil(t, err)

	err = client.DeleteGenerator(addedGenerator.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, len(graphs) == 0)

	server.Shutdown()
	<-done
}
