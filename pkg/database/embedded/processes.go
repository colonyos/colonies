package embedded

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/embedded/index"
	"github.com/colonyos/colonies/pkg/parsers"
)

func (db *EmbeddedDatabase) AddProcess(process *core.Process) error {
	submissionTime := time.Now()
	process.SetSubmissionTime(submissionTime)

	maxWaitTime := process.FunctionSpec.MaxWaitTime
	if maxWaitTime > 0 {
		process.WaitDeadline = time.Now().Add(time.Duration(maxWaitTime) * time.Second)
	}

	// Convert Env to Attributes
	for key, value := range process.FunctionSpec.Env {
		process.Attributes = append(process.Attributes, core.CreateAttribute(
			process.ID,
			process.FunctionSpec.Conditions.ColonyName,
			process.ProcessGraphID,
			core.ENV,
			key,
			value,
		))
	}

	// Store a copy to isolate from caller modifications
	cp := copyProcess(process)
	if err := db.processes.Put(cp.ID, cp); err != nil {
		return err
	}

	colonyName := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Add(colonyName, cp.State, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	if cp.ProcessGraphID != "" {
		db.processesIdx.byGraph.Add(cp.ID, cp.ProcessGraphID)
	}

	// Add attributes
	if err := db.AddAttributes(process.Attributes); err != nil {
		return err
	}

	return nil
}

func (db *EmbeddedDatabase) GetProcesses() ([]*core.Process, error) {
	all := db.processes.All()
	result := make([]*core.Process, 0, len(all))
	for _, p := range all {
		enriched := db.enrichProcess(p)
		result = append(result, enriched)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SubmissionTime.After(result[j].SubmissionTime)
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetProcessByID(processID string) (*core.Process, error) {
	p, ok := db.processes.Get(processID)
	if !ok {
		return nil, nil
	}
	return db.enrichProcess(p), nil
}

func (db *EmbeddedDatabase) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	cutoff := time.Now().Add(-time.Duration(seconds) * time.Second)
	var result []*core.Process
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.FunctionSpec.Conditions.ColonyName == colonyName &&
			p.State == state &&
			!p.SubmissionTime.Before(cutoff) &&
			!p.SubmissionTime.After(time.Now())
	})
	for _, p := range processes {
		result = append(result, db.enrichProcess(p))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SubmissionTime.Before(result[j].SubmissionTime)
	})
	return result, nil
}

func (db *EmbeddedDatabase) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	cutoff := time.Now().Add(-time.Duration(seconds) * time.Second)
	var result []*core.Process
	ids := db.processesIdx.byExecutor.Lookup(executorID)
	for _, id := range ids {
		if p, ok := db.processes.Get(id); ok {
			if p.FunctionSpec.Conditions.ColonyName == colonyName &&
				p.State == state &&
				!p.SubmissionTime.Before(cutoff) &&
				!p.SubmissionTime.After(time.Now()) {
				result = append(result, db.enrichProcess(p))
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].SubmissionTime.Before(result[j].SubmissionTime)
	})
	return result, nil
}

func (db *EmbeddedDatabase) findProcessesByState(colonyName string, state int, executorType string, label string, initiator string, count int, orderByPriorityTime bool) ([]*core.Process, error) {
	var candidates []index.IndexEntry[string]

	db.processesIdx.byColony.AscendFirst(colonyName, state, count*10, func(entry index.IndexEntry[string]) bool {
		candidates = append(candidates, entry)
		return true
	})

	var result []*core.Process
	for _, entry := range candidates {
		if len(result) >= count {
			break
		}
		p, ok := db.processes.Get(entry.PrimaryKey)
		if !ok {
			continue
		}
		if executorType != "" && p.FunctionSpec.Conditions.ExecutorType != executorType {
			continue
		}
		if label != "" && p.FunctionSpec.Label != label {
			continue
		}
		if initiator != "" && p.InitiatorName != initiator {
			continue
		}
		result = append(result, db.enrichProcess(p))
	}

	if !orderByPriorityTime {
		// For non-WAITING states, we need descending order
		// The CompoundIndex gives ascending. For RUNNING sort by StartTime ASC,
		// for SUCCESS/FAILED/CANCELLED sort by EndTime DESC.
		if state == core.RUNNING {
			sort.Slice(result, func(i, j int) bool {
				return result[i].StartTime.Before(result[j].StartTime)
			})
		} else {
			sort.Slice(result, func(i, j int) bool {
				return result[i].EndTime.After(result[j].EndTime)
			})
		}
	}

	return result, nil
}

