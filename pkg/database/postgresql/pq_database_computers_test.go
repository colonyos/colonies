package postgresql

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddComputer(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computerID := core.GenerateRandomID()
	computer := core.CreateComputer(computerID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer)
	assert.Nil(t, err)

	computers, err := db.GetComputers()
	assert.Nil(t, err)

	computerFromDB := computers[0]
	assert.True(t, computer.Equals(computerFromDB))
	assert.True(t, computerFromDB.IsPending())
	assert.False(t, computerFromDB.IsApproved())
	assert.False(t, computerFromDB.IsRejected())
}

func TestAddTwoComputer(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	var computers []*core.Computer
	computers = append(computers, computer1)
	computers = append(computers, computer2)

	computersFromDB, err := db.GetComputers()
	assert.Nil(t, err)
	assert.True(t, core.IsComputerArraysEqual(computers, computersFromDB))
}

func TestGetComputerByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	computerFromDB, err := db.GetComputerByID(computer1.ID)
	assert.Nil(t, err)
	assert.True(t, computer1.Equals(computerFromDB))
}

func TestGetComputerByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)
	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	assert.Nil(t, err)

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	computer3ID := core.GenerateRandomID()
	computer3 := core.CreateComputer(computer3ID, "test_computer", colony2.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer3)
	assert.Nil(t, err)

	var computersColony1 []*core.Computer
	computersColony1 = append(computersColony1, computer1)
	computersColony1 = append(computersColony1, computer2)

	computersColony1FromDB, err := db.GetComputersByColonyID(colony1.ID)
	assert.Nil(t, err)
	assert.True(t, core.IsComputerArraysEqual(computersColony1, computersColony1FromDB))
}

func TestApproveComputer(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	computerID := core.GenerateRandomID()
	computer := core.CreateComputer(computerID, "test_computer", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer)
	assert.Nil(t, err)

	assert.True(t, computer.IsPending())

	err = db.ApproveComputer(computer)
	assert.Nil(t, err)

	assert.False(t, computer.IsPending())
	assert.False(t, computer.IsRejected())
	assert.True(t, computer.IsApproved())

	computerFromDB, err := db.GetComputerByID(computer.ID)
	assert.Nil(t, err)
	assert.True(t, computerFromDB.IsApproved())

	err = db.RejectComputer(computer)
	assert.Nil(t, err)
	assert.True(t, computer.IsRejected())

	computerFromDB, err = db.GetComputerByID(computer.ID)
	assert.Nil(t, err)
	assert.True(t, computer.IsRejected())
}

func TestDeleteComputers(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	computer3ID := core.GenerateRandomID()
	computer3 := core.CreateComputer(computer3ID, "test_computer", colony2.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer3)
	assert.Nil(t, err)

	err = db.DeleteComputerByID(computer2.ID)
	assert.Nil(t, err)

	computerFromDB, err := db.GetComputerByID(computer2.ID)
	assert.Nil(t, err)
	assert.Nil(t, computerFromDB)

	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	computerFromDB, err = db.GetComputerByID(computer2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromDB)

	err = db.DeleteComputersByColonyID(colony1.ID)
	assert.Nil(t, err)

	computerFromDB, err = db.GetComputerByID(computer1.ID)
	assert.Nil(t, err)
	assert.Nil(t, computerFromDB)

	computerFromDB, err = db.GetComputerByID(computer2.ID)
	assert.Nil(t, err)
	assert.Nil(t, computerFromDB)

	computerFromDB, err = db.GetComputerByID(computer3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromDB)
}
