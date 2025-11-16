package database

import "github.com/colonyos/colonies/pkg/core"

type ExecutorDatabase interface {
	AddExecutor(executor *core.Executor) error
	SetAllocations(colonyName string, executorName string, allocations core.Allocations) error
	GetExecutors() ([]*core.Executor, error)
	GetExecutorByID(executorID string) (*core.Executor, error)
	GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error)
	GetExecutorByName(colonyName string, executorName string) (*core.Executor, error)
	GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error)
	ApproveExecutor(executor *core.Executor) error
	RejectExecutor(executor *core.Executor) error
	MarkAlive(executor *core.Executor) error
	RemoveExecutorByName(colonyName string, executorName string) error
	RemoveExecutorsByColonyName(colonyName string) error
	CountExecutors() (int, error)
	CountExecutorsByColonyName(colonyName string) (int, error)
	CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error)
}