func (db *EmbeddedDatabase) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByState(colonyName, core.WAITING, executorType, label, initiator, count, true)
}

func (db *EmbeddedDatabase) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByState(colonyName, core.RUNNING, executorType, label, initiator, count, false)
}

func (db *EmbeddedDatabase) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByState(colonyName, core.SUCCESS, executorType, label, initiator, count, false)
}

func (db *EmbeddedDatabase) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByState(colonyName, core.FAILED, executorType, label, initiator, count, false)
}

func (db *EmbeddedDatabase) FindCancelledProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByState(colonyName, core.CANCELLED, executorType, label, initiator, count, false)
}

func (db *EmbeddedDatabase) FindAllRunningProcesses() ([]*core.Process, error) {
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.State == core.RUNNING
	})
	var result []*core.Process
	for _, p := range processes {
		result = append(result, db.enrichProcess(p))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})
	return result, nil
}

func (db *EmbeddedDatabase) FindAllWaitingProcesses() ([]*core.Process, error) {
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.State == core.WAITING
	})
	var result []*core.Process
	for _, p := range processes {
		result = append(result, db.enrichProcess(p))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].PriorityTime < result[j].PriorityTime
	})
	if len(result) > 1000 {
		result = result[:1000]
	}
	return result, nil
}

func (db *EmbeddedDatabase) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var result []*core.Process

	db.processesIdx.byColony.AscendFirst(colonyName, core.WAITING, count*10, func(entry index.IndexEntry[string]) bool {
		if len(result) >= count {
			return false
		}
		p, ok := db.processes.Get(entry.PrimaryKey)
		if !ok {
			return true
		}
		if !db.matchCandidate(p, executorType, executorLocationName, cpu, memory, storage, nodes, processes, processesPerNode) {
			return true
		}
		// General pool: no target executor names
		if len(p.FunctionSpec.Conditions.ExecutorNames) > 0 {
			return true
		}
		result = append(result, db.enrichProcess(p))
		return true
	})

	return result, nil
}

func (db *EmbeddedDatabase) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var result []*core.Process

	db.processesIdx.byColony.AscendFirst(colonyName, core.WAITING, count*10, func(entry index.IndexEntry[string]) bool {
		if len(result) >= count {
			return false
		}
		p, ok := db.processes.Get(entry.PrimaryKey)
		if !ok {
			return true
		}
		if !db.matchCandidate(p, executorType, executorLocationName, cpu, memory, storage, nodes, processes, processesPerNode) {
			return true
		}
		// Must target this executor by name
		if !containsString(p.FunctionSpec.Conditions.ExecutorNames, executorName) {
			return true
		}
		result = append(result, db.enrichProcess(p))
		return true
	})

	return result, nil
}

