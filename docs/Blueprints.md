# Blueprint Management

ColonyOS provides a Kubernetes-inspired blueprint management system that allows you to define and manage custom resources with declarative specifications, schema validation, and automated reconciliation.

## Overview

The blueprint management system consists of three main components:

1. **BlueprintDefinitions** - Define the schema and structure of custom blueprints (similar to Kubernetes Custom Resource Definitions)
2. **Blueprints** - Instances of custom blueprints based on BlueprintDefinitions
3. **Reconcilers** - Executors that receive reconciliation processes and reconcile the desired state

## Core Concepts

### BlueprintDefinition

A BlueprintDefinition defines:
- The **kind** of blueprint (e.g., `ExecutorDeployment`, `Database`, `MLModel`)
- The **schema** that validates blueprint instances
- The **handler** that specifies which executor type reconciles this blueprint kind

### Blueprint

A Blueprint is an instance of a BlueprintDefinition that contains:
- **Metadata** - Name, namespace, labels, annotations
- **Spec** - The desired state (validated against the BlueprintDefinition schema)
- **Status** - The current state (populated by reconcilers)

### Reconciler

A reconciler is an executor that:
- Registers with a specific `executorType` to handle a blueprint kind
- Receives reconciliation processes assigned by the server when blueprints are created, updated, or deleted
- Processes contain the reconciliation action (create/update/delete) and old/new blueprint state
- Takes actions to make the actual state match the desired state
- Updates the blueprint status with current state information

## Quick Start

### 1. Add a BlueprintDefinition (Colony Owner Only)

BlueprintDefinitions can only be added by colony owners:

```bash
# Set colony owner private key
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}

# Add the BlueprintDefinition
colonies blueprint definition add --spec executor-deployment-definition.json
```

Example BlueprintDefinition:
```json
{
  "metadata": {
    "name": "executor-deployment"
  },
  "spec": {
    "group": "compute.colonies.io",
    "version": "v1",
    "names": {
      "kind": "ExecutorDeployment",
      "plural": "executordeployments",
      "singular": "executordeployment"
    },
    "scope": "Namespaced",
    "schema": {
      "type": "object",
      "properties": {
        "image": {
          "type": "string",
          "description": "Container image to deploy"
        },
        "replicas": {
          "type": "number",
          "description": "Number of instances to run",
          "default": 1
        },
        "executorType": {
          "type": "string",
          "description": "Type of executor instances"
        }
      },
      "required": ["image", "executorType"]
    },
    "handler": {
      "executorType": "docker-reconciler-home-linux-server",
      "functionName": "reconcile"
    }
  }
}
```

### 2. List BlueprintDefinitions

```bash
colonies blueprint definition ls
```

### 3. Get a Specific BlueprintDefinition

```bash
colonies blueprint definition get --name executor-deployment
```

### 4. Add a Blueprint Instance

Members and colony owners can add blueprint instances:

```bash
# Add a blueprint
colonies blueprint add --spec local-docker-executor-deployment.json
```

Example Blueprint:
```json
{
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "docker-executor"
  },
  "handler": {
    "executorType": "docker-reconciler-home-linux-server"
  },
  "spec": {
    "image": "colonyos/dockerexecutor:v1.0.5",
    "replicas": 3,
    "executorType": "container-executor",
    "env": {
      "COLONIES_SERVER_HOST": "colonies-server",
      "COLONIES_SERVER_PORT": "50080"
    },
    "volumes": [
      {
        "host": "/var/run/docker.sock",
        "container": "/var/run/docker.sock"
      }
    ]
  }
}
```

### 5. List Blueprints

```bash
# List all blueprints
colonies blueprint ls

# Filter by kind
colonies blueprint ls --kind ExecutorDeployment
```

### 6. Get Blueprint Status

```bash
colonies blueprint get --name docker-executor
```

This displays:
- **Blueprint metadata** - Name, ID, kind, generation
- **Spec** - Desired configuration (image, replicas, env vars, volumes)
- **Deployment Status** - Running/total instances
- **Instances** - Detailed information about each running instance

