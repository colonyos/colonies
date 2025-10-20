# ResourceDefinition Schema Validation Guide

This guide explains how schemas work in ColonyOS Custom Resource Definitions.

## Overview

ResourceDefinitions can include an **optional** JSON Schema that validates CustomResource instances. This provides:

- ✅ Type safety
- ✅ Required field validation
- ✅ Enum constraints
- ✅ Default values
- ✅ Documentation
- ✅ Nested object validation

## Schema Structure

```json
{
  "schema": {
    "type": "object",
    "properties": {
      "fieldName": {
        "type": "string|integer|boolean|object|array",
        "description": "Field documentation",
        "enum": ["allowed", "values"],
        "default": "defaultValue"
      }
    },
    "required": ["mandatoryField1", "mandatoryField2"]
  }
}
```

## Examples

### 1. Basic Types

```json
{
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Resource name"
      },
      "replicas": {
        "type": "integer",
        "description": "Number of replicas"
      },
      "enabled": {
        "type": "boolean",
        "description": "Whether resource is enabled"
      }
    },
    "required": ["name", "replicas"]
  }
}
```

**Valid Resource:**
```json
{
  "spec": {
    "name": "my-resource",
    "replicas": 3,
    "enabled": true
  }
}
```

**Invalid - Wrong Type:**
```json
{
  "spec": {
    "name": "my-resource",
    "replicas": "three",  ← ✗ Should be integer
    "enabled": true
  }
}
```

**Invalid - Missing Required:**
```json
{
  "spec": {
    "name": "my-resource"
    // ✗ Missing required field "replicas"
  }
}
```

### 2. Enum Constraints

```json
{
  "schema": {
    "properties": {
      "size": {
        "type": "string",
        "enum": ["small", "medium", "large"]
      },
      "environment": {
        "type": "string",
        "enum": ["dev", "staging", "production"]
      }
    }
  }
}
```

**Valid:**
```json
{
  "spec": {
    "size": "medium",
    "environment": "production"
  }
}
```

**Invalid:**
```json
{
  "spec": {
    "size": "extra-large",  ← ✗ Not in enum
    "environment": "prod"    ← ✗ Not in enum (must be "production")
  }
}
```

### 3. Default Values

```json
{
  "schema": {
    "properties": {
      "replicas": {
        "type": "integer",
        "default": 1
      },
      "autoscaling": {
        "type": "boolean",
        "default": false
      }
    }
  }
}
```

**Resource without defaults specified:**
```json
{
  "spec": {
    // If omitted, replicas = 1, autoscaling = false
  }
}
```

### 4. Nested Objects

```json
{
  "schema": {
    "properties": {
      "database": {
        "type": "object",
        "properties": {
          "engine": {
            "type": "string",
            "enum": ["postgresql", "mysql"]
          },
          "version": {
            "type": "string"
          },
          "config": {
            "type": "object",
            "properties": {
              "maxConnections": {
                "type": "integer",
                "default": 100
              }
            }
          }
        },
        "required": ["engine", "version"]
      }
    }
  }
}
```

**Valid:**
```json
{
  "spec": {
    "database": {
      "engine": "postgresql",
      "version": "15.4",
      "config": {
        "maxConnections": 200
      }
    }
  }
}
```

### 5. Arrays

```json
{
  "schema": {
    "properties": {
      "ports": {
        "type": "array",
        "description": "List of exposed ports",
        "items": {
          "type": "integer"
        }
      },
      "endpoints": {
        "type": "array",
        "description": "API endpoints",
        "items": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string"
            },
            "method": {
              "type": "string",
              "enum": ["GET", "POST", "PUT", "DELETE"]
            }
          }
        }
      }
    }
  }
}
```

**Valid:**
```json
{
  "spec": {
    "ports": [8080, 8443, 9090],
    "endpoints": [
      {
        "path": "/api/v1/users",
        "method": "GET"
      },
      {
        "path": "/api/v1/users",
        "method": "POST"
      }
    ]
  }
}
```

### 6. Complex Real-World Example

