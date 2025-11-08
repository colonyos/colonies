package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateResource(t *testing.T) {
	cr := CreateResource("ExecutorDeployment", "test-deploy", "test-colony")

	assert.Equal(t, "ExecutorDeployment", cr.Kind)
	assert.Equal(t, "test-deploy", cr.Metadata.Name)
	assert.Equal(t, "test-colony", cr.Metadata.Namespace)
	assert.NotEmpty(t, cr.ID)
	assert.Equal(t, int64(1), cr.Metadata.Generation)
	assert.NotNil(t, cr.Spec)
	assert.NotNil(t, cr.Status)
	assert.NotNil(t, cr.Metadata.Labels)
	assert.NotNil(t, cr.Metadata.Annotations)
}

func TestResourceSpecOperations(t *testing.T) {
	cr := CreateResource("TestResource", "test", "ns")

	// Test SetSpec
	cr.SetSpec("replicas", 3)
	cr.SetSpec("image", "test-image:latest")
	cr.SetSpec("config", map[string]interface{}{
		"cpu":    "2",
		"memory": "4Gi",
	})

	// Test GetSpec
	replicas, ok := cr.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, 3, replicas)

	image, ok := cr.GetSpec("image")
	assert.True(t, ok)
	assert.Equal(t, "test-image:latest", image)

	config, ok := cr.GetSpec("config")
	assert.True(t, ok)
	configMap := config.(map[string]interface{})
	assert.Equal(t, "2", configMap["cpu"])
	assert.Equal(t, "4Gi", configMap["memory"])

	// Test non-existent key
	_, ok = cr.GetSpec("nonexistent")
	assert.False(t, ok)

	// Check generation incremented
	assert.Equal(t, int64(4), cr.Metadata.Generation)
}

func TestResourceStatusOperations(t *testing.T) {
	cr := CreateResource("TestResource", "test", "ns")

	// Test SetStatus
	cr.SetStatus("phase", "Running")
	cr.SetStatus("ready", 3)
	cr.SetStatus("available", 3)

	// Test GetStatus
	phase, ok := cr.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)

	ready, ok := cr.GetStatus("ready")
	assert.True(t, ok)
	assert.Equal(t, 3, ready)

	// Test non-existent key
	_, ok = cr.GetStatus("nonexistent")
	assert.False(t, ok)

	// Status updates should not increment generation
	assert.Equal(t, int64(1), cr.Metadata.Generation)
}

func TestResourceLabelsAndAnnotations(t *testing.T) {
	cr := CreateResource("TestResource", "test", "ns")

	// Test labels
	cr.Metadata.Labels["app"] = "my-app"
	cr.Metadata.Labels["environment"] = "production"

	assert.Equal(t, "my-app", cr.Metadata.Labels["app"])
	assert.Equal(t, "production", cr.Metadata.Labels["environment"])

	// Test non-existent label
	_, ok := cr.Metadata.Labels["nonexistent"]
	assert.False(t, ok)

	// Test annotations
	cr.Metadata.Annotations["description"] = "Test service"
	cr.Metadata.Annotations["owner"] = "admin"

	assert.Equal(t, "Test service", cr.Metadata.Annotations["description"])
	assert.Equal(t, "admin", cr.Metadata.Annotations["owner"])

	// Test non-existent annotation
	_, ok = cr.Metadata.Annotations["nonexistent"]
	assert.False(t, ok)
}

func TestResourceValidation(t *testing.T) {
	// Valid service
	cr := CreateResource("TestResource", "test", "ns")
	err := cr.Validate()
	assert.NoError(t, err)

	// Missing Kind
	cr3 := CreateResource("", "test", "ns")
	err = cr3.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind is required")

	// Missing Name
	cr4 := CreateResource("TestResource", "", "ns")
	err = cr4.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata.name is required")

	// Missing Namespace
	cr5 := CreateResource("TestResource", "test", "")
	err = cr5.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata.namespace is required")
}

