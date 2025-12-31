package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateBlueprint(t *testing.T) {
	cr := CreateBlueprint("ExecutorDeployment", "test-deploy", "test-colony")

	assert.Equal(t, "ExecutorDeployment", cr.Kind)
	assert.Equal(t, "test-deploy", cr.Metadata.Name)
	assert.Equal(t, "test-colony", cr.Metadata.ColonyName)
	assert.NotEmpty(t, cr.ID)
	assert.Equal(t, int64(1), cr.Metadata.Generation)
	assert.NotNil(t, cr.Spec)
	assert.NotNil(t, cr.Status)
	assert.NotNil(t, cr.Metadata.Labels)
	assert.NotNil(t, cr.Metadata.Annotations)
}

func TestBlueprintSpecOperations(t *testing.T) {
	cr := CreateBlueprint("TestBlueprint", "test", "ns")

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

func TestBlueprintStatusOperations(t *testing.T) {
	cr := CreateBlueprint("TestBlueprint", "test", "ns")

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

func TestBlueprintLabelsAndAnnotations(t *testing.T) {
	cr := CreateBlueprint("TestBlueprint", "test", "ns")

	// Test labels
	cr.Metadata.Labels["app"] = "my-app"
	cr.Metadata.Labels["environment"] = "production"

	assert.Equal(t, "my-app", cr.Metadata.Labels["app"])
	assert.Equal(t, "production", cr.Metadata.Labels["environment"])

	// Test non-existent label
	_, ok := cr.Metadata.Labels["nonexistent"]
	assert.False(t, ok)

	// Test annotations
	cr.Metadata.Annotations["description"] = "Test blueprint"
	cr.Metadata.Annotations["owner"] = "admin"

	assert.Equal(t, "Test blueprint", cr.Metadata.Annotations["description"])
	assert.Equal(t, "admin", cr.Metadata.Annotations["owner"])

	// Test non-existent annotation
	_, ok = cr.Metadata.Annotations["nonexistent"]
	assert.False(t, ok)
}

func TestBlueprintValidation(t *testing.T) {
	// Valid blueprint
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	err := cr.Validate()
	assert.NoError(t, err)

	// Missing Kind
	cr3 := CreateBlueprint("", "test", "ns")
	err = cr3.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind is required")

	// Missing Name
	cr4 := CreateBlueprint("TestBlueprint", "", "ns")
	err = cr4.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata.name is required")

	// Missing Namespace
	cr5 := CreateBlueprint("TestBlueprint", "test", "")
	err = cr5.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metadata.namespace is required")
}

func TestBlueprintJSONConversion(t *testing.T) {
	cr := CreateBlueprint("ExecutorDeployment", "test-deploy", "test-colony")
	cr.SetSpec("runtime", "kubernetes")
	cr.SetSpec("replicas", 3)
	cr.Metadata.Labels["app"] = "test"
	cr.SetStatus("phase", "Running")

	// Convert to JSON
	jsonStr, err := cr.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Convert back from JSON
	cr2, err := ConvertJSONToBlueprint(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, cr.Kind, cr2.Kind)
	assert.Equal(t, cr.Metadata.Name, cr2.Metadata.Name)
	assert.Equal(t, cr.Metadata.ColonyName, cr2.Metadata.ColonyName)

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

func TestCreateBlueprintDefinition(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"executor-deployment",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"blueprint-controller",
		"reconcile_executor_deployment",
	)

	assert.Equal(t, "BlueprintDefinition", crd.Kind)
	assert.Equal(t, "executor-deployment", crd.Metadata.Name)
	assert.Equal(t, "compute.colonies.io", crd.Spec.Group)
	assert.Equal(t, "v1", crd.Spec.Version)
	assert.Equal(t, "ExecutorDeployment", crd.Spec.Names.Kind)
	assert.Equal(t, "ExecutorDeploymentList", crd.Spec.Names.ListKind)
	assert.Equal(t, "executordeployment", crd.Spec.Names.Singular)
	assert.Equal(t, "executordeployments", crd.Spec.Names.Plural)
	assert.Equal(t, "Namespaced", crd.Spec.Scope)
	assert.Equal(t, "blueprint-controller", crd.Spec.Handler.ExecutorType)
	assert.Equal(t, "reconcile_executor_deployment", crd.Spec.Handler.FunctionName)
}

func TestBlueprintDefinitionValidation(t *testing.T) {
	// Valid CRD
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test-controller",
		"reconcile",
	)
	err := crd.Validate()
	assert.NoError(t, err)

	// Missing group
	crd2 := CreateBlueprintDefinition("", "", "v1", "TestBlueprint", "testblueprints", "Namespaced", "test-controller", "reconcile")
	err = crd2.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.group is required")

	// Missing version
	crd3 := CreateBlueprintDefinition("", "test.io", "", "TestBlueprint", "testblueprints", "Namespaced", "test-controller", "reconcile")
	err = crd3.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.version is required")

	// Invalid scope
	crd4 := CreateBlueprintDefinition("", "test.io", "v1", "TestBlueprint", "testblueprints", "Invalid", "test-controller", "reconcile")
	err = crd4.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.scope must be")

	// Missing executor type
	crd5 := CreateBlueprintDefinition("", "test.io", "v1", "TestBlueprint", "testblueprints", "Namespaced", "", "reconcile")
	err = crd5.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "spec.handler.executorType is required")
}

