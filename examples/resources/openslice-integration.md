# OpenSlice Integration with ColonyOS

This document describes how OpenSlice can interface with ColonyOS to provide service catalog and orchestration capabilities for CRD-based deployments.

## Architecture Overview

OpenSlice serves as the **Service Layer** while ColonyOS serves as the **Resource Layer**:

- **OpenSlice**: Service catalog, business logic, customer portal, SLA management
- **ColonyOS**: Resource orchestration, multi-cloud deployment, executor management

## Integration Pattern

### 1. Service Definition in OpenSlice

OpenSlice defines service specifications using TMF APIs (TM Forum):

```json
{
  "serviceSpecification": {
    "name": "Edge AI Inference Service",
    "version": "1.0",
    "description": "AI inference service deployed at network edge",
    "serviceSpecCharacteristic": [
      {
        "name": "model_type",
        "valueType": "string",
        "serviceSpecCharacteristicValue": ["pytorch", "tensorflow", "onnx"]
      },
      {
        "name": "gpu_required",
        "valueType": "boolean"
      },
      {
        "name": "edge_locations",
        "valueType": "array",
        "description": "List of edge deployment locations"
      },
      {
        "name": "qos_class",
        "valueType": "string",
        "serviceSpecCharacteristicValue": ["low-latency", "high-throughput", "best-effort"]
      }
    ]
  }
}
```

### 2. Service Order Translation

When a customer orders the service through OpenSlice, it translates to ColonyOS CRDs:

**OpenSlice Service Order** → **ColonyOS Custom Resources**

#### Service Order in OpenSlice:
```json
{
  "serviceOrder": {
    "externalId": "order-12345",
    "category": "Edge AI",
    "orderItem": [
      {
        "service": {
          "serviceSpecification": {
            "name": "Edge AI Inference Service"
          },
          "serviceCharacteristic": [
            {
              "name": "model_type",
              "value": "pytorch"
            },
            {
              "name": "model_uri",
              "value": "s3://models/object-detection-v2.pt"
            },
            {
              "name": "edge_locations",
              "value": ["stockholm", "london", "paris"]
            },
            {
              "name": "qos_class",
              "value": "low-latency"
            },
            {
              "name": "gpu_required",
              "value": true
            }
          ]
        }
      }
    ]
  }
}
```

#### Generated ColonyOS CRDs:

**1. MLModel Custom Resource** (for each location):
```json
{
  "apiVersion": "ml.colonies.io/v1beta1",
  "kind": "MLModel",
  "metadata": {
    "name": "edge-ai-stockholm",
    "namespace": "edge-colony-stockholm",
    "labels": {
      "openslice.order": "order-12345",
      "openslice.service": "Edge-AI-Inference-Service",
      "location": "stockholm",
      "qos": "low-latency"
    },
    "annotations": {
      "openslice.customer": "customer-789",
      "openslice.sla": "gold-tier"
    }
  },
  "spec": {
    "modelUri": "s3://models/object-detection-v2.pt",
    "framework": "pytorch",
    "servingConfig": {
      "gpuMemory": "8Gi",
      "autoscaling": {
        "minReplicas": 2,
        "maxReplicas": 10,
        "targetLatency": "100ms"
      }
    }
  }
}
```

**2. ExecutorDeployment Custom Resource** (for GPU executors):
```json
{
  "apiVersion": "compute.colonies.io/v1",
  "kind": "ExecutorDeployment",
  "metadata": {
    "name": "gpu-executors-stockholm",
    "namespace": "edge-colony-stockholm",
    "labels": {
      "openslice.order": "order-12345",
      "location": "stockholm"
    }
  },
  "spec": {
    "runtime": "kubernetes",
    "replicas": 3,
    "template": {
      "type": "gpu-executor",
      "capabilities": {
        "hardware": {
          "gpu": {
            "name": "nvidia-t4",
            "count": 1
          }
        }
      }
    },
    "config": {
      "nodeSelector": {
        "location": "stockholm",
        "zone": "edge"
      }
    }
  }
}
```

