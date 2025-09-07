# VelocityDB Makefile Guide

## 🚀 **Quick Start**

```bash
# See all available commands
make help

# Run core tests
make test-core

# Run performance benchmarks  
make bench

# Run full development workflow
make dev
```

## 🧪 **Testing Commands**

### Basic Testing
```bash
make test              # Run all tests
make test-core         # Run only core VelocityDB tests  
make test-schema       # Run only schema tests
make test-short        # Run fast tests only
```

### Advanced Testing
```bash
make test-race         # Run with race detection
make test-coverage     # Generate coverage report
make test-coverage-html # Generate HTML coverage report
make check             # Run all quality checks
```

### Example Output
```bash
$ make test-core
✓ TestVelocityDB_BasicOperations
✓ TestVelocityDB_CompareAndSwap  
✓ TestVelocityDB_ConcurrentAccess
✓ All 9 core tests passed in 0.178s
```

## 📊 **Performance Testing**

### Individual Benchmarks
```bash
make bench-insert      # Benchmark insert operations
make bench-read        # Benchmark read operations  
make bench-cas         # Benchmark CAS operations
make bench-concurrent  # Benchmark concurrent ops
make bench-cache       # Benchmark cache performance
```

### Comprehensive Benchmarking
```bash
make bench-full        # Complete benchmark suite
make bench-verbose     # Detailed benchmark output
make demo-perf         # Performance demonstration
```

### Example Performance Output
```bash
$ make demo-perf
=== Single-threaded Performance ===
BenchmarkVelocityDB_Insert-32    285,624 ops/sec  (19.2μs/op)
BenchmarkVelocityDB_Get-32      3,449,110 ops/sec  (1.7μs/op)
BenchmarkVelocityDB_CAS-32        353,108 ops/sec  (29.7μs/op)

=== Concurrent Performance ===  
BenchmarkVelocityDB_ConcurrentReads-32   49,479,704 ops/sec  (121ns/op)
BenchmarkVelocityDB_ConcurrentWrites-32     248,828 ops/sec  (22μs/op)
```

## 🔧 **Development Workflows**

### Quick Development
```bash
make dev               # Format → Test → Benchmark
make quick             # Format → Core tests only
```

### Pre-commit Workflow  
```bash
make full              # Complete test suite before commit
make ci                # Full CI pipeline
```

### Example Development Workflow
```bash
# Make code changes
make dev               # Quick verification

# Before committing
make full              # Comprehensive testing
```

## 🎯 **Specialized Testing**

### Stress & Load Testing
```bash
make stress-test       # High concurrency tests
make load-test         # Realistic data volumes
make perf-memory       # Memory usage profiling
make perf-cpu          # CPU performance profiling
```

### Quality Assurance
```bash
make lint              # Code linting (requires golangci-lint)
make security          # Security scanning (requires gosec) 
make format            # Code formatting
```

## 📈 **Performance Analysis**

### Custom Performance Tests
```bash
# Configure test parameters
export PERF_INSERTS=100000
export PERF_READS=1000000
export PERF_WORKERS=10

make perf-test         # Run custom performance tests
```

### Memory & CPU Profiling
```bash
make perf-memory       # Generates mem.prof
make perf-cpu          # Generates cpu.prof

# Analyze profiles
go tool pprof mem.prof
go tool pprof cpu.prof
```

## 📊 **Reporting**

### Comprehensive Reports
```bash
make report            # Test coverage + performance summary
make ci-coverage       # CI-friendly coverage report
```

### Example Report Output
```bash
$ make report
================================
 VelocityDB Test & Performance Report
================================

Test Coverage: 47.5% of statements

Performance Summary:
Insert:    285,624 ops/sec  (19.2μs/op, 3.8KB/op)
Get:     3,449,110 ops/sec  (1.7μs/op,  1.0KB/op)
Update:     32,259 ops/sec  (34.2μs/op, 5.1KB/op)
CAS:       353,108 ops/sec  (29.7μs/op, 8.4KB/op)
```

## 🎮 **Demo Commands**

### Interactive Demos
```bash
make demo              # Full VelocityDB demonstration
make demo-perf         # Performance showcase
```

## 🔍 **Continuous Integration**

### CI Pipeline Commands
```bash
make ci                # Full CI: clean → deps → test → race → bench
make ci-coverage       # CI with coverage reporting
```

### GitHub Actions Integration
```yaml
# .github/workflows/test.yml
- name: Run Tests
  run: make ci-coverage
  
- name: Run Benchmarks  
  run: make bench
```

## 🛠️ **Maintenance Commands**

### Cleanup & Setup
```bash
make clean             # Remove build artifacts and temp files
make deps              # Install/update dependencies
make format            # Format all Go code
```

## 💡 **Pro Tips**

### Speed Up Development
```bash
# Use short tests during development
make test-short        # Faster feedback loop

# Focus on specific components
make test-core         # Test only core functionality
make bench-insert      # Benchmark specific operations
```

### Performance Optimization
```bash
# Profile memory usage
make perf-memory

# Analyze CPU bottlenecks  
make perf-cpu

# Test under high load
make stress-test
```

### Quality Assurance
```bash
# Run before every commit
make full

# Check for race conditions
make test-race

# Ensure code quality
make check
```

## 📋 **Command Reference**

| Command | Purpose | Time | Use Case |
|---------|---------|------|----------|
| `make quick` | Fast development test | ~3s | During development |
| `make dev` | Development workflow | ~10s | After code changes |  
| `make test` | All tests | ~15s | Regular testing |
| `make bench` | Performance tests | ~20s | Performance check |
| `make full` | Complete test suite | ~45s | Before commit |
| `make ci` | CI pipeline | ~60s | Automated testing |

## 🎯 **Performance Targets**

Our Makefile helps verify these performance targets:

- **Insert Rate**: >250K ops/sec
- **Read Rate**: >3M ops/sec  
- **CAS Rate**: >300K ops/sec
- **Concurrent Reads**: >40M ops/sec
- **Latency**: <30μs for operations

Run `make demo-perf` to verify your system meets these targets!