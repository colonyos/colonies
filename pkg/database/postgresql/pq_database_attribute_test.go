package postgresql

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	processID := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)

	assert.NotNil(t, attributeFromDB)
	assert.Equal(t, attribute.ID(), attributeFromDB.ID())
	assert.Equal(t, attribute.TargetID(), attributeFromDB.TargetID())
	assert.Equal(t, attribute.AttributeType(), attributeFromDB.AttributeType())
	assert.Equal(t, attribute.Key(), attributeFromDB.Key())
	assert.Equal(t, attribute.Value(), attributeFromDB.Value())
}

func TestGetAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	processID := core.GenerateRandomID()

	attribute := core.CreateAttribute(processID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(processID, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(processID, core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributes(processID, core.IN)
	assert.Nil(t, err)

	counter := 0
	for _, attributeFromDB := range attributesFromDB {
		if attributeFromDB.Key() == "test_key1" && attributeFromDB.Value() == "test_value1" {
			counter++
		}

		if attributeFromDB.Key() == "test_key2" && attributeFromDB.Value() == "test_value2" {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	attributesFromDB, err = db.GetAttributes(processID, core.ERR)
	assert.Nil(t, err)

	counter = 0
	for _, attributeFromDB := range attributesFromDB {
		if attributeFromDB.Key() == "test_key3" && attributeFromDB.Value() == "test_value3" {
			counter++
		}
	}
	assert.Equal(t, 1, counter)

	attributesFromDB, err = db.GetAttributes(processID, core.OUT)
	assert.Len(t, attributesFromDB, 0)
}

func TestUpdateAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	processID := core.GenerateRandomID()

	attribute := core.CreateAttribute(processID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.Equal(t, "test_value1", attributeFromDB.Value())

	attributeFromDB.SetValue("updated_test_value1")
	err = db.UpdateAttribute(attributeFromDB)
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.Equal(t, "updated_test_value1", attributeFromDB.Value())

	// Test update an attribute not added to the database
	nonExistingAttribute := core.CreateAttribute(processID, core.ERR, "test_key2", "test_value2")
	err = db.UpdateAttribute(nonExistingAttribute)
	assert.NotNil(t, err)
}

func TestDeleteAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(processID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID1, core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(processID2, core.OUT, "test_key4", "test_value4")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attribute5 := core.CreateAttribute(processID2, core.ERR, "test_key5", "test_value5")
	err = db.AddAttribute(attribute5)
	assert.Nil(t, err)

	attribute6 := core.CreateAttribute(processID2, core.ERR, "test_key6", "test_value6")
	err = db.AddAttribute(attribute6)
	assert.Nil(t, err)

	attribute7 := core.CreateAttribute(processID2, core.OUT, "test_key7", "test_value7")
	err = db.AddAttribute(attribute7)
	assert.Nil(t, err)

	// Test DeleteAttributesByID

	attributeFromDB, err := db.GetAttributeByID(attribute6.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAttributeByID(attribute6.ID())
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute6.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	// Test DeleteAttributesByProcessID

	err = db.DeleteAttributesByProcessID(processID1, core.IN)
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute1.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute2.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Attribute 3 should still be there since it is of type core.ERR

	// Test DeleteAllAttributesByProcessID

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAllAttributesByProcessID(processID2)
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	// Test DeleteAllAttributes

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.DeleteAllAttributes()
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)
}
