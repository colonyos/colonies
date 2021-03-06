package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAddAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyID := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.True(t, attribute.Equals(attributeFromDB))
}

func TestGetAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyID := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID, colonyID, core.GenerateRandomID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID, colonyID, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID, colonyID, "", core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	var allAttributes []core.Attribute
	allAttributes = append(allAttributes, attribute1)
	allAttributes = append(allAttributes, attribute2)
	allAttributes = append(allAttributes, attribute3)

	var inAttributes []core.Attribute
	inAttributes = append(inAttributes, attribute1)
	inAttributes = append(inAttributes, attribute2)

	var errAttributes []core.Attribute
	errAttributes = append(errAttributes, attribute3)

	attributesFromDB, err := db.GetAttributesByType(processID, core.IN)
	assert.Nil(t, err)
	assert.True(t, core.IsAttributeArraysEqual(inAttributes, attributesFromDB))

	attributesFromDB, err = db.GetAttributesByType(processID, core.ERR)
	assert.Nil(t, err)
	assert.True(t, core.IsAttributeArraysEqual(errAttributes, attributesFromDB))

	attributesFromDB, err = db.GetAttributesByType(processID, core.OUT)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributes(processID)
	assert.True(t, core.IsAttributeArraysEqual(allAttributes, attributesFromDB))
}

func TestUpdateAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyID := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.Equal(t, "test_value1", attributeFromDB.Value)

	attributeFromDB.SetValue("updated_test_value1")
	err = db.UpdateAttribute(attributeFromDB)
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.Equal(t, "updated_test_value1", attributeFromDB.Value)

	// Test update an attribute not added to the database
	nonExistingAttribute := core.CreateAttribute(processID, colonyID, "", core.ERR, "test_key2", "test_value2")
	err = db.UpdateAttribute(nonExistingAttribute)
	assert.NotNil(t, err)
}

func TestDeleteAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	colonyID := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID1, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyID, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID1, colonyID, "", core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(processID2, colonyID, "", core.OUT, "test_key4", "test_value4")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attribute5 := core.CreateAttribute(processID2, colonyID, "", core.ERR, "test_key5", "test_value5")
	err = db.AddAttribute(attribute5)
	assert.Nil(t, err)

	attribute6 := core.CreateAttribute(processID2, colonyID, core.GenerateRandomID(), core.ERR, "test_key6", "test_value6")
	err = db.AddAttribute(attribute6)
	assert.Nil(t, err)

	attribute7 := core.CreateAttribute(processID2, colonyID, "", core.OUT, "test_key7", "test_value7")
	err = db.AddAttribute(attribute7)
	assert.Nil(t, err)

	// Test DeleteAttributesByID

	attributeFromDB, err := db.GetAttributeByID(attribute6.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAttributeByID(attribute6.ID)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute6.ID)
	assert.NotNil(t, err)

	// Test DeleteAttributesByProcessID

	err = db.DeleteAttributesByTargetID(processID1, core.IN)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute1.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Attribute 3 should still be there since it is of type core.ERR

	// Test DeleteAllAttributesByProcessID

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAllAttributesByTargetID(processID2)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute4.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute5.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute7.ID)
	assert.NotNil(t, err)

	// Test DeleteAllAttributes

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAllAttributes()
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute3.ID)
	assert.NotNil(t, err)
}

func TestDeleteAllAttributesByProcessGraphID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(processID1, colonyID, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyID, processGraphID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID2, colonyID, processGraphID2, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)

	err = db.DeleteAllAttributesByProcessGraphID(processGraphID1)
	assert.Nil(t, err)

	attributesFromDB, err = db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)
}

func TestDeleteAllAttributesInProcesssGraphByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(processID1, colonyID, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyID, processGraphID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID2, colonyID, processGraphID2, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(processID2, colonyID, "", core.IN, "test_key3", "test_value2")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	err = db.DeleteAllAttributesInProcessGraphsByColonyID(colonyID)
	assert.Nil(t, err)

	attributesFromDB, err = db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)
}