func (db *EmbeddedDatabase) SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error) {
	db.processes.Lock()
	defer db.processes.Unlock()

	// Find best candidate: combines FindCandidatesByName + FindCandidates logic with OR
	var bestProcess *core.Process

	idx := db.processesIdx.byColony.Get(colonyName, core.WAITING)
	if idx == nil {
		return nil, nil
	}

	idx.AscendFirst(count*10, func(entry index.IndexEntry[string]) bool {
		p, ok := db.processes.GetUnlocked(entry.PrimaryKey)
		if !ok {
			return true
		}
		if !db.matchCandidateUnlocked(p, executorType, executorLocation, cpu, memory, storage, nodes, processes, processesPerNode) {
			return true
		}

		// Match by executor name OR general pool
		named := containsString(p.FunctionSpec.Conditions.ExecutorNames, executorName)
		general := len(p.FunctionSpec.Conditions.ExecutorNames) == 0
		if !named && !general {
			return true
		}

		bestProcess = p
		return false // stop at first match
	})

	if bestProcess == nil {
		return nil, nil
	}

	// Atomically assign via copy-on-write
	now := time.Now()
	oldState := bestProcess.State

	cp := *bestProcess
	cp.IsAssigned = true
	cp.StartTime = now
	cp.AssignedExecutorID = executorID
	cp.State = core.RUNNING
	if cp.FunctionSpec.MaxExecTime > 0 {
		cp.ExecDeadline = now.Add(time.Duration(cp.FunctionSpec.MaxExecTime) * time.Second)
	}

	if err := db.processes.PutUnlocked(cp.ID, &cp); err != nil {
		return nil, err
	}

	// Update index
	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.RUNNING, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byExecutor.Add(cp.ID, executorID)

	// Update attribute state
	db.setAttributeState(cp.ID, core.RUNNING)

	return db.enrichProcess(&cp), nil
}

func (db *EmbeddedDatabase) Assign(executorID string, process *core.Process) error {
	p, ok := db.processes.Get(process.ID)
	if !ok {
		return errors.New("Process with Id <" + process.ID + "> not found")
	}

	if p.IsAssigned {
		return errors.New("Process already assigned")
	}

	if p.State == core.CANCELLED {
		return errors.New("Cannot assign cancelled process")
	}

	startTime := time.Now()
	oldState := p.State

	cp := *p // copy-on-write: never modify stored pointer in-place
	cp.IsAssigned = true
	cp.StartTime = startTime
	cp.AssignedExecutorID = executorID
	cp.State = core.RUNNING
	if cp.FunctionSpec.MaxExecTime > 0 {
		cp.ExecDeadline = time.Now().Add(time.Duration(cp.FunctionSpec.MaxExecTime) * time.Second)
	}

	if err := db.processes.Put(process.ID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.RUNNING, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byExecutor.Add(cp.ID, executorID)

	db.setAttributeState(process.ID, core.RUNNING)

	process.SetStartTime(startTime)
	process.Assign()
	process.SetAssignedExecutorID(executorID)
	process.SetState(core.RUNNING)

	return nil
}

func (db *EmbeddedDatabase) Unassign(process *core.Process) error {
	p, ok := db.processes.Get(process.ID)
	if !ok {
		return errors.New("Process with Id <" + process.ID + "> not found")
	}

	endTime := time.Now()
	oldState := p.State
	oldExecutorID := p.AssignedExecutorID

	cp := *p
	cp.IsAssigned = false
	cp.EndTime = endTime
	cp.State = core.WAITING
	cp.Retries = process.Retries + 1
	cp.AssignedExecutorID = ""
	if cp.FunctionSpec.MaxWaitTime > 0 {
		cp.WaitDeadline = time.Now().Add(time.Duration(cp.FunctionSpec.MaxWaitTime) * time.Second)
	}

	if err := db.processes.Put(process.ID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.WAITING, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	if oldExecutorID != "" {
		db.processesIdx.byExecutor.Remove(cp.ID, oldExecutorID)
	}

	db.setAttributeState(process.ID, core.PENDING)

	process.SetEndTime(endTime)
	process.Unassign()
	process.SetState(core.WAITING)

	return nil
}

func (db *EmbeddedDatabase) MarkSuccessful(processID string) (float64, float64, error) {
	p, ok := db.processes.Get(processID)
	if !ok {
		return 0, 0, errors.New("Process with Id <" + processID + "> not found")
	}

	if p.State == core.FAILED {
		return 0, 0, errors.New("Tried to set failed process as successful")
	}
	if p.State == core.CANCELLED {
		return 0, 0, errors.New("Tried to set cancelled process as successful")
	}
	if p.State == core.WAITING {
		return 0, 0, errors.New("Tried to set waiting process as successful without being running")
	}

	endTime := time.Now()
	oldState := p.State

	cp := *p
	cp.EndTime = endTime
	cp.State = core.SUCCESS

	if err := db.processes.Put(processID, &cp); err != nil {
		return 0, 0, err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.SUCCESS, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})

	db.setAttributeState(processID, core.SUCCESS)

	return cp.WaitingTime().Seconds(), cp.ProcessingTime().Seconds(), nil
}

func (db *EmbeddedDatabase) MarkFailed(processID string, errs []string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}

	if p.State == core.SUCCESS {
		return errors.New("Tried to set successful process as failed")
	}
	if p.State == core.FAILED {
		return errors.New("Tried to set failed process as failed")
	}
	if p.State == core.CANCELLED {
		return errors.New("Tried to set cancelled process as failed")
	}

	endTime := time.Now()
	oldState := p.State

	cp := *p
	cp.EndTime = endTime
	cp.State = core.FAILED
	cp.Errors = errs

	if err := db.processes.Put(processID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.FAILED, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})

	db.setAttributeState(processID, core.FAILED)

	return nil
}