func TestResourceJSONConversion(t *testing.T) {
	cr := CreateResource("ExecutorDeployment", "test-deploy", "test-colony")
	cr.SetSpec("runtime", "kubernetes")
	cr.SetSpec("replicas", 3)
	cr.Metadata.Labels["app"] = "test"
	cr.SetStatus("phase", "Running")

	// Convert to JSON
	jsonStr, err := cr.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Convert back from JSON
	cr2, err := ConvertJSONToResource(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, cr.Kind, cr2.Kind)
	assert.Equal(t, cr.Metadata.Name, cr2.Metadata.Name)
	assert.Equal(t, cr.Metadata.Namespace, cr2.Metadata.Namespace)

	runtime, ok := cr2.GetSpec("runtime")
	assert.True(t, ok)
	assert.Equal(t, "kubernetes", runtime)

	replicas, ok := cr2.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(3), replicas) // JSON numbers are float64

	assert.Equal(t, "test", cr2.Metadata.Labels["app"])

	phase, ok := cr2.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)
}

func TestCreateResourceDefinition(t *testing.T) {
	crd := CreateResourceDefinition(
		"executordeployments.compute.colonies.io",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"service-controller",
		"reconcile_executor_deployment",
	)

	assert.Equal(t, "ResourceDefinition", crd.Kind)
	assert.Equal(t, "executordeployments.compute.colonies.io", crd.Metadata.Name)
	assert.Equal(t, "compute.colonies.io", crd.Spec.Group)
	assert.Equal(t, "v1", crd.Spec.Version)
	assert.Equal(t, "ExecutorDeployment", crd.Spec.Names.Kind)
	assert.Equal(t, "ExecutorDeploymentList", crd.Spec.Names.ListKind)
	assert.Equal(t, "executordeployment", crd.Spec.Names.Singular)
	assert.Equal(t, "executordeployments", crd.Spec.Names.Plural)
	assert.Equal(t, "Namespaced", crd.Spec.Scope)
	assert.Equal(t, "service-controller", crd.Spec.Handler.ExecutorType)
	assert.Equal(t, "reconcile_executor_deployment", crd.Spec.Handler.FunctionName)
}

func TestResourceDefinitionValidation(t *testing.T) {
	// Valid CRD
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)
	err := crd.Validate()
	assert.NoError(t, err)

	// Missing group
	crd2 := CreateResourceDefinition("", "", "v1", "TestResource", "testresources", "Namespaced", "test-controller", "reconcile")
	err = crd2.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.group is required")

	// Missing version
	crd3 := CreateResourceDefinition("", "test.io", "", "TestResource", "testresources", "Namespaced", "test-controller", "reconcile")
	err = crd3.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.version is required")

	// Invalid scope
	crd4 := CreateResourceDefinition("", "test.io", "v1", "TestResource", "testresources", "Invalid", "test-controller", "reconcile")
	err = crd4.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.scope must be")

	// Missing executor type
	crd5 := CreateResourceDefinition("", "test.io", "v1", "TestResource", "testresources", "Namespaced", "", "reconcile")
	err = crd5.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.handler.executorType is required")
}

func TestResourceDefinitionJSONConversion(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	// Convert to JSON
	jsonStr, err := crd.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Convert back from JSON
	crd2, err := ConvertJSONToResourceDefinition(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, crd.Metadata.Name, crd2.Metadata.Name)
	assert.Equal(t, crd.Spec.Group, crd2.Spec.Group)
	assert.Equal(t, crd.Spec.Version, crd2.Spec.Version)
	assert.Equal(t, crd.Spec.Names.Kind, crd2.Spec.Names.Kind)
	assert.Equal(t, crd.Spec.Handler.ExecutorType, crd2.Spec.Handler.ExecutorType)
}

func TestResourceDefinitionHelperMethods(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	assert.Equal(t, "test.io/v1", crd.GetAPIVersion())
}

func TestResourceInFunctionSpec(t *testing.T) {
	cr := CreateResource("TestResource", "test-service", "test-colony")
	cr.SetSpec("replicas", 3)
	cr.SetSpec("image", "test:latest")

	// Create a FunctionSpec with the Service attached
	funcSpec := CreateEmptyFunctionSpec()
	funcSpec.Service = cr

	// Verify the service is properly attached
	assert.NotNil(t, funcSpec.Service)
	assert.Equal(t, "TestResource", funcSpec.Service.Kind)
	assert.Equal(t, "test-service", funcSpec.Service.Metadata.Name)
	assert.Equal(t, "test-colony", funcSpec.Service.Metadata.Namespace)

	replicas, ok := funcSpec.Service.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, 3, replicas)
}

