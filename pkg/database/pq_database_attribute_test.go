package database

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
)

func TestAddAttribute(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	taskID := core.GenerateRandomID()

	attribute := core.CreateAttribute(taskID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	CheckError(t, err)

	attributeFromDB, err := db.GetAttribute(taskID, "test_key1", core.IN)
	CheckError(t, err)

	if attributeFromDB == nil {
		Fatal(t, "expected an attribute")
	}
	if attributeFromDB.ID() != attribute.ID() {
		Fatal(t, "invalid attribute id")
	}
	if attributeFromDB.TargetID() != attribute.TargetID() {
		Fatal(t, "invalid attribute target id")
	}
	if attributeFromDB.AttributeType() != attribute.AttributeType() {
		Fatal(t, "invalid attribute type")
	}
	if attributeFromDB.Key() != attribute.Key() {
		Fatal(t, "invalid attribute key")
	}
	if attributeFromDB.Value() != attribute.Value() {
		Fatal(t, "invalid attribute value")
	}
}

func TestGetAttributes(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	taskID := core.GenerateRandomID()

	attribute := core.CreateAttribute(taskID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	CheckError(t, err)

	attribute = core.CreateAttribute(taskID, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	CheckError(t, err)

	attribute = core.CreateAttribute(taskID, core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute)
	CheckError(t, err)

	attributesFromDB, err := db.GetAttributes(taskID, core.IN)
	CheckError(t, err)

	counter := 0
	for _, attributeFromDB := range attributesFromDB {
		if attributeFromDB.Key() == "test_key1" && attributeFromDB.Value() == "test_value1" {
			counter++
		}

		if attributeFromDB.Key() == "test_key2" && attributeFromDB.Value() == "test_value2" {
			counter++
		}
	}

	if counter != 2 {
		Fatal(t, "expected 2 in attributes")
	}

	attributesFromDB, err = db.GetAttributes(taskID, core.ERR)
	CheckError(t, err)

	counter = 0
	for _, attributeFromDB := range attributesFromDB {
		if attributeFromDB.Key() == "test_key3" && attributeFromDB.Value() == "test_value3" {
			counter++
		}
	}

	if counter != 1 {
		Fatal(t, "expected 1 err attribute")
	}

	attributesFromDB, err = db.GetAttributes(taskID, core.OUT)
	if len(attributesFromDB) != 0 {
		Fatal(t, "expected 0 out attributes")
	}
}

func TestUpdateAttribute(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	taskID := core.GenerateRandomID()

	attribute := core.CreateAttribute(taskID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	CheckError(t, err)

	attributeFromDB, err := db.GetAttribute(taskID, "test_key1", core.IN)
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected an attribute")
	}
	if attributeFromDB.Value() != "test_value1" {
		Fatal(t, "invalid attribute value")
	}

	attributeFromDB.SetValue("updated_test_value1")
	err = db.UpdateAttribute(attributeFromDB)
	CheckError(t, err)

	attributeFromDB, err = db.GetAttribute(taskID, "test_key1", core.IN)
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected an attribute")
	}
	if attributeFromDB.Value() != "updated_test_value1" {
		Fatal(t, "invalid updated attribute value")
	}

	// Test update an attribute not added to the database
	nonExistingAttribute := core.CreateAttribute(taskID, core.ERR, "test_key2", "test_value2")
	err = db.UpdateAttribute(nonExistingAttribute)
	if err == nil {
		Fatal(t, "updated an attribute does not exists")
	}
}

func TestDeleteAttributes(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	taskID1 := core.GenerateRandomID()
	taskID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(taskID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	CheckError(t, err)

	attribute2 := core.CreateAttribute(taskID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	CheckError(t, err)

	attribute3 := core.CreateAttribute(taskID1, core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute3)
	CheckError(t, err)

	attribute4 := core.CreateAttribute(taskID2, core.OUT, "test_key4", "test_value4")
	err = db.AddAttribute(attribute4)
	CheckError(t, err)

	attribute5 := core.CreateAttribute(taskID2, core.ERR, "test_key5", "test_value5")
	err = db.AddAttribute(attribute5)
	CheckError(t, err)

	attribute6 := core.CreateAttribute(taskID2, core.ERR, "test_key6", "test_value6")
	err = db.AddAttribute(attribute6)
	CheckError(t, err)

	attribute7 := core.CreateAttribute(taskID2, core.OUT, "test_key7", "test_value7")
	err = db.AddAttribute(attribute7)
	CheckError(t, err)

	// Test DeleteAttributesByID

	attributeFromDB, err := db.GetAttributeByID(attribute6.ID())
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected attribute to be in database")
	}

	err = db.DeleteAttributeByID(attribute6.ID())
	CheckError(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute6.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	// Test DeleteAttributesByTaskID

	err = db.DeleteAttributesByTaskID(taskID1, core.IN)
	CheckError(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute1.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute2.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	CheckError(t, err)
	if attributeFromDB == nil { // Attribute 3 should still be there since it is of type core.ERR
		Fatal(t, "expected attribute to be in database")
	}

	// Test DeleteAllAttributesByTaskID

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID())
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected attribute to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID())
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected attribute to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID())
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected attribute to be in database")
	}

	err = db.DeleteAllAttributesByTaskID(taskID2)
	CheckError(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}

	// Test DeleteAllAttributes

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	CheckError(t, err)
	if attributeFromDB == nil {
		Fatal(t, "expected attribute to be in database")
	}

	err = db.DeleteAllAttributes()
	CheckError(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID())
	CheckError(t, err)
	if attributeFromDB != nil {
		Fatal(t, "expected attribute not to be in database")
	}
}