func TestBlueprintDefinitionJSONConversion(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	// Convert to JSON
	jsonStr, err := crd.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Convert back from JSON
	crd2, err := ConvertJSONToBlueprintDefinition(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, crd.Metadata.Name, crd2.Metadata.Name)
	assert.Equal(t, crd.Spec.Group, crd2.Spec.Group)
	assert.Equal(t, crd.Spec.Version, crd2.Spec.Version)
	assert.Equal(t, crd.Spec.Names.Kind, crd2.Spec.Names.Kind)
	assert.Equal(t, crd.Spec.Handler.ExecutorType, crd2.Spec.Handler.ExecutorType)
}

func TestBlueprintDefinitionHelperMethods(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	assert.Equal(t, "test.io/v1", crd.GetAPIVersion())
}

func TestBlueprintInFunctionSpec(t *testing.T) {
	cr := CreateBlueprint("TestBlueprint", "test-blueprint", "test-colony")
	cr.SetSpec("replicas", 3)
	cr.SetSpec("image", "test:latest")

	// Create a FunctionSpec with the Blueprint attached
	funcSpec := CreateEmptyFunctionSpec()
	funcSpec.Blueprint = cr

	// Verify the blueprint is properly attached
	assert.NotNil(t, funcSpec.Blueprint)
	assert.Equal(t, "TestBlueprint", funcSpec.Blueprint.Kind)
	assert.Equal(t, "test-blueprint", funcSpec.Blueprint.Metadata.Name)
	assert.Equal(t, "test-colony", funcSpec.Blueprint.Metadata.ColonyName)

	replicas, ok := funcSpec.Blueprint.GetSpec("replicas")
	assert.True(t, ok)
	assert.Equal(t, 3, replicas)
}

