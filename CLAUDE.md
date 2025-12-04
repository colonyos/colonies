# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Style Guidelines

- Do not use emojis in code, comments, commit messages, or documentation
- Keep log messages concise and informative
- Use structured logging with fields (e.g., `log.WithFields`) for context

## Development Commands

### Building
```bash
make build              # Build the main colonies binary and libraries
make container          # Build Docker container
make install            # Install binaries to /usr/local/bin
```

**IMPORTANT: Container Image Name**
When building Docker containers, NEVER change the image name. Always use:
```bash
BUILD_IMAGE=colonyos/colonies:latest make container
```
The deployment docker-compose files expect `colonyos/colonies:latest`. Using a different image name will cause the container to not be updated when restarting.

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

## Debugging Reconciliation Mechanism

The reconciliation system uses blueprints to manage container deployments. When debugging reconciliation issues:

### Viewing Reconciler Logs

```bash
# View logs from a specific reconciler executor
colonies log get -e local-docker-reconciler

# View logs for a specific reconciliation process
colonies log get -p <processID>

# Follow logs in real-time (requires process ID)
colonies log get -p <processID> --follow
```

### Triggering Reconciliation

```bash
# Normal reconciliation (checks if changes needed)
colonies blueprint reconcile --name <blueprint-name>

# Force reconciliation (recreates all containers with fresh images)
colonies blueprint reconcile --name <blueprint-name> --force
```

### Checking Blueprint and Process Status

```bash
# List all blueprints
colonies blueprint ls

# Get blueprint details
colonies blueprint get --name <blueprint-name>

# Check running processes
colonies process ps

# Get process details (includes output and errors)
colonies process get -p <processID>
```

### Docker Reconciler Container Logs

```bash
# Find reconciler containers
docker ps | grep reconciler

# View container logs directly
docker logs <container-name>

# Follow container logs
docker logs -f <container-name>
```

### Common Debugging Scenarios

1. **Process not closing**: Check if executor received the close message via `colonies log get -e <executor>`. Verify channels are cleaned up when process closes.

2. **Containers not starting**: Check reconciler logs for image pull errors or container start failures. Use `colonies log get -p <processID>` to see detailed reconciliation steps.

3. **Force reconcile not working**: Ensure the `--force` flag is being passed. Check logs for "Force flag enabled" message.

4. **Executor not registering**: Check reconciler logs for registration errors. Verify colony owner key has permissions.

### Key Files for Reconciliation

- `pkg/server/handlers/blueprint/handlers.go`: Blueprint CRUD and reconciliation triggering
- `pkg/channel/router.go`: Channel management (cleanup on process close)
- Docker reconciler repo (`colonyos/executors/docker-reconciler`):
  - `pkg/executor/process_handler.go`: Handles reconcile/cleanup processes
  - `pkg/reconciler/reconciler.go`: Core reconciliation logic
  - `pkg/reconciler/executor_deployment.go`: ExecutorDeployment handling
