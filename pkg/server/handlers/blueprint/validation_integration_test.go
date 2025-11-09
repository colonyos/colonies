package blueprint

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSchemaValidation_Integration(t *testing.T) {
	// Create a BlueprintDefinition with schema
	sd := core.CreateBlueprintDefinition(
		"deployments.compute.io",
		"compute.io",
		"v1",
		"Deployment",
		"deployments",
		"Namespaced",
		"controller",
		"reconcile",
	)
	sd.Metadata.Namespace = "test-colony"
	sd.Spec.Schema = &core.ValidationSchema{
		Type: "object",
		Properties: map[string]core.SchemaProperty{
			"image": {
				Type:        "string",
				Description: "Container image",
			},
			"replicas": {
				Type:        "number",
				Description: "Number of replicas",
			},
			"protocol": {
				Type: "string",
				Enum: []interface{}{"TCP", "UDP"},
			},
		},
		Required: []string{"image"},
	}

	// Test 1: Valid service should pass validation
	validBlueprint := core.CreateBlueprint("Deployment", "valid-deployment", "test-colony")
	validBlueprint.SetSpec("image", "nginx:1.21")
	validBlueprint.SetSpec("replicas", 3)
	validBlueprint.SetSpec("protocol", "TCP")

	err := core.ValidateBlueprintAgainstSchema(validBlueprint, sd.Spec.Schema)
	assert.NoError(t, err, "Valid service should pass validation")

	// Test 2: Blueprint missing required field should fail
	invalidBlueprint := core.CreateBlueprint("Deployment", "invalid-deployment", "test-colony")
	invalidBlueprint.SetSpec("replicas", 3) // Missing required 'image'

	err = core.ValidateBlueprintAgainstSchema(invalidBlueprint, sd.Spec.Schema)
	assert.Error(t, err, "Blueprint missing required field should fail validation")
	assert.Contains(t, err.Error(), "required field 'image' is missing")

	// Test 3: Blueprint with invalid type should fail
	invalidTypeBlueprint := core.CreateBlueprint("Deployment", "invalid-type", "test-colony")
	invalidTypeBlueprint.SetSpec("image", "nginx:1.21")
	invalidTypeBlueprint.SetSpec("replicas", "not-a-number") // Should be number

	err = core.ValidateBlueprintAgainstSchema(invalidTypeBlueprint, sd.Spec.Schema)
	assert.Error(t, err, "Blueprint with invalid type should fail validation")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Blueprint with invalid enum value should fail
	invalidEnumBlueprint := core.CreateBlueprint("Deployment", "invalid-enum", "test-colony")
	invalidEnumBlueprint.SetSpec("image", "nginx:1.21")
	invalidEnumBlueprint.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	err = core.ValidateBlueprintAgainstSchema(invalidEnumBlueprint, sd.Spec.Schema)
	assert.Error(t, err, "Blueprint with invalid enum value should fail validation")
	assert.Contains(t, err.Error(), "must be one of")
}