func TestComplexBlueprintScenario(t *testing.T) {
	// Create a CRD for ExecutorDeployment
	crd := CreateBlueprintDefinition(
		"executor-deployment",
		"compute.colonies.io",
		"v1",
		"ExecutorDeployment",
		"executordeployments",
		"Namespaced",
		"blueprint-controller",
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

	// Create a custom blueprint instance
	cr := CreateBlueprint("ExecutorDeployment", "ml-executors", "ml-colony")
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

	cr2, err := ConvertJSONToBlueprint(jsonStr)
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")

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
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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

	// Valid blueprint
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("runtime", "docker")
	cr.SetSpec("replicas", 3)
	cr.SetSpec("enabled", true)

	err := cr.ValidateAgainstSD(crd)
	assert.NoError(t, err)
}

func TestSchemaValidationMissingRequired(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	// Not setting runtime

	err := cr.ValidateAgainstSD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required field 'runtime' is missing")
}

func TestSchemaValidationInvalidType(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("replicas", "not-a-number")

	err := cr.ValidateAgainstSD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}

func TestSchemaValidationInvalidEnum(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("runtime", "invalid-runtime")

	err := cr.ValidateAgainstSD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
	assert.Contains(t, err.Error(), "must be one of")
}

func TestSchemaValidationNestedObject(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("config", map[string]interface{}{
		"cpu":    "2",
		"memory": "4Gi",
	})

	err := cr.ValidateAgainstSD(crd)
	assert.NoError(t, err)
}

func TestSchemaValidationArray(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
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
	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("ports", []interface{}{80, 443, 8080})

	err := cr.ValidateAgainstSD(crd)
	assert.NoError(t, err)

	// Invalid array item type
	cr2 := CreateBlueprint("TestBlueprint", "test2", "ns")
	cr2.SetSpec("ports", []interface{}{80, "not-a-number", 8080})

	err = cr2.ValidateAgainstSD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}

func TestSchemaValidationKindMismatch(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test-controller",
		"reconcile",
	)

	// Wrong kind
	cr := CreateBlueprint("WrongBlueprint", "test", "ns")

	err := cr.ValidateAgainstSD(crd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kind mismatch")
}

func TestSchemaValidationNoSchema(t *testing.T) {
	crd := CreateBlueprintDefinition(
		"testblueprints.test.io",
		"test.io",
		"v1",
		"TestBlueprint",
		"testblueprints",
		"Namespaced",
		"test-controller",
		"reconcile",
	)
	// No schema defined

	cr := CreateBlueprint("TestBlueprint", "test", "ns")
	cr.SetSpec("anything", "goes")

	// Should pass validation when no schema is defined
	err := cr.ValidateAgainstSD(crd)
	assert.NoError(t, err)
}

// Reconciliation tests

func TestCreateReconciliationCreate(t *testing.T) {
	// Test create action (old is nil, new exists)
	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("replicas", 3)

	reconciliation := CreateReconciliation(nil, newBlueprint)

	assert.Nil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationCreate, reconciliation.Action)
	assert.Nil(t, reconciliation.Diff)
}

func TestCreateReconciliationDelete(t *testing.T) {
	// Test delete action (old exists, new is nil)
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)

	reconciliation := CreateReconciliation(oldBlueprint, nil)

	assert.NotNil(t, reconciliation.Old)
	assert.Nil(t, reconciliation.New)
	assert.Equal(t, ReconciliationDelete, reconciliation.Action)
	assert.Nil(t, reconciliation.Diff)
}

func TestCreateReconciliationUpdate(t *testing.T) {
	// Test update action (both exist with changes)
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)
	oldBlueprint.SetSpec("image", "nginx:1.0")

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("replicas", 5)
	newBlueprint.SetSpec("image", "nginx:2.0")

	reconciliation := CreateReconciliation(oldBlueprint, newBlueprint)

	assert.NotNil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationUpdate, reconciliation.Action)
	assert.NotNil(t, reconciliation.Diff)
	assert.True(t, reconciliation.Diff.HasChanges)
	assert.Equal(t, 2, len(reconciliation.Diff.SpecChanges))
}

func TestCreateReconciliationNoop(t *testing.T) {
	// Test noop action (both exist with no changes)
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)
	oldBlueprint.SetSpec("image", "nginx:1.0")

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("replicas", 3)
	newBlueprint.SetSpec("image", "nginx:1.0")

	reconciliation := CreateReconciliation(oldBlueprint, newBlueprint)

	assert.NotNil(t, reconciliation.Old)
	assert.NotNil(t, reconciliation.New)
	assert.Equal(t, ReconciliationNoop, reconciliation.Action)
	assert.NotNil(t, reconciliation.Diff)
	assert.False(t, reconciliation.Diff.HasChanges)
}

