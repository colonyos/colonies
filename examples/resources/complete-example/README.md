# Complete ExecutorDeployment ResourceDefinition Example

This is a **complete, working example** demonstrating the Resource Definition pattern in ColonyOS, from definition to implementation.

## 📁 What's Included

| File | Purpose |
|------|---------|
| `executor-deployment-definition.json` | ResourceDefinition definition for ExecutorDeployment resource type |
| `ml-executor-deployment.json` | Example CustomResource instance |
| `controller.go` | Full controller implementation that reconciles resources |
| `main.go` | CLI tool for registration, submission, and running controller |
| `docker-compose.yml` | Complete test environment (ColonyOS + TimescaleDB) |
| `run-demo.sh` | Quick demo script (no server required) |
| `USAGE.md` | Detailed usage guide and examples |

## 🚀 Quick Start

### Option 1: Demo Mode (No Server Required)

```bash
# Run the complete demo
./run-demo.sh
```

This demonstrates the entire ResourceDefinition workflow without needing a running ColonyOS server.

### Option 2: Full Integration Test

```bash
# 1. Start ColonyOS
docker-compose up -d

# 2. Build the example
go build -o crd-example .

# 3. Register the ResourceDefinition
./crd-example -mode register-crd

# 4. Start the controller (in a new terminal)
./crd-example -mode controller

# 5. Submit a resource (in another terminal)
./crd-example -mode submit

# 6. Watch the controller logs to see reconciliation!
```

## 📋 What This Example Shows

### 1. **Resource Definition**
```json
{
  "kind": "CustomResourceDefinition",
  "spec": {
    "names": {
      "kind": "ExecutorDeployment"
    },
    "handler": {
      "executorType": "executor-deployment-controller",
      "functionName": "reconcile_executor_deployment"
    }
  }
}
```

### 2. **CustomResource Instance**
```json
{
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "ml-training-executors"
  },
  "spec": {
    "runtime": "docker",
    "replicas": 3,
    "executorType": "ml-executor"
  }
}
```

### 3. **Controller Implementation**
The controller (`controller.go`):
- ✅ Polls ColonyOS for reconciliation work
- ✅ Extracts CustomResource from Process
- ✅ Compares desired state vs current state
- ✅ Deploys Docker containers as executors
- ✅ Scales up/down based on replica count
- ✅ Updates resource status
- ✅ Handles multiple runtime backends (Docker, K8s, local)

### 4. **Complete Workflow**
```
User submits CustomResource
        ↓
ColonyOS Server:
  - Looks up Resource Definition
  - Converts to Process (by attaching Resource to FunctionSpec)
  - Adds tracking attributes
  - Queues Process
        ↓
Controller receives Process
        ↓
Controller extracts CustomResource
        ↓
Controller reconciles (deploy/scale)
        ↓
Controller updates status
        ↓
Process completes
```

## 🎯 Key Features Demonstrated

| Feature | Implementation |
|---------|----------------|
| **Resource Definition Registration** | Validates and stores resource definitions |
| **Resource Validation** | Checks required fields and types |
| **Process Conversion** | CustomResource → FunctionSpec → Process |
| **Reconciliation Loop** | Controller continuously reconciles state |
| **Multi-Runtime Support** | Docker, Kubernetes, local processes |
| **Status Updates** | Bidirectional state synchronization |
| **Scaling** | Automatic scale up/down based on replicas |
| **Error Handling** | Graceful failure and retry logic |

## 📖 Example Output

### Submitting a Resource
```
$ ./crd-example -mode submit
Submitting CustomResource from file: ml-executor-deployment.json
Resource Details:
  API Version: compute.colonies.io/v1
  Kind: ExecutorDeployment
  Name: ml-training-executors
  Namespace: dev
  Spec:
    runtime: docker
    replicas: 3
    executorType: ml-executor
    image: colonyos/ml-executor:latest

✓ Resource converted to Process
```

