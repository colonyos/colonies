package embedded

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func copyFunction(f *core.Function) *core.Function {
	cp := *f
	if f.Args != nil {
		cp.Args = make([]*core.FunctionArg, len(f.Args))
		for i, arg := range f.Args {
			if arg != nil {
				argCopy := *arg
				if arg.Enum != nil {
					argCopy.Enum = make([]string, len(arg.Enum))
					copy(argCopy.Enum, arg.Enum)
				}
				cp.Args[i] = &argCopy
			}
		}
	}
	return &cp
}

func (db *EmbeddedDatabase) AddFunction(function *core.Function) error {
	if _, ok := db.functions.Get(function.FunctionID); ok {
		return errors.New("Function with ID <" + function.FunctionID + "> already exists")
	}

	cp := copyFunction(function)
	if err := db.functions.Put(cp.FunctionID, cp); err != nil {
		return err
	}

	db.functionsIdx.byColony.Add(cp.FunctionID, cp.ColonyName)
	db.functionsIdx.byExecutor.Add(cp.FunctionID, cp.ColonyName+":"+cp.ExecutorName)
	db.functionsIdx.byFullName.Add(cp.FunctionID, cp.ColonyName+":"+cp.ExecutorName+":"+cp.FuncName)

	return nil
}

func (db *EmbeddedDatabase) GetFunctionByID(functionID string) (*core.Function, error) {
	f, ok := db.functions.Get(functionID)
	if !ok {
		return nil, nil
	}
	return copyFunction(f), nil
}

func (db *EmbeddedDatabase) GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) {
	ids := db.functionsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	var functions []*core.Function
	for _, id := range ids {
		if f, ok := db.functions.Get(id); ok {
			functions = append(functions, copyFunction(f))
		}
	}
	return functions, nil
}

func (db *EmbeddedDatabase) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	ids := db.functionsIdx.byColony.Lookup(colonyName)
	var functions []*core.Function
	for _, id := range ids {
		if f, ok := db.functions.Get(id); ok {
			functions = append(functions, copyFunction(f))
		}
	}
	return functions, nil
}

func (db *EmbeddedDatabase) GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error) {
	ids := db.functionsIdx.byFullName.Lookup(colonyName + ":" + executorName + ":" + name)
	if len(ids) == 0 {
		return nil, nil
	}
	f, ok := db.functions.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyFunction(f), nil
}

func (db *EmbeddedDatabase) UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error {
	ids := db.functionsIdx.byFullName.Lookup(colonyName + ":" + executorName + ":" + name)
	if len(ids) == 0 {
		return nil
	}
	f, ok := db.functions.Get(ids[0])
	if !ok {
		return nil
	}

	cp := copyFunction(f)
	cp.Counter = counter
	cp.MinWaitTime = minWaitTime
	cp.MaxWaitTime = maxWaitTime
	cp.MinExecTime = minExecTime
	cp.MaxExecTime = maxExecTime
	cp.AvgWaitTime = avgWaitTime
	cp.AvgExecTime = avgExecTime

	return db.functions.Put(cp.FunctionID, cp)
}

func (db *EmbeddedDatabase) ResetFunctionStatsByColonyName(colonyName string) error {
	ids := db.functionsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		f, ok := db.functions.Get(id)
		if !ok {
			continue
		}
		cp := copyFunction(f)
		cp.Counter = 0
		cp.MinWaitTime = 0
		cp.MaxWaitTime = 0
		cp.MinExecTime = 0
		cp.MaxExecTime = 0
		cp.AvgWaitTime = 0
		cp.AvgExecTime = 0
		if err := db.functions.Put(cp.FunctionID, cp); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveFunctionByID(functionID string) error {
	f, ok := db.functions.Get(functionID)
	if !ok {
		return nil
	}

	db.functionsIdx.byColony.Remove(functionID, f.ColonyName)
	db.functionsIdx.byExecutor.Remove(functionID, f.ColonyName+":"+f.ExecutorName)
	db.functionsIdx.byFullName.Remove(functionID, f.ColonyName+":"+f.ExecutorName+":"+f.FuncName)

	return db.functions.Delete(functionID)
}

func (db *EmbeddedDatabase) RemoveFunctionByName(colonyName string, executorName string, name string) error {
	ids := db.functionsIdx.byFullName.Lookup(colonyName + ":" + executorName + ":" + name)
	for _, id := range ids {
		if err := db.RemoveFunctionByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveFunctionsByExecutorName(colonyName string, executorName string) error {
	ids := db.functionsIdx.byExecutor.Lookup(colonyName + ":" + executorName)
	for _, id := range ids {
		if err := db.RemoveFunctionByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveFunctionsByColonyName(colonyName string) error {
	ids := db.functionsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if err := db.RemoveFunctionByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveFunctions() error {
	all := db.functions.All()
	for _, f := range all {
		if err := db.RemoveFunctionByID(f.FunctionID); err != nil {
			return err
		}
	}
	return nil
}