func TestBlueprintDiffSpecChanges(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)
	oldBlueprint.SetSpec("image", "nginx:1.0")
	oldBlueprint.SetSpec("port", 8080)

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("replicas", 5)           // Modified
	newBlueprint.SetSpec("image", "nginx:1.0")    // Unchanged
	newBlueprint.SetSpec("command", []string{"run"}) // Added
	// port removed

	diff := oldBlueprint.Diff(newBlueprint)

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

func TestBlueprintDiffStatusChanges(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetStatus("phase", "Pending")
	oldBlueprint.SetStatus("ready", 0)

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetStatus("phase", "Running")
	newBlueprint.SetStatus("ready", 3)

	diff := oldBlueprint.Diff(newBlueprint)

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

func TestBlueprintDiffMetadataChanges(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.Metadata.Labels["app"] = "old-app"
	oldBlueprint.Metadata.Labels["version"] = "1.0"
	oldBlueprint.Metadata.Annotations["description"] = "old description"

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.Metadata.Labels["app"] = "new-app"
	newBlueprint.Metadata.Labels["environment"] = "production"
	newBlueprint.Metadata.Annotations["description"] = "new description"

	diff := oldBlueprint.Diff(newBlueprint)

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

func TestBlueprintDiffHelperMethods(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)
	oldBlueprint.SetStatus("phase", "Running")
	oldBlueprint.Metadata.Labels["app"] = "test"

	// Test OnlyStatusChanged
	newBlueprint1 := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint1.SetSpec("replicas", 3)
	newBlueprint1.SetStatus("phase", "Succeeded")
	newBlueprint1.Metadata.Labels["app"] = "test"

	diff1 := oldBlueprint.Diff(newBlueprint1)
	assert.True(t, diff1.OnlyStatusChanged())
	assert.False(t, diff1.OnlyMetadataChanged())

	// Test OnlyMetadataChanged
	newBlueprint2 := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint2.SetSpec("replicas", 3)
	newBlueprint2.SetStatus("phase", "Running")
	newBlueprint2.Metadata.Labels["app"] = "new-app"

	diff2 := oldBlueprint.Diff(newBlueprint2)
	assert.True(t, diff2.OnlyMetadataChanged())
	assert.False(t, diff2.OnlyStatusChanged())

	// Test mixed changes
	newBlueprint3 := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint3.SetSpec("replicas", 5)
	newBlueprint3.SetStatus("phase", "Failed")
	newBlueprint3.Metadata.Labels["app"] = "new-app"

	diff3 := oldBlueprint.Diff(newBlueprint3)
	assert.False(t, diff3.OnlyStatusChanged())
	assert.False(t, diff3.OnlyMetadataChanged())
}

func TestBlueprintDiffComplexChanges(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("config", map[string]interface{}{
		"cpu":    "2",
		"memory": "4Gi",
	})

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("config", map[string]interface{}{
		"cpu":    "4",
		"memory": "4Gi",
	})

	diff := oldBlueprint.Diff(newBlueprint)

	assert.True(t, diff.HasChanges)
	assert.Equal(t, 1, len(diff.SpecChanges))

	// The entire config object changed
	configChange := diff.GetFieldChange("spec.config")
	assert.NotNil(t, configChange)
	assert.Equal(t, ChangeModified, configChange.Type)
}

func TestHasFieldChange(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("replicas", 3)
	oldBlueprint.SetStatus("phase", "Running")

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("replicas", 5)
	newBlueprint.SetStatus("phase", "Running")

	diff := oldBlueprint.Diff(newBlueprint)

	assert.True(t, diff.HasFieldChange("spec.replicas"))
	assert.False(t, diff.HasFieldChange("spec.image"))
	assert.False(t, diff.HasFieldChange("status.phase"))
}

func TestGetFieldChange(t *testing.T) {
	oldBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	oldBlueprint.SetSpec("image", "nginx:1.0")

	newBlueprint := CreateBlueprint("TestBlueprint", "test", "ns")
	newBlueprint.SetSpec("image", "nginx:2.0")

	diff := oldBlueprint.Diff(newBlueprint)

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

func TestDeepEqualWithJSONSerializableValues(t *testing.T) {
	// Test equal values
	a := map[string]interface{}{"key": "value", "num": 42}
	b := map[string]interface{}{"key": "value", "num": 42}
	assert.True(t, deepEqual(a, b), "Equal maps should return true")

	// Test unequal values
	c := map[string]interface{}{"key": "different", "num": 42}
	assert.False(t, deepEqual(a, c), "Different maps should return false")

	// Test with nested structures
	nested1 := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	}
	nested2 := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	}
	assert.True(t, deepEqual(nested1, nested2), "Equal nested structures should return true")

	// Test with different nested values
	nested3 := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "different",
		},
	}
	assert.False(t, deepEqual(nested1, nested3), "Different nested structures should return false")
}