### 3. OpenSlice Adapter Implementation

Create an adapter service that translates OpenSlice service orders to ColonyOS CRDs:

```go
// pkg/adapters/openslice/adapter.go
package openslice

import (
    "github.com/colonyos/colonies/pkg/core"
    "github.com/colonyos/colonies/pkg/client"
)

type OpenSliceAdapter struct {
    coloniesClient *client.ColoniesClient
    templateEngine *TemplateEngine
}

// TranslateServiceOrder converts an OpenSlice service order to ColonyOS CRDs
func (a *OpenSliceAdapter) TranslateServiceOrder(order *ServiceOrder) ([]*core.CustomResource, error) {
    var resources []*core.CustomResource

    // Extract service characteristics
    characteristics := extractCharacteristics(order)

    // Generate ResourceDefinitions based on service type
    switch order.ServiceType {
    case "Edge-AI-Inference":
        resources = a.generateEdgeAIResources(characteristics)
    case "5G-Network-Slice":
        resources = a.generateNetworkSliceResources(characteristics)
    case "NFV-Service-Chain":
        resources = a.generateServiceChainResources(characteristics)
    default:
        return nil, fmt.Errorf("unsupported service type: %s", order.ServiceType)
    }

    // Add OpenSlice metadata to all resources
    for _, r := range resources {
        r.SetLabel("openslice.order", order.ID)
        r.SetLabel("openslice.customer", order.CustomerID)
        r.SetAnnotation("openslice.service-spec", order.ServiceSpecification)
    }

    return resources, nil
}

func (a *OpenSliceAdapter) generateEdgeAIResources(char ServiceCharacteristics) []*core.CustomResource {
    var resources []*core.CustomResource

    // Get edge locations
    locations := char.GetStringArray("edge_locations")
    modelURI := char.GetString("model_uri")
    framework := char.GetString("model_type")
    gpuRequired := char.GetBool("gpu_required")
    qosClass := char.GetString("qos_class")

    // Create MLModel resource for each location
    for _, location := range locations {
        colonyName := fmt.Sprintf("edge-colony-%s", location)

        // Create ML Model resource
        mlModel := core.CreateCustomResource(
            "ml.colonies.io/v1beta1",
            "MLModel",
            fmt.Sprintf("edge-ai-%s", location),
            colonyName,
        )

        mlModel.SetSpec("modelUri", modelURI)
        mlModel.SetSpec("framework", framework)
        mlModel.SetSpec("servingConfig", map[string]interface{}{
            "autoscaling": map[string]interface{}{
                "minReplicas": a.getMinReplicasForQoS(qosClass),
                "maxReplicas": a.getMaxReplicasForQoS(qosClass),
                "targetLatency": a.getTargetLatencyForQoS(qosClass),
            },
        })

        mlModel.SetLabel("location", location)
        mlModel.SetLabel("qos", qosClass)

        resources = append(resources, mlModel)

        // Create executor deployment if GPU required
        if gpuRequired {
            execDeploy := a.createGPUExecutorDeployment(location, colonyName, qosClass)
            resources = append(resources, execDeploy)
        }
    }

    return resources
}

// StatusCallback handles status updates from ColonyOS and reports back to OpenSlice
func (a *OpenSliceAdapter) StatusCallback(resource *core.CustomResource) error {
    // Extract OpenSlice order ID
    orderID, ok := resource.GetLabel("openslice.order")
    if !ok {
        return fmt.Errorf("no openslice.order label found")
    }

    // Get resource status
    phase, _ := resource.GetStatus("phase")

    // Map to OpenSlice service state
    serviceState := a.mapToServiceState(phase)

    // Update OpenSlice via API
    return a.updateOpenSliceServiceState(orderID, serviceState, resource.Status)
}

func (a *OpenSliceAdapter) mapToServiceState(phase interface{}) string {
    switch phase {
    case "Pending":
        return "acknowledged"
    case "Running":
        return "active"
    case "Failed":
        return "terminated"
    default:
        return "inProgress"
    }
}
```

