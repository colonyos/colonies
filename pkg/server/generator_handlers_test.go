package server

import (
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestHandleAddGeneratorHTTPRequest(t *testing.T) {
	server, controllerMock, _, _, ctx, w := setupFakeServer()

	recoveredID := "invalid_id"
	payloadType := "invalid_payload_type"
	jsonString := "invalid json string"
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	generator := utils.FakeGenerator(t, core.GenerateRandomID())
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	msg = rpc.CreateAddGeneratorMsg(nil)
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	msg = rpc.CreateAddGeneratorMsg(generator)
	generator.WorkflowSpec = "invalid_spec"
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	colonyName := core.GenerateRandomID()
	workflowSpec := core.CreateWorkflowSpec(colonyName)
	funcSpec1 := utils.CreateTestFunctionSpec(colonyName)
	funcSpec1.NodeName = "task1"
	funcSpec2 := utils.CreateTestFunctionSpec(colonyName)
	funcSpec2.NodeName = "task2"
	funcSpec2.AddDependency("task10") // Error
	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator = core.CreateGenerator(colonyName, "test_genname"+core.GenerateRandomID(), jsonStr, 10, -1)
	msg = rpc.CreateAddGeneratorMsg(generator)
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	generator = utils.FakeGenerator(t, core.GenerateRandomID())
	controllerMock.returnError = "addGenerator"
	msg = rpc.CreateAddGeneratorMsg(generator)
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())
	controllerMock.returnError = ""

	generator = utils.FakeGenerator(t, core.GenerateRandomID())
	msg = rpc.CreateAddGeneratorMsg(generator)
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleAddGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())
}

func TestHandleGetGeneratorHTTPRequest(t *testing.T) {
	server, controllerMock, _, dbMock, ctx, w := setupFakeServer()

	recoveredID := "invalid_id"
	payloadType := "invalid_payload_type"
	jsonString := "invalid json string"
	server.handleGetGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	//generator := utils.FakeGenerator(t, core.GenerateRandomID())
	msg := rpc.CreateGetGeneratorMsg("invalid_id")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleGetGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	controllerMock.returnError = "getGenerator"
	msg = rpc.CreateGetGeneratorMsg("invalid_id")
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleGetGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())
	controllerMock.returnError = ""

	msg = rpc.CreateGetGeneratorMsg("invalid_id")
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleGetGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

	controllerMock.returnValue = "getGenerator"
	dbMock.returnError = "CountGeneratorArgs"
	msg = rpc.CreateGetGeneratorMsg("invalid_id")
	payloadType = msg.MsgType
	jsonString, err = msg.ToJSON()
	assert.Nil(t, err)
	ctx, w = getTestGinContext()
	server.handleGetGeneratorHTTPRequest(ctx, recoveredID, payloadType, jsonString)
	assertRPCError(t, w.Body.String())

}

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

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
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

	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 7)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 7)

	server.Shutdown()
	<-done
}

func TestAddGeneratorTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}

// Tests that generator worker works correctly if it run before it has been packed
// Also, tests that the generator keeps on working
func TestAddGeneratorTimeout2(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
	generator.Trigger = 10
	generator.Timeout = 1
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	// Sleep so that generator worker runs before it has been packed
	time.Sleep(1 * time.Second)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	err = client.PackGenerator(addedGenerator.ID, "arg1", env.executorPrvKey)
	assert.Nil(t, err)

	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	server.Shutdown()
	<-done
}

// This test tests a combination of trigger and timeout triggered workflows
func TestAddGeneratorTimeout3(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
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

	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 1)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	// The second workflow should be submitted by the timemout trigger
	WaitForProcessGraphs(t, client, colonyName, addedGenerator.ID, env.executorPrvKey, 2)

	graphs, err = client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 2)

	process, err := client.Assign(env.colonyName, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 10)

	process, err = client.Assign(env.colonyName, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, process.FunctionSpec.Args, 2)

	server.Shutdown()
	<-done
}

func TestGetGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
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

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
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

	colonyName := env.colonyName

	generator1 := utils.FakeGenerator(t, colonyName)
	addedGenerator1, err := client.AddGenerator(generator1, env.executorPrvKey)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyName)
	addedGenerator2, err := client.AddGenerator(generator2, env.executorPrvKey)
	assert.Nil(t, err)

	generator3 := utils.FakeGenerator(t, colonyName)
	addedGenerator3, err := client.AddGenerator(generator3, env.executorPrvKey)
	assert.Nil(t, err)

	generatorsFromServer, err := client.GetGenerators(colonyName, 100, env.executorPrvKey)
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

	generatorsFromServer, err = client.GetGenerators(colonyName, 1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromServer, 1)

	server.Shutdown()
	<-done
}

func TestRemoveGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyName := env.colonyName

	generator := utils.FakeGenerator(t, colonyName)
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	err = client.RemoveGenerator(addedGenerator.ID, env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	graphs, err := client.GetWaitingProcessGraphs(colonyName, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, len(graphs) == 0)

	server.Shutdown()
	<-done
}