func TestComplexCustomResourceScenario(t *testing.T) {
	// Create a CRD for ExecutorDeployment
	crd := CreateResourceDefinition(
		"executordeployments.compute.colonies.io",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"service-controller",
		"reconcile_executor_deployment",
	)
	crd.Spec.Handler.ReconcileInterval = 30

	// Add schema validation
	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"runtime": {
				Type:        "string",
				Description: "Target runtime platform",
				Enum:        []interface{}{"kubernetes", "docker", "hpc"},
			},
			"replicas": {
				Type:        "integer",
				Description: "Number of executor instances",
				Default:     1,
			},
		},
		Required: []string{"runtime"},
	}

	err := crd.Validate()
	assert.NoError(t, err)

	// Create a custom service instance
	cr := CreateResource("ExecutorDeployment", "ml-executors", "ml-colony")
	cr.Metadata.Labels["app"] = "ml-training"
	cr.Metadata.Labels["environment"] = "production"
	cr.Metadata.Annotations["description"] = "GPU-enabled ML training executors"

	// Set complex spec
	cr.SetSpec("runtime", "kubernetes")
	cr.SetSpec("replicas", 5)
	cr.SetSpec("template", map[string]interface{}{
		"type": "gpu-executor",
		"capabilities": map[string]interface{}{
			"gpu": "nvidia-a100",
		},
	})
	cr.SetSpec("config", map[string]interface{}{
		"image": "ml-executor:latest",
		"services": map[string]interface{}{
			"gpu":    2,
			"memory": "64Gi",
		},
	})

	// Set status
	cr.SetStatus("phase", "Running")
	cr.SetStatus("ready", 5)
	cr.SetStatus("available", 5)
	cr.SetStatus("lastUpdateTime", time.Now().Format(time.RFC3339))

	// Validate
	err = cr.Validate()
	assert.NoError(t, err)

	// Convert to JSON and back
	jsonStr, err := cr.ToJSON()
	assert.NoError(t, err)

	cr2, err := ConvertJSONToResource(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, cr.Kind, cr2.Kind)

	// Verify all data made it through
	runtime, ok := cr2.GetSpec("runtime")
	assert.True(t, ok)
	assert.Equal(t, "kubernetes", runtime)

	replicas, ok := cr2.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, float64(5), replicas) // JSON numbers are float64

	phase, ok := cr2.GetStatus("phase")
	assert.True(t, ok)
	assert.Equal(t, "Running", phase)

	assert.Equal(t, "ml-training", cr2.Metadata.Labels["app"])
}

func TestUpdateGeneration(t *testing.T) {
	cr := CreateResource("TestResource", "test", "ns")

	initialGen := cr.Metadata.Generation
	assert.Equal(t, int64(1), initialGen)

	// Spec changes should increment generation
	cr.SetSpec("value1", "test1")
	assert.Equal(t, int64(2), cr.Metadata.Generation)

	cr.SetSpec("value2", "test2")
	assert.Equal(t, int64(3), cr.Metadata.Generation)

	// Status changes should NOT increment generation
	cr.SetStatus("ready", true)
	assert.Equal(t, int64(3), cr.Metadata.Generation)

	cr.SetStatus("phase", "Running")
	assert.Equal(t, int64(3), cr.Metadata.Generation)
}

func TestSchemaValidation(t *testing.T) {
	// Create CRD with schema
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"runtime": {
				Type:        "string",
				Description: "Runtime type",
				Enum:        []interface{}{"docker", "kubernetes", "local"},
			},
			"replicas": {
				Type:        "integer",
				Description: "Number of replicas",
			},
			"enabled": {
				Type:        "boolean",
				Description: "Is enabled",
			},
		},
		Required: []string{"runtime", "replicas"},
	}

	// Valid service
	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("runtime", "docker")
	cr.SetSpec("replicas", 3)
	cr.SetSpec("enabled", true)

	err := cr.ValidateAgainstRD(crd)
	assert.NoError(t, err)
}