func (db *EmbeddedDatabase) MarkCancelled(processID string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}

	if p.State == core.SUCCESS {
		return errors.New("Tried to cancel successful process")
	}
	if p.State == core.FAILED {
		return errors.New("Tried to cancel failed process")
	}
	if p.State == core.CANCELLED {
		return errors.New("Tried to cancel already cancelled process")
	}

	endTime := time.Now()
	oldState := p.State

	cp := *p
	cp.EndTime = endTime
	cp.State = core.CANCELLED

	if err := db.processes.Put(processID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.CANCELLED, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})

	db.setAttributeState(processID, core.CANCELLED)

	return nil
}

func (db *EmbeddedDatabase) ResetProcess(process *core.Process) error {
	p, ok := db.processes.Get(process.ID)
	if !ok {
		return errors.New("Process with Id <" + process.ID + "> not found")
	}

	oldState := p.State
	oldPriorityTime := p.PriorityTime
	submissionTime := time.Now()

	cp := *p
	cp.IsAssigned = false
	cp.SetSubmissionTime(submissionTime)
	cp.StartTime = time.Time{}
	cp.EndTime = time.Time{}
	cp.AssignedExecutorID = ""
	cp.State = core.WAITING
	if cp.FunctionSpec.MaxWaitTime > 0 {
		cp.WaitDeadline = time.Now().Add(time.Duration(cp.FunctionSpec.MaxWaitTime) * time.Second)
	}

	if err := db.processes.Put(process.ID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    oldPriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, core.WAITING, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})

	process.SetSubmissionTime(submissionTime)
	process.SetStartTime(time.Time{})
	process.SetEndTime(time.Time{})
	process.AssignedExecutorID = ""
	process.Unassign()
	process.SetState(core.WAITING)

	return nil
}

func (db *EmbeddedDatabase) SetInput(processID string, input []interface{}) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.Input = input
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) SetOutput(processID string, output []interface{}) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.Output = output
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) SetErrors(processID string, errs []string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.Errors = errs
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) SetProcessState(processID string, state int) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}

	oldState := p.State

	cp := *p
	cp.State = state

	if err := db.processes.Put(processID, &cp); err != nil {
		return err
	}

	cn := cp.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, oldState, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})
	db.processesIdx.byColony.Add(cn, state, index.IndexEntry[string]{
		SortKey:    cp.PriorityTime,
		PrimaryKey: cp.ID,
	})

	db.setAttributeState(processID, state)

	return nil
}

func (db *EmbeddedDatabase) SetParents(processID string, parents []string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.Parents = parents
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) SetChildren(processID string, children []string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.Children = children
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) SetWaitForParents(processID string, waitForParent bool) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return errors.New("Process with Id <" + processID + "> not found")
	}
	cp := *p
	cp.WaitForParents = waitForParent
	return db.processes.Put(processID, &cp)
}

