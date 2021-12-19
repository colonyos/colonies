package postgresql

import (
	"colonies/pkg/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddProcess(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	computer1ID := core.GenerateRandomID()
	computer2ID := core.GenerateRandomID()

	process := core.CreateProcess(colonyID, []string{computer1ID, computer2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, colonyID, processFromDB.TargetColonyID())
	assert.Contains(t, processFromDB.TargetComputerIDs(), computer1ID)
	assert.Contains(t, processFromDB.TargetComputerIDs(), computer2ID)
}

func TestDeleteProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	computer1ID := core.GenerateRandomID()
	computer2ID := core.GenerateRandomID()

	process1 := core.CreateProcess(colonyID, []string{computer1ID, computer2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := core.CreateProcess(colonyID, []string{computer1ID, computer2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := core.CreateProcess(colonyID, []string{computer1ID, computer2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	numberOfProcesses, err := db.NumberOfWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	err = db.DeleteProcessByID(process1.ID())
	assert.Nil(t, err)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfProcesses)
}

func TestDeleteAllProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	process1 := core.CreateProcess(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(process1.ID(), "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)
}

func TestDeleteProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	process1 := core.CreateProcess(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := core.CreateProcess(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(process2.ID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteProcessByID(process1.ID())
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(process1.ID(), "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttribute(process2.ID(), "test_key2", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Not deleted as it belongs to process 2
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	computer := core.CreateComputer(colony.ID(), "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer)
	assert.Nil(t, err)

	process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, processFromDB.Status())
	assert.False(t, processFromDB.Assigned())

	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.True(t, processFromDB.Assigned())
	assert.False(t, int64(processFromDB.StartTime().Sub(processFromDB.SubmissionTime())) < 0)
	assert.Equal(t, core.RUNNING, processFromDB.Status())

	err = db.UnassignComputer(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID())
	assert.Nil(t, err)
	assert.False(t, processFromDB.Assigned())
	assert.False(t, int64(processFromDB.EndTime().Sub(processFromDB.StartTime())) < 0)
}

func TestMarkSuccessful(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	computer := core.CreateComputer(colony.ID(), "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer)
	assert.Nil(t, err)

	process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.Status())

	err = db.MarkSuccessful(process)
	assert.NotNil(t, err) // Not possible to set waiting process to successfull

	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, process.Status())

	err = db.MarkSuccessful(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.SUCCESS, processFromDB.Status())

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set successful process as failed
}

func TestMarkFailed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	computer := core.CreateComputer(colony.ID(), "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer)
	assert.Nil(t, err)

	process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.Status())

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set waiting process to failed

	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, processFromDB.Status())

	err = db.MarkFailed(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.FAILED, processFromDB.Status())

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set successful process as failed
}

func TestReset(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	computer := core.CreateComputer(colony.ID(), "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer)
	assert.Nil(t, err)

	process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	process = core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	process = core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignComputer(computer.ID(), process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err := db.NumberOfFailedProcesses()
	assert.Equal(t, 3, numberOfFailedProcesses)

	err = db.ResetProcess(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Equal(t, 2, numberOfFailedProcesses)

	err = db.ResetAllProcesses(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Equal(t, 0, numberOfFailedProcesses)
}

func TestFindUnassignedProcesses1(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	process1 := core.CreateProcess(colony.ID(), []string{computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(colony.ID(), []string{computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID(), computer2.ID(), 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID(), processsFromDB[0].ID())

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID(), computer2.ID(), 2)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 2)
}

func TestFindUnassignedProcesses2(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(colony.ID(), []string{computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := core.CreateProcess(colony.ID(), []string{computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID(), computer2.ID(), 2)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 2)

	counter := 0
	for _, processFromDB := range processsFromDB {
		if processFromDB.ID() == process1.ID() {
			counter++
		}

		if processFromDB.ID() == process2.ID() {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}

func TestFindUnassignedProcesses3(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	// Here, we are testing that the order of targetComputerIDs strings does not matter

	process1 := core.CreateProcess(colony.ID(), []string{computer1.ID(), computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(colony.ID(), []string{computer1.ID(), computer2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID(), computer1.ID(), 1)
	assert.Nil(t, err)

	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID(), processsFromDB[0].ID())

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID(), computer2.ID(), 1)
	assert.Nil(t, err)

	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID(), processsFromDB[0].ID())
}

func TestFindProcessAssigned(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	process1 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	numberOfProcesses, err := db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	numberOfRunningProcesses, err := db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfRunningProcesses)

	numberOfSuccesfulProcesses, err := db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err := db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfFailedProcesses)

	processsFromDB1, err := db.FindUnassignedProcesses(colony.ID(), computer1.ID(), 1)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID(), processsFromDB1[0].ID())
	assert.Len(t, processsFromDB1, 1)

	err = db.AssignComputer(computer1.ID(), processsFromDB1[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfRunningProcesses)

	processsFromDB2, err := db.FindUnassignedProcesses(colony.ID(), computer1.ID(), 1)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID(), processsFromDB2[0].ID())

	err = db.AssignComputer(computer1.ID(), processsFromDB2[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfRunningProcesses)

	err = db.MarkSuccessful(processsFromDB1[0])
	assert.Nil(t, err)

	err = db.MarkFailed(processsFromDB2[0])
	assert.Nil(t, err)

	numberOfSuccesfulProcesses, err = db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfFailedProcesses)
}

func TestFindWaitingProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	computerID := core.GenerateRandomID()
	computer := core.CreateComputer(computerID, "test_computer", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddComputer(computer)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	waitingProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		//	err = db.AssignComputer(computerID, process)
		//assert.Nil(t, err)
		waitingProcessIDs[process.ID()] = true
	}
	waitingProcessIDsFromDB, err := db.FindWaitingProcesses(colony.ID(), 20)
	assert.Nil(t, err)

	// Create some running processes
	runningProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignComputer(computerID, process)
		assert.Nil(t, err)
		runningProcessIDs[process.ID()] = true
	}
	runningProcessIDsFromDB, err := db.FindRunningProcesses(colony.ID(), 20)
	assert.Nil(t, err)

	// Create some successful processes
	successfulProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignComputer(computerID, process)
		assert.Nil(t, err)
		err = db.MarkSuccessful(process)
		assert.Nil(t, err)
		successfulProcessIDs[process.ID()] = true
	}
	successfulProcessIDsFromDB, err := db.FindSuccessfulProcesses(colony.ID(), 20)
	assert.Nil(t, err)

	// Create some successful processes
	failedProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := core.CreateProcess(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignComputer(computerID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process)
		assert.Nil(t, err)
		failedProcessIDs[process.ID()] = true
	}
	failedProcessIDsFromDB, err := db.FindFailedProcesses(colony.ID(), 20)
	assert.Nil(t, err)

	// Now, lets to some checks
	counter := 0
	for _, processFromDB := range waitingProcessIDsFromDB {
		if waitingProcessIDs[processFromDB.ID()] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range runningProcessIDsFromDB {
		if runningProcessIDs[processFromDB.ID()] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range successfulProcessIDsFromDB {
		if successfulProcessIDs[processFromDB.ID()] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range failedProcessIDsFromDB {
		if failedProcessIDs[processFromDB.ID()] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	numberOfProcesses, err := db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 40, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

}
