package embedded

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/embedded/index"
)

// copyExecutor creates a deep copy of an Executor.
func copyExecutor(e *core.Executor) *core.Executor {
	cp := *e
	if e.Capabilities.Hardware != nil {
		cp.Capabilities.Hardware = make([]core.Hardware, len(e.Capabilities.Hardware))
		for i, hw := range e.Capabilities.Hardware {
			cp.Capabilities.Hardware[i] = hw
			if hw.Network != nil {
				cp.Capabilities.Hardware[i].Network = make([]string, len(hw.Network))
				copy(cp.Capabilities.Hardware[i].Network, hw.Network)
			}
		}
	}
	if e.Capabilities.Software != nil {
		cp.Capabilities.Software = make([]core.Software, len(e.Capabilities.Software))
		copy(cp.Capabilities.Software, e.Capabilities.Software)
	}
	if e.Allocations.Projects != nil {
		cp.Allocations.Projects = make(map[string]core.Project, len(e.Allocations.Projects))
		for k, v := range e.Allocations.Projects {
			cp.Allocations.Projects[k] = v
		}
	}
	return &cp
}

func (db *EmbeddedDatabase) AddExecutor(executor *core.Executor) error {
	if executor == nil {
		return errors.New("Executor is nil")
	}

	existingExecutor, err := db.GetExecutorByName(executor.ColonyName, executor.Name)
	if err != nil {
		return err
	}

	if existingExecutor != nil {
		if existingExecutor.State == core.UNREGISTERED {
			// Reactivate the executor
			db.executorsIdx.byColony.Remove(existingExecutor.ID, existingExecutor.ColonyName)
			db.executorsIdx.byName.Remove(existingExecutor.ID, existingExecutor.ColonyName+":"+existingExecutor.Name)
			if existingExecutor.BlueprintID != "" {
				db.executorsIdx.byBlueprint.Remove(existingExecutor.ID, existingExecutor.BlueprintID)
			}
			if err := db.executors.Delete(existingExecutor.ID); err != nil {
				return err
			}

			executor.State = core.PENDING
			executor.CommissionTime = time.Now()
			cp := copyExecutor(executor)
			if err := db.executors.Put(cp.ID, cp); err != nil {
				return err
			}

			db.executorsIdx.byColony.Add(cp.ID, cp.ColonyName)
			db.executorsIdx.byName.Add(cp.ID, cp.ColonyName+":"+cp.Name)
			if cp.BlueprintID != "" {
				db.executorsIdx.byBlueprint.Add(cp.ID, cp.BlueprintID)
			}
			return nil
		}
		return errors.New("Executor with name <" + executor.Name + "> already exists in Colony with name <" + executor.ColonyName + ">")
	}

	executor.CommissionTime = time.Now()
	cp := copyExecutor(executor)
	if err := db.executors.Put(cp.ID, cp); err != nil {
		return err
	}

	db.executorsIdx.byColony.Add(cp.ID, cp.ColonyName)
	db.executorsIdx.byName.Add(cp.ID, cp.ColonyName+":"+cp.Name)
	if cp.BlueprintID != "" {
		db.executorsIdx.byBlueprint.Add(cp.ID, cp.BlueprintID)
	}

	return nil
}

func (db *EmbeddedDatabase) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	e, err := db.getExecutorByNameInternal(colonyName, executorName)
	if err != nil {
		return err
	}
	if e == nil {
		return errors.New("Executor with name <" + executorName + "> does not exists in Colony with name <" + colonyName + ">")
	}

	cp := *e
	cp.Allocations = allocations
	return db.executors.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) GetExecutors() ([]*core.Executor, error) {
	stored := db.executors.Filter(func(e *core.Executor) bool {
		return e.State != core.UNREGISTERED
	})
	result := make([]*core.Executor, len(stored))
	for i, e := range stored {
		result[i] = copyExecutor(e)
	}
	return result, nil
}

func (db *EmbeddedDatabase) GetExecutorByID(executorID string) (*core.Executor, error) {
	e, ok := db.executors.Get(executorID)
	if !ok {
		return nil, nil
	}
	return copyExecutor(e), nil
}

func (db *EmbeddedDatabase) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	ids := db.executorsIdx.byColony.Lookup(colonyName)
	var executors []*core.Executor
	for _, id := range ids {
		if e, ok := db.executors.Get(id); ok {
			executors = append(executors, copyExecutor(e))
		}
	}
	return executors, nil
}

