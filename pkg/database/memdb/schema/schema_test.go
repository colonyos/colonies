package schema

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchema_Validation(t *testing.T) {
	// Create test schema
	schema := NewSchema("users").
		AddField("name", StringType, true, true, false, nil).
		AddField("age", IntType, true, true, false, 0).
		AddField("email", StringType, false, false, true, nil).
		AddField("active", BoolType, false, false, false, true).
		AddField("created_at", TimeType, false, false, false, nil)

	t.Run("ValidDocument", func(t *testing.T) {
		doc := map[string]interface{}{
			"name":       "John Doe",
			"age":        30,
			"email":      "john@example.com",
			"active":     true,
			"created_at": time.Now().Format(time.RFC3339),
		}

		err := schema.Validate(doc)
		assert.NoError(t, err)
	})

	t.Run("MissingRequiredField", func(t *testing.T) {
		doc := map[string]interface{}{
			"age": 30, // Missing required "name" field
		}

		err := schema.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field 'name' is missing")
	})

	t.Run("WrongFieldType", func(t *testing.T) {
		doc := map[string]interface{}{
			"name": "John Doe",
			"age":  "thirty", // Should be int, not string
		}

		err := schema.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected integer")
	})

	t.Run("InvalidTimeFormat", func(t *testing.T) {
		doc := map[string]interface{}{
			"name":       "John Doe",
			"age":        30,
			"created_at": "invalid-time",
		}

		err := schema.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected RFC3339 time format")
	})

	t.Run("ValidTimeObject", func(t *testing.T) {
		doc := map[string]interface{}{
			"name":       "John Doe",
			"age":        30,
			"created_at": time.Now(),
		}

		err := schema.Validate(doc)
		assert.NoError(t, err)
	})

	t.Run("JSONNumberHandling", func(t *testing.T) {
		// JSON unmarshaling turns numbers into float64
		doc := map[string]interface{}{
			"name": "John Doe",
			"age":  float64(30), // JSON number
		}

		err := schema.Validate(doc)
		assert.NoError(t, err)
	})

	t.Run("InvalidJSONFloat", func(t *testing.T) {
		doc := map[string]interface{}{
			"name": "John Doe",
			"age":  30.5, // Should be whole number
		}

		err := schema.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected integer")
	})
}

func TestSchema_ApplyDefaults(t *testing.T) {
	schema := NewSchema("config").
		AddField("name", StringType, true, false, false, nil).
		AddField("timeout", IntType, false, false, false, 30).
		AddField("enabled", BoolType, false, false, false, true).
		AddField("description", StringType, false, false, false, "Default description")

	t.Run("ApplyDefaults", func(t *testing.T) {
		doc := map[string]interface{}{
			"name": "test-config",
		}

		result := schema.ApplyDefaults(doc)
		
		assert.Equal(t, "test-config", result["name"])
		assert.Equal(t, 30, result["timeout"])
		assert.Equal(t, true, result["enabled"])
		assert.Equal(t, "Default description", result["description"])
	})

	t.Run("PreserveExistingValues", func(t *testing.T) {
		doc := map[string]interface{}{
			"name":    "test-config",
			"timeout": 60, // Override default
			"enabled": false, // Override default
		}

		result := schema.ApplyDefaults(doc)
		
		assert.Equal(t, "test-config", result["name"])
		assert.Equal(t, 60, result["timeout"]) // Preserved
		assert.Equal(t, false, result["enabled"]) // Preserved
		assert.Equal(t, "Default description", result["description"]) // Applied default
	})
}

