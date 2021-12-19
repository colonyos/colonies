package database

import "colonies/pkg/core"

type Database interface {
	// Colony functions ...
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	DeleteColonyByID(colonyID string) error

	// Computer functions ...
	AddComputer(computer *core.Computer) error
	GetComputers() ([]*core.Computer, error)
	GetComputerByID(computerID string) (*core.Computer, error)
	GetComputersByColonyID(colonyID string) ([]*core.Computer, error)
	ApproveComputer(computer *core.Computer) error
	RejectComputer(computer *core.Computer) error
	DeleteComputerByID(computerID string) error
	DeleteComputersByColonyID(colonyID string) error

	// process functions ...
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyID string, count int) ([]*core.Process, error)
	FindSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyID string, count int) ([]*core.Process, error)
	FindUnassignedProcesses(colonyID string, computerID string, count int) ([]*core.Process, error)
	DeleteProcessByID(processID string) error
	DeleteAllProcesses() error
	ResetProcess(process *core.Process) error
	ResetAllProcesses(process *core.Process) error
	AssignComputer(computerID string, process *core.Process) error
	UnassignComputer(process *core.Process) error
	MarkSuccessful(process *core.Process) error
	MarkFailed(process *core.Process) error
	NumberOfProcesses() (int, error)
	NumberOfWaitingProcesses() (int, error)
	NumberOfRunningProcesses() (int, error)
	NumberOfSuccessfulProcesses() (int, error)
	NumberOfFailedProcesses() (int, error)

	// Attribute functions
	AddAttribute(attribute *core.Attribute) error
	GetAttributeByID(attributeID string) (*core.Attribute, error)
	GetAttribute(targetID string, key string, attributeType int) (*core.Attribute, error)
	GetAttributes(targetID string, attributeType int) ([]*core.Attribute, error)
	UpdateAttribute(attribute *core.Attribute) error
	DeleteAttributeByID(attributeID string) error
	DeleteAttributesByProcessID(targetID string, attributeType int) error
	DeleteAllAttributesByProcessID(targetID string) error
	DeleteAllAttributes() error
}
