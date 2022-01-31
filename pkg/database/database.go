package database

import "github.com/colonyos/colonies/pkg/core"

type Database interface {
	// Colony functions ...
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	DeleteColonyByID(colonyID string) error

	// Runtime functions ...
	AddRuntime(runtime *core.Runtime) error
	GetRuntimes() ([]*core.Runtime, error)
	GetRuntimeByID(runtimeID string) (*core.Runtime, error)
	GetRuntimesByColonyID(colonyID string) ([]*core.Runtime, error)
	ApproveRuntime(runtime *core.Runtime) error
	RejectRuntime(runtime *core.Runtime) error
	DeleteRuntimeByID(runtimeID string) error
	DeleteRuntimesByColonyID(colonyID string) error

	// process functions ...
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyID string, count int) ([]*core.Process, error)
	FindSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyID string, count int) ([]*core.Process, error)
	FindUnassignedProcesses(colonyID string, runtimeID string, runtimeType string, count int) ([]*core.Process, error)
	DeleteProcessByID(processID string) error
	DeleteAllProcesses() error
	ResetProcess(process *core.Process) error
	ResetAllProcesses(process *core.Process) error
	AssignRuntime(runtimeID string, process *core.Process) error
	UnassignRuntime(process *core.Process) error
	MarkSuccessful(process *core.Process) error
	MarkFailed(process *core.Process) error
	NrOfProcesses() (int, error)
	NrOfWaitingProcesses() (int, error)
	NrOfRunningProcesses() (int, error)
	NrOfSuccessfulProcesses() (int, error)
	NrOfFailedProcesses() (int, error)
	NrWaitingProcessesForColony(colonyID string) (int, error)
	NrRunningProcessesForColony(colonyID string) (int, error)
	NrSuccessfulProcessesForColony(colonyID string) (int, error)
	NrFailedProcessesForColony(colonyID string) (int, error)

	// Attribute functions
	AddAttribute(attribute *core.Attribute) error
	AddAttributes(attribute []*core.Attribute) error
	GetAttributeByID(attributeID string) (*core.Attribute, error)
	GetAttribute(targetID string, key string, attributeType int) (*core.Attribute, error)
	GetAttributes(targetID string) ([]*core.Attribute, error)
	GetAttributesByType(targetID string, attributeType int) ([]*core.Attribute, error)
	UpdateAttribute(attribute *core.Attribute) error
	DeleteAttributeByID(attributeID string) error
	DeleteAttributesByProcessID(targetID string, attributeType int) error
	DeleteAllAttributesByProcessID(targetID string) error
	DeleteAllAttributes() error
}