Example output:
```
╭──────────────────────────────────────────╮
│ Deployment Status                        │
├───────────────────┬──────────────────────┤
│ Running Instances │ 3                    │
│ Total Instances   │ 3                    │
│ Last Updated      │ 2025-11-08T15:17:00Z │
╰───────────────────┴──────────────────────╯

╭─────────────────────────────────────────────────────────────────────╮
│ Instances                                                           │
├──────────────┬────────┬───────────┬─────────┬──────────┬────────────┤
│ NAME         │ ID     │ TYPE      │ STATE   │ IMAGE    │ LAST CHECK │
├──────────────┼────────┼───────────┼─────────┼──────────┼────────────┤
│ executor-a1b │ 5884aa │ container │ running │ nginx... │ 15:17:00   │
│ executor-c2d │ d55e6c │ container │ running │ nginx... │ 15:17:00   │
│ executor-e3f │ 2efcd9 │ container │ running │ nginx... │ 15:17:00   │
╰──────────────┴────────┴───────────┴─────────┴──────────┴────────────╯
```

### 7. Update a Blueprint

```bash
# Modify a field (e.g., scale replicas)
colonies blueprint set --name docker-executor --key spec.replicas --value 5

# Or update entire spec
colonies blueprint update --spec updated-deployment.json
```

### 8. Remove a Blueprint

```bash
colonies blueprint remove --name docker-executor
```

## Authorization

- **BlueprintDefinition operations**: Only colony owners can create/update/delete BlueprintDefinitions
- **Blueprint operations**: Both members and colony owners can manage Blueprints
- **Read operations**: Members and colony owners can read both BlueprintDefinitions and Blueprints

## Schema Validation

### Server-Side Validation

The ColonyOS server **automatically validates** all blueprint instances against their BlueprintDefinition schema before saving them. This ensures data integrity and catches errors early.

Validation checks:
- **Required fields** - All fields in `schema.required` must be present
- **Type validation** - Values match their declared types (string, number, boolean, object, array)
- **Enum constraints** - Values are within allowed enum values
- **Array item validation** - Each array item matches the item schema
- **Nested object validation** - Recursive validation of nested structures

If validation fails, the server returns HTTP 400 Bad Request with a detailed error message.

### Example Validation Error

```bash
$ colonies blueprint add --spec invalid-blueprint.json

Error: blueprint validation failed: required field 'image' is missing
```

### Schema Features

The schema system supports full JSON Schema features:

**Basic Types:**
```json
{
  "properties": {
    "name": { "type": "string" },
    "replicas": { "type": "number" },
    "enabled": { "type": "boolean" }
  }
}
```

**Enum Constraints:**
```json
{
  "properties": {
    "size": {
      "type": "string",
      "enum": ["small", "medium", "large"]
    }
  }
}
```

**Default Values:**
```json
{
  "properties": {
    "replicas": {
      "type": "number",
      "default": 1
    }
  }
}
```

**Nested Objects:**
```json
{
  "properties": {
    "database": {
      "type": "object",
      "properties": {
        "engine": {
          "type": "string",
          "enum": ["postgresql", "mysql"]
        },
        "port": {
          "type": "number",
          "default": 5432
        }
      },
      "required": ["engine"]
    }
  }
}
```

**Arrays:**
```json
{
  "properties": {
    "ports": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "port": { "type": "number" },
          "protocol": {
            "type": "string",
            "enum": ["TCP", "UDP"]
          }
        }
      }
    }
  }
}
```

### Optional Schema

Schemas are **optional**. BlueprintDefinitions without schemas accept any structure in the blueprint spec. This is useful for:
- Rapid prototyping where structure is evolving
- Highly dynamic configurations
- When validation happens in reconciler code

## Reconciliation

### How Reconciliation Works

1. User creates/updates/deletes a blueprint
2. Server validates the blueprint against the schema
3. Server creates a reconciliation process with:
   - `reconciliation.action` - "create", "update", or "delete"
   - `reconciliation.old` - Previous blueprint state (null for create)
   - `reconciliation.new` - New desired state (null for delete)
4. Process is assigned to an executor matching the handler's `executorType`
5. Reconciler receives the process and reconciles the state
6. Reconciler updates blueprint status with current state

### Reconciliation Process Example

```json
{
  "functionSpec": {
    "funcName": "reconcile",
    "reconciliation": {
      "action": "update",
      "old": {
        "kind": "ExecutorDeployment",
        "metadata": { "name": "docker-executor" },
        "spec": { "replicas": 1 }
      },
      "new": {
        "kind": "ExecutorDeployment",
        "metadata": { "name": "docker-executor" },
        "spec": { "replicas": 3 }
      }
    }
  }
}
```

