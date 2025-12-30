package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
)

type Config struct {
	ServerHost      string
	ServerPort      int
	Insecure        bool
	ColonyName      string
	ColonyPrvKey    string
	ServerPrvKey    string
	NumExecutors    int
	NumProcesses    int
	QueueDepth      int
	OutputFile      string
	ExecutorType    string
	CleanupAfter    bool
	SetupColony     bool
}

type Result struct {
	Timestamp       time.Time
	ExecutorID      int
	LatencyMs       float64
	Success         bool
}

type Stats struct {
	TotalAssigns    int64
	SuccessAssigns  int64
	FailedAssigns   int64
	TotalLatencyMs  float64
	MinLatencyMs    float64
	MaxLatencyMs    float64
	Latencies       []float64
}

func main() {
	config := parseFlags()

	fmt.Println("=== Colonies Assign Performance Test ===")
	fmt.Printf("Server: %s:%d\n", config.ServerHost, config.ServerPort)
	fmt.Printf("Colony: %s\n", config.ColonyName)
	fmt.Printf("Executors: %d\n", config.NumExecutors)
	fmt.Printf("Processes: %d\n", config.NumProcesses)
	fmt.Printf("Queue Depth: %d\n", config.QueueDepth)
	fmt.Printf("Output: %s\n", config.OutputFile)
	fmt.Println()

	// Create client
	c := client.CreateColoniesClient(config.ServerHost, config.ServerPort, config.Insecure, false)

	// Setup colony if requested
	if config.SetupColony {
		fmt.Println("Setting up colony...")
		config = setupColony(c, config)
	}

	if config.ColonyPrvKey == "" {
		log.Fatal("Colony private key required. Use --setup to create a colony, or provide --colonyprvkey")
	}

	// Setup: Create executors
	fmt.Println("Setting up executors...")
	executors := setupExecutors(c, config)
	fmt.Printf("Created %d executors\n", len(executors))

	// Use first executor's key for submitting (executors are colony members)
	executorPrvKey := executors[0].PrvKey

	// Setup: Submit processes to create queue
	fmt.Println("Submitting processes...")
	submitProcesses(c, config, executorPrvKey)
	fmt.Printf("Submitted %d processes\n", config.NumProcesses)

	// Wait for queue to fill
	fmt.Printf("Waiting for queue depth of %d...\n", config.QueueDepth)
	waitForQueueDepth(c, config, executorPrvKey)

	// Run benchmark
	fmt.Println("Starting benchmark...")
	results := runBenchmark(c, config, executors)

	// Calculate and print stats
	stats := calculateStats(results)
	printStats(stats, config)

	// Save results to CSV
	saveResults(results, stats, config)

	// Cleanup
	if config.CleanupAfter {
		fmt.Println("Cleaning up...")
		cleanup(c, config, executors)
	}

	fmt.Println("Done!")
}

func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.ServerHost, "host", getEnv("COLONIES_SERVER_HOST", "localhost"), "Colonies server host")
	flag.IntVar(&config.ServerPort, "port", getEnvInt("COLONIES_SERVER_PORT", 50080), "Colonies server port")
	flag.BoolVar(&config.Insecure, "insecure", getEnvBool("COLONIES_TLS", true) == false, "Use insecure HTTP")
	flag.StringVar(&config.ColonyName, "colony", getEnv("COLONIES_COLONY_NAME", "perf-test"), "Colony name")
	flag.StringVar(&config.ColonyPrvKey, "colonyprvkey", getEnv("COLONIES_COLONY_PRVKEY", ""), "Colony private key")
	flag.StringVar(&config.ServerPrvKey, "serverprvkey", getEnv("COLONIES_SERVER_PRVKEY", ""), "Server private key (for colony setup)")
	flag.IntVar(&config.NumExecutors, "executors", 10, "Number of concurrent executors")
	flag.IntVar(&config.NumProcesses, "processes", 1000, "Number of processes to submit")
	flag.IntVar(&config.QueueDepth, "queue-depth", 100, "Wait for this many waiting processes before starting")
	flag.StringVar(&config.OutputFile, "output", "benchmark_results.csv", "Output CSV file")
	flag.StringVar(&config.ExecutorType, "executor-type", "perf-test", "Executor type for the test")
	flag.BoolVar(&config.CleanupAfter, "cleanup", true, "Cleanup executors and processes after test")
	flag.BoolVar(&config.SetupColony, "setup", false, "Create colony if it doesn't exist (requires --serverprvkey)")

	flag.Parse()

	if config.SetupColony && config.ServerPrvKey == "" {
		log.Fatal("Server private key required for setup (--serverprvkey or COLONIES_SERVER_PRVKEY)")
	}

	return config
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return defaultVal
}

