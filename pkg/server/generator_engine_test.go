package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorEngineSyncStates(t *testing.T) {
	env, _, server, _, done := setupTestEnv2(t)

	engine := createGeneratorEngine(server.db, server.controller)
	assert.Len(t, engine.states, 0)

	colonyID := env.colonyID
	generator := utils.FakeGenerator(t, colonyID)
	err := server.db.AddGenerator(generator)
	assert.Nil(t, err)

	engine.syncStatesFromDB()
	assert.Len(t, engine.states, 1)

	generatorFromEngine := engine.getGenerator(generator.ID)
	assert.True(t, generator.Equals(generatorFromEngine))

	generatorFromStates := engine.states[generator.ID]
	assert.True(t, generatorFromStates.generator.Equals(generator))

	err = server.db.DeleteGeneratorByID(generator.ID)

	engine.syncStatesFromDB()
	assert.Len(t, engine.states, 0)

	server.Shutdown()
	<-done
}

func TestGeneratorEngineTriggerByTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	engine := createGeneratorEngine(server.db, server.controller)
	assert.Len(t, engine.states, 0)

	colonyID := env.colonyID
	generator := utils.FakeGenerator(t, colonyID)
	err := server.db.AddGenerator(generator)
	assert.Nil(t, err)

	engine.syncStatesFromDB()

	engine.increaseCounter(generator.ID) // Counter needs to be at least 1 to trigger timeoput

	engine.triggerGenerators()
	time.Sleep(2 * time.Second)
	engine.triggerGenerators()

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}

func TestGeneratorEngineTriggerByCounter(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	engine := createGeneratorEngine(server.db, server.controller)
	assert.Len(t, engine.states, 0)

	colonyID := env.colonyID
	generator := utils.FakeGenerator(t, colonyID)
	err := server.db.AddGenerator(generator)
	assert.Nil(t, err)

	engine.syncStatesFromDB()

	for i := 0; i <= 10; i++ {
		engine.increaseCounter(generator.ID)
	}

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}