### Controller Reconciling
```
$ ./crd-example -mode controller
Starting controller:
  Colony: dev
  Executor Type: executor-deployment-controller
Controller is running. Press Ctrl+C to stop.

Assigned process: abc123 (label: ExecutorDeployment/ml-training-executors)
Reconciling ExecutorDeployment/ml-training-executors
  Runtime: docker
  Replicas: 3
  ExecutorType: ml-executor
  Current replicas: 0, Desired: 3
Scaling up: deploying 3 new Docker containers
Started container: ml-training-executors-abc12345-0 (ID: def456789012)
Started container: ml-training-executors-abc12345-1 (ID: ghi789012345)
Started container: ml-training-executors-abc12345-2 (ID: jkl012345678)
Reconciliation complete: 3 replicas running
```

### Verifying Deployment
```bash
$ docker ps | grep ml-training-executors
def456789012   colonyos/ml-executor:latest   "..."   ml-training-executors-abc12345-0
ghi789012345   colonyos/ml-executor:latest   "..."   ml-training-executors-abc12345-1
jkl012345678   colonyos/ml-executor:latest   "..."   ml-training-executors-abc12345-2
```

## 🔧 Customization

### Change Runtime

**Docker:**
```json
{"spec": {"runtime": "docker", "replicas": 3}}
```

**Kubernetes:**
```json
{"spec": {"runtime": "kubernetes", "replicas": 5}}
```

**Local Processes:**
```json
{"spec": {"runtime": "local", "replicas": 2}}
```

### Add Environment Variables

```json
{
  "spec": {
    "config": {
      "env": {
        "MODEL_CACHE_DIR": "/models",
        "BATCH_SIZE": "32"
      }
    }
  }
}
```

### Add Volume Mounts

```json
{
  "spec": {
    "config": {
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

## 🏗️ Architecture

```
┌─────────────┐
│    User     │
└──────┬──────┘
       │ Submit CustomResource
       ▼
┌─────────────────────┐
│   ColonyOS Server   │
│  ┌──────────────┐  │
│  │ ResourceDefinition Registry │  │
│  └──────────────┘  │
│  ┌──────────────┐  │
│  │Process Queue │  │
│  └──────────────┘  │
└──────┬──────────────┘
       │ Assign Process
       ▼
┌─────────────────────┐
│    Controller       │
│  ┌──────────────┐  │
│  │ Reconcile    │  │
│  │ Loop         │  │
│  └──────────────┘  │
└──────┬──────────────┘
       │ Deploy
       ▼
┌─────────────────────┐
│  Infrastructure     │
│  ┌─────┐ ┌──────┐ │
│  │Docker│ │ K8s  │ │
│  └─────┘ └──────┘ │
└─────────────────────┘
```

## 📚 Learning Path

1. **Start with demo** - Run `./run-demo.sh` to see the workflow
2. **Read USAGE.md** - Comprehensive guide with examples
3. **Study controller.go** - See how reconciliation works
4. **Run full test** - Deploy with Docker Compose
5. **Modify examples** - Change replicas, runtime, config
6. **Create your own CRD** - Follow the pattern for new resource types

## 🎓 Advanced Topics

### Creating Custom CRDs

See `USAGE.md` section "Creating Your Own CRD" for:
- Database deployment CRD
- ML model serving CRD
- Workflow orchestration CRD

### OpenSlice Integration

See `../openslice-integration.md` for:
- Telecom/5G use cases
- Service catalog integration
- NFV orchestration patterns

## 🐛 Troubleshooting

| Problem | Solution |
|---------|----------|
| Controller not receiving work | Check executor type matches ResourceDefinition handler |
| Docker containers not starting | Verify Docker daemon is running |
| Permission denied on Docker socket | Add user to docker group or run as root |
| Resource validation fails | Check all required fields are present |

## 📝 License

Part of the ColonyOS project.

## 🤝 Contributing

This example demonstrates the ResourceDefinition pattern. To extend it:

1. Add new runtime backends (HPC, Lambda, etc.)
2. Implement ResourceDefinition storage in database
3. Add REST API for ResourceDefinition management
4. Implement resource watching/streaming
5. Add validation webhooks

## 🔗 See Also

- [ColonyOS Documentation](https://colonyos.io/docs)
- [CRD Core Implementation](../../../pkg/core/custom_resource.go)
- [Example ResourceDefinitions and Resources](../)
- [OpenSlice Integration Guide](../openslice-integration.md)