func TestDeepEqualWithUnmarshalableValues(t *testing.T) {
	// Create values that can't be JSON marshaled (channels)
	ch1 := make(chan int)
	ch2 := make(chan int)

	// Same channel should be equal (reflect.DeepEqual fallback)
	assert.True(t, deepEqual(ch1, ch1), "Same channel should be equal via reflect.DeepEqual")

	// Different channels should not be equal
	assert.False(t, deepEqual(ch1, ch2), "Different channels should not be equal")

	// Note: Functions are NOT tested here because reflect.DeepEqual returns false
	// for all function comparisons in Go, even for the same function reference.
	// This is by design in Go's reflect package.
}

func TestDeepEqualDoesNotReturnTrueForDifferentUnmarshalableValues(t *testing.T) {
	// This test verifies the bug fix: before the fix, different unmarshable values
	// would both marshal to empty strings and incorrectly compare as equal

	// Create two different channels
	ch1 := make(chan int)
	ch2 := make(chan int)

	// Before the fix: both would marshal to "" and return true (BUG!)
	// After the fix: falls back to reflect.DeepEqual and returns false
	result := deepEqual(ch1, ch2)
	assert.False(t, result, "Different unmarshable values should NOT be equal (bug fix verification)")

	// Also test with a mix: one marshable, one not
	normalValue := map[string]string{"key": "value"}
	unmarshalableValue := make(chan int)

	// These should definitely not be equal
	assert.False(t, deepEqual(normalValue, unmarshalableValue),
		"Marshable and unmarshable values should not be equal")
}

func TestDeepEqualWithBlueprintSpec(t *testing.T) {
	// Test with actual blueprint-like structures
	spec1 := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.21",
		"env": map[string]interface{}{
			"PORT": "8080",
		},
	}
	spec2 := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.21",
		"env": map[string]interface{}{
			"PORT": "8080",
		},
	}
	spec3 := map[string]interface{}{
		"replicas": 5, // Different
		"image":    "nginx:1.21",
		"env": map[string]interface{}{
			"PORT": "8080",
		},
	}

	assert.True(t, deepEqual(spec1, spec2), "Identical specs should be equal")
	assert.False(t, deepEqual(spec1, spec3), "Different specs should not be equal")
}

func TestConvertBlueprintArrayToJSON(t *testing.T) {
	bp1 := CreateBlueprint("TestKind", "test1", "ns")
	bp1.SetSpec("replicas", 3)
	bp2 := CreateBlueprint("TestKind", "test2", "ns")
	bp2.SetSpec("replicas", 5)

	blueprints := []*Blueprint{bp1, bp2}

	jsonStr, err := ConvertBlueprintArrayToJSON(blueprints)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "test1")
	assert.Contains(t, jsonStr, "test2")
}

