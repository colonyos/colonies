package schema

import (
	"fmt"
	"reflect"
	"time"
)

// FieldType represents the type of a field
type FieldType int

const (
	StringType FieldType = iota
	IntType
	FloatType
	BoolType
	TimeType
	ArrayType
	ObjectType
)

func (f FieldType) String() string {
	switch f {
	case StringType:
		return "string"
	case IntType:
		return "int"
	case FloatType:
		return "float"
	case BoolType:
		return "bool"
	case TimeType:
		return "time"
	case ArrayType:
		return "array"
	case ObjectType:
		return "object"
	default:
		return "unknown"
	}
}

// Field defines a field in a document schema
type Field struct {
	Name     string    `json:"name"`
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
	Indexed  bool      `json:"indexed"`
	Unique   bool      `json:"unique"`
	Default  interface{} `json:"default,omitempty"`
}

// Schema defines the structure of documents in a collection
type Schema struct {
	Name    string   `json:"name"`
	Fields  []*Field `json:"fields"`
	Indexes []string `json:"indexes"`
}

// Validate checks if a document conforms to the schema
func (s *Schema) Validate(doc map[string]interface{}) error {
	// Check required fields
	for _, field := range s.Fields {
		if field.Required {
			if _, exists := doc[field.Name]; !exists {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	// Validate field types
	for fieldName, value := range doc {
		field := s.getField(fieldName)
		if field == nil {
			continue // Allow unknown fields for flexibility
		}

		if err := s.validateFieldType(field, value); err != nil {
			return fmt.Errorf("field '%s': %w", fieldName, err)
		}
	}

	return nil
}

// getField returns the field definition for a given field name
func (s *Schema) getField(name string) *Field {
	for _, field := range s.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

// validateFieldType checks if a value matches the expected field type
func (s *Schema) validateFieldType(field *Field, value interface{}) error {
	if value == nil {
		if field.Required {
			return fmt.Errorf("required field cannot be nil")
		}
		return nil
	}

	switch field.Type {
	case StringType:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case IntType:
		switch value.(type) {
		case int, int32, int64:
			// OK
		case float64:
			// JSON numbers are float64, check if it's actually an integer
			if f, ok := value.(float64); ok && f == float64(int64(f)) {
				// OK - it's a whole number
			} else {
				return fmt.Errorf("expected integer, got float with decimal places")
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case FloatType:
		switch value.(type) {
		case float32, float64, int, int32, int64:
			// OK - numbers can be converted
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case BoolType:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case TimeType:
		switch v := value.(type) {
		case time.Time:
			// OK
		case string:
			// Try to parse as RFC3339
			if _, err := time.Parse(time.RFC3339, v); err != nil {
				return fmt.Errorf("expected RFC3339 time format, got invalid string: %v", err)
			}
		default:
			return fmt.Errorf("expected time.Time or RFC3339 string, got %T", value)
		}
	case ArrayType:
		if reflect.TypeOf(value).Kind() != reflect.Slice && reflect.TypeOf(value).Kind() != reflect.Array {
			return fmt.Errorf("expected array, got %T", value)
		}
	case ObjectType:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}

	return nil
}

// ApplyDefaults applies default values to a document
func (s *Schema) ApplyDefaults(doc map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Copy existing fields
	for k, v := range doc {
		result[k] = v
	}

	// Apply defaults for missing fields
	for _, field := range s.Fields {
		if _, exists := result[field.Name]; !exists && field.Default != nil {
			result[field.Name] = field.Default
		}
	}

	return result
}

// GetIndexedFields returns all fields that should be indexed
func (s *Schema) GetIndexedFields() []string {
	var indexed []string
	for _, field := range s.Fields {
		if field.Indexed {
			indexed = append(indexed, field.Name)
		}
	}
	return indexed
}

// GetUniqueFields returns all fields that must be unique
func (s *Schema) GetUniqueFields() []string {
	var unique []string
	for _, field := range s.Fields {
		if field.Unique {
			unique = append(unique, field.Name)
		}
	}
	return unique
}

// NewSchema creates a new schema
func NewSchema(name string) *Schema {
	return &Schema{
		Name:    name,
		Fields:  make([]*Field, 0),
		Indexes: make([]string, 0),
	}
}

// AddField adds a field to the schema
func (s *Schema) AddField(name string, fieldType FieldType, required, indexed, unique bool, defaultValue interface{}) *Schema {
	field := &Field{
		Name:     name,
		Type:     fieldType,
		Required: required,
		Indexed:  indexed,
		Unique:   unique,
		Default:  defaultValue,
	}
	s.Fields = append(s.Fields, field)
	return s
}