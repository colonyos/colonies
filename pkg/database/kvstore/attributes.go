package kvstore

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// AttributeDatabase Interface Implementation
// =====================================

// AddAttribute adds an attribute to the database
func (db *KVStoreDatabase) AddAttribute(attribute core.Attribute) error {
	// Store attribute at /attributes/{attributeID}
	attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
	
	// Check if attribute already exists
	if db.store.Exists(attributePath) {
		return fmt.Errorf("attribute with ID %s already exists", attribute.ID)
	}

	err := db.store.Put(attributePath, attribute)
	if err != nil {
		return fmt.Errorf("failed to add attribute %s: %w", attribute.ID, err)
	}

	return nil
}

// AddAttributes adds multiple attributes to the database
func (db *KVStoreDatabase) AddAttributes(attributes []core.Attribute) error {
	for _, attribute := range attributes {
		attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
		
		// Check if attribute already exists
		if db.store.Exists(attributePath) {
			return fmt.Errorf("attribute with ID %s already exists", attribute.ID)
		}

		err := db.store.Put(attributePath, attribute)
		if err != nil {
			return fmt.Errorf("failed to add attribute %s: %w", attribute.ID, err)
		}
	}

	return nil
}

// GetAttributeByID retrieves an attribute by ID
func (db *KVStoreDatabase) GetAttributeByID(attributeID string) (core.Attribute, error) {
	attributePath := fmt.Sprintf("/attributes/%s", attributeID)
	
	if !db.store.Exists(attributePath) {
		return core.Attribute{}, fmt.Errorf("attribute with ID %s not found", attributeID)
	}

	attributeInterface, err := db.store.Get(attributePath)
	if err != nil {
		return core.Attribute{}, fmt.Errorf("failed to get attribute %s: %w", attributeID, err)
	}

	attribute, ok := attributeInterface.(core.Attribute)
	if !ok {
		return core.Attribute{}, fmt.Errorf("stored object is not an attribute")
	}

	return attribute, nil
}

// GetAttributesByColonyName retrieves all attributes for a colony
func (db *KVStoreDatabase) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	// Search for attributes by colony name - use the correct struct field name
	attributes, err := db.store.FindRecursive("/attributes", "targetcolonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find attributes for colony %s: %w", colonyName, err)
	}

	var result []core.Attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			result = append(result, attribute)
		}
	}

	return result, nil
}

// GetAttribute retrieves an attribute by target ID, key, and type
func (db *KVStoreDatabase) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	// Search for attributes by target ID
	attributes, err := db.store.FindRecursive("/attributes", "targetid", targetID)
	if err != nil {
		return core.Attribute{}, fmt.Errorf("failed to find attributes for target %s: %w", targetID, err)
	}

	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			if attribute.Key == key && attribute.AttributeType == attributeType {
				return attribute, nil
			}
		}
	}

	return core.Attribute{}, fmt.Errorf("attribute with targetID %s, key %s, type %d not found", targetID, key, attributeType)
}

// GetAttributes retrieves all attributes for a target ID
func (db *KVStoreDatabase) GetAttributes(targetID string) ([]core.Attribute, error) {
	// Search for attributes by target ID
	attributes, err := db.store.FindRecursive("/attributes", "targetid", targetID)
	if err != nil {
		// Return empty slice when no attributes found, like PostgreSQL
		return []core.Attribute{}, nil
	}

	var result []core.Attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			result = append(result, attribute)
		}
	}

	return result, nil
}

// GetAttributesByType retrieves attributes by target ID and type
func (db *KVStoreDatabase) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	// Search for attributes by target ID
	attributes, err := db.store.FindRecursive("/attributes", "targetid", targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find attributes for target %s: %w", targetID, err)
	}

	var result []core.Attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok && attribute.AttributeType == attributeType {
			result = append(result, attribute)
		}
	}

	return result, nil
}

// UpdateAttribute updates an existing attribute
func (db *KVStoreDatabase) UpdateAttribute(attribute core.Attribute) error {
	attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
	
	if !db.store.Exists(attributePath) {
		return fmt.Errorf("attribute with ID %s not found", attribute.ID)
	}

	err := db.store.Put(attributePath, attribute)
	if err != nil {
		return fmt.Errorf("failed to update attribute %s: %w", attribute.ID, err)
	}

	return nil
}