func TestConvertJSONToBlueprintArray(t *testing.T) {
	bp1 := CreateBlueprint("TestKind", "test1", "ns")
	bp1.SetSpec("replicas", 3)
	bp2 := CreateBlueprint("TestKind", "test2", "ns")
	bp2.SetSpec("replicas", 5)

	blueprints := []*Blueprint{bp1, bp2}

	jsonStr, err := ConvertBlueprintArrayToJSON(blueprints)
	assert.NoError(t, err)

	parsed, err := ConvertJSONToBlueprintArray(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "test1", parsed[0].Metadata.Name)
	assert.Equal(t, "test2", parsed[1].Metadata.Name)

	// Test invalid JSON
	_, err = ConvertJSONToBlueprintArray("invalid json")
	assert.Error(t, err)
}

func TestConvertBlueprintDefinitionArrayToJSON(t *testing.T) {
	sd1 := CreateBlueprintDefinition("sd1", "test.io", "v1", "Kind1", "kinds1", "Namespaced", "executor1", "func1")
	sd2 := CreateBlueprintDefinition("sd2", "test.io", "v1", "Kind2", "kinds2", "Namespaced", "executor2", "func2")

	sds := []*BlueprintDefinition{sd1, sd2}

	jsonStr, err := ConvertBlueprintDefinitionArrayToJSON(sds)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "Kind1")
	assert.Contains(t, jsonStr, "Kind2")
}

func TestConvertJSONToBlueprintDefinitionArray(t *testing.T) {
	sd1 := CreateBlueprintDefinition("sd1", "test.io", "v1", "Kind1", "kinds1", "Namespaced", "executor1", "func1")
	sd2 := CreateBlueprintDefinition("sd2", "test.io", "v1", "Kind2", "kinds2", "Namespaced", "executor2", "func2")

	sds := []*BlueprintDefinition{sd1, sd2}

	jsonStr, err := ConvertBlueprintDefinitionArrayToJSON(sds)
	assert.NoError(t, err)

	parsed, err := ConvertJSONToBlueprintDefinitionArray(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "Kind1", parsed[0].Spec.Names.Kind)
	assert.Equal(t, "Kind2", parsed[1].Spec.Names.Kind)

	// Test invalid JSON
	_, err = ConvertJSONToBlueprintDefinitionArray("invalid json")
	assert.Error(t, err)
}

func TestCreateBlueprintHistory(t *testing.T) {
	bp := CreateBlueprint("TestKind", "test-name", "test-colony")
	bp.SetSpec("replicas", 3)
	bp.SetSpec("config", map[string]interface{}{"key": "value"})
	bp.SetStatus("phase", "Running")

	history := CreateBlueprintHistory(bp, "user-123", "create")

	assert.NotEmpty(t, history.ID)
	assert.Equal(t, bp.ID, history.BlueprintID)
	assert.Equal(t, "TestKind", history.Kind)
	assert.Equal(t, "test-colony", history.Namespace)
	assert.Equal(t, "test-name", history.Name)
	assert.Equal(t, bp.Metadata.Generation, history.Generation)
	assert.Equal(t, "user-123", history.ChangedBy)
	assert.Equal(t, "create", history.ChangeType)
	assert.NotNil(t, history.Spec)
	assert.NotNil(t, history.Status)
}

func TestBlueprintHistoryToJSON(t *testing.T) {
	bp := CreateBlueprint("TestKind", "test-name", "test-colony")
	bp.SetSpec("replicas", 3)
	history := CreateBlueprintHistory(bp, "user-123", "update")

	jsonStr, err := history.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "user-123")
	assert.Contains(t, jsonStr, "update")
}

