package server

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

// DatabaseMock implements all database interfaces for testing
type DatabaseMock struct {
	ReturnError string
	ReturnValue string
}

// Implement key database interface methods as no-ops for testing
// ColonyDatabase interface
func (db *DatabaseMock) AddColony(colony *core.Colony) error { return nil }
func (db *DatabaseMock) GetColonyByName(name string) (*core.Colony, error) { return nil, nil }
func (db *DatabaseMock) GetColonyByID(colonyID string) (*core.Colony, error) { return nil, nil }
func (db *DatabaseMock) GetColonies() ([]*core.Colony, error) { return nil, nil }
func (db *DatabaseMock) RemoveColonyByName(name string) error { return nil }
func (db *DatabaseMock) RemoveColonyByID(colonyID string) error { return nil }

// ExecutorDatabase interface  
func (db *DatabaseMock) AddExecutor(executor *core.Executor) error { return nil }
func (db *DatabaseMock) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) { return nil, nil }
func (db *DatabaseMock) GetExecutorByID(executorID string) (*core.Executor, error) { return nil, nil }
func (db *DatabaseMock) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) { return nil, nil }
func (db *DatabaseMock) RemoveExecutorByName(colonyName string, executorName string) error { return nil }
func (db *DatabaseMock) RemoveExecutorByID(executorID string) error { return nil }
func (db *DatabaseMock) RemoveExecutorsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) UpdateExecutor(executor *core.Executor) error { return nil }
func (db *DatabaseMock) SetExecutorState(executorID string, state int) error { return nil }
func (db *DatabaseMock) SetExecutorApprovalState(executorID string, state int) error { return nil }
func (db *DatabaseMock) SetExecutorCommissionTime(executorID string) error { return nil }
func (db *DatabaseMock) CountExecutors() (int, error) { return 0, nil }

// ProcessDatabase interface
func (db *DatabaseMock) AddProcess(process *core.Process) error { return nil }
func (db *DatabaseMock) GetProcessByID(processID string) (*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindAllProcesses() ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindProcessesByColonyName(colonyName string, count int, state int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindProcessesByExecutorID(executorID string, count int, state int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindWaitingProcesses(colonyName string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindRunningProcesses(colonyName string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindSuccessfulProcesses(colonyName string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindFailedProcesses(colonyName string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) RemoveProcessByID(processID string) error { return nil }
func (db *DatabaseMock) RemoveAllWaitingProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllFailedProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) SetOutput(processID string, output []interface{}) error { return nil }
func (db *DatabaseMock) MarkSuccessful(processID string, output []interface{}) error { return nil }
func (db *DatabaseMock) MarkFailed(processID string, errs []string) error { return nil }
func (db *DatabaseMock) MarkRunning(processID string, executorID string) error { return nil }
func (db *DatabaseMock) SetWaitingTime(processID string) error { return nil }
func (db *DatabaseMock) SetWaitForParents(processID string, waitForParents bool) error { return nil }
func (db *DatabaseMock) SetProcessState(processID string, state int) error { return nil }
func (db *DatabaseMock) Assign(executorID string, colonyName string, cpu int64, memory int64) (*core.Process, error) { return nil, nil }
func (db *DatabaseMock) Unassign(processID string) error { return nil }
func (db *DatabaseMock) ResetProcess(processID string) error { return nil }
func (db *DatabaseMock) CountProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountWaitingProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountRunningProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountSuccessfulProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountFailedProcesses() (int, error) { return 0, nil }

// UserDatabase interface
func (db *DatabaseMock) AddUser(user *core.User) error { return nil }
func (db *DatabaseMock) GetUserByName(colonyName string, name string) (*core.User, error) { return nil, nil }
func (db *DatabaseMock) GetUserByID(colonyName string, userID string) (*core.User, error) { return nil, nil }
func (db *DatabaseMock) GetUsersByColonyName(colonyName string) ([]*core.User, error) { return nil, nil }
func (db *DatabaseMock) RemoveUserByName(colonyName string, name string) error { return nil }
func (db *DatabaseMock) RemoveUserByID(colonyName string, userID string) error { return nil }
func (db *DatabaseMock) RemoveUsersByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountUsers() (int, error) { return 0, nil }

// Implement basic database interface methods
func (db *DatabaseMock) CreateTables() error { return nil }
func (db *DatabaseMock) DropTables() error { return nil }
func (db *DatabaseMock) Close() error { return nil }

// ValidatorMock implements the security.Validator interface for testing
type ValidatorMock struct {
	ReturnError string
}

func (v *ValidatorMock) RequireMembership(recoveredID string, colonyName string, executorMayJoin bool) error {
	if v.ReturnError != "" {
		return errors.New(v.ReturnError)
	}
	return nil
}

func (v *ValidatorMock) RequireColonyOwner(recoveredID string, colonyName string) error {
	if v.ReturnError != "" {
		return errors.New(v.ReturnError)
	}
	return nil
}

func (v *ValidatorMock) RequireExecutorMembership(recoveredID string, colonyName string, targetExecutorID string) error {
	if v.ReturnError != "" {
		return errors.New(v.ReturnError)
	}
	return nil
}

func (v *ValidatorMock) RequireServerOwner(recoveredID string, serverID string) error {
	if v.ReturnError != "" {
		return errors.New(v.ReturnError)
	}
	return nil
}

func (v *ValidatorMock) ParseSignature(payload string, signature string) (string, error) {
	if v.ReturnError != "" {
		return "", errors.New(v.ReturnError)
	}
	return "test-id", nil
}

func (v *ValidatorMock) GetExecutorDB() interface{} {
	return nil
}

func (v *ValidatorMock) GetUserDB() interface{} {
	return nil
}

func (v *ValidatorMock) GetColonyDB() interface{} {
	return nil
}