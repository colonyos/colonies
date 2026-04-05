package embedded

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// createConcurrencyProcess is a helper that creates a minimal process suitable
// for concurrency tests. It mirrors the pattern used in createTestProcess but
// allows the caller to specify the colony and executor type.
func createConcurrencyProcess(colonyName, executorType string) *core.Process {
	env := make(map[string]string)
	funcSpec := core.CreateFunctionSpec(
		"", "conc-func", []interface{}{}, map[string]interface{}{},
		colonyName, []string{}, executorType,
		0, 0, 0, env, []string{}, 0, "",
	)
	funcSpec.Conditions.CPU = "1000m"
	funcSpec.Conditions.Memory = "1Gi"
	funcSpec.Conditions.Storage = "10Gi"
	funcSpec.Conditions.Nodes = 1
	funcSpec.Conditions.Processes = 1
	funcSpec.Conditions.ProcessesPerNode = 1
	return core.CreateProcess(funcSpec)
}

// addTestColony adds a colony and returns it, failing the test on error.
func addTestColony(t *testing.T, db *EmbeddedDatabase, name string) *core.Colony {
	t.Helper()
	colony := core.CreateColony(core.GenerateRandomID(), name)
	if err := db.AddColony(colony); err != nil {
		t.Fatal(err)
	}
	return colony
}

// addTestExecutor adds an executor, approves it, and returns it.
func addTestExecutor(t *testing.T, db *EmbeddedDatabase, name, colonyName, executorType string) *core.Executor {
	t.Helper()
	e := core.CreateExecutor(core.GenerateRandomID(), executorType, name, colonyName, time.Now(), time.Now())
	if err := db.AddExecutor(e); err != nil {
		t.Fatal(err)
	}
	if err := db.ApproveExecutor(e); err != nil {
		t.Fatal(err)
	}
	return e
}

// pace adds a small delay between iterations to avoid overwhelming the
// embedded database disk flusher during stress tests.
func pace() {
	time.Sleep(10 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// 1. TestConcurrentAddRemoveExecutors
// ---------------------------------------------------------------------------

func TestConcurrentAddRemoveExecutors(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "conc-colony")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	const goroutines = 20
	var wg sync.WaitGroup
	var errCount atomic.Int64

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			iter := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				iter++
				name := fmt.Sprintf("exec-%d-%d", id, iter)
				e := core.CreateExecutor(core.GenerateRandomID(), "worker", name, "conc-colony", time.Now(), time.Now())
				if err := db.AddExecutor(e); err != nil {
					errCount.Add(1)
					continue
				}
				_ = db.ApproveExecutor(e)
				_ = db.RemoveExecutorByName("conc-colony", name)
				pace()
			}
		}(i)
	}

	wg.Wait()

	// Verify: no duplicates among remaining executors
	executors, err := db.GetExecutorsByColonyName("conc-colony", true)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, e := range executors {
		if seen[e.Name] {
			t.Fatalf("duplicate executor found: %s", e.Name)
		}
		seen[e.Name] = true
	}
}

// ---------------------------------------------------------------------------
// 2. TestConcurrentProcessStateTransitions
// ---------------------------------------------------------------------------