func TestConvertJSONToBlueprintHistory(t *testing.T) {
	bp := CreateBlueprint("TestKind", "test-name", "test-colony")
	bp.SetSpec("replicas", 3)
	history := CreateBlueprintHistory(bp, "user-123", "delete")

	jsonStr, err := history.ToJSON()
	assert.NoError(t, err)

	parsed, err := ConvertJSONToBlueprintHistory(jsonStr)
	assert.NoError(t, err)
	assert.Equal(t, history.ID, parsed.ID)
	assert.Equal(t, history.BlueprintID, parsed.BlueprintID)
	assert.Equal(t, "user-123", parsed.ChangedBy)
	assert.Equal(t, "delete", parsed.ChangeType)

	// Test invalid JSON
	_, err = ConvertJSONToBlueprintHistory("invalid json")
	assert.Error(t, err)
}

func TestConvertBlueprintHistoryArrayToJSON(t *testing.T) {
	bp := CreateBlueprint("TestKind", "test-name", "test-colony")
	bp.SetSpec("replicas", 3)

	history1 := CreateBlueprintHistory(bp, "user-1", "create")
	history2 := CreateBlueprintHistory(bp, "user-2", "update")

	histories := []*BlueprintHistory{history1, history2}

	jsonStr, err := ConvertBlueprintHistoryArrayToJSON(histories)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "user-1")
	assert.Contains(t, jsonStr, "user-2")
}

func TestConvertJSONToBlueprintHistoryArray(t *testing.T) {
	bp := CreateBlueprint("TestKind", "test-name", "test-colony")
	bp.SetSpec("replicas", 3)

	history1 := CreateBlueprintHistory(bp, "user-1", "create")
	history2 := CreateBlueprintHistory(bp, "user-2", "update")

	histories := []*BlueprintHistory{history1, history2}

	jsonStr, err := ConvertBlueprintHistoryArrayToJSON(histories)
	assert.NoError(t, err)

	parsed, err := ConvertJSONToBlueprintHistoryArray(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "user-1", parsed[0].ChangedBy)
	assert.Equal(t, "user-2", parsed[1].ChangedBy)

	// Test invalid JSON
	_, err = ConvertJSONToBlueprintHistoryArray("invalid json")
	assert.Error(t, err)
}

func TestCopyMapAndCopySlice(t *testing.T) {
	// Test copyMap with nested structures
	original := map[string]interface{}{
		"simple":  "value",
		"number":  42,
		"boolean": true,
		"nested": map[string]interface{}{
			"inner": "innerValue",
		},
		"array": []interface{}{"a", "b", "c"},
	}

	copied := copyMap(original)

	// Verify copy is not nil
	assert.NotNil(t, copied)

	// Verify values are equal
	assert.Equal(t, original["simple"], copied["simple"])
	assert.Equal(t, original["number"], copied["number"])

	// Verify nested map is a deep copy
	originalNested := original["nested"].(map[string]interface{})
	copiedNested := copied["nested"].(map[string]interface{})
	assert.Equal(t, originalNested["inner"], copiedNested["inner"])

	// Modify original to verify deep copy
	originalNested["inner"] = "modified"
	assert.NotEqual(t, originalNested["inner"], copiedNested["inner"])

	// Test nil map
	assert.Nil(t, copyMap(nil))
}

func TestCopySliceNested(t *testing.T) {
	original := []interface{}{
		"string",
		42,
		map[string]interface{}{"key": "value"},
		[]interface{}{1, 2, 3},
	}

	copied := copySlice(original)

	assert.NotNil(t, copied)
	assert.Len(t, copied, 4)
	assert.Equal(t, "string", copied[0])
	assert.Equal(t, 42, copied[1])

	// Verify nested map is a deep copy
	originalMap := original[2].(map[string]interface{})
	copiedMap := copied[2].(map[string]interface{})
	assert.Equal(t, originalMap["key"], copiedMap["key"])

	// Modify original to verify deep copy
	originalMap["key"] = "modified"
	assert.NotEqual(t, originalMap["key"], copiedMap["key"])

	// Test nil slice
	assert.Nil(t, copySlice(nil))
}

