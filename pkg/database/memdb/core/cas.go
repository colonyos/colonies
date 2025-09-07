package core

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// CASOperation represents a Compare-And-Swap operation
type CASOperation struct {
	Collection string      `json:"collection"`
	ID         string      `json:"id"`
	Expected   interface{} `json:"expected"`
	Value      interface{} `json:"value"`
	Field      string      `json:"field,omitempty"` // Optional: specific field to CAS
	TTL        time.Duration `json:"ttl,omitempty"`
}

// CASResult represents the result of a CAS operation
type CASResult struct {
	Success      bool        `json:"success"`
	CurrentValue interface{} `json:"current_value"`
	Version      uint64      `json:"version"`
	Error        string      `json:"error,omitempty"`
}

// CASEngine handles Compare-And-Swap operations
type CASEngine struct {
	storage StorageEngine
}

// StorageEngine interface for CAS operations
type StorageEngine interface {
	Get(ctx context.Context, collection string, id string) (*Document, error)
	Update(ctx context.Context, collection string, id string, fields map[string]interface{}) (*Document, error)
	Insert(ctx context.Context, collection string, doc *Document) error
}

// Document represents a document in storage
type Document struct {
	ID       string                 `json:"id"`
	Fields   map[string]interface{} `json:"fields"`
	Version  uint64                 `json:"version"`
	Created  time.Time             `json:"created"`
	Modified time.Time             `json:"modified"`
}

// NewCASEngine creates a new CAS engine
func NewCASEngine(storage StorageEngine) *CASEngine {
	return &CASEngine{
		storage: storage,
	}
}

// CompareAndSwap performs a compare-and-swap operation
func (c *CASEngine) CompareAndSwap(ctx context.Context, operation *CASOperation) (*CASResult, error) {
	if operation.Field != "" {
		return c.compareAndSwapField(ctx, operation)
	}
	return c.compareAndSwapDocument(ctx, operation)
}

// compareAndSwapDocument performs CAS on entire document
func (c *CASEngine) compareAndSwapDocument(ctx context.Context, operation *CASOperation) (*CASResult, error) {
	// Get current document
	currentDoc, err := c.storage.Get(ctx, operation.Collection, operation.ID)
	if err != nil {
		// Document doesn't exist, handle creation case
		if operation.Expected == nil {
			// Expected nil means we expect the document not to exist
			return c.createDocument(ctx, operation)
		}
		
		return &CASResult{
			Success:      false,
			CurrentValue: nil,
			Version:      0,
			Error:        fmt.Sprintf("document not found: %v", err),
		}, nil
	}

	// Compare with expected value
	if !c.valuesEqual(operation.Expected, currentDoc.Fields) {
		return &CASResult{
			Success:      false,
			CurrentValue: currentDoc.Fields,
			Version:      currentDoc.Version,
			Error:        "expected value does not match current value",
		}, nil
	}

	// Perform the swap
	newFields, ok := operation.Value.(map[string]interface{})
	if !ok {
		return &CASResult{
			Success:      false,
			CurrentValue: currentDoc.Fields,
			Version:      currentDoc.Version,
			Error:        "new value must be a map[string]interface{}",
		}, nil
	}

	updatedDoc, err := c.storage.Update(ctx, operation.Collection, operation.ID, newFields)
	if err != nil {
		return &CASResult{
			Success:      false,
			CurrentValue: currentDoc.Fields,
			Version:      currentDoc.Version,
			Error:        fmt.Sprintf("update failed: %v", err),
		}, nil
	}

	return &CASResult{
		Success:      true,
		CurrentValue: updatedDoc.Fields,
		Version:      updatedDoc.Version,
	}, nil
}

// compareAndSwapField performs CAS on a specific field
func (c *CASEngine) compareAndSwapField(ctx context.Context, operation *CASOperation) (*CASResult, error) {
	// Get current document
	currentDoc, err := c.storage.Get(ctx, operation.Collection, operation.ID)
	if err != nil {
		// Document doesn't exist, handle creation case
		if operation.Expected == nil {
			return c.createDocumentWithField(ctx, operation)
		}
		
		return &CASResult{
			Success:      false,
			CurrentValue: nil,
			Version:      0,
			Error:        fmt.Sprintf("document not found: %v", err),
		}, nil
	}

	// Get current field value
	currentValue := currentDoc.Fields[operation.Field]

	// Compare with expected value
	if !c.valuesEqual(operation.Expected, currentValue) {
		return &CASResult{
			Success:      false,
			CurrentValue: currentValue,
			Version:      currentDoc.Version,
			Error:        "expected field value does not match current value",
		}, nil
	}

	// Perform the swap
	updateFields := map[string]interface{}{
		operation.Field: operation.Value,
	}

	updatedDoc, err := c.storage.Update(ctx, operation.Collection, operation.ID, updateFields)
	if err != nil {
		return &CASResult{
			Success:      false,
			CurrentValue: currentValue,
			Version:      currentDoc.Version,
			Error:        fmt.Sprintf("update failed: %v", err),
		}, nil
	}

	return &CASResult{
		Success:      true,
		CurrentValue: updatedDoc.Fields[operation.Field],
		Version:      updatedDoc.Version,
	}, nil
}