// getExecutorByNameInternal returns the stored pointer (no copy) for internal use.
func (db *EmbeddedDatabase) getExecutorByNameInternal(colonyName string, executorName string) (*core.Executor, error) {
	ids := db.executorsIdx.byName.Lookup(colonyName + ":" + executorName)
	if len(ids) == 0 {
		return nil, nil
	}
	e, ok := db.executors.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (db *EmbeddedDatabase) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	e, err := db.getExecutorByNameInternal(colonyName, executorName)
	if err != nil || e == nil {
		return nil, err
	}
	return copyExecutor(e), nil
}

func (db *EmbeddedDatabase) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	ids := db.executorsIdx.byBlueprint.Lookup(blueprintID)
	var executors []*core.Executor
	for _, id := range ids {
		if e, ok := db.executors.Get(id); ok {
			executors = append(executors, copyExecutor(e))
		}
	}
	return executors, nil
}

func (db *EmbeddedDatabase) ApproveExecutor(executor *core.Executor) error {
	e, ok := db.executors.Get(executor.ID)
	if !ok {
		return nil
	}

	cp := *e
	cp.State = core.APPROVED
	executor.Approve()
	return db.executors.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) RejectExecutor(executor *core.Executor) error {
	e, ok := db.executors.Get(executor.ID)
	if !ok {
		return nil
	}

	cp := *e
	cp.State = core.REJECTED
	executor.Reject()
	return db.executors.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) MarkAlive(executor *core.Executor) error {
	e, ok := db.executors.Get(executor.ID)
	if !ok {
		return nil
	}

	cp := *e
	cp.LastHeardFromTime = time.Now()
	return db.executors.Put(cp.ID, &cp)
}

func (db *EmbeddedDatabase) RemoveExecutorByName(colonyName string, executorName string) error {
	e, err := db.getExecutorByNameInternal(colonyName, executorName)
	if err != nil {
		return err
	}
	if e == nil {
		return errors.New("Executor <" + executorName + "> does not exists")
	}

	// Mark as UNREGISTERED instead of deleting
	cp := *e
	cp.State = core.UNREGISTERED
	if err := db.executors.Put(cp.ID, &cp); err != nil {
		return err
	}

	// Move running processes assigned to this executor back to WAITING
	ids := db.processesIdx.byExecutor.Lookup(e.ID)
	for _, id := range ids {
		if p, ok := db.processes.Get(id); ok {
			if p.State == core.RUNNING {
				oldState := p.State
				cn := p.FunctionSpec.Conditions.ColonyName
				pcp := *p
				db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
					SortKey:    pcp.PriorityTime,
					PrimaryKey: pcp.ID,
				})
				pcp.IsAssigned = false
				pcp.StartTime = time.Time{}
				pcp.EndTime = time.Time{}
				pcp.AssignedExecutorID = ""
				pcp.State = core.WAITING
				db.processes.Put(pcp.ID, &pcp)
				db.processesIdx.byColony.Add(cn, core.WAITING, index.IndexEntry[string]{
					SortKey:    pcp.PriorityTime,
					PrimaryKey: pcp.ID,
				})
				db.processesIdx.byExecutor.Remove(pcp.ID, e.ID)
			}
		}
	}

	// Remove functions for this executor
	return db.RemoveFunctionsByExecutorName(colonyName, executorName)
}

func (db *EmbeddedDatabase) RemoveExecutorsByColonyName(colonyName string) error {
	ids := db.executorsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if e, ok := db.executors.Get(id); ok {
			db.executorsIdx.byName.Remove(id, e.ColonyName+":"+e.Name)
			if e.BlueprintID != "" {
				db.executorsIdx.byBlueprint.Remove(id, e.BlueprintID)
			}
		}
		db.executorsIdx.byColony.Remove(id, colonyName)
		if err := db.executors.Delete(id); err != nil {
			return err
		}
	}

	return db.RemoveFunctionsByColonyName(colonyName)
}

func (db *EmbeddedDatabase) CountExecutors() (int, error) {
	executors, err := db.GetExecutors()
	if err != nil {
		return -1, err
	}
	return len(executors), nil
}

func (db *EmbeddedDatabase) CountExecutorsByColonyName(colonyName string) (int, error) {
	executors, err := db.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return -1, err
	}
	return len(executors), nil
}

func (db *EmbeddedDatabase) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	count := db.executors.Count(func(e *core.Executor) bool {
		return e.ColonyName == colonyName && e.State == state
	})
	return count, nil
}

func (db *EmbeddedDatabase) UpdateExecutorCapabilities(colonyName string, executorName string, capabilities core.Capabilities) error {
	e, err := db.getExecutorByNameInternal(colonyName, executorName)
	if err != nil {
		return err
	}
	if e == nil {
		return errors.New("Executor with name <" + executorName + "> does not exist in Colony with name <" + colonyName + ">")
	}

	cp := *e
	cp.Capabilities = capabilities
	return db.executors.Put(cp.ID, &cp)
}