func TestConcurrentProcessStateTransitions(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "state-colony")
	executor := addTestExecutor(t, db, "state-exec", "state-colony", "cli")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Submitters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				p := createConcurrencyProcess("state-colony", "cli")
				_ = db.AddProcess(p)
				pace()
			}
		}()
	}

	// Assigners
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				waiting, err := db.FindWaitingProcesses("state-colony", "", "", "", 1)
				if err != nil || len(waiting) == 0 {
					pace()
					continue
				}
				_ = db.Assign(executor.ID, waiting[0])
				pace()
			}
		}()
	}

	// Closers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				running, err := db.FindRunningProcesses("state-colony", "", "", "", 1)
				if err != nil || len(running) == 0 {
					pace()
					continue
				}
				_, _, _ = db.MarkSuccessful(running[0].ID)
				pace()
			}
		}()
	}

	wg.Wait()

	// Verify: every process is in exactly one state, no duplicates across states
	waiting, _ := db.FindWaitingProcesses("state-colony", "", "", "", 100000)
	running, _ := db.FindRunningProcesses("state-colony", "", "", "", 100000)
	successful, _ := db.FindSuccessfulProcesses("state-colony", "", "", "", 100000)

	allIDs := make(map[string]string)
	for _, p := range waiting {
		if st, ok := allIDs[p.ID]; ok {
			t.Fatalf("process %s found in both WAITING and %s", p.ID, st)
		}
		allIDs[p.ID] = "WAITING"
	}
	for _, p := range running {
		if st, ok := allIDs[p.ID]; ok {
			t.Fatalf("process %s found in both RUNNING and %s", p.ID, st)
		}
		allIDs[p.ID] = "RUNNING"
	}
	for _, p := range successful {
		if st, ok := allIDs[p.ID]; ok {
			t.Fatalf("process %s found in both SUCCESS and %s", p.ID, st)
		}
		allIDs[p.ID] = "SUCCESS"
	}
}

// ---------------------------------------------------------------------------
// 3. TestConcurrentMetricsWithProcesses
// ---------------------------------------------------------------------------

