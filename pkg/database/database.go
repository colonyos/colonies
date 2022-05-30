package database

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

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
	MarkAlive(runtime *core.Runtime) error
	DeleteRuntimeByID(runtimeID string) error
	DeleteRuntimesByColonyID(colonyID string) error

	// Process functions ...
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindProcessesForColony(colonyID string, seconds int, state int) ([]*core.Process, error)
	FindProcessesForRuntime(colonyID string, runtimeID string, seconds int, state int) ([]*core.Process, error)
	FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyID string, count int) ([]*core.Process, error)
	FindAllRunningProcesses() ([]*core.Process, error)
	FindSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyID string, count int) ([]*core.Process, error)
	FindUnassignedProcesses(colonyID string, runtimeID string, runtimeType string, count int) ([]*core.Process, error)
	DeleteProcessByID(processID string) error
	DeleteAllProcesses() error
	DeleteAllProcessesForColony(colonyID string) error
	ResetProcess(process *core.Process) error
	SetProcessState(processID string, state int) error
	SetRuntimeGroup(processID string, runtimeGroup string) error
	SetWaitForParents(processID string, waitingForParent bool) error
	SetDeadline(process *core.Process, deadline time.Time) error
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
	DeleteAllAttributesByColonyID(colonyID string) error
	DeleteAttributesByTargetID(targetID string, attributeType int) error
	DeleteAllAttributesByTargetID(targetID string) error
	DeleteAllAttributes() error

	// ProcessGraph functions
	AddProcessGraph(processGraph *core.ProcessGraph) error
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	SetProcessGraphState(processGraphID string, state int) error
	FindWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	NrOfWaitingProcessGraphs() (int, error)
	NrOfRunningProcessGraphs() (int, error)
	NrOfSuccessfulProcessGraphs() (int, error)
	NrOfFailedProcessGraphs() (int, error)
	NrOfWaitingProcessGraphsForColony(colonyID string) (int, error)
	NrOfRunningProcessGraphsForColony(colonyID string) (int, error)
	NrOfSuccessfulProcessGraphsForColony(colonyID string) (int, error)
	NrOfFailedProcessGraphsForColony(colonyID string) (int, error)

	// TODO: Implement support deleting process graphs
}