### 4. Deployment Workflow

```
┌─────────────┐
│  Customer   │
│   Portal    │
└──────┬──────┘
       │ 1. Service Order
       ▼
┌─────────────────────┐
│    OpenSlice        │
│  Service Catalog    │
│     & BSS/OSS       │
└──────┬──────────────┘
       │ 2. Validate & Approve
       │ 3. Generate CRDs
       ▼
┌─────────────────────┐
│ OpenSlice Adapter   │
│  (Translation)      │
└──────┬──────────────┘
       │ 4. Submit CRDs
       ▼
┌─────────────────────┐
│    ColonyOS         │
│  Resource Manager   │
└──────┬──────────────┘
       │ 5. Create Processes
       │ 6. Assign to Controllers
       ▼
┌─────────────────────┐
│  Resource           │
│  Controllers        │
│  (Executors)        │
└──────┬──────────────┘
       │ 7. Deploy Infrastructure
       ▼
┌─────────────────────┐
│   K8s / Edge /      │
│   HPC / Cloud       │
└──────┬──────────────┘
       │ 8. Status Updates
       │
       ▼
    (feedback loop back to OpenSlice)
```

## Use Case Examples

### 1. 5G Network Slice with Edge Compute

**OpenSlice Service**: "5G Low-Latency Gaming Slice"

**Generated ColonyOS CRDs**:
- `NetworkSlice` - Defines 5G slice parameters
- `ExecutorDeployment` - Deploys game server executors at edge
- `Database` - Redis cache at each edge location
- `Workflow` - Traffic routing and load balancing configuration

### 2. Multi-Tenant NFV Service Chain

**OpenSlice Service**: "vCPE Service Chain"

**Generated ColonyOS CRDs**:
- `ServiceChain` - Defines VNF ordering (Firewall → Router → NAT)
- `VNF` - Each virtual network function deployment
- `ExecutorDeployment` - NFV executors on appropriate infrastructure

### 3. Edge AI Video Analytics

**OpenSlice Service**: "Smart City Video Analytics"

**Generated ColonyOS CRDs**:
- `MLModel` - Computer vision models at each camera location
- `ExecutorDeployment` - GPU executors for inference
- `Workflow` - Data ingestion → Processing → Storage pipeline
- `Database` - TimescaleDB for analytics storage

## Implementation Considerations

### 1. Mapping Strategy
- Define templates for common service patterns
- Use OpenSlice service characteristics to parameterize CRDs
- Maintain bidirectional mapping (order ↔ resources)

### 2. State Synchronization
- ColonyOS resource status → OpenSlice service state
- Handle partial failures gracefully
- Implement reconciliation loops

### 3. Multi-Tenancy
- Use ColonyOS namespaces for customer isolation
- Map OpenSlice customers to ColonyOS colonies
- Implement RBAC policies

### 4. Lifecycle Management
- Service activation → ResourceDefinition creation
- Service modification → ResourceDefinition updates
- Service termination → ResourceDefinition deletion
- Cleanup and garbage collection

## Alternative: Lightweight Approaches

If OpenSlice is too heavyweight, consider:

1. **Custom API Gateway**: Simple REST API that generates CRDs
2. **GitOps**: Store ResourceDefinitions in Git, use ArgoCD/Flux style reconciliation
3. **Terraform Provider**: ColonyOS provider for infrastructure as code
4. **Direct CLI/SDK**: Users create ResourceDefinitions directly via colonies CLI

## Conclusion

OpenSlice + ColonyOS makes excellent sense for:
- **Telecom operators** deploying 5G services
- **Service providers** with complex service catalogs
- **NFV/Edge computing** scenarios
- Organizations needing **business logic separation** from infrastructure

For simpler use cases, direct ResourceDefinition management via API/CLI may be sufficient.