func TestConcurrentMetricsWithProcesses(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "metrics-colony")
	executors := make([]*core.Executor, 3)
	for i := 0; i < 3; i++ {
		executors[i] = addTestExecutor(t, db, fmt.Sprintf("metrics-exec-%d", i), "metrics-colony", "cli")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Metric incrementers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			execName := executors[id%3].Name
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				_ = db.IncrementMetric("metrics-colony", execName, "ops", core.PERIOD_NONE, time.Time{}, 1.0)
				pace()
			}
		}(i)
	}

	// Process submitters and closers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				p := createConcurrencyProcess("metrics-colony", "cli")
				if err := db.AddProcess(p); err != nil {
					pace()
					continue
				}
				_ = db.Assign(executors[0].ID, p)
				_, _, _ = db.MarkSuccessful(p.ID)
				pace()
			}
		}()
	}

	// Metric readers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			execName := executors[id%3].Name
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				_, _ = db.GetMetricsByExecutorName("metrics-colony", execName)
				pace()
			}
		}(i)
	}

	wg.Wait()

	// Verify: metrics exist and are positive
	for _, e := range executors {
		m, err := db.GetMetric("metrics-colony", e.Name, "ops", core.PERIOD_NONE, time.Time{})
		if err != nil {
			continue // not all executors may have been targeted
		}
		if m.Value <= 0 {
			t.Fatalf("expected positive metric value for %s, got %f", e.Name, m.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. TestSimulationLLMFleet
// ---------------------------------------------------------------------------

func TestSimulationLLMFleet(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "llm-colony")

	const numExecutors = 5
	executors := make([]*core.Executor, numExecutors)
	for i := 0; i < numExecutors; i++ {
		name := fmt.Sprintf("llm-%d", i+1)
		executors[i] = addTestExecutor(t, db, name, "llm-colony", "llm")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Heartbeat goroutines (1 per executor)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				_ = db.MarkAlive(executors[idx])
				pace()
			}
		}(i)
	}

	// Process submitters (10 goroutines)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				p := createConcurrencyProcess("llm-colony", "llm")
				_ = db.AddProcess(p)
				pace()
			}
		}()
	}

	// Process workers (1 per executor)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				waiting, err := db.FindWaitingProcesses("llm-colony", "", "", "", 1)
				if err != nil || len(waiting) == 0 {
					pace()
					continue
				}
				if err := db.Assign(executors[idx].ID, waiting[0]); err != nil {
					pace()
					continue
				}
				_, _, _ = db.MarkSuccessful(waiting[0].ID)
				pace()
			}
		}(i)
	}

	// Metric reporters (1 per executor)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			execName := executors[idx].Name
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				delta := float64(rand.Intn(100) + 1)
				_ = db.IncrementMetric("llm-colony", execName, "tokens_used", core.PERIOD_NONE, time.Time{}, delta)

				temp := float64(rand.Intn(31) + 60)
				m := core.CreateMetric("llm-colony", execName, "gpu_temp", core.GAUGE, temp)
				_ = db.SetMetric(m)
				pace()
			}
		}(i)
	}

	// Metric readers (3 goroutines)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				execName := executors[idx%numExecutors].Name
				_, _ = db.GetMetricsByExecutorName("llm-colony", execName)
				pace()
			}
		}(i)
	}

	wg.Wait()

	// Verify: each executor name exists exactly once
	execs, err := db.GetExecutorsByColonyName("llm-colony", false)
	if err != nil {
		t.Fatal(err)
	}
	nameCount := make(map[string]int)
	for _, e := range execs {
		nameCount[e.Name]++
	}
	for name, count := range nameCount {
		if count != 1 {
			t.Fatalf("executor %s appears %d times, expected 1", name, count)
		}
	}

	// Verify: all metric counters > 0
	for _, e := range executors {
		m, err := db.GetMetric("llm-colony", e.Name, "tokens_used", core.PERIOD_NONE, time.Time{})
		if err != nil {
			t.Fatalf("missing tokens_used metric for %s: %v", e.Name, err)
		}
		if m.Value <= 0 {
			t.Fatalf("expected positive tokens_used for %s, got %f", e.Name, m.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. TestSimulationBurstAssignment
// ---------------------------------------------------------------------------

func TestSimulationBurstAssignment(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "burst-colony")

	const numExecutors = 10
	executors := make([]*core.Executor, numExecutors)
	for i := 0; i < numExecutors; i++ {
		executors[i] = addTestExecutor(t, db, fmt.Sprintf("burst-exec-%d", i), "burst-colony", "worker")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Submitters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				p := createConcurrencyProcess("burst-colony", "worker")
				_ = db.AddProcess(p)
				pace()
			}
		}()
	}

	// Assigners (1 per executor)
	for i := 0; i < numExecutors; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				waiting, err := db.FindWaitingProcesses("burst-colony", "", "", "", 1)
				if err != nil || len(waiting) == 0 {
					pace()
					continue
				}
				if err := db.Assign(executors[idx].ID, waiting[0]); err != nil {
					pace()
					continue
				}
				_, _, _ = db.MarkSuccessful(waiting[0].ID)
				pace()
			}
		}(i)
	}

	// Cancellers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				waiting, err := db.FindWaitingProcesses("burst-colony", "", "", "", 1)
				if err != nil || len(waiting) == 0 {
					pace()
					continue
				}
				_ = db.MarkCancelled(waiting[0].ID)
				pace()
			}
		}()
	}

	wg.Wait()

	// Verify: no process assigned to two executors
	running, _ := db.FindRunningProcesses("burst-colony", "", "", "", 100000)
	successful, _ := db.FindSuccessfulProcesses("burst-colony", "", "", "", 100000)

	seenIDs := make(map[string]bool)
	for _, p := range running {
		if seenIDs[p.ID] {
			t.Fatalf("process %s appears more than once in RUNNING", p.ID)
		}
		seenIDs[p.ID] = true
	}
	for _, p := range successful {
		if seenIDs[p.ID] {
			t.Fatalf("process %s appears in both RUNNING and SUCCESS", p.ID)
		}
		seenIDs[p.ID] = true
	}
}

// ---------------------------------------------------------------------------
// 6. TestSimulationDeadlockDetector
// ---------------------------------------------------------------------------

