package server

import (
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TEST ERROR
// time="2023-02-22T22:06:06+01:00" level=error msg="Failed to iterate processgraph, process is nil"
//
//	test_utils.go:375:
//	    	Error Trace:	test_utils.go:375
//	    	            				generator_handlers_test.go:33
//	    	Error:      	Expected nil, but got: &errors.errorString{s:"Failed to iterate processgraph, process is nil"}
//	    	Test:       	TestAddGeneratorCounter
func TestAddGeneratorCounter(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.Trigger = 10
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	doneInc := make(chan bool)
	go func() {
		for i := 0; i < 73; i++ {
			err = client.PackGenerator(addedGenerator.ID, "arg"+strconv.Itoa(i), env.executorPrvKey)
			assert.Nil(t, err)
		}
		doneInc <- true
	}()
	<-doneInc

	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 7)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 7)

	server.Shutdown()
	<-done
}

func TestAddGeneratorTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}

// Tests that generator worker works correctly if it run before it has been packed
// Also, tests that the generator keeps on working
func TestAddGeneratorTimeout2(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Sleep so that generator worker runs before it has been packed
	time.Sleep(1 * time.Second)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	server.Shutdown()
	<-done
}

// This test tests a combination of trigger and timeout triggered workflows
func TestAddGeneratorTimeout3(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Since we are packing 12 args and the trigger is set to 10, a workflow should immediately by submitted
	for i := 0; i < 12; i++ {
		err = client.PackGenerator(addedGenerator.ID, "arg", env.executorPrvKey)
		assert.Nil(t, err)
	}

	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// The second workflow should be submitted by the timemout trigger
	WaitForProcessGraphs(t, client, colonyID, addedGenerator.ID, env.executorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 10)

	process, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 2)

	server.Shutdown()
	<-done
}

func TestGetGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)

	generatorFromServer, err := client.GetGenerator(addedGenerator.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, generatorFromServer.ID, addedGenerator.ID)

	server.Shutdown()
	<-done
}

func TestResolveGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	generator.Name = "test_generator_name"
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)

	generatorFromServer, err := client.ResolveGenerator("test_generator_name", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, generatorFromServer.ID, addedGenerator.ID)

	server.Shutdown()
	<-done
}

func TestGetGenerators(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator1 := utils.FakeGenerator(t, colonyID)
	addedGenerator1, err := client.AddGenerator(generator1, env.executorPrvKey)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyID)
	addedGenerator2, err := client.AddGenerator(generator2, env.executorPrvKey)
	assert.Nil(t, err)

	generator3 := utils.FakeGenerator(t, colonyID)
	addedGenerator3, err := client.AddGenerator(generator3, env.executorPrvKey)
	assert.Nil(t, err)

	generatorsFromServer, err := client.GetGenerators(colonyID, 100, env.executorPrvKey)
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

	generatorsFromServer, err = client.GetGenerators(colonyID, 1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromServer, 1)

	server.Shutdown()
	<-done
}

func TestDeleteGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGenerator(t, colonyID)
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.DeleteGenerator(addedGenerator.ID, env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, len(graphs) == 0)

	server.Shutdown()
	<-done
}