func TestSchema_FieldTypes(t *testing.T) {
	schema := NewSchema("types_test").
		AddField("string_field", StringType, false, false, false, nil).
		AddField("int_field", IntType, false, false, false, nil).
		AddField("float_field", FloatType, false, false, false, nil).
		AddField("bool_field", BoolType, false, false, false, nil).
		AddField("time_field", TimeType, false, false, false, nil).
		AddField("array_field", ArrayType, false, false, false, nil).
		AddField("object_field", ObjectType, false, false, false, nil)

	tests := []struct {
		name    string
		doc     map[string]interface{}
		wantErr bool
	}{
		{
			name: "ValidTypes",
			doc: map[string]interface{}{
				"string_field": "test",
				"int_field":    42,
				"float_field":  3.14,
				"bool_field":   true,
				"time_field":   time.Now(),
				"array_field":  []string{"a", "b", "c"},
				"object_field": map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "WrongStringType",
			doc: map[string]interface{}{
				"string_field": 123,
			},
			wantErr: true,
		},
		{
			name: "WrongIntType",
			doc: map[string]interface{}{
				"int_field": "not_a_number",
			},
			wantErr: true,
		},
		{
			name: "WrongFloatType",
			doc: map[string]interface{}{
				"float_field": "not_a_number",
			},
			wantErr: true,
		},
		{
			name: "WrongBoolType",
			doc: map[string]interface{}{
				"bool_field": "true",
			},
			wantErr: true,
		},
		{
			name: "WrongArrayType",
			doc: map[string]interface{}{
				"array_field": "not_an_array",
			},
			wantErr: true,
		},
		{
			name: "WrongObjectType",
			doc: map[string]interface{}{
				"object_field": "not_an_object",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schema.Validate(tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchema_GetIndexedFields(t *testing.T) {
	schema := NewSchema("indexed_test").
		AddField("name", StringType, false, true, false, nil).
		AddField("age", IntType, false, false, false, nil).
		AddField("email", StringType, false, true, false, nil).
		AddField("active", BoolType, false, true, false, nil)

	indexed := schema.GetIndexedFields()
	assert.Len(t, indexed, 3)
	assert.Contains(t, indexed, "name")
	assert.Contains(t, indexed, "email")
	assert.Contains(t, indexed, "active")
	assert.NotContains(t, indexed, "age")
}

func TestSchema_GetUniqueFields(t *testing.T) {
	schema := NewSchema("unique_test").
		AddField("id", StringType, false, false, true, nil).
		AddField("email", StringType, false, false, true, nil).
		AddField("name", StringType, false, false, false, nil)

	unique := schema.GetUniqueFields()
	assert.Len(t, unique, 2)
	assert.Contains(t, unique, "id")
	assert.Contains(t, unique, "email")
	assert.NotContains(t, unique, "name")
}

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		fieldType FieldType
		expected  string
	}{
		{StringType, "string"},
		{IntType, "int"},
		{FloatType, "float"},
		{BoolType, "bool"},
		{TimeType, "time"},
		{ArrayType, "array"},
		{ObjectType, "object"},
		{FieldType(999), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.fieldType.String())
	}
}

func TestSchema_AllowUnknownFields(t *testing.T) {
	// Schema validation should allow unknown fields for flexibility
	schema := NewSchema("flexible").
		AddField("required_field", StringType, true, false, false, nil)

	doc := map[string]interface{}{
		"required_field": "value",
		"unknown_field":  "should_be_allowed",
		"another_unknown": 123,
	}

	err := schema.Validate(doc)
	assert.NoError(t, err, "Unknown fields should be allowed for flexibility")
}

func TestSchema_NilValues(t *testing.T) {
	schema := NewSchema("nil_test").
		AddField("optional_field", StringType, false, false, false, nil).
		AddField("required_field", StringType, true, false, false, nil)

	t.Run("NilOptionalField", func(t *testing.T) {
		doc := map[string]interface{}{
			"required_field": "value",
			"optional_field": nil,
		}

		err := schema.Validate(doc)
		assert.NoError(t, err)
	})

	t.Run("NilRequiredField", func(t *testing.T) {
		doc := map[string]interface{}{
			"required_field": nil,
		}

		err := schema.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field cannot be nil")
	})
}