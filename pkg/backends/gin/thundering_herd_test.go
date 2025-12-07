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
			_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
			if err == nil {
				// Executor woke up from signal
				atomic.AddInt32(&wakeUpCount, 1)
			}
		}(i)
	}

	// Give time for all executors to register
	time.Sleep(100 * time.Millisecond)

	// Verify all executors are registered as listeners
	allListeners, listeners, _ := handler.NumberOfListeners(executorType, core.WAITING, "")
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
				_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
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
			_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
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

// TestProcessWithLocationWakesGeneralExecutor verifies that a process WITH a location
// can wake an executor that registered WITHOUT a location filter (accepts any location).
func TestProcessWithLocationWakesGeneralExecutor(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	var wakeUpCount int32
	var wg sync.WaitGroup

	// Create a process WITH location
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = "datacenter-1"
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executor WITHOUT location filter (accepts any location)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Executor waiting without location filter
		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
		if err == nil {
			atomic.AddInt32(&wakeUpCount, 1)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify executor registered
	_, listeners, _ := handler.NumberOfListeners(executorType, core.WAITING, "")
	assert.Equal(t, 1, listeners, "Executor should be registered without location")

	// Signal process with location
	handler.Signal(process)

	wg.Wait()

	assert.Equal(t, int32(1), wakeUpCount,
		"Process with location should wake executor without location filter")
}

// TestProcessWithLocationWakesLocationSpecificExecutor verifies that a process WITH a location
// wakes an executor that registered WITH the same location filter.
func TestProcessWithLocationWakesLocationSpecificExecutor(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	location := "datacenter-1"
	var wakeUpCount int32
	var wg sync.WaitGroup

	// Create a process WITH location
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = location
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executor WITH matching location filter
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Executor waiting with location filter
		_, err := handler.WaitForProcess(executorType, core.WAITING, "", location, ctx)
		if err == nil {
			atomic.AddInt32(&wakeUpCount, 1)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Verify executor registered with location
	_, listeners, _ := handler.NumberOfListeners(executorType, core.WAITING, location)
	assert.Equal(t, 1, listeners, "Executor should be registered with location")

	// Signal process with location
	handler.Signal(process)

	wg.Wait()

	assert.Equal(t, int32(1), wakeUpCount,
		"Process with location should wake executor with matching location filter")
}

// TestProcessWithoutLocationWakesAnyExecutor verifies that a process WITHOUT
// a location can wake executors at any location (since "no location" means "can run anywhere").
// This is critical for production where executors have locations but processes don't specify one.
func TestProcessWithoutLocationWakesAnyExecutor(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	var totalWakeUp int32
	var wg sync.WaitGroup

	// Create a process WITHOUT location
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	// No location set - means "can run anywhere"
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executor WITH location filter (simulates production scenario)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "datacenter-1", ctx)
		if err == nil {
			atomic.AddInt32(&totalWakeUp, 1)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Signal process without location
	handler.Signal(process)

	wg.Wait()

	assert.Equal(t, int32(1), totalWakeUp,
		"Process without location should wake location-specific executor")
}

// TestProcessWithLocationDoesNotWakeDifferentLocationExecutor verifies that a process
// with location X does NOT wake an executor filtering for location Y.
func TestProcessWithLocationDoesNotWakeDifferentLocationExecutor(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	var wakeUpCount int32
	var wg sync.WaitGroup

	// Create a process with location "datacenter-1"
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = "datacenter-1"
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executor filtering for DIFFERENT location "datacenter-2"
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "datacenter-2", ctx)
		if err == nil {
			atomic.AddInt32(&wakeUpCount, 1)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Signal process with datacenter-1
	handler.Signal(process)

	wg.Wait()

	assert.Equal(t, int32(0), wakeUpCount,
		"Process with location should NOT wake executor filtering for different location")
}

// TestMixedLocationExecutors verifies correct signal routing when executors have
// different location filters and a process with location is signaled.
func TestMixedLocationExecutors(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	var generalWakeUp int32
	var matchingWakeUp int32
	var differentWakeUp int32
	var wg sync.WaitGroup

	// Create a process with location "datacenter-1"
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = "datacenter-1"
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start executor WITHOUT location filter (general)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
		if err == nil {
			atomic.AddInt32(&generalWakeUp, 1)
		}
	}()

	// Start executor WITH matching location filter
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "datacenter-1", ctx)
		if err == nil {
			atomic.AddInt32(&matchingWakeUp, 1)
		}
	}()

	// Start executor WITH different location filter
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := handler.WaitForProcess(executorType, core.WAITING, "", "datacenter-2", ctx)
		if err == nil {
			atomic.AddInt32(&differentWakeUp, 1)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Signal process with datacenter-1
	handler.Signal(process)

	wg.Wait()

	t.Logf("General wake-up: %d, Matching wake-up: %d, Different wake-up: %d",
		generalWakeUp, matchingWakeUp, differentWakeUp)

	// Only ONE executor should wake up (either general or matching, not both)
	// to prevent thundering herd
	totalWakeUps := generalWakeUp + matchingWakeUp + differentWakeUp
	assert.Equal(t, int32(1), totalWakeUps,
		"Only ONE executor should wake up to prevent thundering herd")
	assert.Equal(t, int32(0), differentWakeUp,
		"Executor with different location should NOT wake up")
}

// TestMultipleSignalsWithLocation verifies round-robin works correctly
// when multiple signals are sent for processes with location.
func TestMultipleSignalsWithLocation(t *testing.T) {
	handler := CreateTestableEventHandler(nil)
	defer handler.Stop()

	executorType := "test-executor"
	numExecutors := 3
	numSignals := 3
	var totalWakeUps int32
	var wg sync.WaitGroup

	// Create a process with location
	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Conditions.LocationName = "datacenter-1"
	process := core.CreateProcess(funcSpec)
	process.State = core.WAITING

	// Start multiple general executors (no location filter)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numSignals; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := handler.WaitForProcess(executorType, core.WAITING, "", "", ctx)
				cancel()
				if err == nil {
					atomic.AddInt32(&totalWakeUps, 1)
				} else {
					break
				}
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Send multiple signals
	for i := 0; i < numSignals; i++ {
		handler.Signal(process)
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()

	t.Logf("Total wake-ups: %d (signals=%d, executors=%d)", totalWakeUps, numSignals, numExecutors)

	assert.Equal(t, int32(numSignals), totalWakeUps,
		"Each signal should wake exactly 1 executor (1:1 ratio with location)")
}
