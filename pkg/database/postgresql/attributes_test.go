package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAttributeClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.NotNil(t, err)

	attribute1 := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.IN, "test_key1", "test_value1")
	attribute2 := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.OUT, "test_key2", "test_value2")
	attributes := []core.Attribute{attribute1, attribute2}
	err = db.AddAttributes(attributes)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetAttributesByColonyName("invalid_name")
	assert.NotNil(t, err)

	_, err = db.GetAttribute(core.GenerateRandomID(), "test_key1", core.IN)
	assert.NotNil(t, err)

	_, err = db.GetAttributes("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetAttributesByType("invalid_id", 1)
	assert.NotNil(t, err)

	err = db.UpdateAttribute(attribute)
	assert.NotNil(t, err)

	err = db.RemoveAttributeByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesByColonyNameWithState("invalid_name", 10)
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesByProcessGraphID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesInProcessGraphsByColonyName("invalid")
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesInProcessGraphsByColonyNameWithState("invalid", -1)
	assert.NotNil(t, err)

	err = db.RemoveAttributesByTargetID("invalid_id", -1)
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesByTargetID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllAttributes()
	assert.NotNil(t, err)
}

func TestAddAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.True(t, attribute.Equals(attributeFromDB))
}

func TestAddAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")
	attribute2 := core.CreateAttribute(processID, colonyName, "", core.OUT, "test_key2", "test_value2")
	attributes := []core.Attribute{attribute1, attribute2}

	err = db.AddAttributes(attributes)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(processID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.True(t, attribute1.Equals(attributeFromDB))

	attributeFromDB, err = db.GetAttribute(processID, "test_key2", core.OUT)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)
	assert.True(t, attribute2.Equals(attributeFromDB))

	attributesFromDB, err := db.GetAttributesByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)
}

func TestGetAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID, colonyName, core.GenerateRandomID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID, colonyName, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID, colonyName, "", core.ERR, "test_key3", "test_value3")
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

	attributesFromDB, err := db.GetAttributesByType("invalid_id", core.IN)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributesByType("invalid_id", 20)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributesByType(processID, core.IN)
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

func TestGetAttributesByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	process1ID := core.GenerateRandomID()
	process2ID := core.GenerateRandomID()
	process3ID := core.GenerateRandomID()
	colony1Name := core.GenerateRandomID()
	colony2Name := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(process1ID, colony1Name, core.GenerateRandomID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(process1ID, colony1Name, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(process2ID, colony1Name, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(process3ID, colony2Name, "", core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributesByColonyName("invalid_name")
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributesByColonyName(colony1Name)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 3)

	attributesFromDB, err = db.GetAttributesByColonyName(colony2Name)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)
}

func TestUpdateAttribute(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key1", "test_value1")
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
	nonExistingAttribute := core.CreateAttribute(processID, colonyName, "", core.ERR, "test_key2", "test_value2")
	err = db.UpdateAttribute(nonExistingAttribute)
	assert.NotNil(t, err)
}

func TestSetAttributeState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	process1ID := core.GenerateRandomID()
	process2ID := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(process1ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(process1ID, colonyName, "", core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(process2ID, colonyName, "", core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttributeByID(attribute1.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 0)

	attributeFromDB, err = db.GetAttributeByID(attribute2.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 0)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 0)

	err = db.SetAttributeState(process1ID, core.SUCCESS)
	assert.Nil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute1.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 2)

	attributeFromDB, err = db.GetAttributeByID(attribute2.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 2)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.State, 0)
}

func TestRemoveAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	colonyName := core.GenerateRandomID()
	attribute1 := core.CreateAttribute(processID1, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyName, core.GenerateRandomID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID1, colonyName, "", core.ERR, "test_key3", "test_value3")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(processID2, colonyName, "", core.OUT, "test_key4", "test_value4")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attribute5 := core.CreateAttribute(processID2, colonyName, "", core.ERR, "test_key5", "test_value5")
	err = db.AddAttribute(attribute5)
	assert.Nil(t, err)

	attribute6 := core.CreateAttribute(processID2, colonyName, core.GenerateRandomID(), core.ERR, "test_key6", "test_value6")
	err = db.AddAttribute(attribute6)
	assert.Nil(t, err)

	attribute7 := core.CreateAttribute(processID2, colonyName, "", core.OUT, "test_key7", "test_value7")
	err = db.AddAttribute(attribute7)
	assert.Nil(t, err)

	// Test RemoveAttributesByID

	attributeFromDB, err := db.GetAttributeByID(attribute6.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.RemoveAttributeByID(attribute6.ID)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute6.ID)
	assert.NotNil(t, err)

	// Test RemoveAttributesByProcessID

	err = db.RemoveAttributesByTargetID(processID1, core.IN)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute1.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Attribute 3 should still be there since it is of type core.ERR

	// Test RemoveAllAttributesByProcessID

	attributeFromDB, err = db.GetAttributeByID(attribute4.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute5.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttributeByID(attribute7.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.RemoveAllAttributesByTargetID(processID2)
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute4.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute5.ID)
	assert.NotNil(t, err)

	_, err = db.GetAttributeByID(attribute7.ID)
	assert.NotNil(t, err)

	// Test RemoveAllAttributes

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB)

	err = db.RemoveAllAttributes()
	assert.Nil(t, err)

	_, err = db.GetAttributeByID(attribute3.ID)
	assert.NotNil(t, err)
}

func TestRemoveAttributesByColonyNameWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	process5 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process5)
	assert.Nil(t, err)

	process6 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process6.ProcessGraphID = core.GenerateRandomID() // Should not be removed
	err = db.AddProcess(process6)
	assert.Nil(t, err)

	attribute1 := core.CreateAttribute(process1.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(process2.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(process3.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(process4.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attribute5 := core.CreateAttribute(process5.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute5)
	assert.Nil(t, err)

	attribute6 := core.CreateAttribute(process6.ID, colonyName, process6.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute6)
	assert.Nil(t, err)

	err = db.SetProcessState(process1.ID, core.WAITING)
	assert.Nil(t, err)

	err = db.SetProcessState(process2.ID, core.RUNNING)
	assert.Nil(t, err)

	err = db.SetProcessState(process3.ID, core.SUCCESS)
	assert.Nil(t, err)

	err = db.SetProcessState(process4.ID, core.FAILED)
	assert.Nil(t, err)

	err = db.SetProcessState(process5.ID, core.FAILED)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttributeByID(attribute1.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB, attribute1)

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.WAITING)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute1.ID)
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.RUNNING)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.ID, attribute3.ID)

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.FAILED)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributesFromDB, err := db.GetAttributesByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2) // 1 successful process and 1 process with process graph == 2 processes

	defer db.Close()
}

func TestRemoveAttributesInProcessGraphByColonyNameWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process1.ProcessGraphID = core.GenerateRandomID()
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process2.ProcessGraphID = core.GenerateRandomID()
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process3.ProcessGraphID = core.GenerateRandomID()
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process4.ProcessGraphID = core.GenerateRandomID()
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	process5 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	process5.ProcessGraphID = core.GenerateRandomID()
	err = db.AddProcess(process5)
	assert.Nil(t, err)

	process6 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process6) // Should not be removed
	assert.Nil(t, err)

	attribute1 := core.CreateAttribute(process1.ID, colonyName, process1.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(process2.ID, colonyName, process2.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(process3.ID, colonyName, process3.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(process4.ID, colonyName, process4.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attribute5 := core.CreateAttribute(process5.ID, colonyName, process5.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute5)
	assert.Nil(t, err)

	attribute6 := core.CreateAttribute(process6.ID, colonyName, process6.ProcessGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute6)
	assert.Nil(t, err)

	err = db.SetProcessState(process1.ID, core.WAITING)
	assert.Nil(t, err)

	err = db.SetProcessState(process2.ID, core.RUNNING)
	assert.Nil(t, err)

	err = db.SetProcessState(process3.ID, core.SUCCESS)
	assert.Nil(t, err)

	err = db.SetProcessState(process4.ID, core.FAILED)
	assert.Nil(t, err)

	err = db.SetProcessState(process5.ID, core.FAILED)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttributeByID(attribute1.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB, attribute1)

	err = db.RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName, core.WAITING)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute1.ID)
	assert.NotNil(t, err)

	err = db.RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName, core.RUNNING)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributeFromDB, err = db.GetAttributeByID(attribute3.ID)
	assert.Nil(t, err)
	assert.Equal(t, attributeFromDB.ID, attribute3.ID)

	err = db.RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName, core.FAILED)
	assert.Nil(t, err)
	_, err = db.GetAttributeByID(attribute2.ID)
	assert.NotNil(t, err)

	attributesFromDB, err := db.GetAttributesByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2) // 1 running process and 1 process with no process graph == 2 processes

	defer db.Close()
}

func TestRemoveAllAttributesByProcessGraphID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(processID1, colonyName, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyName, processGraphID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID2, colonyName, processGraphID2, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)

	err = db.RemoveAllAttributesByProcessGraphID(processGraphID1)
	assert.Nil(t, err)

	attributesFromDB, err = db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)
}

func TestRemoveAllAttributesInProcesssGraphByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	processID1 := core.GenerateRandomID()
	processID2 := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()

	attribute1 := core.CreateAttribute(processID1, colonyName, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	attribute2 := core.CreateAttribute(processID1, colonyName, processGraphID1, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	attribute3 := core.CreateAttribute(processID2, colonyName, processGraphID2, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	attribute4 := core.CreateAttribute(processID2, colonyName, "", core.IN, "test_key3", "test_value2")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	attributesFromDB, err := db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 2)

	err = db.RemoveAllAttributesInProcessGraphsByColonyName(colonyName)
	assert.Nil(t, err)

	attributesFromDB, err = db.GetAttributes(processID1)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 0)

	attributesFromDB, err = db.GetAttributes(processID2)
	assert.Nil(t, err)
	assert.Len(t, attributesFromDB, 1)
}