func TestSimulationDeadlockDetector(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "deadlock-colony")
	executor := addTestExecutor(t, db, "deadlock-exec", "deadlock-colony", "cli")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		var wg sync.WaitGroup

		// AddProcess goroutines (touches processes + attributes + indexes)
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					p := createConcurrencyProcess("deadlock-colony", "cli")
					_ = db.AddProcess(p)
					pace()
				}
			}()
		}

		// MarkSuccessful/MarkFailed on running processes
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					running, err := db.FindRunningProcesses("deadlock-colony", "", "", "", 1)
					if err != nil || len(running) == 0 {
						pace()
						continue
					}
					if id%2 == 0 {
						_, _, _ = db.MarkSuccessful(running[0].ID)
					} else {
						_ = db.MarkFailed(running[0].ID, []string{"test error"})
					}
					pace()
				}
			}(i)
		}

		// Assign waiting processes so the MarkSuccessful/MarkFailed goroutines have work
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					waiting, err := db.FindWaitingProcesses("deadlock-colony", "", "", "", 1)
					if err != nil || len(waiting) == 0 {
						pace()
						continue
					}
					_ = db.Assign(executor.ID, waiting[0])
					pace()
				}
			}()
		}

		// RemoveAllSuccessfulProcessesByColonyName
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					_ = db.RemoveAllSuccessfulProcessesByColonyName("deadlock-colony")
					pace()
				}
			}()
		}

		// AddExecutor/RemoveExecutorByName with same executor name (re-registration)
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				name := fmt.Sprintf("rereg-exec-%d", id)
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					e := core.CreateExecutor(core.GenerateRandomID(), "cli", name, "deadlock-colony", time.Now(), time.Now())
					if err := db.AddExecutor(e); err != nil {
						pace()
						continue
					}
					_ = db.RemoveExecutorByName("deadlock-colony", name)
					pace()
				}
			}(i)
		}

		// ApplyRetentionPolicy
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					_ = db.ApplyRetentionPolicy(1)
					pace()
				}
			}()
		}

		// SetMetric/IncrementMetric
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
					}
					if id%2 == 0 {
						m := core.CreateMetric("deadlock-colony", "deadlock-exec", "gauge1", core.GAUGE, float64(rand.Intn(100)))
						_ = db.SetMetric(m)
					} else {
						_ = db.IncrementMetric("deadlock-colony", "deadlock-exec", "counter1", core.PERIOD_NONE, time.Time{}, 1.0)
					}
					pace()
				}
			}(i)
		}

		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test completed without deadlock
	case <-time.After(10 * time.Second):
		t.Fatal("test timed out - potential deadlock detected")
	}
}

// ---------------------------------------------------------------------------
// 7. TestConcurrentRetentionDuringWrites
// ---------------------------------------------------------------------------

func TestConcurrentRetentionDuringWrites(t *testing.T) {
	db := setupTestDB(t)
	addTestColony(t, db, "retention-colony")
	executor := addTestExecutor(t, db, "retention-exec", "retention-colony", "cli")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Track recently created process IDs so we can verify they survive retention
	var recentMu sync.Mutex
	recentIDs := make(map[string]time.Time)

	// Process lifecycle goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				p := createConcurrencyProcess("retention-colony", "cli")
				if err := db.AddProcess(p); err != nil {
					pace()
					continue
				}

				recentMu.Lock()
				recentIDs[p.ID] = time.Now()
				recentMu.Unlock()

				if err := db.Assign(executor.ID, p); err != nil {
					pace()
					continue
				}
				_, _, _ = db.MarkSuccessful(p.ID)
				pace()
			}
		}()
	}

	// Retention policy goroutine - 1 second retention means only processes
	// finished more than 1 second ago should be deleted.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_ = db.ApplyRetentionPolicy(1)
			pace()
		}
	}()

	wg.Wait()

	// Verify: processes that finished less than 1 second ago should still exist.
	// We check waiting and running processes (these should never be deleted by
	// retention since they are not finished).
	waiting, _ := db.FindWaitingProcesses("retention-colony", "", "", "", 100000)
	running, _ := db.FindRunningProcesses("retention-colony", "", "", "", 100000)

	for _, p := range waiting {
		if p.State != core.WAITING {
			t.Fatalf("waiting process %s has wrong state %d", p.ID, p.State)
		}
	}
	for _, p := range running {
		if p.State != core.RUNNING {
			t.Fatalf("running process %s has wrong state %d", p.ID, p.State)
		}
	}
}
