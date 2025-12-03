# BlueprintDefinition Schema Validation Guide

This guide explains how schema validation works in ColonyOS services.

## Overview

BlueprintDefinitions can include an **optional** JSON Schema that validates blueprint instances on the server side. This provides:

- Type safety
- Required field validation
- Enum constraints
- Default values (documentation only - not auto-applied)
- Documentation
- Nested object validation
- Array item validation

## How Validation Works

### Server-Side Enforcement

**All blueprint validation happens on the server**, not in the CLI. When you create or update a service:

1. Client sends blueprint JSON to server
2. Server looks up the BlueprintDefinition for the service's `kind`
3. If a schema is defined, server validates the blueprint spec against it
4. If validation fails, server returns HTTP 400 Bad Request
5. If validation passes, server saves the service

This architecture ensures:
- Data integrity regardless of which client is used
- Consistent validation across all API consumers
- No way to bypass validation

### Validation Flow

```
User submits Service
        ↓
CLI sends JSON to server
        ↓
Server receives service
        ↓
Server finds BlueprintDefinition by kind
        ↓
[Server validates against schema]
        ↓
Validation passes → Save blueprint → Trigger reconciliation
        ↓
Validation fails → Return 400 error with details
```

## Schema Structure

```json
{
  "schema": {
    "type": "object",
    "properties": {
      "fieldName": {
        "type": "string|number|boolean|object|array",
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

**Schema:**
```json
{
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Blueprint name"
      },
      "replicas": {
        "type": "number",
        "description": "Number of replicas"
      },
      "enabled": {
        "type": "boolean",
        "description": "Whether blueprint is enabled"
      }
    },
    "required": ["name", "replicas"]
  }
}
```

**Valid Service:**
```json
{
  "spec": {
    "name": "my-service",
    "replicas": 3,
    "enabled": true
  }
}
```

**Invalid - Wrong Type:**
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: field 'replicas' must be a number, got string
```

```json
{
  "spec": {
    "name": "my-service",
    "replicas": "three",  ← ✗ Should be number
    "enabled": true
  }
}
```

**Invalid - Missing Required:**
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: required field 'replicas' is missing
```

```json
{
  "spec": {
    "name": "my-service"
    // ✗ Missing required field "replicas"
  }
}
```

### 2. Enum Constraints

**Schema:**
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
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: field 'size' must be one of [small medium large], got extra-large
```

```json
{
  "spec": {
    "size": "extra-large",  ← ✗ Not in enum
    "environment": "prod"    ← ✗ Not in enum (must be "production")
  }
}
```

### 3. Nested Objects

**Schema:**
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
                "type": "number",
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

**Invalid - Missing Nested Required:**
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: required field 'version' is missing
```

**Invalid - Wrong Enum in Nested Object:**
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: field 'engine' must be one of [postgresql mysql], got mongodb
```

### 4. Arrays

**Schema:**
```json
{
  "schema": {
    "properties": {
      "ports": {
        "type": "array",
        "description": "List of exposed ports",
        "items": {
          "type": "number"
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

**Invalid - Wrong Array Item Type:**
```bash
$ colonies blueprint add --spec invalid.json