func (db *EmbeddedDatabase) RemoveProcessByID(processID string) error {
	p, ok := db.processes.Get(processID)
	if !ok {
		return nil
	}
	db.removeProcessFromIndexes(p)
	db.RemoveAllAttributesByTargetID(processID)
	return db.processes.Delete(processID)
}

func (db *EmbeddedDatabase) RemoveAllProcesses() error {
	for _, p := range db.processes.All() {
		db.removeProcessFromIndexes(p)
		db.processes.Delete(p.ID)
	}
	db.RemoveAllAttributes()
	return nil
}

func (db *EmbeddedDatabase) removeProcessesByColonyNameAndState(colonyName string, state int) error {
	// Remove processes that belong to colony, have matching state, and are NOT in a process graph
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.FunctionSpec.Conditions.ColonyName == colonyName &&
			p.ProcessGraphID == "" &&
			p.State == state
	})
	for _, p := range processes {
		db.removeProcessFromIndexes(p)
		db.RemoveAllAttributesByTargetID(p.ID)
		db.processes.Delete(p.ID)
	}
	db.RemoveAllAttributesByColonyNameWithState(colonyName, state)
	return nil
}

func (db *EmbeddedDatabase) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyNameAndState(colonyName, core.WAITING)
}

func (db *EmbeddedDatabase) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyNameAndState(colonyName, core.RUNNING)
}

func (db *EmbeddedDatabase) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyNameAndState(colonyName, core.SUCCESS)
}

func (db *EmbeddedDatabase) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyNameAndState(colonyName, core.FAILED)
}

func (db *EmbeddedDatabase) RemoveAllCancelledProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyNameAndState(colonyName, core.CANCELLED)
}

func (db *EmbeddedDatabase) RemoveAllProcessesByColonyName(colonyName string) error {
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.FunctionSpec.Conditions.ColonyName == colonyName && p.ProcessGraphID == ""
	})
	for _, p := range processes {
		db.removeProcessFromIndexes(p)
		db.RemoveAllAttributesByTargetID(p.ID)
		db.processes.Delete(p.ID)
	}
	db.RemoveAllAttributesByColonyName(colonyName)
	return nil
}

func (db *EmbeddedDatabase) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	ids := db.processesIdx.byGraph.Lookup(processGraphID)
	for _, id := range ids {
		if p, ok := db.processes.Get(id); ok {
			db.removeProcessFromIndexes(p)
			db.processes.Delete(id)
		}
	}
	db.RemoveAllAttributesByProcessGraphID(processGraphID)
	return nil
}

func (db *EmbeddedDatabase) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	db.RemoveAllAttributesInProcessGraphsByColonyName(colonyName)
	processes := db.processes.Filter(func(p *core.Process) bool {
		return p.FunctionSpec.Conditions.ColonyName == colonyName && p.ProcessGraphID != ""
	})
	for _, p := range processes {
		db.removeProcessFromIndexes(p)
		db.processes.Delete(p.ID)
	}
	return nil
}

func (db *EmbeddedDatabase) CountProcesses() (int, error) {
	return db.processes.Len(), nil
}

func (db *EmbeddedDatabase) CountWaitingProcesses() (int, error) {
	return db.processesIdx.byColony.CountBySubgroup(core.WAITING), nil
}

func (db *EmbeddedDatabase) CountRunningProcesses() (int, error) {
	return db.processesIdx.byColony.CountBySubgroup(core.RUNNING), nil
}

func (db *EmbeddedDatabase) CountSuccessfulProcesses() (int, error) {
	return db.processesIdx.byColony.CountBySubgroup(core.SUCCESS), nil
}

func (db *EmbeddedDatabase) CountFailedProcesses() (int, error) {
	return db.processesIdx.byColony.CountBySubgroup(core.FAILED), nil
}

func (db *EmbeddedDatabase) CountCancelledProcesses() (int, error) {
	return db.processesIdx.byColony.CountBySubgroup(core.CANCELLED), nil
}

func (db *EmbeddedDatabase) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	return db.processesIdx.byColony.Count(colonyName, core.WAITING), nil
}

