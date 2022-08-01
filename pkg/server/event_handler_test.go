package server

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type retValues struct {
	process *core.Process
	err     error
}

func TestEventHandler(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerTimeout(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerStop(t *testing.T) {
	handler := createEventHandler(nil)

	handler.stop()

	stopped := false
	for {
		stopped = handler.hasStopped()
		if stopped {
			break
		}
	}
	assert.True(t, stopped)
}

func TestEventHandlerTimeout2(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type2" // NOTE: we are signaling to another target
		process.State = core.WAITING
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerTimeout3(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
		process.State = core.RUNNING // NOTE: we are signaling to another target
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerMany(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)

	var retChans []chan retValues

	for i := 0; i < 10; i++ {
		retChan := make(chan retValues)
		retChans = append(retChans, retChan)
		go func() {
			process, err := handler.waitForProcess("test_runtime_type", core.WAITING, "", ctx)
			retChan <- retValues{process: process, err: err}
		}()
	}

	go func() {
		handler.waitForProcess("test_runtime_type2", core.WAITING, "", ctx)
	}()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		handler.signal(process)
	}()

	// Wait for listers
	for _, retChan := range retChans {
		retVal := <-retChan
		assert.Nil(t, retVal.err) // OK
		assert.True(t, process.Equals(retVal.process))
	}
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 1) // Note 1, as test_runtime_type2 was never signaled
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerUpdate(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, process.ID, ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerUpdateTimeout(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		// Wait for another process with random ID, i.e. we will time out
		process, err := handler.waitForProcess("test_runtime_type", core.WAITING, core.GenerateRandomID(), ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // OK
	allListeners, listeners, processIDs := handler.numberOfListeners("test_runtime_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerSubscribe(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_runtime_type", core.WAITING, "", ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.signal(process)
	retVal := <-retChan
	assert.Nil(t, retVal.err)
	assert.True(t, retVal.process.Equals(process))
}

func TestEventHandlerSubscribeCancel(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_runtime_type", core.WAITING, "", ctx) // Subscribe to all processes
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	cancelCtx()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Timeout
}

func TestEventHandlerSubscribeTimeout(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_runtime_type", core.WAITING, "", ctx) // Subscribe to all processes
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Timeout
}

func TestEventHandlerSubscribeProcessID(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_runtime_type", core.WAITING, process.ID, ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.signal(process)
	retVal := <-retChan
	assert.Nil(t, retVal.err)
	assert.True(t, retVal.process.Equals(process))
}

func TestEventHandlerSubscribeProcessIDFailed(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.ProcessSpec.Conditions.RuntimeType = "test_runtime_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_runtime_type", core.WAITING, core.GenerateRandomID(), ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.signal(process)
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
}
