package core

import (
	"testing"
)

func TestValidateServiceAgainstSchema_RequiredFields(t *testing.T) {
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
	service := CreateService("Test", "test-service", "default")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test 2: Missing required field
	invalidService := CreateService("Test", "invalid-service", "default")
	invalidService.SetSpec("replicas", 3) // Missing 'image'

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field")
	}
}

func TestValidateServiceAgainstSchema_TypeValidation(t *testing.T) {
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
	service := CreateService("Test", "test", "default")
	service.SetSpec("name", "test-name")
	service.SetSpec("count", 5)
	service.SetSpec("enabled", true)

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test invalid string type
	invalidService := CreateService("Test", "invalid", "default")
	invalidService.SetSpec("name", 123) // Should be string

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid string type")
	}

	// Test invalid number type
	invalidService2 := CreateService("Test", "invalid2", "default")
	invalidService2.SetSpec("count", "not-a-number") // Should be number

	err = ValidateServiceAgainstSchema(invalidService2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid number type")
	}

	// Test invalid boolean type
	invalidService3 := CreateService("Test", "invalid3", "default")
	invalidService3.SetSpec("enabled", "not-a-boolean") // Should be boolean

	err = ValidateServiceAgainstSchema(invalidService3, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid boolean type")
	}
}

func TestValidateServiceAgainstSchema_EnumValidation(t *testing.T) {
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
	service := CreateService("Test", "test", "default")
	service.SetSpec("protocol", "TCP")

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid enum, got error: %v", err)
	}

	// Test invalid enum value
	invalidService := CreateService("Test", "invalid", "default")
	invalidService.SetSpec("protocol", "HTTP") // Not in enum

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestValidateServiceAgainstSchema_ArrayValidation(t *testing.T) {
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
	service := CreateService("Test", "test", "default")
	service.SetSpec("ports", []interface{}{80, 443, 8080})

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid array, got error: %v", err)
	}

	// Test invalid array item type
	invalidService := CreateService("Test", "invalid", "default")
	invalidService.SetSpec("ports", []interface{}{80, "not-a-number", 8080})

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid array item type")
	}

	// Test non-array value
	invalidService2 := CreateService("Test", "invalid2", "default")
	invalidService2.SetSpec("ports", "not-an-array")

	err = ValidateServiceAgainstSchema(invalidService2, schema)
	if err == nil {
		t.Error("Expected validation to fail for non-array value")
	}
}

func TestValidateServiceAgainstSchema_NestedObjectValidation(t *testing.T) {
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
	service := CreateService("Test", "test", "default")
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}
	service.SetSpec("config", config)

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid nested object, got error: %v", err)
	}

	// Test invalid nested field type
	invalidService := CreateService("Test", "invalid", "default")
	invalidConfig := map[string]interface{}{
		"host": "localhost",
		"port": "not-a-number", // Should be number
	}
	invalidService.SetSpec("config", invalidConfig)

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid nested field type")
	}
}

func TestValidateServiceAgainstSchema_NoSchema(t *testing.T) {
	// Test with nil schema - should pass
	service := CreateService("Test", "test", "default")
	service.SetSpec("anything", "goes")

	err := ValidateServiceAgainstSchema(service, nil)
	if err != nil {
		t.Errorf("Expected validation to pass with nil schema, got error: %v", err)
	}
}

func TestValidateServiceAgainstSchema_ComplexExample(t *testing.T) {
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
	service := CreateService("ExecutorDeployment", "web-server", "production")
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

	err := ValidateServiceAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for complex valid service, got error: %v", err)
	}

	// Test missing required field
	invalidService := CreateService("ExecutorDeployment", "invalid", "production")
	invalidService.SetSpec("image", "nginx:1.21")
	// Missing executorType

	err = ValidateServiceAgainstSchema(invalidService, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field 'executorType'")
	}

	// Test invalid port protocol enum
	invalidService2 := CreateService("ExecutorDeployment", "invalid2", "production")
	invalidService2.SetSpec("image", "nginx:1.21")
	invalidService2.SetSpec("executorType", "container-executor")
	invalidService2.SetSpec("ports", []interface{}{
		map[string]interface{}{
			"name":     "http",
			"port":     80,
			"protocol": "HTTP", // Invalid, should be TCP or UDP
		},
	})

	err = ValidateServiceAgainstSchema(invalidService2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value in nested array")
	}
}
