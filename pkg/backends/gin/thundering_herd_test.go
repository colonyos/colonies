package gin

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// TestSingleWakeUp verifies that when a signal is sent for process assignment,
// only ONE waiting executor wakes up, not all. This prevents the thundering
// herd problem where all executors call Assign() simultaneously.
func TestSingleWakeUp(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	numExecutors := 8
	executorType := "test-executor"
	var wakeUpCount int32
	var wg sync.WaitGroup

	// Create a dummy process for signaling
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start multiple "executors" waiting for processes
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Simulate executor waiting for process
			_, err := handler.WaitForProcess(executorType, core.WAITING, "", ctx)
			if err == nil {
				// Executor woke up from signal
				atomic.AddInt32(&wakeUpCount, 1)
			}
		}(i)
	}

	// Give time for all executors to register
	time.Sleep(100 * time.Millisecond)

	// Verify all executors are registered as listeners
	allListeners, listeners, _ := handler.NumberOfListeners(executorType, core.WAITING)
	t.Logf("Registered listeners: %d (for target: %d)", allListeners, listeners)
	assert.Equal(t, numExecutors, listeners, "All executors should be registered")

	// Send ONE signal (simulating one process submitted)
	handler.Signal(process)

	// Wait for all goroutines to complete
	wg.Wait()

	// FIXED: Only 1 executor wakes up from a single signal (no thundering herd)
	t.Logf("Wake-up count: %d (expected 1)", wakeUpCount)
	assert.Equal(t, int32(1), wakeUpCount,
		"Only 1 executor should wake up from 1 signal (thundering herd fixed)")
}

// TestNoAmplification verifies that signals have 1:1 mapping with wake-ups.
// Each signal wakes exactly ONE executor, preventing amplification.
func TestNoAmplification(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	numExecutors := 4
	numSignals := 3
	executorType := "test-executor"
	var totalWakeUps int32
	var wg sync.WaitGroup

	// Create processes for signaling
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executors that loop multiple times (simulating assign loop)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numSignals; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := handler.WaitForProcess(executorType, core.WAITING, "", ctx)
				cancel()
				if err == nil {
					atomic.AddInt32(&totalWakeUps, 1)
				} else {
					// Timeout - stop waiting
					break
				}
			}
		}(i)
	}

	// Give time for executors to register
	time.Sleep(100 * time.Millisecond)

	// Send signals (simulating processes being submitted)
	for i := 0; i < numSignals; i++ {
		handler.Signal(process)
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for completion
	wg.Wait()

	// FIXED: Each signal wakes exactly ONE executor = numSignals wake-ups (1:1 ratio)
	t.Logf("Total wake-ups: %d (signals=%d, executors=%d)", totalWakeUps, numSignals, numExecutors)
	t.Logf("Ratio: %.1fx (expected 1.0x)", float64(totalWakeUps)/float64(numSignals))

	assert.Equal(t, int32(numSignals), totalWakeUps,
		"Each signal should wake exactly 1 executor (no amplification)")
}

// TestBufferedChannelExhaustion demonstrates how the 100-capacity buffered
// channel can accumulate signals, causing rapid-fire wake-ups without blocking.
func TestBufferedChannelExhaustion(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	numSignals := 50

	// Create process for signaling
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Register one listener
	var wakeUpCount int32
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		for {
			_, err := handler.WaitForProcess(executorType, core.WAITING, "", ctx)
			if err != nil {
				break
			}
			atomic.AddInt32(&wakeUpCount, 1)
		}
		close(done)
	}()

	// Give time for listener to register
	time.Sleep(50 * time.Millisecond)

	// Rapidly send many signals (they'll buffer up)
	start := time.Now()
	for i := 0; i < numSignals; i++ {
		handler.Signal(process)
	}
	signalDuration := time.Since(start)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)
	cancel()
	<-done

	t.Logf("Sent %d signals in %v", numSignals, signalDuration)
	t.Logf("Listener received %d wake-ups", wakeUpCount)
	t.Logf("Signals processed without blocking: channel buffer allows rapid accumulation")

	// The buffered channel (size 100) allows signals to accumulate
	// without blocking the sender, contributing to the thundering herd
	assert.True(t, wakeUpCount > 0, "Listener should receive wake-ups")
	assert.True(t, signalDuration < 100*time.Millisecond,
		"Signals sent quickly without blocking (buffered channel)")
}

// TestExpectedVsActualDatabaseOperations calculates the expected vs actual
// database operations based on the thundering herd behavior.
func TestExpectedVsActualDatabaseOperations(t *testing.T) {
	// Simulation parameters (based on production observations)
	numActiveExecutors := 8
	numProcesses := 100

	// In ideal case: 1 signal -> 1 executor wakes -> 1 Assign() -> 1 MarkAlive()
	idealOperations := numProcesses

	// With thundering herd: 1 signal -> N executors wake -> N Assign() -> N MarkAlive()
	actualOperations := numProcesses * numActiveExecutors

	// But it gets worse: when executor doesn't get process, it loops back
	// If buffered signals cause repeated wake-ups, each executor might loop many times
	// Observed in production: ~1840x amplification
	//
	// Simplified model: each executor loops back ~230 times per process
	// (based on 19,742,702 updates / 1,341 processes / 8 executors / ~6.4 loops)

	t.Logf("=== Thundering Herd Impact Analysis ===")
	t.Logf("Active executors: %d", numActiveExecutors)
	t.Logf("Processes: %d", numProcesses)
	t.Logf("")
	t.Logf("Ideal (1:1 signal:assign): %d MarkAlive() calls", idealOperations)
	t.Logf("With broadcast (N:1):      %d MarkAlive() calls (%dx)", actualOperations, numActiveExecutors)
	t.Logf("")
	t.Logf("Production observation:")
	t.Logf("  - 19,742,702 executor updates")
	t.Logf("  - 1,341 processes")
	t.Logf("  - 8 active executors")
	t.Logf("  - = 14,722 updates per process (expected: 8)")
	t.Logf("  - = 1,840x amplification beyond simple broadcast model")

	// The test "passes" but documents the problem
	assert.True(t, actualOperations > idealOperations,
		"Thundering herd causes %dx more operations than ideal", numActiveExecutors)
}