func TestSchemaValidationMissingRequired(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"runtime": {
				Type: "string",
			},
		},
		Required: []string{"runtime"},
	}

	// Missing required field
	cr := CreateResource("TestResource", "test", "ns")
	// Not setting runtime

	err := cr.ValidateAgainstRD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field 'runtime' is missing")
}

func TestSchemaValidationInvalidType(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"replicas": {
				Type: "integer",
			},
		},
	}

	// Wrong type (string instead of integer)
	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("replicas", "not-a-number")

	err := cr.ValidateAgainstRD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}

func TestSchemaValidationInvalidEnum(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"runtime": {
				Type: "string",
				Enum: []interface{}{"docker", "kubernetes"},
			},
		},
		Required: []string{"runtime"},
	}

	// Invalid enum value
	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("runtime", "invalid-runtime")

	err := cr.ValidateAgainstRD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
	assert.Contains(t, err.Error(), "must be one of")
}

func TestSchemaValidationNestedObject(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"config": {
				Type: "object",
				Properties: map[string]SchemaProperty{
					"cpu": {
						Type: "string",
					},
					"memory": {
						Type: "string",
					},
				},
			},
		},
	}

	// Valid nested object
	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("config", map[string]interface{}{
		"cpu":    "2",
		"memory": "4Gi",
	})

	err := cr.ValidateAgainstRD(crd)
	assert.NoError(t, err)
}

func TestSchemaValidationArray(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	crd.Spec.Schema = &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"ports": {
				Type: "array",
				Items: &SchemaProperty{
					Type: "integer",
				},
			},
		},
	}

	// Valid array
	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("ports", []interface{}{80, 443, 8080})

	err := cr.ValidateAgainstRD(crd)
	assert.NoError(t, err)

	// Invalid array item type
	cr2 := CreateResource("TestResource", "test2", "ns")
	cr2.SetSpec("ports", []interface{}{80, "not-a-number", 8080})

	err = cr2.ValidateAgainstRD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}

func TestSchemaValidationKindMismatch(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	// Wrong kind
	cr := CreateResource("WrongResource", "test", "ns")

	err := cr.ValidateAgainstRD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind mismatch")
}

func TestSchemaValidationNoSchema(t *testing.T) {
	crd := CreateResourceDefinition(
		"testresources.test.io",
		"test.io",
		"v1",
		"TestResource",
		"testresources",
		"Namespaced",
		"test-controller",
		"reconcile",
	)
	// No schema defined

	cr := CreateResource("TestResource", "test", "ns")
	cr.SetSpec("anything", "goes")

	// Should pass validation when no schema is defined
	err := cr.ValidateAgainstRD(crd)
	assert.NoError(t, err)
}

// Reconciliation tests

func TestCreateReconciliationCreate(t *testing.T) {
	// Test create action (old is nil, new exists)
	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 3)

	reconciliation := CreateReconciliation(nil, newResource)

	assert.Nil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationCreate, reconciliation.Action)
	assert.Nil(t, reconciliation.Diff)
}

func TestCreateReconciliationDelete(t *testing.T) {
	// Test delete action (old exists, new is nil)
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)

	reconciliation := CreateReconciliation(oldResource, nil)

	assert.NotNil(t, reconciliation.Old)
	assert.Nil(t, reconciliation.New)
	assert.Equal(t, ReconciliationDelete, reconciliation.Action)
	assert.Nil(t, reconciliation.Diff)
}

func TestCreateReconciliationUpdate(t *testing.T) {
	// Test update action (both exist with changes)
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)
	oldResource.SetSpec("image", "nginx:1.0")

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 5)
	newResource.SetSpec("image", "nginx:2.0")

	reconciliation := CreateReconciliation(oldResource, newResource)

	assert.NotNil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationUpdate, reconciliation.Action)
	assert.NotNil(t, reconciliation.Diff)
	assert.True(t, reconciliation.Diff.HasChanges)
	assert.Equal(t, 2, len(reconciliation.Diff.SpecChanges))
}

