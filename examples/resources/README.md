# Custom Resources Examples

This directory contains examples of Custom Resource Definitions (CRDs) and Custom Resource instances for ColonyOS.

## Overview

ColonyOS supports a Kubernetes-inspired Custom Resource Definition (CRD) system that allows you to extend the platform with your own resource types. This enables:

- **Declarative Infrastructure**: Define executors, services, and workflows as resources
- **Operator Pattern**: Special executors act as controllers/operators to manage resources
- **Extensibility**: Create new resource types without modifying ColonyOS core
- **Runtime Flexibility**: Support any deployment target (K8s, Docker, HPC, Lambda, etc.)

## Structure

- `definitions/` - Custom Resource Definition examples
- `instances/` - Custom Resource instance examples

## Basic Concepts

### Custom Resource Definition (CRD)

A ResourceDefinition defines a new resource type. It specifies:
- The schema (group, version, kind)
- Which executor type handles the resource
- Optional validation schema

### Custom Resource (CR)

A CR is an instance of a ResourceDefinition. It contains:
- `apiVersion` and `kind` (defined by the ResourceDefinition)
- `metadata` (name, namespace, labels, annotations)
- `spec` (desired state - completely flexible)
- `status` (current state - managed by controller)

## Resource Processing Flow

1. User creates a CustomResource instance
2. ColonyOS creates a Process with the resource data
3. Process is assigned to an executor matching the ResourceDefinition's handler type
4. Executor reconciles the resource (creates, updates, or deletes infrastructure)
5. Executor updates the resource status

## Examples

### 1. Executor Deployment CRD
Defines how to deploy executors across different runtimes.

### 2. ML Model CRD
Defines how to deploy machine learning models as services.

### 3. HPC Job CRD
Defines batch jobs for HPC systems (Slurm, PBS, etc.).

### 4. Workflow CRD
Defines complex multi-step workflows.

### 5. Database CRD
Defines database provisioning and management.

## Usage

### Registering a ResourceDefinition

```bash
colonies crd create -f definitions/executor-deployment-crd.json
colonies crd list
```

### Creating a Custom Resource

```bash
colonies resource create -f instances/ml-executor-deployment.json
colonies resource get ml-executors -n ml-colony
colonies resource list -n ml-colony
```

### Implementing a Resource Controller

A controller is a special executor that watches for custom resources:

```go
type ResourceController struct {
    coloniesClient *client.ColoniesClient
}

func (rc *ResourceController) Run() {
    // Poll for processes with resource reconciliation
    for {
        funcSpec := rc.coloniesClient.Assign(...)

        // Extract resource from kwargs
        apiVersion := funcSpec.KwArgs["apiVersion"].(string)
        kind := funcSpec.KwArgs["kind"].(string)
        spec := funcSpec.KwArgs["spec"].(map[string]interface{})

        // Reconcile resource
        err := rc.reconcile(apiVersion, kind, spec)

        // Update status
        if err != nil {
            rc.coloniesClient.Failed(...)
        } else {
            rc.coloniesClient.Success(...)
        }
    }
}
```

## Best Practices

1. **Use meaningful names**: Group names should be domain-like (e.g., `compute.colonies.io`)
2. **Version your APIs**: Start with `v1alpha1`, move to `v1beta1`, then `v1`
3. **Validate schemas**: Define optional validation schemas in CRDs
4. **Use labels**: Add labels for filtering and organization
5. **Update status**: Controllers should update resource status with current state
6. **Idempotent reconciliation**: Controllers should be idempotent
7. **Use generations**: Track resource versions with `metadata.generation`

## Advanced Patterns

### Composite Resources

Create resources that spawn other resources:

```json
{
  "apiVersion": "app.colonies.io/v1",
  "kind": "Application",
  "spec": {
    "components": [
      {"kind": "Database", "spec": {...}},
      {"kind": "WebService", "spec": {...}},
      {"kind": "WorkerPool", "spec": {...}}
    ]
  }
}
```

### Status Conditions

Use structured status conditions:

```json
{
  "status": {
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "lastTransitionTime": "2025-10-20T10:00:00Z",
        "reason": "AllReplicasReady"
      }
    ]
  }
}
```

### Owner References

Track resource ownership for cleanup:

```json
{
  "metadata": {
    "annotations": {
      "ownerReferences": "[{\"kind\":\"Application\",\"name\":\"my-app\"}]"
    }
  }
}
```

## See Also

- [ColonyOS Documentation](https://colonyos.io/docs)
- [Kubernetes ResourceDefinition Documentation](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
