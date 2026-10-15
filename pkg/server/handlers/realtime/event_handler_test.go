package realtime_test

import (
	"context"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	backendGin "github.com/colonyos/colonies/pkg/backends/gin"
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

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.Signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerTimeout(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerStop(t *testing.T) {
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)

	handler.Stop()

	stopped := false
	for {
		stopped = handler.HasStopped()
		if stopped {
			break
		}
	}
	assert.True(t, stopped)
}

func TestEventHandlerTimeout2(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type2" // NOTE: we are Signaling to another target
		process.State = core.WAITING
		handler.Signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerTimeout3(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		process := utils.CreateTestProcess(core.GenerateRandomID())
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
		process.State = core.RUNNING // NOTE: we are Signaling to another target
		handler.Signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
	assert.Equal(t, allListeners, 0)
	assert.Equal(t, listeners, 0)
	assert.Equal(t, processIDs, 0)
}

func TestEventHandlerMany(t *testing.T) {
	// Test that multiple listeners waiting for a SPECIFIC processID all receive the signal.
	// This tests Pass 1 (broadcast) of the thundering herd prevention mechanism.
	// Note: General listeners (empty processID) only wake ONE at a time via Pass 2.
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 2000*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)

	var retChans []chan retValues

	for i := 0; i < 10; i++ {
		retChan := make(chan retValues)
		retChans = append(retChans, retChan)
		go func() {
			// Use specific processID so all listeners receive the signal (Pass 1 broadcast)
			process, err := handler.WaitForProcess("test_executor_type", core.WAITING, process.ID, "", ctx)
			retChan <- retValues{process: process, err: err}
		}()
	}

	go func() {
		handler.WaitForProcess("test_executor_type2", core.WAITING, "", "", ctx)
	}()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		handler.Signal(process)
	}()

	// Wait for listeners - all should receive the signal since they're waiting for specific processID
	for _, retChan := range retChans {
		retVal := <-retChan
		assert.Nil(t, retVal.err) // OK
		assert.True(t, process.Equals(retVal.process))
	}
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
	assert.Equal(t, 1, allListeners) // Note 1, as test_executor_type2 was never Signaled
	assert.Equal(t, 0, listeners)
	assert.Equal(t, 0, processIDs)
}

func TestEventHandlerThunderingHerdPrevention(t *testing.T) {
	// Test that general listeners (empty processID) only wake ONE at a time.
	// This tests Pass 2 (single wake-up) of the thundering herd prevention mechanism.
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)

	// Create 5 general listeners (empty processID)
	successCount := int32(0)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := handler.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)

	// Signal once - only ONE listener should wake up
	handler.Signal(process)

	// Wait for all goroutines to finish (either success or timeout)
	wg.Wait()

	// Exactly ONE should have succeeded (thundering herd prevention)
	assert.Equal(t, int32(1), successCount)
}

func TestEventHandlerUpdate(t *testing.T) {
	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING

	ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
	defer cancelCtx()

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, process.ID, "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.Signal(process)
	}()
	retVal := <-retChan
	assert.Nil(t, retVal.err) // OK
	assert.True(t, process.Equals(retVal.process))
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
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

	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	retChan := make(chan retValues)
	go func() {
		// Wait for another process with random ID, i.e. we will time out
		process, err := handler.WaitForProcess("test_executor_type", core.WAITING, core.GenerateRandomID(), "", ctx)
		retChan <- retValues{process: process, err: err}
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		handler.Signal(process)
	}()
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // OK
	allListeners, listeners, processIDs := handler.NumberOfListeners("test_executor_type", core.WAITING, "")
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
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	processChan, errChan := handler.Subscribe("test_executor_type", core.WAITING, "", "", ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.Signal(process)
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
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	processChan, errChan := handler.Subscribe("test_executor_type", core.WAITING, "", "", ctx) // Subscribe to all processes
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
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	processChan, errChan := handler.Subscribe("test_executor_type", core.WAITING, "", "", ctx) // Subscribe to all processes
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
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	processChan, errChan := handler.Subscribe("test_executor_type", core.WAITING, process.ID, "", ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.Signal(process)
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
	factory := backendGin.NewFactory()
	handler := factory.CreateTestableEventHandler(nil)
	processChan, errChan := handler.Subscribe("test_executor_type", core.WAITING, core.GenerateRandomID(), "", ctx)
	go func() {
		select {
		case err := <-errChan:
			retChan <- retValues{process: nil, err: err}
		case process := <-processChan:
			retChan <- retValues{process: process, err: nil}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	handler.Signal(process)
	retVal := <-retChan
	assert.NotNil(t, retVal.err) // Not OK, will timeout
}

// waitForRelayServers blocks until every node's relay port accepts
// connections. The relay servers listen in a goroutine, so signalling before
// they are bound would silently lose the broadcast.
func waitForRelayServers(t *testing.T, nodes ...cluster.Node) {
	t.Helper()
	for _, node := range nodes {
		addr := net.JoinHostPort(node.Host, strconv.Itoa(node.RelayPort))
		assert.Eventually(t, func() bool {
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err != nil {
				return false
			}
			conn.Close()
			return true
		}, 10*time.Second, 20*time.Millisecond, "relay server %s did not start", node.Name)
	}
}

func TestEventHandleRelayServer(t *testing.T) {
	// Kernel-assigned ports keep this test from colliding with other test
	// processes and with earlier iterations of itself in the same process
	reserved := utils.ReservePortsOrPanic(12)
	nodes := make([]cluster.Node, 3)
	for i, name := range []string{"etcd1", "etcd2", "etcd3"} {
		nodes[i] = cluster.Node{
			Name:           name,
			Host:           "localhost",
			EtcdClientPort: reserved[4*i].Port(),
			EtcdPeerPort:   reserved[4*i+1].Port(),
			RelayPort:      reserved[4*i+2].Port(),
			APIPort:        reserved[4*i+3].Port(),
		}
	}
	node1, node2, node3 := nodes[0], nodes[1], nodes[2]

	config := cluster.Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	utils.ReleasePorts(reserved)
	relayServer1 := cluster.CreateRelayServer(node1, config)
	relayServer2 := cluster.CreateRelayServer(node2, config)
	relayServer3 := cluster.CreateRelayServer(node3, config)
	defer relayServer1.Shutdown()
	defer relayServer2.Shutdown()
	defer relayServer3.Shutdown()
	waitForRelayServers(t, node1, node2, node3)

	factory1 := backendGin.NewFactory()
	handler1 := factory1.CreateTestableEventHandler(relayServer1)
	factory2 := backendGin.NewFactory()
	handler2 := factory2.CreateTestableEventHandler(relayServer2)
	factory3 := backendGin.NewFactory()
	handler3 := factory3.CreateTestableEventHandler(relayServer3)

	retChan1 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler1.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan1 <- retValues{process: process, err: err}
	}()

	retChan2 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler2.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan2 <- retValues{process: process, err: err}
	}()

	retChan3 := make(chan retValues)
	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3000*time.Millisecond)
		defer cancelCtx()
		process, err := handler3.WaitForProcess("test_executor_type", core.WAITING, "", "", ctx)
		retChan3 <- retValues{process: process, err: err}
	}()

	time.Sleep(100 * time.Millisecond)

	process := utils.CreateTestProcess(core.GenerateRandomID())
	process.FunctionSpec.Conditions.ExecutorType = "test_executor_type"
	process.State = core.WAITING
	go func() {
		handler3.Signal(process)
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
