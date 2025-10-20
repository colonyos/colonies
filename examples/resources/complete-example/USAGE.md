# Complete ResourceDefinition Usage Example

This is a **complete, working example** showing how to use Custom Resource Definitions (CRDs) in ColonyOS.

## What This Example Demonstrates

### 1. **Resource Definition** (`executor-deployment-definition.json`)
Defines a new resource type `ExecutorDeployment` that allows declarative deployment of executors.

### 2. **Custom Resource Instance** (`ml-executor-deployment.json`)
An actual deployment specification for 3 ML training executors running in Docker.

### 3. **Resource Controller** (`controller.go`)
A ColonyOS executor that:
- Polls for reconciliation work
- Extracts CustomResource from Process
- Reconciles desired state vs actual state
- Deploys/scales executors using Docker
- Updates resource status
- Reports back to ColonyOS

### 4. **Integration** (`main.go`)
Shows how to:
- Register CRDs
- Submit CustomResources
- Convert resources to Processes
- Run the controller

## Quick Start - Demo Mode

Run the demo without needing a ColonyOS server:

```bash
# Make the script executable
chmod +x run-demo.sh

# Run the demo
./run-demo.sh
```

This will show:
- ✓ ResourceDefinition registration
- ✓ CustomResource creation
- ✓ Process conversion
- ✓ Controller reconciliation logic
- ✓ Status updates

## Full Setup - With ColonyOS Server

### Prerequisites
- Docker and Docker Compose
- Go 1.21+

### Step 1: Start ColonyOS

```bash
# Start all services (TimescaleDB + ColonyOS)
docker-compose up -d

# Wait for services to be healthy
docker-compose ps

# Check logs
docker-compose logs -f colonies-server
```

### Step 2: Build the Example

```bash
go mod tidy
go build -o crd-example .
```

### Step 3: Register the ResourceDefinition

```bash
./crd-example -mode register-crd
```

Output:
```
Registering ResourceDefinition from file: executor-deployment-definition.json
ResourceDefinition Details:
  Name: executordeployments.compute.colonies.io
  Group: compute.colonies.io
  Version: v1
  Kind: ExecutorDeployment
  Handler Type: executor-deployment-controller
  Handler Function: reconcile_executor_deployment
✓ ResourceDefinition registered successfully
```

### Step 4: Start the Controller

In a **new terminal**:

```bash
export COLONIES_SERVER_HOST=localhost
export COLONIES_SERVER_PORT=50080
export COLONIES_INSECURE=true
export COLONIES_COLONY_NAME=dev

./crd-example -mode controller
```

Output:
```
Starting controller:
  Colony: dev
  Executor Name: executor-deployment-controller
  Executor Type: executor-deployment-controller
Controller is running. Press Ctrl+C to stop.
```

### Step 5: Submit a CustomResource

In **another terminal**:

```bash
./crd-example -mode submit -resource ml-executor-deployment.json
```

### Step 6: Watch the Magic Happen

Back in the controller terminal, you'll see:

```
Assigned process: a1b2c3d4 (label: ExecutorDeployment/ml-training-executors)
Reconciling ExecutorDeployment/ml-training-executors
  Runtime: docker
  Replicas: 3
  ExecutorType: ml-executor
  Image: colonyos/ml-executor:latest
  Current replicas: 0, Desired: 3
Scaling up: deploying 3 new Docker containers
Started container: ml-training-executors-a1b2c3d4-0 (ID: abc123456789)
Started container: ml-training-executors-a1b2c3d4-1 (ID: def456789012)
Started container: ml-training-executors-a1b2c3d4-2 (ID: ghi789012345)
Reconciliation complete: 3 replicas running
```

### Step 7: Verify Deployment

```bash
# List running containers
docker ps | grep ml-training-executors

# Should see 3 containers running
```

## How It Works

### 1. Resource Submission Flow

```
User submits CustomResource
        ↓
ColonyOS Server:
  - Looks up ResourceDefinition for this resource type
  - Calls resource.attaches Resource to FunctionSpec
  - Creates Process from FunctionSpec
  - Adds resource tracking attributes
  - Submits Process to queue
        ↓
Process enters queue with:
  - FuncName: "reconcile_executor_deployment"
  - ExecutorType: "executor-deployment-controller"
  - KwArgs: {entire CustomResource}
```

### 2. Controller Reconciliation Loop

```go
for {
    // Get work from ColonyOS
    process = client.Assign(executorType)

    // Extract CustomResource from process
    resource = extractFromKwArgs(process)

    // Reconcile
    currentState = getCurrentState(resource)
    desiredState = resource.Spec

    if currentState != desiredState {
        // Scale up/down
        deploy/remove executors
    }

    // Update status
    resource.Status = newState

    // Complete process
    client.Close(process.ID)
}
```

### 3. Reconciliation Logic

The controller:

1. **Compares states**:
   - Desired: `spec.replicas = 3`
   - Current: `2 containers running`

2. **Takes action**:
   - Starts 1 new Docker container

3. **Updates status**:
   ```json
   {
     "phase": "Running",
     "ready": 3,
     "deployedExecutors": ["id1", "id2", "id3"]
   }
   ```

## Supported Runtimes

### Docker (Implemented)
```json
{
  "spec": {
    "runtime": "docker",
    "replicas": 3,
    "image": "colonyos/ml-executor:latest"
  }
}
```

