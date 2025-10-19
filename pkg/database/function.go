package database

import "github.com/colonyos/colonies/pkg/core"

type FunctionDatabase interface {
	AddFunction(function *core.Function) error
	GetFunctionByID(functionID string) (*core.Function, error)
	GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
	GetFunctionsByColonyName(colonyName string) ([]*core.Function, error)
	GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error)
	UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error
	RemoveFunctionByID(functionID string) error
	RemoveFunctionByName(colonyName string, executorName string, name string) error
	RemoveFunctionsByExecutorName(colonyName string, executorName string) error
	RemoveFunctionsByColonyName(colonyName string) error
	RemoveFunctions() error
}