package resource

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSchemaValidation_Integration(t *testing.T) {
	// Create a ResourceDefinition with schema
	rd := core.CreateResourceDefinition(
		"deployments.compute.io",
		"compute.io",
		"v1",
		"Deployment",
		"deployments",
		"Namespaced",
		"controller",
		"reconcile",
	)
	rd.Metadata.Namespace = "test-colony"
	rd.Spec.Schema = &core.ValidationSchema{
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

	// Test 1: Valid resource should pass validation
	validResource := core.CreateResource("Deployment", "valid-deployment", "test-colony")
	validResource.SetSpec("image", "nginx:1.21")
	validResource.SetSpec("replicas", 3)
	validResource.SetSpec("protocol", "TCP")

	err := core.ValidateResourceAgainstSchema(validResource, rd.Spec.Schema)
	assert.NoError(t, err, "Valid resource should pass validation")

	// Test 2: Resource missing required field should fail
	invalidResource := core.CreateResource("Deployment", "invalid-deployment", "test-colony")
	invalidResource.SetSpec("replicas", 3) // Missing required 'image'

	err = core.ValidateResourceAgainstSchema(invalidResource, rd.Spec.Schema)
	assert.Error(t, err, "Resource missing required field should fail validation")
	assert.Contains(t, err.Error(), "required field 'image' is missing")

	// Test 3: Resource with invalid type should fail
	invalidTypeResource := core.CreateResource("Deployment", "invalid-type", "test-colony")
	invalidTypeResource.SetSpec("image", "nginx:1.21")
	invalidTypeResource.SetSpec("replicas", "not-a-number") // Should be number

	err = core.ValidateResourceAgainstSchema(invalidTypeResource, rd.Spec.Schema)
	assert.Error(t, err, "Resource with invalid type should fail validation")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Resource with invalid enum value should fail
	invalidEnumResource := core.CreateResource("Deployment", "invalid-enum", "test-colony")
	invalidEnumResource.SetSpec("image", "nginx:1.21")
	invalidEnumResource.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	err = core.ValidateResourceAgainstSchema(invalidEnumResource, rd.Spec.Schema)
	assert.Error(t, err, "Resource with invalid enum value should fail validation")
	assert.Contains(t, err.Error(), "must be one of")
}
