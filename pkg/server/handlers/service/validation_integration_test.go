package service

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSchemaValidation_Integration(t *testing.T) {
	// Create a ServiceDefinition with schema
	sd := core.CreateServiceDefinition(
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
	validService := core.CreateService("Deployment", "valid-deployment", "test-colony")
	validService.SetSpec("image", "nginx:1.21")
	validService.SetSpec("replicas", 3)
	validService.SetSpec("protocol", "TCP")

	err := core.ValidateServiceAgainstSchema(validService, sd.Spec.Schema)
	assert.NoError(t, err, "Valid service should pass validation")

	// Test 2: Service missing required field should fail
	invalidService := core.CreateService("Deployment", "invalid-deployment", "test-colony")
	invalidService.SetSpec("replicas", 3) // Missing required 'image'

	err = core.ValidateServiceAgainstSchema(invalidService, sd.Spec.Schema)
	assert.Error(t, err, "Service missing required field should fail validation")
	assert.Contains(t, err.Error(), "required field 'image' is missing")

	// Test 3: Service with invalid type should fail
	invalidTypeService := core.CreateService("Deployment", "invalid-type", "test-colony")
	invalidTypeService.SetSpec("image", "nginx:1.21")
	invalidTypeService.SetSpec("replicas", "not-a-number") // Should be number

	err = core.ValidateServiceAgainstSchema(invalidTypeService, sd.Spec.Schema)
	assert.Error(t, err, "Service with invalid type should fail validation")
	assert.Contains(t, err.Error(), "must be a number")

	// Test 4: Service with invalid enum value should fail
	invalidEnumService := core.CreateService("Deployment", "invalid-enum", "test-colony")
	invalidEnumService.SetSpec("image", "nginx:1.21")
	invalidEnumService.SetSpec("protocol", "HTTP") // Not in enum [TCP, UDP]

	err = core.ValidateServiceAgainstSchema(invalidEnumService, sd.Spec.Schema)
	assert.Error(t, err, "Service with invalid enum value should fail validation")
	assert.Contains(t, err.Error(), "must be one of")
}
