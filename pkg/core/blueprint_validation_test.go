package core

import (
	"testing"
)

func TestValidateBlueprintAgainstSchema_RequiredFields(t *testing.T) {
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
	service := CreateBlueprint("Test", "test-service", "default")
	service.SetSpec("image", "nginx:1.21")
	service.SetSpec("replicas", 3)

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test 2: Missing required field
	invalidBlueprint := CreateBlueprint("Test", "invalid-service", "default")
	invalidBlueprint.SetSpec("replicas", 3) // Missing 'image'

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field")
	}
}

func TestValidateBlueprintAgainstSchema_TypeValidation(t *testing.T) {
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
	service := CreateBlueprint("Test", "test", "default")
	service.SetSpec("name", "test-name")
	service.SetSpec("count", 5)
	service.SetSpec("enabled", true)

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Test invalid string type
	invalidBlueprint := CreateBlueprint("Test", "invalid", "default")
	invalidBlueprint.SetSpec("name", 123) // Should be string

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid string type")
	}

	// Test invalid number type
	invalidBlueprint2 := CreateBlueprint("Test", "invalid2", "default")
	invalidBlueprint2.SetSpec("count", "not-a-number") // Should be number

	err = ValidateBlueprintAgainstSchema(invalidBlueprint2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid number type")
	}

	// Test invalid boolean type
	invalidBlueprint3 := CreateBlueprint("Test", "invalid3", "default")
	invalidBlueprint3.SetSpec("enabled", "not-a-boolean") // Should be boolean

	err = ValidateBlueprintAgainstSchema(invalidBlueprint3, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid boolean type")
	}
}

func TestValidateBlueprintAgainstSchema_EnumValidation(t *testing.T) {
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
	service := CreateBlueprint("Test", "test", "default")
	service.SetSpec("protocol", "TCP")

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid enum, got error: %v", err)
	}

	// Test invalid enum value
	invalidBlueprint := CreateBlueprint("Test", "invalid", "default")
	invalidBlueprint.SetSpec("protocol", "HTTP") // Not in enum

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestValidateBlueprintAgainstSchema_ArrayValidation(t *testing.T) {
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
	service := CreateBlueprint("Test", "test", "default")
	service.SetSpec("ports", []interface{}{80, 443, 8080})

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid array, got error: %v", err)
	}

	// Test invalid array item type
	invalidBlueprint := CreateBlueprint("Test", "invalid", "default")
	invalidBlueprint.SetSpec("ports", []interface{}{80, "not-a-number", 8080})

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid array item type")
	}

	// Test non-array value
	invalidBlueprint2 := CreateBlueprint("Test", "invalid2", "default")
	invalidBlueprint2.SetSpec("ports", "not-an-array")

	err = ValidateBlueprintAgainstSchema(invalidBlueprint2, schema)
	if err == nil {
		t.Error("Expected validation to fail for non-array value")
	}
}

func TestValidateBlueprintAgainstSchema_NestedObjectValidation(t *testing.T) {
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
	service := CreateBlueprint("Test", "test", "default")
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}
	service.SetSpec("config", config)

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for valid nested object, got error: %v", err)
	}

	// Test invalid nested field type
	invalidBlueprint := CreateBlueprint("Test", "invalid", "default")
	invalidConfig := map[string]interface{}{
		"host": "localhost",
		"port": "not-a-number", // Should be number
	}
	invalidBlueprint.SetSpec("config", invalidConfig)

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid nested field type")
	}
}

func TestValidateBlueprintAgainstSchema_NoSchema(t *testing.T) {
	// Test with nil schema - should pass
	service := CreateBlueprint("Test", "test", "default")
	service.SetSpec("anything", "goes")

	err := ValidateBlueprintAgainstSchema(service, nil)
	if err != nil {
		t.Errorf("Expected validation to pass with nil schema, got error: %v", err)
	}
}

func TestValidateBlueprintAgainstSchema_ComplexExample(t *testing.T) {
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
	service := CreateBlueprint("ExecutorDeployment", "web-server", "production")
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

	err := ValidateBlueprintAgainstSchema(service, schema)
	if err != nil {
		t.Errorf("Expected validation to pass for complex valid service, got error: %v", err)
	}

	// Test missing required field
	invalidBlueprint := CreateBlueprint("ExecutorDeployment", "invalid", "production")
	invalidBlueprint.SetSpec("image", "nginx:1.21")
	// Missing executorType

	err = ValidateBlueprintAgainstSchema(invalidBlueprint, schema)
	if err == nil {
		t.Error("Expected validation to fail for missing required field 'executorType'")
	}

	// Test invalid port protocol enum
	invalidBlueprint2 := CreateBlueprint("ExecutorDeployment", "invalid2", "production")
	invalidBlueprint2.SetSpec("image", "nginx:1.21")
	invalidBlueprint2.SetSpec("executorType", "container-executor")
	invalidBlueprint2.SetSpec("ports", []interface{}{
		map[string]interface{}{
			"name":     "http",
			"port":     80,
			"protocol": "HTTP", // Invalid, should be TCP or UDP
		},
	})

	err = ValidateBlueprintAgainstSchema(invalidBlueprint2, schema)
	if err == nil {
		t.Error("Expected validation to fail for invalid enum value in nested array")
	}
}