Error: blueprint validation failed: field 'ports[1]' must be a number, got string
```

```json
{
  "spec": {
    "ports": [8080, "8443", 9090]  ← ✗ Array item must be number
  }
}
```

### 5. Complex Real-World Example

**BlueprintDefinition:**
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
        "type": "number",
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
                "type": "number",
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
            "type": "number",
            "default": 1
          },
          "maxReplicas": {
            "type": "number",
            "default": 10
          },
          "targetCPU": {
            "type": "number",
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

**Valid Service:**
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

BlueprintDefinitions can work **without** a schema for maximum flexibility:

```json
{
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

Then your Blueprint can have any structure:

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
- Validation happens in the reconciler code

## When to Use Schema

### Use Schema When:

1. **Well-defined structure** - You know exactly what fields are needed
2. **User-facing services** - Help users avoid mistakes early
3. **Multiple teams** - Enforce consistency across teams
4. **Production systems** - Prevent configuration errors before deployment
5. **Self-service** - Users create blueprints via UI/API
6. **Clear errors** - Get immediate feedback on what's wrong

### Skip Schema When:

1. **Rapid prototyping** - Structure is still evolving
2. **Highly dynamic** - Runtime-specific configurations vary widely
3. **Expert users only** - Reconciler does comprehensive validation
4. **Custom DSL** - You have your own validation logic
5. **Pass-through** - Spec is passed directly to external system

## Validation Implementation

The validation is implemented in the server at:
- `pkg/server/handlers/service/handlers.go:HandleAddResource()` (lines 401-407)
- `pkg/core/service.go:ValidateResourceAgainstSchema()` (line 678+)

The validator checks:

```go
func ValidateResourceAgainstSchema(blueprint *Service, schema *ValidationSchema) error {
    // Check required fields
    for _, requiredField := range schema.Required {
        if _, ok := service.Spec[requiredField]; !ok {
            return fmt.Errorf("required field '%s' is missing", requiredField)
        }
    }

    // Validate each field in the spec
    for fieldName, fieldValue := range service.Spec {
        if schemaProp, ok := schema.Properties[fieldName]; ok {
            if err := validateField(fieldName, fieldValue, &schemaProp); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### Validated Properties

- **Type matching** - string, number, boolean, object, array
- **Required fields** - All fields in `required` array must be present
- **Enum values** - Value must be in the enum list
- **Nested objects** - Recursive validation of object properties
- **Array items** - Each array element validated against item schema

### Not Currently Validated

- **String patterns/regex** - Not yet implemented
- **Number ranges** - min/max not yet implemented
- **String length** - minLength/maxLength not yet implemented
- **Array length** - minItems/maxItems not yet implemented
- **Unique items** - uniqueItems not yet implemented
- **Default application** - Defaults are documentation only, not auto-applied

These can be added in future enhancements if needed.

## Best Practices

1. **Start simple** - Add schema incrementally as requirements become clear
2. **Use descriptions** - Document each field to help users
3. **Set sensible defaults** - Reduce configuration burden (documentation)
4. **Use enums** - Constrain to valid values early to prevent errors
5. **Make optional when possible** - Only require what's truly essential
6. **Nest logically** - Group related fields in objects
7. **Test validation** - Verify schema catches invalid configurations
8. **Version your schemas** - Use different BlueprintDefinition names for breaking changes

## Testing Your Schema

Create invalid test cases to verify your schema catches errors:

```bash
# Test missing required field
cat > test-missing-required.json <<EOF
{
  "kind": "ExecutorDeployment",
  "metadata": { "name": "test" },
  "spec": {
    "replicas": 1
    # Missing required "image" field
  }
}
EOF

colonies blueprint add --spec test-missing-required.json
# Expected: Error: blueprint validation failed: required field 'image' is missing

# Test wrong type
cat > test-wrong-type.json <<EOF
{
  "kind": "ExecutorDeployment",
  "metadata": { "name": "test" },
  "spec": {
    "image": "nginx:latest",
    "replicas": "three"  # Should be number
  }
}
EOF

colonies blueprint add --spec test-wrong-type.json
# Expected: Error: blueprint validation failed: field 'replicas' must be a number, got string

# Test invalid enum
cat > test-invalid-enum.json <<EOF
{
  "kind": "MyResource",
  "metadata": { "name": "test" },
  "spec": {
    "size": "extra-large"  # Not in enum [small, medium, large]
  }
}
EOF

colonies blueprint add --spec test-invalid-enum.json
# Expected: Error: blueprint validation failed: field 'size' must be one of [small medium large], got extra-large
```

## See Also

- [JSON Schema Documentation](https://json-schema.org/)
- [Kubernetes CRD Schema](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#specifying-a-structural-schema)
- [OpenAPI 3.0 Schema](https://spec.openapis.org/oas/v3.0.0#schema-object)
- [Blueprint Management Guide](Services.md)