func (db *EmbeddedDatabase) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	return db.processesIdx.byColony.Count(colonyName, core.RUNNING), nil
}

func (db *EmbeddedDatabase) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	return db.processesIdx.byColony.Count(colonyName, core.SUCCESS), nil
}

func (db *EmbeddedDatabase) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	return db.processesIdx.byColony.Count(colonyName, core.FAILED), nil
}

func (db *EmbeddedDatabase) CountCancelledProcessesByColonyName(colonyName string) (int, error) {
	return db.processesIdx.byColony.Count(colonyName, core.CANCELLED), nil
}

// copyProcess creates a deep copy of a Process so that the stored pointer
// is never shared with callers. This prevents data races between concurrent
// goroutines reading and writing the same process.
func copyProcess(p *core.Process) *core.Process {
	cp := *p // shallow struct copy

	// Deep copy slices
	if p.Attributes != nil {
		cp.Attributes = make([]core.Attribute, len(p.Attributes))
		copy(cp.Attributes, p.Attributes)
	}
	if p.Parents != nil {
		cp.Parents = make([]string, len(p.Parents))
		copy(cp.Parents, p.Parents)
	}
	if p.Children != nil {
		cp.Children = make([]string, len(p.Children))
		copy(cp.Children, p.Children)
	}
	if p.Errors != nil {
		cp.Errors = make([]string, len(p.Errors))
		copy(cp.Errors, p.Errors)
	}
	if p.Input != nil {
		cp.Input = make([]interface{}, len(p.Input))
		copy(cp.Input, p.Input)
	}
	if p.Output != nil {
		cp.Output = make([]interface{}, len(p.Output))
		copy(cp.Output, p.Output)
	}

	// Deep copy FunctionSpec reference types
	if p.FunctionSpec.Args != nil {
		cp.FunctionSpec.Args = make([]interface{}, len(p.FunctionSpec.Args))
		copy(cp.FunctionSpec.Args, p.FunctionSpec.Args)
	}
	if p.FunctionSpec.KwArgs != nil {
		cp.FunctionSpec.KwArgs = make(map[string]interface{}, len(p.FunctionSpec.KwArgs))
		for k, v := range p.FunctionSpec.KwArgs {
			cp.FunctionSpec.KwArgs[k] = v
		}
	}
	if p.FunctionSpec.Env != nil {
		cp.FunctionSpec.Env = make(map[string]string, len(p.FunctionSpec.Env))
		for k, v := range p.FunctionSpec.Env {
			cp.FunctionSpec.Env[k] = v
		}
	}
	if p.FunctionSpec.Channels != nil {
		cp.FunctionSpec.Channels = make([]string, len(p.FunctionSpec.Channels))
		copy(cp.FunctionSpec.Channels, p.FunctionSpec.Channels)
	}
	if p.FunctionSpec.Conditions.ExecutorNames != nil {
		cp.FunctionSpec.Conditions.ExecutorNames = make([]string, len(p.FunctionSpec.Conditions.ExecutorNames))
		copy(cp.FunctionSpec.Conditions.ExecutorNames, p.FunctionSpec.Conditions.ExecutorNames)
	}
	if p.FunctionSpec.Conditions.Dependencies != nil {
		cp.FunctionSpec.Conditions.Dependencies = make([]string, len(p.FunctionSpec.Conditions.Dependencies))
		copy(cp.FunctionSpec.Conditions.Dependencies, p.FunctionSpec.Conditions.Dependencies)
	}
	if p.FunctionSpec.Filesystem.SnapshotMounts != nil {
		cp.FunctionSpec.Filesystem.SnapshotMounts = make([]core.SnapshotMount, len(p.FunctionSpec.Filesystem.SnapshotMounts))
		copy(cp.FunctionSpec.Filesystem.SnapshotMounts, p.FunctionSpec.Filesystem.SnapshotMounts)
	}
	if p.FunctionSpec.Filesystem.SyncDirMounts != nil {
		cp.FunctionSpec.Filesystem.SyncDirMounts = make([]core.SyncDirMount, len(p.FunctionSpec.Filesystem.SyncDirMounts))
		copy(cp.FunctionSpec.Filesystem.SyncDirMounts, p.FunctionSpec.Filesystem.SyncDirMounts)
	}

	return &cp
}

