package controllers

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestPackGeneratorErrorPaths(t *testing.T) {
	controller, dbMock := createFakeColoniesController(t)
	defer controller.Stop()

	for _, failing := range []string{"AddGeneratorArg", "SetGeneratorFirstPack"} {
		dbMock.ReturnError = failing
		for i := 0; i < 3; i++ {
			done := make(chan error, 1)
			go func() { done <- controller.PackGenerator("gen_id", "colony", "arg") }()
			select {
			case err := <-done:
				assert.NotNil(t, err, failing)
			case <-time.After(5 * time.Second):
				t.Fatalf("PackGenerator blocked after %s failure", failing)
			}
		}
	}

	dbMock.ReturnError = ""
	assert.Nil(t, controller.PackGenerator("gen_id", "colony", "arg"))
	_, err := controller.GetGenerators("colony", 10)
	assert.Nil(t, err)
}

func TestPackGeneratorSetsFirstPackOnce(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_PACK_FIRSTPACK")
	assert.Nil(t, err)
	defer db.Close()

	controller := createTestColoniesController(t, db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()
	generator := utils.FakeGeneratorSingleProcess(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	generator.Trigger = 1000
	assert.Nil(t, db.AddGenerator(generator))

	assert.Nil(t, controller.PackGenerator(generator.ID, colonyName, "arg1"))
	generatorFromDB, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	firstPack := generatorFromDB.FirstPack
	assert.True(t, firstPack.Unix() > 0)

	time.Sleep(1100 * time.Millisecond)
	assert.Nil(t, controller.PackGenerator(generator.ID, colonyName, "arg2"))
	generatorFromDB, err = db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.True(t, generatorFromDB.FirstPack.Equal(firstPack))

	count, err := db.CountGeneratorArgs(generator.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestPackGeneratorNotBlockedByQueue(t *testing.T) {
	controller, _ := createFakeColoniesController(t)
	defer controller.Stop()

	release := make(chan struct{})
	started := make(chan struct{})
	controller.cmdQueue <- &command{handler: func(cmd *command) {
		close(started)
		<-release
	}}
	<-started
	defer close(release)

	done := make(chan error, 3)
	go func() { done <- controller.PackGenerator("gen_id", "colony", "arg") }()
	go func() { _, err := controller.GetGenerator("gen_id"); done <- err }()
	go func() { _, err := controller.ResolveGenerator("colony", "name"); done <- err }()
	for i := 0; i < 3; i++ {
		select {
		case err := <-done:
			assert.Nil(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("generator call blocked behind busy cmdQueue worker")
		}
	}
}

func TestPackGeneratorConcurrentWithTrigger(t *testing.T) {
	db, err := postgresql.PrepareTestsWithPrefix("TEST_PACK_CONCURRENT")
	assert.Nil(t, err)
	defer db.Close()

	controller := createTestColoniesController(t, db)
	defer controller.Stop()

	colonyName := core.GenerateRandomID()
	executor, _, err := utils.CreateTestExecutorWithKey(colonyName)
	assert.Nil(t, err)
	assert.Nil(t, db.AddExecutor(executor))

	generator := utils.FakeGeneratorSingleProcess(t, colonyName, executor.ID, executor.Name)
	generator.ID = core.GenerateRandomID()
	generator.Trigger = 5
	assert.Nil(t, db.AddGenerator(generator))

	const packs = 103
	var wg sync.WaitGroup
	for i := 0; i < packs; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			assert.Nil(t, controller.PackGenerator(generator.ID, colonyName, "arg"+strconv.Itoa(i)))
			if i%10 == 0 {
				controller.TriggerGenerators()
			}
		}(i)
	}
	wg.Wait()

	controller.TriggerGenerators()
	_, err = controller.GetGenerators(colonyName, 10)
	assert.Nil(t, err)

	remaining, err := db.CountGeneratorArgs(generator.ID)
	assert.Nil(t, err)
	assert.Less(t, remaining, generator.Trigger)

	processes, err := db.FindWaitingProcesses(colonyName, "test_executor_type", "", "", 1000)
	assert.Nil(t, err)

	seen := make(map[string]int)
	for _, process := range processes {
		assert.Len(t, process.FunctionSpec.Args, generator.Trigger)
		for _, arg := range process.FunctionSpec.Args {
			seen[arg.(string)]++
		}
	}
	for arg, n := range seen {
		assert.Equal(t, 1, n, "arg consumed more than once: "+arg)
	}
	assert.Equal(t, packs, len(seen)+remaining)
}
