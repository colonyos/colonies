package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAttributeClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.IN, "test_key1", "test_value1")
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute1 := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.IN, "test_key1", "test_value1")
	attribute2 := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.OUT, "test_key2", "test_value2")
	attributes := []core.Attribute{attribute1, attribute2}
	err = db.AddAttributes(attributes)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.GetAttributesByColonyName("invalid_name")
	assert.Nil(t, err) // Returns empty slice

	_, err = db.GetAttribute(core.GenerateRandomID(), "test_key1", core.IN)
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.GetAttributes("invalid_id")
	assert.Nil(t, err) // Returns empty slice

	_, err = db.GetAttributesByType("invalid_id", 1)
	assert.Nil(t, err) // Returns empty slice

	err = db.UpdateAttribute(attribute)
	assert.Nil(t, err)

	err = db.RemoveAttributeByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveAllAttributesByColonyName("invalid_name")
	assert.Nil(t, err) // No error when nothing to remove

	err = db.RemoveAllAttributesByColonyNameWithState("invalid_name", 10)
	assert.Nil(t, err)

	err = db.RemoveAllAttributesByProcessGraphID("invalid_id")
	assert.Nil(t, err)

	err = db.RemoveAllAttributesInProcessGraphsByColonyName("invalid")
	assert.Nil(t, err)

	err = db.RemoveAllAttributesInProcessGraphsByColonyNameWithState("invalid", -1)
	assert.Nil(t, err)

	err = db.RemoveAttributesByTargetID("invalid_id", -1)
	assert.Nil(t, err)

	err = db.RemoveAllAttributesByTargetID("invalid_id")
	assert.Nil(t, err)

	err = db.RemoveAllAttributes()
	assert.Nil(t, err)
}

func TestAddAttribute(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")

	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.True(t, attribute.Equals(attributeFromDB))
}

func TestAddAttributes(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")
	attribute2 := core.CreateAttribute(processID, colonyName, "", core.OUT, "test_key2", "test_value2")
	attributes := []core.Attribute{attribute1, attribute2}

	err = db.AddAttributes(attributes)
	assert.Nil(t, err)

	// Verify both attributes were added
	attributeFromDB1, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.True(t, attribute1.Equals(attributeFromDB1))

	attributeFromDB2, err := db.GetAttribute(processID, "test_key2", core.OUT)
	assert.Nil(t, err)
	assert.True(t, attribute2.Equals(attributeFromDB2))

	// Test adding empty slice
	err = db.AddAttributes([]core.Attribute{})
	assert.Nil(t, err)
}

func TestGetAttributeByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")

	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	// Get by ID
	attributeFromDB, err := db.GetAttributeByID(attribute.ID)
	assert.Nil(t, err)
	assert.True(t, attribute.Equals(attributeFromDB))

	// Test non-existing ID
	_, err = db.GetAttributeByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestGetAttributesByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.GenerateRandomID()
	colony2 := core.GenerateRandomID()

	// Add attributes to different colonies
	attr1 := core.CreateAttribute(core.GenerateRandomID(), colony1, "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(core.GenerateRandomID(), colony1, "", core.OUT, "key2", "value2")
	attr3 := core.CreateAttribute(core.GenerateRandomID(), colony2, "", core.IN, "key3", "value3")

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)
	err = db.AddAttribute(attr3)
	assert.Nil(t, err)

	// Get attributes by colony
	attrs1, err := db.GetAttributesByColonyName(colony1)
	assert.Nil(t, err)
	assert.Len(t, attrs1, 2)

	attrs2, err := db.GetAttributesByColonyName(colony2)
	assert.Nil(t, err)
	assert.Len(t, attrs2, 1)

	// Test invalid colony
	attrsInvalid, err := db.GetAttributesByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Empty(t, attrsInvalid)
}

func TestGetAttributes(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	targetID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()

	// Add multiple attributes for same target
	attr1 := core.CreateAttribute(targetID, colonyName, "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(targetID, colonyName, "", core.OUT, "key2", "value2")
	attr3 := core.CreateAttribute(core.GenerateRandomID(), colonyName, "", core.IN, "key3", "value3") // Different target

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)
	err = db.AddAttribute(attr3)
	assert.Nil(t, err)

	// Get attributes by target ID
	attrs, err := db.GetAttributes(targetID)
	assert.Nil(t, err)
	assert.Len(t, attrs, 2)

	// Verify correct attributes returned
	keys := make([]string, len(attrs))
	for i, attr := range attrs {
		keys[i] = attr.Key
		assert.Equal(t, attr.TargetID, targetID)
	}
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")

	// Test non-existing target
	attrsEmpty, err := db.GetAttributes("non_existing_target")
	assert.Nil(t, err)
	assert.Empty(t, attrsEmpty)
}

func TestGetAttributesByType(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	targetID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()

	// Add attributes of different types
	attr1 := core.CreateAttribute(targetID, colonyName, "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(targetID, colonyName, "", core.IN, "key2", "value2")
	attr3 := core.CreateAttribute(targetID, colonyName, "", core.OUT, "key3", "value3")

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)
	err = db.AddAttribute(attr3)
	assert.Nil(t, err)

	// Get attributes by type
	inAttrs, err := db.GetAttributesByType(targetID, core.IN)
	assert.Nil(t, err)
	assert.Len(t, inAttrs, 2)

	outAttrs, err := db.GetAttributesByType(targetID, core.OUT)
	assert.Nil(t, err)
	assert.Len(t, outAttrs, 1)
	assert.Equal(t, outAttrs[0].Key, "key3")

	// Test non-existing type
	noneAttrs, err := db.GetAttributesByType(targetID, 999)
	assert.Nil(t, err)
	assert.Empty(t, noneAttrs)
}

