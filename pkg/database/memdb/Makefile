# VelocityDB Makefile
# Production-grade testing and benchmarking for VelocityDB

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=velocitydb
BINARY_UNIX=$(BINARY_NAME)_unix

# Test parameters
TEST_TIMEOUT=30s
BENCH_TIME=10s
BENCH_COUNT=3
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Performance test parameters
PERF_INSERTS=100000
PERF_READS=1000000
PERF_WORKERS=10

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help all clean deps test test-verbose test-coverage test-race bench bench-verbose bench-full perf-test demo check format lint security

## Help
help: ## Show this help message
	@echo "$(BLUE)VelocityDB - High-Performance In-Memory Database$(NC)"
	@echo ""
	@echo "$(GREEN)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

## Development
all: clean deps test bench ## Clean, install dependencies, run tests and benchmarks

clean: ## Clean build artifacts and test files
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -rf /tmp/velocitydb_* /tmp/simple_memdb_* /tmp/memdb_*
	@echo "$(GREEN)✓ Clean complete$(NC)"

deps: ## Download and install dependencies
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

format: ## Format Go code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## Testing
test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -v ./...
	@echo "$(GREEN)✓ All tests passed$(NC)"

test-verbose: ## Run tests with verbose output
	@echo "$(YELLOW)Running tests with verbose output...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -v -count=1 ./...

test-short: ## Run only fast tests
	@echo "$(YELLOW)Running short tests...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -short -v ./...
	@echo "$(GREEN)✓ Short tests passed$(NC)"

test-core: ## Run only core VelocityDB tests
	@echo "$(YELLOW)Running core VelocityDB tests...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -v -run TestVelocityDB .
	@echo "$(GREEN)✓ Core tests passed$(NC)"

test-schema: ## Run only schema validation tests
	@echo "$(YELLOW)Running schema tests...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -v ./schema/...
	@echo "$(GREEN)✓ Schema tests passed$(NC)"

test-race: ## Run tests with race detection
	@echo "$(YELLOW)Running race condition tests...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -race -v ./...
	@echo "$(GREEN)✓ Race condition tests passed$(NC)"

test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GOTEST) -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "$(BLUE)Coverage report:$(NC)"
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	@echo "$(GREEN)✓ Coverage report generated$(NC)"

test-coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(YELLOW)Generating HTML coverage report...$(NC)"
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)✓ HTML coverage report generated: $(COVERAGE_HTML)$(NC)"

## Benchmarking
bench: ## Run performance benchmarks
	@echo "$(YELLOW)Running performance benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem -run=^$$ .
	@echo "$(GREEN)✓ Benchmarks complete$(NC)"

bench-verbose: ## Run benchmarks with verbose output
	@echo "$(YELLOW)Running verbose benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ -v .

bench-insert: ## Benchmark insert operations only
	@echo "$(YELLOW)Benchmarking insert operations...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Insert -benchmem -benchtime=$(BENCH_TIME) -run=^$$ .

bench-read: ## Benchmark read operations only
	@echo "$(YELLOW)Benchmarking read operations...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Get -benchmem -benchtime=$(BENCH_TIME) -run=^$$ .

bench-cas: ## Benchmark CAS operations only
	@echo "$(YELLOW)Benchmarking CAS operations...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_CAS -benchmem -benchtime=$(BENCH_TIME) -run=^$$ .

bench-concurrent: ## Benchmark concurrent operations
	@echo "$(YELLOW)Benchmarking concurrent operations...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Concurrent -benchmem -benchtime=$(BENCH_TIME) -run=^$$ .

bench-cache: ## Benchmark cache performance
	@echo "$(YELLOW)Benchmarking cache performance...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Cache -benchmem -benchtime=$(BENCH_TIME) -run=^$$ .

bench-full: ## Run comprehensive benchmark suite
	@echo "$(YELLOW)Running comprehensive benchmark suite...$(NC)"
	@echo "$(BLUE)=== Insert Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Insert -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== Read Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Get -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== Update Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Update -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== CAS Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_CAS -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== Concurrent Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Concurrent -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== Cache Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Cache -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(BLUE)=== Mixed Workload ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Mixed -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) -run=^$$ .
	@echo "$(GREEN)✓ Comprehensive benchmarks complete$(NC)"

## Performance Testing
perf-test: ## Run custom performance tests
	@echo "$(YELLOW)Running custom performance tests...$(NC)"
	@echo "$(BLUE)Testing with $(PERF_INSERTS) inserts, $(PERF_READS) reads, $(PERF_WORKERS) workers$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Insert -benchtime=$(PERF_INSERTS)x -run=^$$ .
	$(GOTEST) -bench=BenchmarkVelocityDB_Get -benchtime=$(PERF_READS)x -run=^$$ .
	$(GOTEST) -bench=BenchmarkVelocityDB_ConcurrentReads -benchtime=$(BENCH_TIME) -run=^$$ .
	@echo "$(GREEN)✓ Performance tests complete$(NC)"