type ExecutorInfo struct {
	Name   string
	PrvKey string
}

func setupColony(c *client.ColoniesClient, config Config) Config {
	cryptoInstance := crypto.CreateCrypto()

	// Generate colony key if not provided
	if config.ColonyPrvKey == "" {
		prvKey, err := cryptoInstance.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("Failed to generate colony key: %v", err)
		}
		config.ColonyPrvKey = prvKey
	}

	// Generate colony ID from private key
	colonyID, err := cryptoInstance.GenerateID(config.ColonyPrvKey)
	if err != nil {
		log.Fatalf("Failed to generate colony ID: %v", err)
	}

	// Try to get existing colony
	_, err = c.GetColonyByName(config.ColonyName, config.ColonyPrvKey)
	if err == nil {
		fmt.Printf("Colony '%s' already exists\n", config.ColonyName)
		return config
	}

	// Create new colony
	colony := core.CreateColony(colonyID, config.ColonyName)

	_, err = c.AddColony(colony, config.ServerPrvKey)
	if err != nil {
		log.Fatalf("Failed to create colony: %v", err)
	}

	fmt.Printf("Created colony '%s' (ID: %s)\n", config.ColonyName, colonyID)
	fmt.Printf("Colony private key: %s\n", config.ColonyPrvKey)

	return config
}

func setupExecutors(c *client.ColoniesClient, config Config) []ExecutorInfo {
	executors := make([]ExecutorInfo, config.NumExecutors)
	cryptoInstance := crypto.CreateCrypto()

	for i := 0; i < config.NumExecutors; i++ {
		prvKey, err := cryptoInstance.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		// Derive executor ID from private key
		executorID, err := cryptoInstance.GenerateID(prvKey)
		if err != nil {
			log.Fatalf("Failed to generate executor ID: %v", err)
		}

		name := fmt.Sprintf("perf-executor-%d-%d", time.Now().UnixNano(), i)
		executor := &core.Executor{
			ID:         executorID,
			Name:       name,
			Type:       config.ExecutorType,
			ColonyName: config.ColonyName,
		}

		_, err = c.AddExecutor(executor, config.ColonyPrvKey)
		if err != nil {
			log.Fatalf("Failed to add executor: %v", err)
		}

		err = c.ApproveExecutor(config.ColonyName, name, config.ColonyPrvKey)
		if err != nil {
			log.Fatalf("Failed to approve executor: %v", err)
		}

		executors[i] = ExecutorInfo{
			Name:   name,
			PrvKey: prvKey,
		}
	}

	return executors
}

