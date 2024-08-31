package server

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
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
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
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
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
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
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type2" // NOTE: we are signaling to another target
		process.State = core.WAITING
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
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
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
		process.State = core.RUNNING // NOTE: we are signaling to another target
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerMany(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)

	var retChans []chan retValues

	for i := 0; i < 10; i++ {
		retChan := make(chan retValues)
		retChans = append(retChans, retChan)
		go func() {
			process, err := handler.waitForProcess("test_executor_type", core.WAITING, "", ctx)
			retChan <- retValues{process: process, err: err}
		}()
	}

	go func() {
		handler.waitForProcess("test_executor_type2", core.WAITING, "", ctx)
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
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
	assert.Equal(t, allListeners, 1) // Note 1, as test_executor_type2 was never signaled
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerUpdate(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, process.ID, ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerUpdateTimeout(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	handler := createEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		// Wait for another process with random ID, i.e. we will time out
		process, err := handler.waitForProcess("test_executor_type", core.WAITING, core.GenerateRandomID(), ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // OK
	allListeners, listeners, processIDs := handler.numberOfListeners("test_executor_type", core.WAITING)
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerSubscribe(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_executor_type", core.WAITING, "", ctx)
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
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_executor_type", core.WAITING, "", ctx) // Subscribe to all processes
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
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_executor_type", core.WAITING, "", ctx) // Subscribe to all processes
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
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_executor_type", core.WAITING, process.ID, ctx)
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
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	retChan := make(chan retValues)
	handler := createEventHandler(nil)
	processChan, errChan := handler.subscribe("test_executor_type", core.WAITING, core.GenerateRandomID(), ctx)
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

func TestEventHandleRelayServer(t *testing.T) {
	node1 := cluster.Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := cluster.Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := cluster.Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := cluster.Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterServer1 := cluster.CreateClusterServer(node1, config, ".")
	clusterServer2 := cluster.CreateClusterServer(node2, config, ".")
	clusterServer3 := cluster.CreateClusterServer(node3, config, ".")

	defer clusterServer1.Shutdown()
	defer clusterServer2.Shutdown()
	defer clusterServer3.Shutdown()

	relay1 := clusterServer1.Relay()
	relay2 := clusterServer2.Relay()
	relay3 := clusterServer3.Relay()

	handler1 := createEventHandler(relay1)
	handler2 := createEventHandler(relay2)
	handler3 := createEventHandler(relay3)

	retChan1 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler1.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan1 <- retValues{process: process, err: err}
	}()

	retChan2 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler2.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan2 <- retValues{process: process, err: err}
	}()

	retChan3 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler3.waitForProcess("test_executor_type", core.WAITING, "", ctx)
		retChan3 <- retValues{process: process, err: err}
	}()

	time.Sleep(100 * time.Millisecond)

	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	go func() {
		handler3.signal(process)
	}()

	retVal1 := <-retChan1
	assert.True(t, process.Equals(retVal1.process))
	assert.Nil(t, retVal1.err)
	retVal2 := <-retChan2
	assert.Nil(t, retVal2.err)
	assert.True(t, process.Equals(retVal2.process))
	retVal3 := <-retChan3
	assert.Nil(t, retVal3.err)
	assert.True(t, process.Equals(retVal3.process))
}
