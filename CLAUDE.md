# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building
```bash
make build              # Build the main colonies binary and libraries
make container          # Build Docker container  
make install            # Install binaries to /usr/local/bin
```

### Testing
```bash
make test              # Run all tests (requires grc for colored output)
make github_test       # Run tests without grc (for CI)
```

### Development Environment
```bash
docker-compose up -d   # Start Colonies server with dependencies (TimescaleDB, MinIO)
docker-compose down    # Stop all services
docker-compose logs -f # View logs
```

### Coverage
```bash
make coverage         # Generate test coverage reports
```

## Architecture Overview
ColonyOS is a distributed meta-orchestrator framework that creates compute continuums across different platforms. This repository contains the core Colonies server implementation.

### Core Components
- **Colony**: A distributed runtime environment consisting of networked Executors
- **Process**: A computational workload defined by a FunctionSpec, with states (WAITING, RUNNING, SUCCESS, FAILED)
- **Executor**: Distributed workers that pull and execute processes, can run anywhere on the Internet
- **FunctionSpec**: Specification defining what computation to run and execution conditions

### Key Packages
- `pkg/core/`: Core domain models (Process, Executor, Colony, FunctionSpec)
- `pkg/service/`: HTTP RPC service implementation 
- `pkg/client/`: Go SDK for Colonies API
- `pkg/database/postgresql/`: PostgreSQL database layer with TimescaleDB support
- `pkg/security/`: Zero-trust security protocol implementation
- `pkg/scheduler/`: Process scheduling and assignment logic
- `internal/cli/`: Command-line interface implementation using Cobra

### Architecture Patterns
- **Zero-trust security**: All communication is cryptographically signed
- **Process graphs**: Workflows as DAGs with parent-child relationships
- **Distributed scheduling**: Processes assigned to available Executors based on conditions
- **Meta-orchestration**: Coordinates workloads across multiple platforms without direct control

### Database
The system uses PostgreSQL with TimescaleDB for time-series data. Database interactions are abstracted through interfaces in `pkg/database/`.

### CLI Structure
The main binary is built from `cmd/main.go` which delegates to `internal/cli/` for all command handling. Commands are organized by domain (process, executor, colony, etc.).

### Testing Philosophy
Tests are co-located with source files using `_test.go` suffix. The test suite covers reliability, crypto, core domain models, database layer, RPC protocol, security, and scheduling components.
