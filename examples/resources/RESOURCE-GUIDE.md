# Custom Resource Definition (CRD) Guide

This guide explains the relationship between ResourceDefinitions and Resources, and how to use them in ColonyOS.

## 📚 Core Concepts

### ResourceDefinition (CRD)
A ResourceDefinition is a **blueprint** that defines a NEW resource type. Think of it like a class definition in programming.

**Key Points:**
- Created **first** (before any instances)
- Registered with the ColonyOS server
- Defines the schema, handler, and behavior for a resource type
- Analogous to Kubernetes CRDs

### Resource (CR)
A Resource is an **instance** of a resource type defined by a ResourceDefinition. Think of it like an object instance of a class.

**Key Points:**
- Created **second** (after ResourceDefinition is registered)
- References the ResourceDefinition implicitly via `apiVersion` + `kind`
- Must conform to the ResourceDefinition's schema (if one is defined)
- Server validates it against the ResourceDefinition before processing

## 🔄 The Complete Workflow

### Step 1: Define and Register the ResourceDefinition

```go
// Create the ResourceDefinition (defines "ExecutorDeployment" as a new resource type)
crd := core.CreateResourceDefinition(
    "executordeployments.compute.colonies.io",  // metadata.name
    "compute.colonies.io",                       // group
    "v1",                                        // version
    "ExecutorDeployment",                        // kind
    "executordeployments",                       // plural
    "Namespaced",                                // scope
    "executor-deployment-controller",            // handler executor type
    "reconcile_executor_deployment",             // handler function
)

// Add schema validation (optional but recommended)
crd.Spec.Schema = &core.ValidationSchema{
    Type: "object",
    Properties: map[string]core.SchemaProperty{
        "runtime": {
            Type: "string",
            Enum: []interface{}{"docker", "kubernetes", "local"},
        },
        "replicas": {
            Type: "integer",
        },
    },
    Required: []string{"runtime", "replicas"},
}

// Validate the ResourceDefinition
if err := crd.Validate(); err != nil {
    log.Fatal(err)
}

// Register with server (future implementation)
// server.RegisterCRD(crd)
```

### Step 2: Create Resource Instances

```go
// Create an instance of the ExecutorDeployment resource type
cr := core.CreateResource(
    "compute.colonies.io/v1",      // apiVersion (matches CRD's group/version)
    "ExecutorDeployment",           // kind (matches CRD's kind)
    "ml-training-executors",        // name of THIS instance
    "my-colony",                    // namespace
)

// Set the spec (this is what the controller will reconcile)
cr.SetSpec("runtime", "docker")
cr.SetSpec("replicas", 3)
cr.SetSpec("image", "ml-executor:latest")

// Basic validation (checks required fields exist)
if err := cr.Validate(); err != nil {
    log.Fatal(err)
}

// Schema validation (validates against ResourceDefinition schema)
if err := cr.ValidateAgainstRD(crd); err != nil {
    log.Fatal(err)  // Will fail if spec doesn't match schema
}
```

### Step 3: Server Processing

When you submit a Resource to the server:

```go
// Server receives the Resource
receivedCR := customResourceFromClient

// 1. Look up the ResourceDefinition by apiVersion + kind
crd := server.FindCRD(receivedCR.APIVersion, receivedCR.Kind)
if crd == nil {
    return errors.New("CRD not found")
}

// 2. Validate the Resource against the ResourceDefinition
if err := receivedCR.ValidateAgainstRD(crd); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// 3. Create FunctionSpec and attach Resource
funcSpec := core.CreateEmptyFunctionSpec()
funcSpec.FuncName = crd.Spec.Handler.FunctionName
funcSpec.Conditions.ExecutorType = crd.Spec.Handler.ExecutorType
funcSpec.Conditions.ColonyName = receivedCR.Metadata.Namespace
funcSpec.Label = receivedCR.Kind + "/" + receivedCR.Metadata.Name
funcSpec.Resource = receivedCR

// 4. Create Process from FunctionSpec
process := core.CreateProcessFromFunctionSpec(funcSpec)

// 5. Submit to queue for controller to handle
server.SubmitProcess(process)
```

### Step 4: Controller Reconciliation

The controller receives and reconciles the resource:

```go
// Controller polls for work
process := client.Assign(crd.Spec.Handler.ExecutorType)

// Extract Resource from process kwargs
cr := extractResourceFromProcess(process)

// Reconcile: compare desired state (cr.Spec) with current state
currentState := getCurrentState()
desiredState := cr.Spec

if currentState != desiredState {
    // Take action to make current == desired
    deploy() // or scale(), remove(), etc.
}

// Update status
cr.SetStatus("phase", "Running")
cr.SetStatus("ready", currentReplicas)

// Complete the process
client.Close(process.ID)
```

## 🔗 How ResourceDefinitions and Resources are Linked

Resources don't have a direct Go reference to their CRD. Instead, they're linked by **matching fields**:

| Resource Field | ResourceDefinition Field | Example |
|---------------------|-----------|---------|
| `apiVersion` | `spec.group/spec.version` | `compute.colonies.io/v1` |
| `kind` | `spec.names.kind` | `ExecutorDeployment` |

The server performs the lookup:

```go
func (s *Server) FindCRD(apiVersion, kind string) *ResourceDefinition {
    for _, crd := range s.registeredCRDs {
        if crd.GetAPIVersion() == apiVersion &&
           crd.Spec.Names.Kind == kind {
            return crd
        }
    }
    return nil
}
```

## ✅ Schema Validation

