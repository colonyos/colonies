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
	if attributeFromDB.TaskID() != attribute.TaskID() {
		Fatal(t, "invalid attribute task id")
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