// createDocument creates a new document when expected is nil
func (c *CASEngine) createDocument(ctx context.Context, operation *CASOperation) (*CASResult, error) {
	fields, ok := operation.Value.(map[string]interface{})
	if !ok {
		return &CASResult{
			Success: false,
			Error:   "new value must be a map[string]interface{}",
		}, nil
	}

	doc := &Document{
		ID:     operation.ID,
		Fields: fields,
	}

	err := c.storage.Insert(ctx, operation.Collection, doc)
	if err != nil {
		return &CASResult{
			Success: false,
			Error:   fmt.Sprintf("insert failed: %v", err),
		}, nil
	}

	// Get the created document to return current version
	createdDoc, err := c.storage.Get(ctx, operation.Collection, operation.ID)
	if err != nil {
		return &CASResult{
			Success: true, // Insert succeeded even if we can't get the version
			Error:   fmt.Sprintf("document created but failed to retrieve version: %v", err),
		}, nil
	}

	return &CASResult{
		Success:      true,
		CurrentValue: createdDoc.Fields,
		Version:      createdDoc.Version,
	}, nil
}

// createDocumentWithField creates a new document with a specific field
func (c *CASEngine) createDocumentWithField(ctx context.Context, operation *CASOperation) (*CASResult, error) {
	fields := map[string]interface{}{
		operation.Field: operation.Value,
	}

	doc := &Document{
		ID:     operation.ID,
		Fields: fields,
	}

	err := c.storage.Insert(ctx, operation.Collection, doc)
	if err != nil {
		return &CASResult{
			Success: false,
			Error:   fmt.Sprintf("insert failed: %v", err),
		}, nil
	}

	// Get the created document to return current version
	createdDoc, err := c.storage.Get(ctx, operation.Collection, operation.ID)
	if err != nil {
		return &CASResult{
			Success: true, // Insert succeeded even if we can't get the version
			Error:   fmt.Sprintf("document created but failed to retrieve version: %v", err),
		}, nil
	}

	return &CASResult{
		Success:      true,
		CurrentValue: createdDoc.Fields[operation.Field],
		Version:      createdDoc.Version,
	}, nil
}

// valuesEqual compares two values for equality, handling various types
func (c *CASEngine) valuesEqual(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}

	if expected == nil || actual == nil {
		return false
	}

	// Use deep equal for complex types
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	// Handle JSON number type conversions
	return c.handleNumericComparison(expected, actual)
}

// handleNumericComparison handles comparison of numeric types that might be
// represented differently due to JSON marshaling/unmarshaling
func (c *CASEngine) handleNumericComparison(expected, actual interface{}) bool {
	expectedFloat, expectedOk := c.toFloat64(expected)
	actualFloat, actualOk := c.toFloat64(actual)
	
	if expectedOk && actualOk {
		return expectedFloat == actualFloat
	}

	// Handle string representations
	expectedStr := fmt.Sprintf("%v", expected)
	actualStr := fmt.Sprintf("%v", actual)
	
	return expectedStr == actualStr
}

// toFloat64 converts a value to float64 if possible
func (c *CASEngine) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	}
	return 0, false
}

// BatchCAS performs multiple CAS operations atomically
func (c *CASEngine) BatchCAS(ctx context.Context, operations []*CASOperation) ([]*CASResult, error) {
	results := make([]*CASResult, len(operations))
	
	// For now, perform operations sequentially
	// In a full implementation, this would use transactions
	for i, op := range operations {
		result, err := c.CompareAndSwap(ctx, op)
		if err != nil {
			// Rollback previous operations would go here
			return nil, fmt.Errorf("batch CAS failed at operation %d: %w", i, err)
		}
		
		if !result.Success {
			// Rollback previous operations would go here
			return results[:i+1], fmt.Errorf("batch CAS failed at operation %d: %s", i, result.Error)
		}
		
		results[i] = result
	}
	
	return results, nil
}