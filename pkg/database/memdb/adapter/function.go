package adapter

import (
	"context"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// FunctionDatabase interface implementation

func (a *ColonyOSAdapter) AddFunction(function *core.Function) error {
	doc := &memdb.VelocityDocument{
		ID:     function.FunctionID,
		Fields: a.functionToFields(function),
	}
	
	return a.db.Insert(context.Background(), FunctionsCollection, doc)
}

func (a *ColonyOSAdapter) GetFunctions(colonyName string) ([]*core.Function, error) {
	return a.GetFunctionsByColonyName(colonyName)
}

func (a *ColonyOSAdapter) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	result, err := a.db.List(context.Background(), FunctionsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var functions []*core.Function
	for _, doc := range result {
		function, err := a.fieldsToFunction(doc.Fields)
		if err == nil && a.getFunctionColonyName(function) == colonyName {
			functions = append(functions, function)
		}
	}
	
	return functions, nil
}

func (a *ColonyOSAdapter) GetFunctionByID(functionID string) (*core.Function, error) {
	doc, err := a.db.Get(context.Background(), FunctionsCollection, functionID)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToFunction(doc.Fields)
}

func (a *ColonyOSAdapter) GetFunctionsByExecutorName(colonyName, executorName string) ([]*core.Function, error) {
	functions, err := a.GetFunctionsByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Function
	for _, function := range functions {
		if function.ExecutorName == executorName {
			filtered = append(filtered, function)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) GetFunctionsByExecutorNames(colonyName string, executorNames []string) ([]*core.Function, error) {
	functions, err := a.GetFunctionsByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	nameSet := make(map[string]bool)
	for _, name := range executorNames {
		nameSet[name] = true
	}
	
	var filtered []*core.Function
	for _, function := range functions {
		if nameSet[function.ExecutorName] {
			filtered = append(filtered, function)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error) {
	functions, err := a.GetFunctionsByExecutorName(colonyName, executorName)
	if err != nil {
		return nil, err
	}
	
	for _, function := range functions {
		if function.FuncName == name {
			return function, nil
		}
	}
	
	return nil, fmt.Errorf("function not found")
}

func (a *ColonyOSAdapter) UpdateFunction(function *core.Function) error {
	fields := a.functionToFields(function)
	_, err := a.db.Update(context.Background(), FunctionsCollection, function.FunctionID, fields)
	return err
}

func (a *ColonyOSAdapter) UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error {
	function, err := a.GetFunctionsByExecutorAndName(colonyName, executorName, name)
	if err != nil {
		return err
	}
	
	fields := map[string]interface{}{
		"counter":       counter,
		"min_wait_time": minWaitTime,
		"max_wait_time": maxWaitTime,
		"min_exec_time": minExecTime,
		"max_exec_time": maxExecTime,
		"avg_wait_time": avgWaitTime,
		"avg_exec_time": avgExecTime,
	}
	
	_, err = a.db.Update(context.Background(), FunctionsCollection, function.FunctionID, fields)
	return err
}

func (a *ColonyOSAdapter) RemoveFunctionByID(functionID string) error {
	return a.db.Delete(context.Background(), FunctionsCollection, functionID)
}

func (a *ColonyOSAdapter) RemoveFunctionByName(colonyName string, executorName string, name string) error {
	function, err := a.GetFunctionsByExecutorAndName(colonyName, executorName, name)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), FunctionsCollection, function.FunctionID)
}

func (a *ColonyOSAdapter) RemoveFunctionsByColonyName(colonyName string) error {
	functions, err := a.GetFunctionsByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, function := range functions {
		if err := a.db.Delete(context.Background(), FunctionsCollection, function.FunctionID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveFunctionsByExecutorName(colonyName, executorName string) error {
	functions, err := a.GetFunctionsByExecutorName(colonyName, executorName)
	if err != nil {
		return err
	}
	
	for _, function := range functions {
		if err := a.db.Delete(context.Background(), FunctionsCollection, function.FunctionID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveFunctions() error {
	result, err := a.db.List(context.Background(), FunctionsCollection, 10000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		if err := a.db.Delete(context.Background(), FunctionsCollection, doc.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) CountFunctionsByColonyName(colonyName string) (int, error) {
	functions, err := a.GetFunctionsByColonyName(colonyName)
	if err != nil {
		return 0, err
	}
	
	return len(functions), nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) functionToFields(function *core.Function) map[string]interface{} {
	return map[string]interface{}{
		"function_id":   function.FunctionID,
		"executor_name": function.ExecutorName,
		"executor_type": function.ExecutorType,
		"colony_name":   function.ColonyName,
		"func_name":     function.FuncName,
		"counter":       function.Counter,
		"min_wait_time": function.MinWaitTime,
		"max_wait_time": function.MaxWaitTime,
		"min_exec_time": function.MinExecTime,
		"max_exec_time": function.MaxExecTime,
		"avg_wait_time": function.AvgWaitTime,
		"avg_exec_time": function.AvgExecTime,
	}
}

func (a *ColonyOSAdapter) fieldsToFunction(fields map[string]interface{}) (*core.Function, error) {
	function := &core.Function{}
	
	if functionID, ok := fields["function_id"].(string); ok {
		function.FunctionID = functionID
	}
	if executorName, ok := fields["executor_name"].(string); ok {
		function.ExecutorName = executorName
	}
	if executorType, ok := fields["executor_type"].(string); ok {
		function.ExecutorType = executorType
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		function.ColonyName = colonyName
	}
	if funcName, ok := fields["func_name"].(string); ok {
		function.FuncName = funcName
	}
	if counter, ok := fields["counter"].(int); ok {
		function.Counter = counter
	} else if counter, ok := fields["counter"].(float64); ok {
		function.Counter = int(counter)
	}
	if minWaitTime, ok := fields["min_wait_time"].(float64); ok {
		function.MinWaitTime = minWaitTime
	}
	if maxWaitTime, ok := fields["max_wait_time"].(float64); ok {
		function.MaxWaitTime = maxWaitTime
	}
	if minExecTime, ok := fields["min_exec_time"].(float64); ok {
		function.MinExecTime = minExecTime
	}
	if maxExecTime, ok := fields["max_exec_time"].(float64); ok {
		function.MaxExecTime = maxExecTime
	}
	if avgWaitTime, ok := fields["avg_wait_time"].(float64); ok {
		function.AvgWaitTime = avgWaitTime
	}
	if avgExecTime, ok := fields["avg_exec_time"].(float64); ok {
		function.AvgExecTime = avgExecTime
	}
	
	return function, nil
}

func (a *ColonyOSAdapter) getFunctionColonyName(function *core.Function) string {
	return function.ColonyName
}