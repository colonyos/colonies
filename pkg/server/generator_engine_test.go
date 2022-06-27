package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorEngineSyncStates(t *testing.T) {
	env, _, server, _, done := setupTestEnv2(t)

	engine := createGeneratorEngine(server.db, server)
	assert.Len(t, engine.states, 0)

	colonyID := env.colonyID
	generator := utils.FakeGenerator(t, colonyID)
	err := server.db.AddGenerator(generator)
	assert.Nil(t, err)

	engine.syncStatesFromDB()
	assert.Len(t, engine.states, 1)

	generatorFromStates := engine.states[generator.ID]
	assert.True(t, generatorFromStates.generator.Equals(generator))

	err = server.db.DeleteGeneratorByID(generator.ID)

	engine.syncStatesFromDB()
	assert.Len(t, engine.states, 0)

	server.Shutdown()
	<-done
}

func TestGeneratorEngineTriggerOnCreate(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	engine := createGeneratorEngine(server.db, server)
	assert.Len(t, engine.states, 0)

	colonyID := env.colonyID
	generator := utils.FakeGenerator(t, colonyID)
	err := server.db.AddGenerator(generator)
	assert.Nil(t, err)

	engine.syncStatesFromDB()

	graphs, err := client.GetWaitingProcessGraphs(colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, graphs, 1)

	server.Shutdown()
	<-done
}