func TestCreateReconciliationNoop(t *testing.T) {
	// Test noop action (both exist with no changes)
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)
	oldResource.SetSpec("image", "nginx:1.0")

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 3)
	newResource.SetSpec("image", "nginx:1.0")

	reconciliation := CreateReconciliation(oldResource, newResource)

	assert.NotNil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationNoop, reconciliation.Action)
	assert.NotNil(t, reconciliation.Diff)
	assert.False(t, reconciliation.Diff.HasChanges)
}

func TestResourceDiffSpecChanges(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)
	oldResource.SetSpec("image", "nginx:1.0")
	oldResource.SetSpec("port", 8080)

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 5)           // Modified
	newResource.SetSpec("image", "nginx:1.0")    // Unchanged
	newResource.SetSpec("command", []string{"run"}) // Added
	// port removed

	diff := oldResource.Diff(newResource)

	assert.True(t, diff.HasChanges)
	assert.Equal(t, 3, len(diff.SpecChanges))

	// Check for modified field
	replicasChange := diff.GetFieldChange("spec.replicas")
	assert.NotNil(t, replicasChange)
	assert.Equal(t, ChangeModified, replicasChange.Type)
	assert.Equal(t, 3, replicasChange.OldValue)
	assert.Equal(t, 5, replicasChange.NewValue)

	// Check for added field
	commandChange := diff.GetFieldChange("spec.command")
	assert.NotNil(t, commandChange)
	assert.Equal(t, ChangeAdded, commandChange.Type)
	assert.Nil(t, commandChange.OldValue)

	// Check for removed field
	portChange := diff.GetFieldChange("spec.port")
	assert.NotNil(t, portChange)
	assert.Equal(t, ChangeRemoved, portChange.Type)
	assert.Equal(t, 8080, portChange.OldValue)
	assert.Nil(t, portChange.NewValue)
}

func TestResourceDiffStatusChanges(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetStatus("phase", "Pending")
	oldResource.SetStatus("ready", 0)

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetStatus("phase", "Running")
	newResource.SetStatus("ready", 3)

	diff := oldResource.Diff(newResource)

	assert.True(t, diff.HasChanges)
	assert.Equal(t, 2, len(diff.StatusChanges))
	assert.Equal(t, 0, len(diff.SpecChanges))

	// Check phase change
	phaseChange := diff.GetFieldChange("status.phase")
	assert.NotNil(t, phaseChange)
	assert.Equal(t, ChangeModified, phaseChange.Type)
	assert.Equal(t, "Pending", phaseChange.OldValue)
	assert.Equal(t, "Running", phaseChange.NewValue)
}

func TestResourceDiffMetadataChanges(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.Metadata.Labels["app"] = "old-app"
	oldResource.Metadata.Labels["version"] = "1.0"
	oldResource.Metadata.Annotations["description"] = "old description"

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.Metadata.Labels["app"] = "new-app"
	newResource.Metadata.Labels["environment"] = "production"
	newResource.Metadata.Annotations["description"] = "new description"

	diff := oldResource.Diff(newResource)

	assert.True(t, diff.HasChanges)
	assert.Greater(t, len(diff.MetadataChanges), 0)

	// Check modified label
	assert.True(t, diff.HasFieldChange("metadata.labels.app"))

	// Check added label
	assert.True(t, diff.HasFieldChange("metadata.labels.environment"))

	// Check removed label
	assert.True(t, diff.HasFieldChange("metadata.labels.version"))

	// Check modified annotation
	assert.True(t, diff.HasFieldChange("metadata.annotations.description"))
}