func TestUpdateAttribute(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	targetID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(targetID, colonyName, "", core.IN, "test_key", "original_value")

	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	// Update attribute
	attribute.Value = "updated_value"
	err = db.UpdateAttribute(attribute)
	assert.Nil(t, err)

	// Verify update
	updatedAttr, err := db.GetAttribute(targetID, "test_key", core.IN)
	assert.Nil(t, err)
	assert.Equal(t, updatedAttr.Value, "updated_value")
}

func TestRemoveAttribute(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	targetID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(targetID, colonyName, "", core.IN, "test_key", "test_value")

	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	// Verify attribute exists
	_, err = db.GetAttributeByID(attribute.ID)
	assert.Nil(t, err)

	// Remove attribute
	err = db.RemoveAttributeByID(attribute.ID)
	assert.Nil(t, err)

	// Verify attribute is gone
	_, err = db.GetAttributeByID(attribute.ID)
	assert.NotNil(t, err)

	// Test removing non-existing attribute
	err = db.RemoveAttributeByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveAttributesByTarget(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	targetID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()

	// Add multiple attributes
	attr1 := core.CreateAttribute(targetID, colonyName, "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(targetID, colonyName, "", core.OUT, "key2", "value2")
	attr3 := core.CreateAttribute(targetID, colonyName, "", core.IN, "key3", "value3")

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)
	err = db.AddAttribute(attr3)
	assert.Nil(t, err)

	// Remove attributes by type
	err = db.RemoveAttributesByTargetID(targetID, core.IN)
	assert.Nil(t, err)

	// Verify IN attributes are gone, OUT attributes remain
	inAttrs, err := db.GetAttributesByType(targetID, core.IN)
	assert.Nil(t, err)
	assert.Empty(t, inAttrs)

	outAttrs, err := db.GetAttributesByType(targetID, core.OUT)
	assert.Nil(t, err)
	assert.Len(t, outAttrs, 1)

	// Remove all attributes by target ID
	err = db.RemoveAllAttributesByTargetID(targetID)
	assert.Nil(t, err)

	// Verify all attributes are gone
	allAttrs, err := db.GetAttributes(targetID)
	assert.Nil(t, err)
	assert.Empty(t, allAttrs)
}

func TestRemoveAllAttributes(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add multiple attributes
	attr1 := core.CreateAttribute(core.GenerateRandomID(), "colony1", "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(core.GenerateRandomID(), "colony2", "", core.OUT, "key2", "value2")

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)

	// Remove all attributes
	err = db.RemoveAllAttributes()
	assert.Nil(t, err)

	// Verify all attributes are gone
	attrs1, err := db.GetAttributesByColonyName("colony1")
	assert.Nil(t, err)
	assert.Empty(t, attrs1)

	attrs2, err := db.GetAttributesByColonyName("colony2")
	assert.Nil(t, err)
	assert.Empty(t, attrs2)
}

func TestRemoveAttributesByColony(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add attributes to different colonies
	attr1 := core.CreateAttribute(core.GenerateRandomID(), "colony1", "", core.IN, "key1", "value1")
	attr2 := core.CreateAttribute(core.GenerateRandomID(), "colony1", "", core.OUT, "key2", "value2")
	attr3 := core.CreateAttribute(core.GenerateRandomID(), "colony2", "", core.IN, "key3", "value3")

	err = db.AddAttribute(attr1)
	assert.Nil(t, err)
	err = db.AddAttribute(attr2)
	assert.Nil(t, err)
	err = db.AddAttribute(attr3)
	assert.Nil(t, err)

	// Remove all attributes from colony1
	err = db.RemoveAllAttributesByColonyName("colony1")
	assert.Nil(t, err)

	// Verify colony1 attributes are gone, colony2 attributes remain
	attrs1, err := db.GetAttributesByColonyName("colony1")
	assert.Nil(t, err)
	assert.Empty(t, attrs1)

	attrs2, err := db.GetAttributesByColonyName("colony2")
	assert.Nil(t, err)
	assert.Len(t, attrs2, 1)
}

func TestAttributeComplexScenarios(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test with complex attribute values
	complexValue := `{"nested": {"key": "value"}, "array": [1, 2, 3]}`
	attribute := core.CreateAttribute(core.GenerateRandomID(), "test_colony", "", core.IN, "complex_key", complexValue)

	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	// Retrieve and verify
	retrieved, err := db.GetAttributeByID(attribute.ID)
	assert.Nil(t, err)
	assert.Equal(t, retrieved.Value, complexValue)

	// Test duplicate key but different type - should be allowed
	duplicateKeyDiffType := core.CreateAttribute(attribute.TargetID, "test_colony", "", core.OUT, "complex_key", "different_value")
	err = db.AddAttribute(duplicateKeyDiffType)
	assert.Nil(t, err)

	// Both should exist
	inAttr, err := db.GetAttribute(attribute.TargetID, "complex_key", core.IN)
	assert.Nil(t, err)
	assert.Equal(t, inAttr.Value, complexValue)

	outAttr, err := db.GetAttribute(attribute.TargetID, "complex_key", core.OUT)
	assert.Nil(t, err)
	assert.Equal(t, outAttr.Value, "different_value")
}