// enrichProcess creates a deep copy of the process, attaches attributes,
// and rebuilds the Env map from ENV-type attributes.
func (db *EmbeddedDatabase) enrichProcess(p *core.Process) *core.Process {
	cp := copyProcess(p)

	attrs, _ := db.GetAttributes(cp.ID)
	if len(attrs) == 0 {
		attrs = make([]core.Attribute, 0)
	}
	cp.Attributes = attrs

	env := make(map[string]string)
	for _, a := range attrs {
		if a.AttributeType == core.ENV {
			env[a.Key] = a.Value
		}
	}
	cp.FunctionSpec.Env = env

	if cp.Input == nil {
		cp.Input = make([]interface{}, 0)
	}
	if cp.Output == nil {
		cp.Output = make([]interface{}, 0)
	}
	if cp.Errors == nil {
		cp.Errors = make([]string, 0)
	}
	if cp.Parents == nil {
		cp.Parents = make([]string, 0)
	}
	if cp.Children == nil {
		cp.Children = make([]string, 0)
	}
	if cp.FunctionSpec.Conditions.Dependencies == nil {
		cp.FunctionSpec.Conditions.Dependencies = make([]string, 0)
	}
	if cp.FunctionSpec.Conditions.ExecutorNames == nil {
		cp.FunctionSpec.Conditions.ExecutorNames = make([]string, 0)
	}
	if cp.FunctionSpec.Channels == nil {
		cp.FunctionSpec.Channels = make([]string, 0)
	}

	return cp
}

// removeProcessFromIndexes removes a process from all process indexes.
func (db *EmbeddedDatabase) removeProcessFromIndexes(p *core.Process) {
	cn := p.FunctionSpec.Conditions.ColonyName
	db.processesIdx.byColony.Remove(cn, p.State, index.IndexEntry[string]{
		SortKey:    p.PriorityTime,
		PrimaryKey: p.ID,
	})
	if p.AssignedExecutorID != "" {
		db.processesIdx.byExecutor.Remove(p.ID, p.AssignedExecutorID)
	}
	if p.ProcessGraphID != "" {
		db.processesIdx.byGraph.Remove(p.ID, p.ProcessGraphID)
	}
}

// matchCandidate checks if a process matches the candidate criteria.
func (db *EmbeddedDatabase) matchCandidate(p *core.Process, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int) bool {
	if p.IsAssigned || p.WaitForParents {
		return false
	}
	if p.FunctionSpec.Conditions.ExecutorType != executorType {
		return false
	}

	// Parse resource requirements
	pCPU, err := parsers.ConvertCPUToInt(p.FunctionSpec.Conditions.CPU)
	if err != nil {
		return false
	}
	pMem, err := parsers.ConvertMemoryToBytes(p.FunctionSpec.Conditions.Memory)
	if err != nil {
		return false
	}
	pStorage, err := parsers.ConvertMemoryToBytes(p.FunctionSpec.Conditions.Storage)
	if err != nil {
		return false
	}

	if pCPU > cpu || pMem > memory || pStorage > storage {
		return false
	}
	if p.FunctionSpec.Conditions.Nodes > nodes {
		return false
	}
	if p.FunctionSpec.Conditions.Processes > processes {
		return false
	}
	if p.FunctionSpec.Conditions.ProcessesPerNode > processesPerNode {
		return false
	}

	// Location check
	loc := p.FunctionSpec.Conditions.LocationName
	if loc != "" && !strings.EqualFold(loc, executorLocationName) {
		return false
	}

	return true
}

// matchCandidateUnlocked is the same as matchCandidate but used when the store is already locked.
func (db *EmbeddedDatabase) matchCandidateUnlocked(p *core.Process, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int) bool {
	return db.matchCandidate(p, executorType, executorLocationName, cpu, memory, storage, nodes, processes, processesPerNode)
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