// RemoveAttributeByID removes an attribute by ID
func (db *KVStoreDatabase) RemoveAttributeByID(attributeID string) error {
	attributePath := fmt.Sprintf("/attributes/%s", attributeID)
	
	if !db.store.Exists(attributePath) {
		return fmt.Errorf("attribute with ID %s not found", attributeID)
	}

	err := db.store.Delete(attributePath)
	if err != nil {
		return fmt.Errorf("failed to remove attribute %s: %w", attributeID, err)
	}

	return nil
}

// RemoveAllAttributesByColonyName removes all attributes for a colony
func (db *KVStoreDatabase) RemoveAllAttributesByColonyName(colonyName string) error {
	// Find all attributes for the colony
	attributes, err := db.store.FindRecursive("/attributes", "targetcolonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find attributes for colony %s: %w", colonyName, err)
	}

	// Remove each attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
			err := db.store.Delete(attributePath)
			if err != nil {
				return fmt.Errorf("failed to remove attribute %s: %w", attribute.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllAttributesByColonyNameWithState removes all attributes for a colony with a specific state
func (db *KVStoreDatabase) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	// Find all attributes for the colony
	attributes, err := db.store.FindRecursive("/attributes", "targetcolonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find attributes for colony %s: %w", colonyName, err)
	}

	// Remove attributes with matching state
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok && attribute.State == state {
			attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
			err := db.store.Delete(attributePath)
			if err != nil {
				return fmt.Errorf("failed to remove attribute %s: %w", attribute.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllAttributesByProcessGraphID removes all attributes for a process graph
func (db *KVStoreDatabase) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	// Find all attributes for the process graph
	attributes, err := db.store.FindRecursive("/attributes", "processgraphid", processGraphID)
	if err != nil {
		return fmt.Errorf("failed to find attributes for process graph %s: %w", processGraphID, err)
	}

	// Remove each attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
			err := db.store.Delete(attributePath)
			if err != nil {
				return fmt.Errorf("failed to remove attribute %s: %w", attribute.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllAttributesInProcessGraphsByColonyName removes all attributes in process graphs for a colony
func (db *KVStoreDatabase) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	// This would require finding all process graphs for the colony first
	// For now, implement as finding attributes by colony name
	return db.RemoveAllAttributesByColonyName(colonyName)
}

// RemoveAllAttributesInProcessGraphsByColonyNameWithState removes all attributes in process graphs for a colony with state
func (db *KVStoreDatabase) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	// This would require finding all process graphs for the colony first
	// For now, implement as finding attributes by colony name and state
	return db.RemoveAllAttributesByColonyNameWithState(colonyName, state)
}

// RemoveAttributesByTargetID removes attributes by target ID and type
func (db *KVStoreDatabase) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	// Find all attributes for the target
	attributes, err := db.store.FindRecursive("/attributes", "targetid", targetID)
	if err != nil {
		return fmt.Errorf("failed to find attributes for target %s: %w", targetID, err)
	}

	// Remove attributes with matching type
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok && attribute.AttributeType == attributeType {
			attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
			err := db.store.Delete(attributePath)
			if err != nil {
				return fmt.Errorf("failed to remove attribute %s: %w", attribute.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllAttributesByTargetID removes all attributes for a target ID
func (db *KVStoreDatabase) RemoveAllAttributesByTargetID(targetID string) error {
	// Find all attributes for the target
	attributes, err := db.store.FindRecursive("/attributes", "targetid", targetID)
	if err != nil {
		return fmt.Errorf("failed to find attributes for target %s: %w", targetID, err)
	}

	// Remove each attribute
	for _, searchResult := range attributes {
		if attribute, ok := searchResult.Value.(core.Attribute); ok {
			attributePath := fmt.Sprintf("/attributes/%s", attribute.ID)
			err := db.store.Delete(attributePath)
			if err != nil {
				return fmt.Errorf("failed to remove attribute %s: %w", attribute.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllAttributes removes all attributes from the database
func (db *KVStoreDatabase) RemoveAllAttributes() error {
	attributesPath := "/attributes"
	
	if !db.store.Exists(attributesPath) {
		return nil // No attributes to remove
	}

	err := db.store.Delete(attributesPath)
	if err != nil {
		return fmt.Errorf("failed to remove all attributes: %w", err)
	}

	// Recreate the attributes structure
	err = db.store.CreateArray("/attributes")
	if err != nil {
		return fmt.Errorf("failed to recreate attributes structure: %w", err)
	}

	return nil
}