### Kubernetes (TODO)
```json
{
  "spec": {
    "runtime": "kubernetes",
    "replicas": 5,
    "config": {
      "namespace": "ml-executors",
      "nodeSelector": {"gpu": "true"}
    }
  }
}
```

### Local Processes (Implemented)
```json
{
  "spec": {
    "runtime": "local",
    "replicas": 2,
    "executorType": "test-executor"
  }
}
```

## Modifying the Example

### Change Replicas

Edit `ml-executor-deployment.json`:

```json
{
  "spec": {
    "replicas": 5  // Changed from 3 to 5
  }
}
```

Resubmit:
```bash
./crd-example -mode submit
```

Controller will automatically scale from 3 to 5 replicas!

### Add Custom Configuration

```json
{
  "spec": {
    "runtime": "docker",
    "replicas": 3,
    "config": {
      "env": {
        "MODEL_CACHE_DIR": "/models",
        "BATCH_SIZE": "64",
        "GPU_ENABLED": "true"
      },
      "volumes": [
        {
          "source": "/data/models",
          "target": "/models"
        }
      ]
    }
  }
}
```

The controller will pass these as Docker environment variables and volume mounts.

## Creating Your Own CRD

### 1. Define the ResourceDefinition

```json
{
  "apiVersion": "colonies.io/v1",
  "kind": "CustomResourceDefinition",
  "metadata": {
    "name": "databases.storage.colonies.io"
  },
  "spec": {
    "group": "storage.colonies.io",
    "version": "v1",
    "names": {
      "kind": "Database",
      "plural": "databases"
    },
    "scope": "Namespaced",
    "handler": {
      "executorType": "database-controller",
      "functionName": "reconcile_database"
    }
  }
}
```

### 2. Implement the Controller

```go
type DatabaseController struct {
    coloniesClient *client.ColoniesClient
}

func (c *DatabaseController) reconcile(resource *core.CustomResource) error {
    engine, _ := resource.GetSpec("engine")
    size, _ := resource.GetSpec("size")

    // Deploy database based on spec
    // ...

    // Update status
    resource.SetStatus("endpoint", dbEndpoint)
    resource.SetStatus("ready", true)

    return nil
}
```

### 3. Register and Use

```bash
./crd-example -mode register-crd -crd database-definition.json
./crd-example -mode controller  # Your controller
./crd-example -mode submit -resource my-database.json
```

## Architecture Diagram

```
┌─────────────────────────────────────────────┐
│           User / API                         │
└──────────────────┬──────────────────────────┘
                   │
                   │ Submit CustomResource
                   ▼
┌─────────────────────────────────────────────┐
│         ColonyOS Server                      │
│  ┌──────────────────────────────────────┐  │
│  │  ResourceDefinition Registry (future)                │  │
│  │  - Stores ResourceDefinition definitions             │  │
│  │  - Validates resources                 │  │
│  └──────────────────────────────────────┘  │
│                                              │
│  ┌──────────────────────────────────────┐  │
│  │  Process Queue                        │  │
│  │  - Converts CustomResource → Process  │  │
│  │  - Routes to correct executor type    │  │
│  └──────────────────────────────────────┘  │
└──────────────────┬──────────────────────────┘
                   │
                   │ Assign Process
                   ▼
┌─────────────────────────────────────────────┐
│    Resource Controller (Executor)           │
│  ┌──────────────────────────────────────┐  │
│  │  Reconciliation Loop                  │  │
│  │  1. Get Process                       │  │
│  │  2. Extract CustomResource            │  │
│  │  3. Compare desired vs current        │  │
│  │  4. Take action                       │  │
│  │  5. Update status                     │  │
│  └──────────────────────────────────────┘  │
└──────────────────┬──────────────────────────┘
                   │
                   │ Deploy/Manage
                   ▼
┌─────────────────────────────────────────────┐
│         Infrastructure                       │
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌────────┐ │
│  │Docker│  │ K8s  │  │ HPC  │  │Process │ │
│  └──────┘  └──────┘  └──────┘  └────────┘ │
└─────────────────────────────────────────────┘
```

## Cleanup

```bash
# Stop controller (Ctrl+C in controller terminal)

# Stop and remove Docker containers created by controller
docker ps -a | grep ml-training-executors | awk '{print $1}' | xargs docker rm -f

# Stop ColonyOS services
docker-compose down

# Remove volumes (WARNING: deletes database)
docker-compose down -v
```

## Troubleshooting

### Controller not receiving work
- Check executor type matches ResourceDefinition handler
- Verify colony name is correct
- Check ColonyOS server logs

### Docker containers not starting
- Verify Docker daemon is running
- Check Docker socket permissions
- Review controller logs for errors

### Resource not found
- Ensure ResourceDefinition is registered first
- Validate CustomResource against schema
- Check resource namespace matches colony

## Next Steps

1. **Add Database Storage**: Store ResourceDefinitions and CustomResources in TimescaleDB
2. **Implement CRUD API**: REST API for managing ResourceDefinitions and resources
3. **Add Validation**: Schema validation for CustomResources
4. **Status Updates**: Bidirectional status sync
5. **Multiple Controllers**: Deploy multiple controller replicas
6. **Kubernetes Backend**: Implement actual K8s deployment
7. **Monitoring**: Add metrics and observability

## References

- [ColonyOS Documentation](https://colonyos.io/docs)
- [Kubernetes CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
