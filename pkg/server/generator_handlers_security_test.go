package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	generator := utils.FakeGenerator(t, env.colony1ID)

	_, err := client.AddGenerator(generator, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyID := env.colony1ID
	generator := utils.FakeGenerator(t, colonyID)

	addedGenerator, err := client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetGenerator(addedGenerator.ID, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestResolveGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyID := env.colony1ID
	generator := utils.FakeGenerator(t, colonyID)

	addedGenerator, err := client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.ResolveGenerator(addedGenerator.Name, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(addedGenerator.Name, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(addedGenerator.Name, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(addedGenerator.Name, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetGeneratorsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyID := env.colony1ID
	generator := utils.FakeGenerator(t, colonyID)

	_, err := client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetGenerators(colonyID, 100, env.executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyID, 100, env.colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyID, 100, env.colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyID, 100, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestAddArgGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyID := env.colony1ID
	generator := utils.FakeGenerator(t, colonyID)

	addedGenerator, err := client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.PackGenerator(addedGenerator.ID, "arg", env.executor2PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.colony1PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyID := env.colony1ID
	generator := utils.FakeGenerator(t, colonyID)

	addedGenerator, err := client.AddGenerator(generator, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteGenerator(addedGenerator.ID, env.executor2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteGenerator(addedGenerator.ID, env.colony1PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteGenerator(addedGenerator.ID, env.colony2PrvKey)
	assert.NotNil(t, err)
	err = client.DeleteGenerator(addedGenerator.ID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