The reconciler sees:
- User scaled from 1 to 3 replicas
- Need to start 2 more instances
- Takes action to reach desired state

### Status Updates

Reconcilers update blueprint status to reflect current state:

```json
{
  "status": {
    "instances": [
      {
        "name": "docker-executor-a1b2c",
        "id": "5884aadc788d",
        "type": "container",
        "state": "running",
        "image": "colonyos/dockerexecutor:v1.0.5",
        "lastCheck": "2025-11-08T15:17:00Z"
      }
    ],
    "runningInstances": 3,
    "totalInstances": 3,
    "lastUpdated": "2025-11-08T15:17:00Z"
  }
}
```

The `type` field allows the same blueprint abstraction to work with:
- **Docker containers** - `type: "container"`
- **Kubernetes pods** - `type: "pod"`
- **WebAssembly modules** - `type: "wasm"`
- **Virtual machines** - `type: "vm"`
- **HPC jobs** - `type: "job"`

## Built-in Reconcilers

### Docker Reconciler

The `docker-reconciler` manages Docker container deployments.

**ExecutorType:** `docker-reconciler`

**Supported BlueprintDefinition:** `ExecutorDeployment`

**Features:**
- Deploys Docker containers on the same host as the reconciler
- Scales instances up/down based on replica count
- Manages environment variables and volume mounts
- Tracks instance status (running/stopped)
- Self-registers executor instances with the colony

**Example Usage:**
```bash
# Add ExecutorDeployment BlueprintDefinition
colonies blueprint definition add --spec executor-deployment-definition.json

# Deploy docker executors
colonies blueprint add --spec local-docker-executor-deployment.json

# Scale the deployment
colonies blueprint set --name docker-executor --key spec.replicas --value 5

# Check status
colonies blueprint get --name docker-executor
```

## Creating Custom Reconcilers

To create your own reconciler:

1. **Implement a reconciler** that:
   - Registers as an executor with a specific `executorType`
   - Implements the `reconcile` function
   - Handles create/update/delete actions
   - Updates blueprint status

2. **Define a BlueprintDefinition** with:
   - A unique `kind` for your resource type
   - A schema defining valid configurations
   - A handler pointing to your reconciler's executorType

3. **Deploy your reconciler** as an executor in the colony

Example reconciler skeleton:
```go
func (r *Reconciler) Reconcile(process *core.Process, blueprint *core.Blueprint) error {
    reconciliation := process.FunctionSpec.Reconciliation

    switch reconciliation.Action {
    case "create":
        return r.handleCreate(reconciliation.New)
    case "update":
        return r.handleUpdate(reconciliation.Old, reconciliation.New)
    case "delete":
        return r.handleDelete(reconciliation.Old)
    }

    return nil
}

func (r *Reconciler) CollectStatus(blueprint *core.Blueprint) (map[string]interface{}, error) {
    // Collect current state
    instances := r.listRunningInstances(blueprint)

    return map[string]interface{}{
        "instances": instances,
        "runningInstances": len(instances),
        "totalInstances": len(instances),
        "lastUpdated": time.Now().Format(time.RFC3339),
    }, nil
}
```

## Best Practices

### Schema Design

1. **Start simple** - Add schema fields incrementally
2. **Use descriptions** - Document each field for users
3. **Set sensible defaults** - Reduce configuration burden
4. **Use enums** - Constrain to valid values to prevent errors
5. **Make optional when possible** - Only require essential fields
6. **Nest logically** - Group related fields in objects

### Blueprint Management

1. **Use meaningful names** - Blueprint names should be descriptive
2. **Add labels** - Use labels for grouping and filtering
3. **Version your BlueprintDefinitions** - Use different names for breaking changes
4. **Monitor reconciliation** - Check reconciliation status and process logs
5. **Test validation** - Verify schema catches invalid configurations

### Reconciler Development

1. **Idempotent operations** - Handle being called multiple times safely
2. **Error handling** - Return clear error messages
3. **Status updates** - Always update status to reflect actual state
4. **Graceful cleanup** - Handle delete operations properly
5. **Logging** - Log actions for debugging

## See Also

- [JSON Schema Documentation](https://json-schema.org/)
- [Kubernetes Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Schema Validation Guide](SCHEMA-GUIDE.md)