func TestResourceDiffHelperMethods(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)
	oldResource.SetStatus("phase", "Running")
	oldResource.Metadata.Labels["app"] = "test"

	// Test OnlyStatusChanged
	newResource1 := CreateResource("TestResource", "test", "ns")
	newResource1.SetSpec("replicas", 3)
	newResource1.SetStatus("phase", "Succeeded")
	newResource1.Metadata.Labels["app"] = "test"

	diff1 := oldResource.Diff(newResource1)
	assert.True(t, diff1.OnlyStatusChanged())
	assert.False(t, diff1.OnlyMetadataChanged())

	// Test OnlyMetadataChanged
	newResource2 := CreateResource("TestResource", "test", "ns")
	newResource2.SetSpec("replicas", 3)
	newResource2.SetStatus("phase", "Running")
	newResource2.Metadata.Labels["app"] = "new-app"

	diff2 := oldResource.Diff(newResource2)
	assert.True(t, diff2.OnlyMetadataChanged())
	assert.False(t, diff2.OnlyStatusChanged())

	// Test mixed changes
	newResource3 := CreateResource("TestResource", "test", "ns")
	newResource3.SetSpec("replicas", 5)
	newResource3.SetStatus("phase", "Failed")
	newResource3.Metadata.Labels["app"] = "new-app"

	diff3 := oldResource.Diff(newResource3)
	assert.False(t, diff3.OnlyStatusChanged())
	assert.False(t, diff3.OnlyMetadataChanged())
}

func TestResourceDiffComplexChanges(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("config", map[string]interface{}{
		"cpu":    "2",
		"memory": "4Gi",
	})

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("config", map[string]interface{}{
		"cpu":    "4",
		"memory": "4Gi",
	})

	diff := oldResource.Diff(newResource)

	assert.True(t, diff.HasChanges)
	assert.Equal(t, 1, len(diff.SpecChanges))

	// The entire config object changed
	configChange := diff.GetFieldChange("spec.config")
	assert.NotNil(t, configChange)
	assert.Equal(t, ChangeModified, configChange.Type)
}

func TestReconciliationInFunctionSpec(t *testing.T) {
	// Create a reconciliation
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 5)

	reconciliation := CreateReconciliation(oldResource, newResource)

	// Create a FunctionSpec with reconciliation
	funcSpec := CreateEmptyFunctionSpec()
	funcSpec.FuncName = "reconcile"
	funcSpec.Reconciliation = reconciliation

	assert.NotNil(t, funcSpec.Reconciliation)
	assert.Equal(t, ReconciliationUpdate, funcSpec.Reconciliation.Action)
	assert.True(t, funcSpec.Reconciliation.Diff.HasChanges)

	// Test JSON conversion
	jsonStr, err := funcSpec.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "reconciliation")
	assert.Contains(t, jsonStr, "update")
}

func TestReconciliationJSONConversion(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 5)

	reconciliation := CreateReconciliation(oldResource, newResource)

	// Create FunctionSpec with reconciliation
	funcSpec := CreateEmptyFunctionSpec()
	funcSpec.FuncName = "reconcile"
	funcSpec.Reconciliation = reconciliation

	// Convert to JSON
	jsonStr, err := funcSpec.ToJSON()
	assert.NoError(t, err)

	// Convert back from JSON
	funcSpec2, err := ConvertJSONToFunctionSpec(jsonStr)
	assert.NoError(t, err)

	assert.NotNil(t, funcSpec2.Reconciliation)
	assert.Equal(t, ReconciliationUpdate, funcSpec2.Reconciliation.Action)
	assert.True(t, funcSpec2.Reconciliation.Diff.HasChanges)
}

func TestHasFieldChange(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("replicas", 3)
	oldResource.SetStatus("phase", "Running")

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("replicas", 5)
	newResource.SetStatus("phase", "Running")

	diff := oldResource.Diff(newResource)

	assert.True(t, diff.HasFieldChange("spec.replicas"))
	assert.False(t, diff.HasFieldChange("spec.image"))
	assert.False(t, diff.HasFieldChange("status.phase"))
}

func TestGetFieldChange(t *testing.T) {
	oldResource := CreateResource("TestResource", "test", "ns")
	oldResource.SetSpec("image", "nginx:1.0")

	newResource := CreateResource("TestResource", "test", "ns")
	newResource.SetSpec("image", "nginx:2.0")

	diff := oldResource.Diff(newResource)

	change := diff.GetFieldChange("spec.image")
	assert.NotNil(t, change)
	assert.Equal(t, "spec.image", change.Path)
	assert.Equal(t, "nginx:1.0", change.OldValue)
	assert.Equal(t, "nginx:2.0", change.NewValue)
	assert.Equal(t, ChangeModified, change.Type)

	// Non-existent field
	change2 := diff.GetFieldChange("spec.nonexistent")
	assert.Nil(t, change2)
}