### What Gets Validated?

When you call `cr.ValidateAgainstRD(crd)`, it checks:

1. **API Version Match**: `cr.APIVersion == crd.GetAPIVersion()`
2. **Kind Match**: `cr.Kind == crd.Spec.Names.Kind`
3. **Schema Compliance** (if ResourceDefinition has a schema):
   - Required fields present
   - Correct types (string, integer, boolean, object, array)
   - Enum values valid
   - Nested objects valid
   - Array items valid

### Validation Example

```go
// ResourceDefinition with schema
crd.Spec.Schema = &ValidationSchema{
    Properties: map[string]SchemaProperty{
        "runtime": {
            Type: "string",
            Enum: []interface{}{"docker", "kubernetes"},
        },
        "replicas": {
            Type: "integer",
        },
    },
    Required: []string{"runtime"},
}

// VALID Resource
cr := CreateResource("compute.io/v1", "ExecutorDeployment", "test", "ns")
cr.SetSpec("runtime", "docker")  // ✓ string, in enum
cr.SetSpec("replicas", 3)        // ✓ integer
cr.ValidateAgainstRD(crd)       // ✓ PASS

// INVALID Resource - wrong type
cr2 := CreateResource("compute.io/v1", "ExecutorDeployment", "test", "ns")
cr2.SetSpec("runtime", "docker")
cr2.SetSpec("replicas", "three")  // ✗ string instead of integer
cr2.ValidateAgainstRD(crd)       // ✗ FAIL: "must be an integer"

// INVALID Resource - missing required field
cr3 := CreateResource("compute.io/v1", "ExecutorDeployment", "test", "ns")
cr3.SetSpec("replicas", 3)        // missing "runtime"
cr3.ValidateAgainstRD(crd)       // ✗ FAIL: "required field 'runtime' is missing"

// INVALID Resource - bad enum value
cr4 := CreateResource("compute.io/v1", "ExecutorDeployment", "test", "ns")
cr4.SetSpec("runtime", "aws")     // ✗ not in enum
cr4.SetSpec("replicas", 3)
cr4.ValidateAgainstRD(crd)       // ✗ FAIL: "must be one of [docker kubernetes]"
```

## 🎯 Supported Schema Types

### String
```go
"name": {
    Type: "string",
}
```

### Integer
```go
"replicas": {
    Type: "integer",
}
```

### Number (float)
```go
"cpu": {
    Type: "number",
}
```

### Boolean
```go
"enabled": {
    Type: "boolean",
}
```

### Enum
```go
"runtime": {
    Type: "string",
    Enum: []interface{}{"docker", "kubernetes", "local"},
}
```

### Nested Object
```go
"config": {
    Type: "object",
    Properties: map[string]SchemaProperty{
        "cpu": {Type: "string"},
        "memory": {Type: "string"},
    },
}
```

### Array
```go
"ports": {
    Type: "array",
    Items: &SchemaProperty{
        Type: "integer",
    },
}
```

### Array of Objects
```go
"volumes": {
    Type: "array",
    Items: &SchemaProperty{
        Type: "object",
        Properties: map[string]SchemaProperty{
            "source": {Type: "string"},
            "target": {Type: "string"},
        },
    },
}
```

## 🔧 Helper Function for Safer Creation

You can create a helper to make the relationship more explicit:

```go
// Helper function that auto-fills apiVersion and kind from CRD
func CreateResourceFromCRD(
    crd *ResourceDefinition,
    name, namespace string,
) *Resource {
    cr := CreateResource(
        crd.GetAPIVersion(),  // Auto-filled from CRD
        crd.Spec.Names.Kind,  // Auto-filled from CRD
        name,
        namespace,
    )
    return cr
}

// Usage
cr := CreateResourceFromCRD(crd, "ml-training-executors", "my-colony")
cr.SetSpec("runtime", "docker")
cr.SetSpec("replicas", 3)

// Validation will definitely match since we used the ResourceDefinition to create it
err := cr.ValidateAgainstRD(crd)  // ✓ apiVersion and kind guaranteed to match
```

## 📋 Complete Example

See `complete-example/` directory for a full working example including:

- ResourceDefinition definition with schema
- Resource instances
- Controller implementation
- Schema validation in action

Run the demo:

```bash
cd examples/resources/complete-example
go run . -mode demo
```

## 🚀 Best Practices

1. **Always define schemas** in your ResourceDefinitions for validation
2. **Validate early** - call `ValidateAgainstRD()` before submitting
3. **Use required fields** to enforce mandatory configuration
4. **Use enums** to restrict values to known-good options
5. **Document schemas** in the ResourceDefinition description fields
6. **Version your CRDs** (v1alpha1, v1beta1, v1) as they evolve

## 🔍 Debugging Validation Errors

If validation fails, the error message will tell you exactly what's wrong:

```
spec validation failed: required field 'runtime' is missing
spec validation failed: field 'replicas' must be an integer
spec validation failed: field 'runtime' has invalid value 'aws', must be one of [docker kubernetes]
apiVersion mismatch: resource has 'wrong.io/v1' but ResourceDefinition defines 'compute.io/v1'
kind mismatch: resource has 'Wrong' but ResourceDefinition defines 'ExecutorDeployment'
```

## 📖 See Also

- [SCHEMA-GUIDE.md](./SCHEMA-GUIDE.md) - Detailed schema examples
- [complete-example/README.md](./complete-example/README.md) - Full working example
- [complete-example/USAGE.md](./complete-example/USAGE.md) - Usage guide
