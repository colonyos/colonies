package generator_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	generator := utils.FakeGenerator(t, env.Colony1Name, env.Executor1ID, env.Executor1Name)

	_, err := client.AddGenerator(generator, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyName := env.Colony1Name
	generator := utils.FakeGenerator(t, colonyName, env.Executor1ID, env.Executor1Name)

	addedGenerator, err := client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetGenerator(addedGenerator.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerator(addedGenerator.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestResolveGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyName := env.Colony1Name
	generator := utils.FakeGenerator(t, colonyName, env.Executor1ID, env.Executor1Name)

	addedGenerator, err := client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.ResolveGenerator(env.Colony2Name, addedGenerator.Name, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(env.Colony1Name, addedGenerator.Name, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(env.Colony2Name, addedGenerator.Name, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.ResolveGenerator(env.Colony1Name, addedGenerator.Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetGeneratorsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyName := env.Colony1Name
	generator := utils.FakeGenerator(t, colonyName, env.Executor1ID, env.Executor1Name)

	_, err := client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetGenerators(colonyName, 100, env.Executor2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyName, 100, env.Colony1PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyName, 100, env.Colony2PrvKey)
	assert.NotNil(t, err)
	_, err = client.GetGenerators(colonyName, 100, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestAddArgGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyName := env.Colony1Name
	generator := utils.FakeGenerator(t, colonyName, env.Executor1ID, env.Executor1Name)

	addedGenerator, err := client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.PackGenerator(addedGenerator.ID, "arg", env.Executor2PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.Colony1PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.Colony2PrvKey)
	assert.NotNil(t, err)
	err = client.PackGenerator(addedGenerator.ID, "arg", env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveGeneratorSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	colonyName := env.Colony1Name
	generator := utils.FakeGenerator(t, colonyName, env.Executor1ID, env.Executor1Name)

	addedGenerator, err := client.AddGenerator(generator, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveGenerator(addedGenerator.ID, env.Executor2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveGenerator(addedGenerator.ID, env.Colony1PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveGenerator(addedGenerator.ID, env.Colony2PrvKey)
	assert.NotNil(t, err)
	err = client.RemoveGenerator(addedGenerator.ID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