func submitProcesses(c *client.ColoniesClient, config Config, prvKey string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50) // Limit concurrent submissions

	for i := 0; i < config.NumProcesses; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			funcSpec := &core.FunctionSpec{
				FuncName: "benchmark",
				Args:     []interface{}{idx},
				Conditions: core.Conditions{
					ColonyName:   config.ColonyName,
					ExecutorType: config.ExecutorType,
				},
				MaxExecTime: 300,
				MaxRetries:  0,
				Env:         make(map[string]string),
			}

			_, err := c.Submit(funcSpec, prvKey)
			if err != nil {
				log.Printf("Failed to submit process %d: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()
}

func waitForQueueDepth(c *client.ColoniesClient, config Config, prvKey string) {
	for {
		processes, err := c.GetWaitingProcesses(config.ColonyName, config.ExecutorType, "", "", config.QueueDepth, prvKey)
		if err != nil {
			log.Printf("Error checking queue depth: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if len(processes) >= config.QueueDepth {
			fmt.Printf("Queue depth reached: %d processes waiting\n", len(processes))
			break
		}

		fmt.Printf("Current queue depth: %d, waiting for %d...\n", len(processes), config.QueueDepth)
		time.Sleep(500 * time.Millisecond)
	}
}

func runBenchmark(c *client.ColoniesClient, config Config, executors []ExecutorInfo) []Result {
	var results []Result
	var resultsMu sync.Mutex
	var wg sync.WaitGroup
	var closeWg sync.WaitGroup
	var assignedCount int64
	var failedCount int64

	// Semaphore to limit concurrent Close operations (prevents overwhelming the server)
	closeSemaphore := make(chan struct{}, 100)

	startTime := time.Now()

	// Each executor continuously tries to assign
	for i, exec := range executors {
		wg.Add(1)
		go func(execIdx int, execInfo ExecutorInfo) {
			defer wg.Done()

			for {
				// Check if we've assigned all processes
				if atomic.LoadInt64(&assignedCount)+atomic.LoadInt64(&failedCount) >= int64(config.NumProcesses) {
					break
				}

				// Create a context with client-side timeout (server timeout + buffer)
				// This ensures we don't hang if the server gets stuck
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				start := time.Now()
				process, err := c.AssignWithContext(config.ColonyName, 1, ctx, "", "", execInfo.PrvKey)
				cancel() // Clean up context
				latency := time.Since(start).Seconds() * 1000 // ms

				result := Result{
					Timestamp:  start,
					ExecutorID: execIdx,
					LatencyMs:  latency,
					Success:    err == nil && process != nil,
				}

				if result.Success {
					atomic.AddInt64(&assignedCount, 1)

					// Close the process with rate limiting
					closeWg.Add(1)
					go func(pid string) {
						defer closeWg.Done()
						closeSemaphore <- struct{}{} // Acquire semaphore
						defer func() { <-closeSemaphore }() // Release semaphore

						closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer closeCancel()
						if err := c.CloseWithContext(pid, closeCtx, execInfo.PrvKey); err != nil {
							log.Printf("Warning: Close failed for process %s: %v", pid[:8], err)
						}
					}(process.ID)
				} else {
					// Timeout or no process available
					if atomic.LoadInt64(&assignedCount) >= int64(config.NumProcesses) {
						break
					}
					atomic.AddInt64(&failedCount, 1)
				}

				resultsMu.Lock()
				results = append(results, result)
				resultsMu.Unlock()

				// Progress report
				count := atomic.LoadInt64(&assignedCount)
				if count%100 == 0 && count > 0 {
					elapsed := time.Since(startTime).Seconds()
					rate := float64(count) / elapsed
					fmt.Printf("Progress: %d/%d assigned (%.1f/sec)\n", count, config.NumProcesses, rate)
				}
			}
		}(i, exec)
	}

	wg.Wait()

	// Wait for all Close operations to complete
	fmt.Println("Waiting for Close operations to complete...")
	closeWg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("\nBenchmark completed in %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("Total assigned: %d\n", atomic.LoadInt64(&assignedCount))

	return results
}

func calculateStats(results []Result) Stats {
	stats := Stats{
		MinLatencyMs: float64(^uint(0) >> 1), // Max float
		Latencies:    make([]float64, 0),
	}

	for _, r := range results {
		stats.TotalAssigns++
		if r.Success {
			stats.SuccessAssigns++
			stats.TotalLatencyMs += r.LatencyMs
			stats.Latencies = append(stats.Latencies, r.LatencyMs)

			if r.LatencyMs < stats.MinLatencyMs {
				stats.MinLatencyMs = r.LatencyMs
			}
			if r.LatencyMs > stats.MaxLatencyMs {
				stats.MaxLatencyMs = r.LatencyMs
			}
		} else {
			stats.FailedAssigns++
		}
	}

	if stats.SuccessAssigns == 0 {
		stats.MinLatencyMs = 0
	}

	return stats
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func printStats(stats Stats, config Config) {
	fmt.Println("\n=== Results ===")
	fmt.Printf("Total Attempts:    %d\n", stats.TotalAssigns)
	fmt.Printf("Successful:        %d\n", stats.SuccessAssigns)
	fmt.Printf("Failed/Timeout:    %d\n", stats.FailedAssigns)

	if stats.SuccessAssigns > 0 {
		avgLatency := stats.TotalLatencyMs / float64(stats.SuccessAssigns)

		// Sort for percentiles
		sort.Float64s(stats.Latencies)
		p50 := percentile(stats.Latencies, 0.50)
		p95 := percentile(stats.Latencies, 0.95)
		p99 := percentile(stats.Latencies, 0.99)

		fmt.Printf("\nLatency (ms):\n")
		fmt.Printf("  Min:    %.2f\n", stats.MinLatencyMs)
		fmt.Printf("  Max:    %.2f\n", stats.MaxLatencyMs)
		fmt.Printf("  Avg:    %.2f\n", avgLatency)
		fmt.Printf("  P50:    %.2f\n", p50)
		fmt.Printf("  P95:    %.2f\n", p95)
		fmt.Printf("  P99:    %.2f\n", p99)
	}
}

func saveResults(results []Result, stats Stats, config Config) {
	// Save detailed results
	file, err := os.Create(config.OutputFile)
	if err != nil {
		log.Printf("Failed to create output file: %v", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	writer.Write([]string{"timestamp", "executor_id", "latency_ms", "success"})

	// Data rows
	for _, r := range results {
		writer.Write([]string{
			r.Timestamp.Format(time.RFC3339Nano),
			fmt.Sprintf("%d", r.ExecutorID),
			fmt.Sprintf("%.3f", r.LatencyMs),
			fmt.Sprintf("%t", r.Success),
		})
	}

	fmt.Printf("Results saved to %s\n", config.OutputFile)

	// Save summary
	summaryFile := config.OutputFile[:len(config.OutputFile)-4] + "_summary.csv"
	sfile, err := os.Create(summaryFile)
	if err != nil {
		log.Printf("Failed to create summary file: %v", err)
		return
	}
	defer sfile.Close()

	swriter := csv.NewWriter(sfile)
	defer swriter.Flush()

	sort.Float64s(stats.Latencies)
	avgLatency := float64(0)
	if stats.SuccessAssigns > 0 {
		avgLatency = stats.TotalLatencyMs / float64(stats.SuccessAssigns)
	}

	swriter.Write([]string{"metric", "value"})
	swriter.Write([]string{"executors", fmt.Sprintf("%d", config.NumExecutors)})
	swriter.Write([]string{"processes", fmt.Sprintf("%d", config.NumProcesses)})
	swriter.Write([]string{"total_attempts", fmt.Sprintf("%d", stats.TotalAssigns)})
	swriter.Write([]string{"successful", fmt.Sprintf("%d", stats.SuccessAssigns)})
	swriter.Write([]string{"failed", fmt.Sprintf("%d", stats.FailedAssigns)})
	swriter.Write([]string{"min_latency_ms", fmt.Sprintf("%.3f", stats.MinLatencyMs)})
	swriter.Write([]string{"max_latency_ms", fmt.Sprintf("%.3f", stats.MaxLatencyMs)})
	swriter.Write([]string{"avg_latency_ms", fmt.Sprintf("%.3f", avgLatency)})
	swriter.Write([]string{"p50_latency_ms", fmt.Sprintf("%.3f", percentile(stats.Latencies, 0.50))})
	swriter.Write([]string{"p95_latency_ms", fmt.Sprintf("%.3f", percentile(stats.Latencies, 0.95))})
	swriter.Write([]string{"p99_latency_ms", fmt.Sprintf("%.3f", percentile(stats.Latencies, 0.99))})

	fmt.Printf("Summary saved to %s\n", summaryFile)
}

func cleanup(c *client.ColoniesClient, config Config, executors []ExecutorInfo) {
	// Remove all test processes
	c.RemoveAllProcesses(config.ColonyName, config.ColonyPrvKey)

	// Remove executors
	for _, exec := range executors {
		c.RemoveExecutor(config.ColonyName, exec.Name, config.ColonyPrvKey)
	}
}
