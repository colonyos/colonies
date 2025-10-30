# Resource Management Examples

This directory contains examples for managing custom resources in ColonyOS, similar to Kubernetes Custom Resource Definitions (CRDs).

## Overview

The resource management system consists of two main components:

1. **ResourceDefinitions** - Define the schema and structure of custom resources (like Kubernetes CRDs)
2. **Resources** - Instances of custom resources based on ResourceDefinitions

## Files

- `executor-deployment-definition.json` - Defines the ExecutorDeployment resource type
- `executor-deployment-instance.json` - An instance of an ExecutorDeployment resource
- `gitops-example-definition.json` - Example ResourceDefinition with GitOps configuration
- `example-deployment-resource.json` - Example resource that would be stored in a Git repository
- `GITOPS.md` - Complete guide to using GitOps with ColonyOS resources
- `SCHEMA-GUIDE.md` - Guide to schema validation

## Usage

### 1. Add a ResourceDefinition (Colony Owner Only)

ResourceDefinitions can only be added by colony owners:

```bash
# Set colony owner private key
export COLONIES_PRVKEY=${COLONIES_COLONY_PRVKEY}

# Add the ResourceDefinition
colonies resource definition add --spec examples/deployment/executor-deployment-definition.json
```

### 2. List ResourceDefinitions

```bash
# Members or colony owners can list
colonies resource definition ls
```

### 3. Get a Specific ResourceDefinition

```bash
colonies resource definition get --name executordeployments.compute.colonies.io
```

### 4. Add a Resource Instance

Members can add resource instances:

```bash
# Use member private key
export COLONIES_PRVKEY=${COLONIES_PRVKEY}

# Add a resource
colonies resource add --spec examples/deployment/executor-deployment-instance.json
```

### 5. List Resources

```bash
# List all resources
colonies resource ls

# Filter by kind
colonies resource ls --kind ExecutorDeployment
```

### 6. Get a Specific Resource

```bash
colonies resource get --name web-server-deployment
```

### 7. Update a Resource

```bash
# Modify the JSON file, then:
colonies resource update --spec examples/deployment/executor-deployment-instance.json
```

### 8. Remove a Resource

```bash
colonies resource remove --name web-server-deployment
```

### 9. Sync Resources from Git (GitOps)

If a ResourceDefinition has GitOps configuration, you can sync resources from a Git repository:

```bash
# Dry run to see what would be synced
colonies resource sync --definition deployments.example.io --dry-run

# Sync resources from Git
colonies resource sync --definition deployments.example.io
```

See [GITOPS.md](GITOPS.md) for complete GitOps documentation.

## Authorization

- **ResourceDefinition operations**: Only colony owners can create ResourceDefinitions
- **Resource operations**: Both members and colony owners can manage Resources
- **Read operations**: Members and colony owners can read both ResourceDefinitions and Resources

## Example ResourceDefinition Structure

```json
{
  "metadata": {
    "name": "executordeployments.compute.colonies.io"
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
          "description": "Number of executor replicas to run",
          "default": 1
        },
        "executorType": {
          "type": "string",
          "description": "Type of executor to deploy"
        },
        "cpu": {
          "type": "string",
          "description": "CPU resource request"
        },
        "memory": {
          "type": "string",
          "description": "Memory resource request"
        }
      },
      "required": ["image", "executorType"]
    },
    "handler": {
      "executorType": "executor-controller",
      "functionName": "reconcile"
    }
  }
}
```

### Schema Definition

The `schema` field defines the structure and validation rules for Resource instances:

- **type**: Data type (object, string, number, array, boolean)
- **properties**: Nested field definitions
- **required**: List of required field names
- **description**: Human-readable description
- **default**: Default value if not specified
- **enum**: List of allowed values
- **items**: Schema for array elements

This ensures that Resource instances conform to the expected structure.

## Example Resource Structure

```json
{
  "apiVersion": "compute.colonies.io/v1",
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "web-server-deployment"
  },
  "spec": {
    "image": "nginx:1.21",
    "replicas": 3,
    "executorType": "container-executor",
    "cpu": "500m",
    "memory": "512Mi"
  }
}
```

## Reconciliation

Resources can trigger reconciliation by executors:
- The `handler.executorType` specifies which executor type handles this resource
- The `handler.functionName` specifies the function to call for reconciliation
- Executors can watch for resource changes and reconcile the desired state