**CRD:**
```json
{
  "schema": {
    "type": "object",
    "properties": {
      "runtime": {
        "type": "string",
        "enum": ["kubernetes", "docker", "hpc"],
        "description": "Deployment runtime"
      },
      "replicas": {
        "type": "integer",
        "default": 1,
        "description": "Number of instances"
      },
      "resources": {
        "type": "object",
        "properties": {
          "cpu": {
            "type": "string",
            "description": "CPU request (e.g., '2' or '500m')"
          },
          "memory": {
            "type": "string",
            "description": "Memory request (e.g., '4Gi')"
          },
          "gpu": {
            "type": "object",
            "properties": {
              "enabled": {
                "type": "boolean",
                "default": false
              },
              "type": {
                "type": "string",
                "enum": ["nvidia-t4", "nvidia-a100"]
              },
              "count": {
                "type": "integer",
                "default": 1
              }
            }
          }
        },
        "required": ["cpu", "memory"]
      },
      "autoscaling": {
        "type": "object",
        "properties": {
          "enabled": {
            "type": "boolean",
            "default": false
          },
          "minReplicas": {
            "type": "integer",
            "default": 1
          },
          "maxReplicas": {
            "type": "integer",
            "default": 10
          },
          "targetCPU": {
            "type": "integer",
            "description": "Target CPU utilization percentage",
            "default": 70
          }
        }
      }
    },
    "required": ["runtime", "resources"]
  }
}
```

**Valid Resource:**
```json
{
  "spec": {
    "runtime": "kubernetes",
    "replicas": 3,
    "resources": {
      "cpu": "4",
      "memory": "16Gi",
      "gpu": {
        "enabled": true,
        "type": "nvidia-a100",
        "count": 2
      }
    },
    "autoscaling": {
      "enabled": true,
      "minReplicas": 2,
      "maxReplicas": 10,
      "targetCPU": 80
    }
  }
}
```

## Schema is Optional

ResourceDefinitions can work **without** a schema for maximum flexibility:

```json
{
  "kind": "CustomResourceDefinition",
  "spec": {
    "names": {
      "kind": "FlexibleResource"
    },
    "handler": {
      "executorType": "flexible-controller",
      "functionName": "reconcile"
    }
    // No schema - spec can contain anything!
  }
}
```

Then your CustomResource can have any structure:

```json
{
  "kind": "FlexibleResource",
  "spec": {
    "anything": "goes",
    "custom": {
      "nested": {
        "structure": true
      }
    },
    "values": [1, 2, 3]
  }
}
```

This is useful when:
- Schema is too complex to define upfront
- Different runtime backends need different fields
- You want maximum flexibility
- Validation happens in the controller code

## When to Use Schema

### ✅ Use Schema When:

1. **Well-defined structure** - You know exactly what fields are needed
2. **User-facing resources** - Help users avoid mistakes
3. **Multiple teams** - Enforce consistency across teams
4. **Production systems** - Prevent configuration errors
5. **Self-service** - Users create resources via UI/API

### ⚠️ Skip Schema When:

1. **Rapid prototyping** - Structure is still evolving
2. **Highly dynamic** - Runtime-specific configurations
3. **Expert users** - Controller does validation
4. **Custom DSL** - You have your own validation logic

## Validation Flow

```
User submits CustomResource
        ↓
[Optional: Client-side validation against schema]
        ↓
ColonyOS receives resource
        ↓
[Future: Server-side validation against ResourceDefinition schema]
        ↓
Convert to Process
        ↓
Controller receives Process
        ↓
[Controller can do additional validation]
        ↓
Reconcile
```

## Implementation Status

In the current implementation:

- ✅ **Schema structure defined** - ResourceDefinition can include schema
- ✅ **Types support all JSON Schema features** - Properties, enum, defaults, etc.
- ⚠️ **Validation not yet enforced** - Schema is stored but not automatically validated
- 📝 **Future work** - Add automatic validation before Process creation

To add validation, you would implement:

```go
func (cr *CustomResource) ValidateAgainstSchema(schema *ValidationSchema) error {
    // Check required fields
    for _, required := range schema.Required {
        if _, ok := cr.Spec[required]; !ok {
            return fmt.Errorf("required field '%s' is missing", required)
        }
    }

    // Check types and enums
    for key, prop := range schema.Properties {
        if val, ok := cr.Spec[key]; ok {
            if err := validateProperty(val, prop); err != nil {
                return fmt.Errorf("field '%s': %v", key, err)
            }
        }
    }

    return nil
}
```

## Best Practices

1. **Start simple** - Add schema incrementally
2. **Use descriptions** - Document each field
3. **Set sensible defaults** - Reduce configuration burden
4. **Use enums** - Constrain to valid values
5. **Make optional when possible** - Only require what's essential
6. **Nest logically** - Group related fields in objects
7. **Version your schemas** - Use different ResourceDefinition versions for breaking changes

## Examples in This Repo

- `executor-deployment-crd.json` - Basic schema with enums and required fields
- `mlmodel-crd.json` - Complex nested schema with arrays
- `database-crd.json` - Schema with defaults and validation
- `schema-examples.json` - Comprehensive schema features demonstration

## See Also

- [JSON Schema Documentation](https://json-schema.org/)
- [Kubernetes ResourceDefinition Schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema)
- [OpenAPI 3.0 Schema](https://spec.openapis.org/oas/v3.0.0#schema-object)
