# Blueprints

This directory contains reusable blueprint specifications for common deployments.

## Files

### executor-deployment-definition.json
The **BlueprintDefinition** for ExecutorDeployment kind. This must be registered once (by colony owner) before creating ExecutorDeployment blueprints.

**Register once:**
```bash
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}
colonies blueprint definition add --spec executor-deployment-definition.json
```

### local-docker-executor-deployment.json
Deploys a docker executor specifically on the **local/main node**.

**Key Settings:**
- `executorType`: `docker-reconciler-home-linux-server` - Uses the home Linux server reconciler
- `executorName`: `docker-reconciler-home-linux-server` - Targets the home Linux server
- `replicas`: 1 - Single executor instance

**Deploy:**
```bash
colonies blueprint add --spec local-docker-executor-deployment.json
```

**Result:** The deployment will run specifically on the `docker-reconciler-home-linux-server`.

## Executor Targeting Examples

### Example 1: Target Specific Node (Pinned - Current)
```json
{
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "docker-executor"
  },
  "handler": {
    "executorType": "docker-reconciler-home-linux-server",
    "executorName": "docker-reconciler-home-linux-server"
  },
  "spec": {
    "image": "colonyos/dockerexecutor:latest",
    "executorType": "container-executor"
  }
}
```
✅ Guaranteed deployment on specific node
⚠️ Fails if that reconciler is down

### Example 2: Target Apple Ultra Node
```json
{
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "docker-executor-apple"
  },
  "handler": {
    "executorType": "docker-reconciler-apple-ultra",
    "executorName": "docker-reconciler-apple-ultra"
  },
  "spec": {
    "image": "colonyos/dockerexecutor:latest",
    "executorType": "container-executor"
  }
}
```
✅ Guaranteed deployment on Apple Ultra
⚠️ Fails if that reconciler is down

**Available reconcilers in default setup:**
- `docker-reconciler-home-linux-server` - Home Linux server (Intel i9 + RTX 3080 Ti)
- `docker-reconciler-apple-ultra` - Home Apple Ultra (Mac Studio M2 Ultra)

### Example 3: Target Edge Node
```json
{
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "docker-executor-edge"
  },
  "handler": {
    "executorType": "docker-reconciler-edge",
    "executorName": "docker-reconciler-edge"
  },
  "spec": {
    "image": "colonyos/dockerexecutor:latest",
    "executorType": "container-executor"
  }
}
```

See [../../executors/docker-reconciler/examples/](../../executors/docker-reconciler/examples/) for more examples.

## Usage Workflow

### 1. Register Blueprint Definition (One-time)
```bash
# As colony owner
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}
colonies blueprint definition add --spec executor-deployment-definition.json
```

### 2. Deploy Executor
```bash
# Deploy to any available node
colonies blueprint add --spec local-docker-executor-deployment.json

# Check status
colonies blueprint get --name docker-executor

# Watch reconciliation
colonies process ps
```

### 3. Scale Deployment
```bash
# Scale to 3 replicas
colonies blueprint set --name docker-executor --key spec.replicas --value 3

# Scale down to 1
colonies blueprint set --name docker-executor --key spec.replicas --value 1
```

### 4. Update Image
```bash
colonies blueprint set --name docker-executor \
  --key spec.image --value colonyos/dockerexecutor:v1.0.8
```

### 5. Monitor
```bash
# List all blueprints
colonies blueprint ls

# View history
colonies blueprint history --name docker-executor

# Check running executors
colonies executor ls

# Check running containers (on reconciler node)
docker ps --filter label=colonies.blueprint=docker-executor
```

## Environment Configuration

The example includes complete environment configuration for:
- **ColonyOS Connection**: Server host, port, security
- **Colony Credentials**: Name and private key
- **S3/MinIO Storage**: For file operations
- **Executor Metadata**: Type, capabilities, location

All environment variables can be customized via blueprint updates:
```bash
colonies blueprint set --name docker-executor \
  --key spec.env.EXECUTOR_GPU --value 1
```

## Network Configuration

**Important:** The examples use `COLONIES_SERVER_HOST=colonies-server` which works when:
- Both reconciler and colonies-server are on the same Docker network
- The network has the service name `colonies-server` defined

If running reconcilers outside Docker or on different networks, use:
- `host.docker.internal` (Docker Desktop on Mac/Windows)
- Host IP address (e.g., `192.168.1.100`)
- Never use `localhost` inside containers

## Volumes

The examples mount two volumes:
1. `/var/run/docker.sock` - Required for Docker API access (Docker-in-Docker)
2. `/tmp/colonies` - Shared filesystem for data exchange

**Security Note:** Mounting Docker socket gives container full Docker API access. Use `privileged: true` only when necessary.

## See Also

- [../../docs/Blueprints.md](../../docs/Blueprints.md) - Complete blueprint documentation
- [../../docs/Reconciliation.md](../../docs/Reconciliation.md) - How reconciliation works
- [../../executors/docker-reconciler/README.md](../../executors/docker-reconciler/README.md) - Reconciler documentation
- [../../executors/docker-reconciler/examples/](../../executors/docker-reconciler/examples/) - More examples
