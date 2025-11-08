package core

import (
	"testing"
)

func TestValidateResourceAgainstSchema_RequiredFields(t *testing.T) {
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"image": {
				Type:        "string",
				Description: "Container image",
			},
			"replicas": {
				Type:        "number",
				Description: "Number of replicas",
			},
		},
		Required: []string{"image"},
	}

	// Test 1: Valid service with required field
	service := CreateResource("Test", "test-service", "default")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test 2: Missing required field
	invalidResource := CreateResource("Test", "invalid-service", "default")
	invalidResource.SetSpec("replicas", 3) // Missing 'image'

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field")
	}
}

func TestValidateResourceAgainstSchema_TypeValidation(t *testing.T) {
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"name": {
				Type: "string",
			},
			"count": {
				Type: "number",
			},
			"enabled": {
				Type: "boolean",
			},
		},
	}

	// Test string type
	service := CreateResource("Test", "test", "default")
	service.SetSpec("name", "test-name")
	service.SetSpec("count", 5)
	service.SetSpec("enabled", true)

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test invalid string type
	invalidResource := CreateResource("Test", "invalid", "default")
	invalidResource.SetSpec("name", 123) // Should be string

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid string type")
	}

	// Test invalid number type
	invalidResource2 := CreateResource("Test", "invalid2", "default")
	invalidResource2.SetSpec("count", "not-a-number") // Should be number

	err = ValidateResourceAgainstSchema(invalidResource2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid number type")
	}

	// Test invalid boolean type
	invalidResource3 := CreateResource("Test", "invalid3", "default")
	invalidResource3.SetSpec("enabled", "not-a-boolean") // Should be boolean

	err = ValidateResourceAgainstSchema(invalidResource3, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid boolean type")
	}
}

func TestValidateResourceAgainstSchema_EnumValidation(t *testing.T) {
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"protocol": {
				Type: "string",
				Enum: []interface{}{"TCP", "UDP"},
			},
		},
	}

	// Test valid enum value
	service := CreateResource("Test", "test", "default")
	service.SetSpec("protocol", "TCP")

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid enum, got error: %v", err)
	}

	// Test invalid enum value
	invalidResource := CreateResource("Test", "invalid", "default")
	invalidResource.SetSpec("protocol", "HTTP") // Not in enum

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestValidateResourceAgainstSchema_ArrayValidation(t *testing.T) {
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"ports": {
				Type: "array",
				Items: &SchemaProperty{
					Type: "number",
				},
			},
		},
	}

	// Test valid array
	service := CreateResource("Test", "test", "default")
	service.SetSpec("ports", []interface{}{80, 443, 8080})

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid array, got error: %v", err)
	}

	// Test invalid array item type
	invalidResource := CreateResource("Test", "invalid", "default")
	invalidResource.SetSpec("ports", []interface{}{80, "not-a-number", 8080})

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid array item type")
	}

	// Test non-array value
	invalidResource2 := CreateResource("Test", "invalid2", "default")
	invalidResource2.SetSpec("ports", "not-an-array")

	err = ValidateResourceAgainstSchema(invalidResource2, schema)
	if err == nil {
		t.Error("Expected validation to fail for non-array value")
	}
}

func TestValidateResourceAgainstSchema_NestedObjectValidation(t *testing.T) {
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"config": {
				Type: "object",
				Properties: map[string]SchemaProperty{
					"host": {
						Type: "string",
					},
					"port": {
						Type: "number",
					},
				},
			},
		},
	}

	// Test valid nested object
	service := CreateResource("Test", "test", "default")
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}
	service.SetSpec("config", config)

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid nested object, got error: %v", err)
	}

	// Test invalid nested field type
	invalidResource := CreateResource("Test", "invalid", "default")
	invalidConfig := map[string]interface{}{
		"host": "localhost",
		"port": "not-a-number", // Should be number
	}
	invalidResource.SetSpec("config", invalidConfig)

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid nested field type")
	}
}

func TestValidateResourceAgainstSchema_NoSchema(t *testing.T) {
	// Test with nil schema - should pass
	service := CreateResource("Test", "test", "default")
	service.SetSpec("anything", "goes")

	err := ValidateResourceAgainstSchema(service, nil)
	if err != nil {
		t.Errorf("Expected validation to pass with nil schema, got error: %v", err)
	}
}

func TestValidateResourceAgainstSchema_ComplexExample(t *testing.T) {
	// Test the ExecutorDeployment example schema
	schema := &ValidationSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"image": {
				Type:        "string",
				Description: "Container image to deploy",
			},
			"replicas": {
				Type:        "number",
				Description: "Number of executor replicas to run",
				Default:     1,
			},
			"executorType": {
				Type:        "string",
				Description: "Type of executor to deploy",
			},
			"cpu": {
				Type:        "string",
				Description: "CPU service request",
			},
			"memory": {
				Type:        "string",
				Description: "Memory service request",
			},
			"ports": {
				Type:        "array",
				Description: "Ports to expose",
				Items: &SchemaProperty{
					Type: "object",
					Properties: map[string]SchemaProperty{
						"name": {
							Type: "string",
						},
						"port": {
							Type: "number",
						},
						"protocol": {
							Type: "string",
							Enum: []interface{}{"TCP", "UDP"},
						},
					},
				},
			},
		},
		Required: []string{"image", "executorType"},
	}

	// Test valid complex service
	service := CreateResource("ExecutorDeployment", "web-server", "production")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)
	service.SetSpec("executorType", "container-executor")
	service.SetSpec("cpu", "500m")
	service.SetSpec("memory", "512Mi")
	service.SetSpec("ports", []interface{}{
		map[string]interface{}{
			"name":     "http",
			"port":     80,
			"protocol": "TCP",
		},
		map[string]interface{}{
			"name":     "https",
			"port":     443,
			"protocol": "TCP",
		},
	})

	err := ValidateResourceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for complex valid service, got error: %v", err)
	}

	// Test missing required field
	invalidResource := CreateResource("ExecutorDeployment", "invalid", "production")
	invalidResource.SetSpec("image", "nginx:1.21")
	// Missing executorType

	err = ValidateResourceAgainstSchema(invalidResource, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field 'executorType'")
	}

	// Test invalid port protocol enum
	invalidResource2 := CreateResource("ExecutorDeployment", "invalid2", "production")
	invalidResource2.SetSpec("image", "nginx:1.21")
	invalidResource2.SetSpec("executorType", "container-executor")
	invalidResource2.SetSpec("ports", []interface{}{
		map[string]interface{}{
			"name":     "http",
			"port":     80,
			"protocol": "HTTP", // Invalid, should be TCP or UDP
		},
	})

	err = ValidateResourceAgainstSchema(invalidResource2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value in nested array")
	}
}