perf-memory: ## Run memory usage benchmarks
	@echo "$(YELLOW)Running memory performance tests...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_MemoryUsage -benchmem -memprofile=mem.prof -run=^$$ .
	@echo "$(BLUE)Memory profile saved to mem.prof$(NC)"
	@echo "$(GREEN)✓ Memory benchmarks complete$(NC)"

perf-cpu: ## Run CPU performance benchmarks
	@echo "$(YELLOW)Running CPU performance tests...$(NC)"
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof -run=^$$ .
	@echo "$(BLUE)CPU profile saved to cpu.prof$(NC)"
	@echo "$(GREEN)✓ CPU benchmarks complete$(NC)"

## Demos
demo: ## Run VelocityDB demo
	@echo "$(YELLOW)Running VelocityDB demo...$(NC)"
	cd example && $(GOCMD) run velocity_main.go
	@echo "$(GREEN)✓ Demo complete$(NC)"

demo-perf: ## Run performance demonstration
	@echo "$(YELLOW)Running performance demo...$(NC)"
	@echo "$(BLUE)=== Single-threaded Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Insert -benchtime=5s -run=^$$ . | grep BenchmarkVelocityDB_Insert
	$(GOTEST) -bench=BenchmarkVelocityDB_Get -benchtime=5s -run=^$$ . | grep BenchmarkVelocityDB_Get
	$(GOTEST) -bench=BenchmarkVelocityDB_CAS -benchtime=5s -run=^$$ . | grep BenchmarkVelocityDB_CAS
	@echo "$(BLUE)=== Concurrent Performance ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_ConcurrentReads -benchtime=5s -run=^$$ . | grep BenchmarkVelocityDB_ConcurrentReads
	$(GOTEST) -bench=BenchmarkVelocityDB_ConcurrentWrites -benchtime=5s -run=^$$ . | grep BenchmarkVelocityDB_ConcurrentWrites
	@echo "$(GREEN)✓ Performance demo complete$(NC)"

## Quality Assurance
check: test test-race lint ## Run all quality checks

lint: ## Run Go linting (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Linting complete$(NC)"; \
	else \
		echo "$(RED)golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

security: ## Run security scan (requires gosec)
	@echo "$(YELLOW)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
		echo "$(GREEN)✓ Security scan complete$(NC)"; \
	else \
		echo "$(RED)gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

## Continuous Integration
ci: clean deps test test-race bench ## Run full CI pipeline

ci-coverage: clean deps test-coverage ## Run CI with coverage reporting
	@echo "$(BLUE)Final coverage report:$(NC)"
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1

## Stress Testing
stress-test: ## Run stress tests with high concurrency
	@echo "$(YELLOW)Running stress tests...$(NC)"
	@echo "$(BLUE)=== High Concurrency Test ===$(NC)"
	$(GOTEST) -timeout 60s -run TestVelocityDB_ConcurrentAccess -v .
	@echo "$(BLUE)=== Extended Benchmark Run ===$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_ConcurrentReads -benchtime=30s -run=^$$ .
	$(GOTEST) -bench=BenchmarkVelocityDB_ConcurrentWrites -benchtime=30s -run=^$$ .
	@echo "$(GREEN)✓ Stress tests complete$(NC)"

load-test: ## Run load tests with realistic data volumes
	@echo "$(YELLOW)Running load tests...$(NC)"
	$(GOTEST) -bench=BenchmarkVelocityDB_Insert -benchtime=100000x -run=^$$ .
	$(GOTEST) -bench=BenchmarkVelocityDB_Get -benchtime=1000000x -run=^$$ .
	$(GOTEST) -bench=BenchmarkVelocityDB_CAS -benchtime=50000x -run=^$$ .
	@echo "$(GREEN)✓ Load tests complete$(NC)"

## Reporting
report: test-coverage bench ## Generate comprehensive test and performance report
	@echo "$(YELLOW)Generating comprehensive report...$(NC)"
	@echo "$(BLUE)================================$(NC)"
	@echo "$(BLUE) VelocityDB Test & Performance Report$(NC)"
	@echo "$(BLUE)================================$(NC)"
	@echo ""
	@echo "$(GREEN)Test Coverage:$(NC)"
	$(GOCMD) tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo ""
	@echo "$(GREEN)Performance Summary:$(NC)"
	$(GOTEST) -bench=. -benchmem -run=^$$ . | grep "BenchmarkVelocityDB"
	@echo ""
	@echo "$(GREEN)✓ Report generated$(NC)"

## Development Workflows
dev: ## Development workflow: format, test, bench
	@$(MAKE) format
	@$(MAKE) test-short
	@$(MAKE) bench-insert

quick: ## Quick development test
	@$(MAKE) format
	@$(MAKE) test-core

full: ## Full test suite (use before committing)
	@$(MAKE) clean
	@$(MAKE) deps
	@$(MAKE) format
	@$(MAKE) test
	@$(MAKE) test-race
	@$(MAKE) bench
	@echo "$(GREEN)✓ Full test suite complete - ready to commit$(NC)"

## Default target
default: help