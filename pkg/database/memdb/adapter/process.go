package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// ProcessDatabase interface implementation

func (a *ColonyOSAdapter) AddProcess(process *core.Process) error {
	doc := &memdb.VelocityDocument{
		ID:     process.ID,
		Fields: a.processToFields(process),
	}
	
	return a.db.Insert(context.Background(), ProcessesCollection, doc)
}

func (a *ColonyOSAdapter) GetProcesses() ([]*core.Process, error) {
	result, err := a.db.List(context.Background(), ProcessesCollection, 10000, 0)
	if err != nil {
		return nil, err
	}
	
	processes := make([]*core.Process, 0, len(result))
	for _, doc := range result {
		process, err := a.fieldsToProcess(doc.Fields)
		if err == nil {
			processes = append(processes, process)
		}
	}
	
	return processes, nil
}

func (a *ColonyOSAdapter) GetProcessByID(processID string) (*core.Process, error) {
	doc, err := a.db.Get(context.Background(), ProcessesCollection, processID)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToProcess(doc.Fields)
}

// Simplified implementations for process queries (in practice, these would use more sophisticated filtering)
func (a *ColonyOSAdapter) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Process
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName && process.State == state {
			filtered = append(filtered, process)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Process
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName && 
		   process.AssignedExecutorID == executorID && 
		   process.State == state {
			filtered = append(filtered, process)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.WAITING, count)
}

func (a *ColonyOSAdapter) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.RUNNING, count)
}

func (a *ColonyOSAdapter) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.SUCCESS, count)
}

func (a *ColonyOSAdapter) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.FAILED, count)
}

func (a *ColonyOSAdapter) findProcessesByState(colonyName string, state int, count int) ([]*core.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Process
	for _, process := range processes {
		if len(filtered) >= count && count > 0 {
			break
		}
		if a.getProcessColonyName(process) == colonyName && process.State == state {
			filtered = append(filtered, process)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) FindAllRunningProcesses() ([]*core.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Process
	for _, process := range processes {
		if process.State == core.RUNNING {
			filtered = append(filtered, process)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) FindAllWaitingProcesses() ([]*core.Process, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Process
	for _, process := range processes {
		if process.State == core.WAITING {
			filtered = append(filtered, process)
		}
	}
	
	return filtered, nil
}

// Simplified candidate finding - in practice would check resource requirements
func (a *ColonyOSAdapter) FindCandidates(colonyName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.WAITING, count)
}

func (a *ColonyOSAdapter) FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return a.findProcessesByState(colonyName, core.WAITING, count)
}

func (a *ColonyOSAdapter) RemoveProcessByID(processID string) error {
	return a.db.Delete(context.Background(), ProcessesCollection, processID)
}

func (a *ColonyOSAdapter) RemoveAllProcesses() error {
	// In practice, this would be implemented more efficiently
	processes, err := a.GetProcesses()
	if err != nil {
		return err
	}
	
	for _, process := range processes {
		if err := a.RemoveProcessByID(process.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	return a.removeProcessesByColonyAndState(colonyName, core.WAITING)
}

func (a *ColonyOSAdapter) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	return a.removeProcessesByColonyAndState(colonyName, core.RUNNING)
}

func (a *ColonyOSAdapter) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	return a.removeProcessesByColonyAndState(colonyName, core.SUCCESS)
}

func (a *ColonyOSAdapter) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	return a.removeProcessesByColonyAndState(colonyName, core.FAILED)
}

func (a *ColonyOSAdapter) RemoveAllProcessesByColonyName(colonyName string) error {
	processes, err := a.GetProcesses()
	if err != nil {
		return err
	}
	
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName {
			if err := a.RemoveProcessByID(process.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	processes, err := a.GetProcesses()
	if err != nil {
		return err
	}
	
	for _, process := range processes {
		if process.ProcessGraphID == processGraphID {
			if err := a.RemoveProcessByID(process.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	processes, err := a.GetProcesses()
	if err != nil {
		return err
	}
	
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName && process.ProcessGraphID != "" {
			if err := a.RemoveProcessByID(process.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) removeProcessesByColonyAndState(colonyName string, state int) error {
	processes, err := a.GetProcesses()
	if err != nil {
		return err
	}
	
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName && process.State == state {
			if err := a.RemoveProcessByID(process.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) ResetProcess(process *core.Process) error {
	process.State = core.WAITING
	process.AssignedExecutorID = ""
	process.IsAssigned = false
	process.StartTime = time.Time{}
	process.EndTime = time.Time{}
	
	fields := a.processToFields(process)
	_, err := a.db.Update(context.Background(), ProcessesCollection, process.ID, fields)
	return err
}

func (a *ColonyOSAdapter) SetInput(processID string, input []interface{}) error {
	fields := map[string]interface{}{
		"input": input,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetOutput(processID string, output []interface{}) error {
	fields := map[string]interface{}{
		"output": output,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetErrors(processID string, errs []string) error {
	fields := map[string]interface{}{
		"errors": errs,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetProcessState(processID string, state int) error {
	fields := map[string]interface{}{
		"state": state,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetParents(processID string, parents []string) error {
	fields := map[string]interface{}{
		"parents": parents,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetChildren(processID string, children []string) error {
	fields := map[string]interface{}{
		"children": children,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

func (a *ColonyOSAdapter) SetWaitForParents(processID string, waitingForParent bool) error {
	fields := map[string]interface{}{
		"wait_for_parents": waitingForParent,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

// Assign uses simple check and update for process assignment  
// In a real implementation, this would use proper CAS at the database level
func (a *ColonyOSAdapter) Assign(executorID string, process *core.Process) error {
	// Get current process state
	currentProcess, err := a.GetProcessByID(process.ID)
	if err != nil {
		return err
	}
	
	// Check if already assigned
	if currentProcess.IsAssigned || currentProcess.AssignedExecutorID != "" {
		return fmt.Errorf("process already assigned")
	}
	
	// Update with assignment
	fields := map[string]interface{}{
		"assigned_executor_id": executorID,
		"is_assigned":         true,
		"state":              core.RUNNING,
		"start_time":         time.Now(),
	}
	
	_, err = a.db.Update(context.Background(), ProcessesCollection, process.ID, fields)
	return err
}

func (a *ColonyOSAdapter) Unassign(process *core.Process) error {
	fields := map[string]interface{}{
		"assigned_executor_id": "",
		"is_assigned":          false,
		"state":               core.WAITING,
		"start_time":          time.Time{},
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, process.ID, fields)
	return err
}

func (a *ColonyOSAdapter) MarkSuccessful(processID string) (float64, float64, error) {
	fields := map[string]interface{}{
		"state":    core.SUCCESS,
		"end_time": time.Now(),
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return 0.0, 0.0, err // Return 0 for wait and exec time for now
}

func (a *ColonyOSAdapter) MarkFailed(processID string, errs []string) error {
	fields := map[string]interface{}{
		"state":    core.FAILED,
		"end_time": time.Now(),
		"errors":   errs,
	}
	
	_, err := a.db.Update(context.Background(), ProcessesCollection, processID, fields)
	return err
}

// Count methods
func (a *ColonyOSAdapter) CountProcesses() (int, error) {
	result, err := a.db.List(context.Background(), ProcessesCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

func (a *ColonyOSAdapter) CountWaitingProcesses() (int, error) {
	return a.countProcessesByState(core.WAITING)
}

func (a *ColonyOSAdapter) CountRunningProcesses() (int, error) {
	return a.countProcessesByState(core.RUNNING)
}

func (a *ColonyOSAdapter) CountSuccessfulProcesses() (int, error) {
	return a.countProcessesByState(core.SUCCESS)
}

func (a *ColonyOSAdapter) CountFailedProcesses() (int, error) {
	return a.countProcessesByState(core.FAILED)
}

func (a *ColonyOSAdapter) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	return a.countProcessesByColonyAndState(colonyName, core.WAITING)
}

func (a *ColonyOSAdapter) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	return a.countProcessesByColonyAndState(colonyName, core.RUNNING)
}

func (a *ColonyOSAdapter) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	return a.countProcessesByColonyAndState(colonyName, core.SUCCESS)
}

func (a *ColonyOSAdapter) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	return a.countProcessesByColonyAndState(colonyName, core.FAILED)
}

func (a *ColonyOSAdapter) countProcessesByState(state int) (int, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, process := range processes {
		if process.State == state {
			count++
		}
	}
	
	return count, nil
}

func (a *ColonyOSAdapter) countProcessesByColonyAndState(colonyName string, state int) (int, error) {
	processes, err := a.GetProcesses()
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, process := range processes {
		if a.getProcessColonyName(process) == colonyName && process.State == state {
			count++
		}
	}
	
	return count, nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) processToFields(process *core.Process) map[string]interface{} {
	fields := map[string]interface{}{
		"id":                      process.ID,
		"initiator_id":            process.InitiatorID,
		"initiator_name":          process.InitiatorName,
		"assigned_executor_id":    process.AssignedExecutorID,
		"is_assigned":             process.IsAssigned,
		"state":                   process.State,
		"priority_time":           process.PriorityTime,
		"submission_time":         process.SubmissionTime,
		"start_time":              process.StartTime,
		"end_time":                process.EndTime,
		"wait_deadline":           process.WaitDeadline,
		"exec_deadline":           process.ExecDeadline,
		"retries":                 process.Retries,
		"wait_for_parents":        process.WaitForParents,
		"parents":                 process.Parents,
		"children":                process.Children,
		"process_graph_id":        process.ProcessGraphID,
		"input":                   process.Input,
		"output":                  process.Output,
		"errors":                  process.Errors,
	}
	
	// Serialize FunctionSpec
	if funcSpecData, err := json.Marshal(process.FunctionSpec); err == nil {
		fields["function_spec"] = string(funcSpecData)
	}
	
	// Serialize Attributes
	if attributeData, err := json.Marshal(process.Attributes); err == nil {
		fields["attributes"] = string(attributeData)
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToProcess(fields map[string]interface{}) (*core.Process, error) {
	process := &core.Process{}
	
	if id, ok := fields["id"].(string); ok {
		process.ID = id
	}
	if initiatorID, ok := fields["initiator_id"].(string); ok {
		process.InitiatorID = initiatorID
	}
	if initiatorName, ok := fields["initiator_name"].(string); ok {
		process.InitiatorName = initiatorName
	}
	if executorID, ok := fields["assigned_executor_id"].(string); ok {
		process.AssignedExecutorID = executorID
	}
	if isAssigned, ok := fields["is_assigned"].(bool); ok {
		process.IsAssigned = isAssigned
	}
	if state, ok := fields["state"].(int); ok {
		process.State = state
	} else if state, ok := fields["state"].(float64); ok {
		process.State = int(state)
	}
	if priorityTime, ok := fields["priority_time"].(int64); ok {
		process.PriorityTime = priorityTime
	} else if priorityTime, ok := fields["priority_time"].(float64); ok {
		process.PriorityTime = int64(priorityTime)
	}
	if submissionTime, ok := fields["submission_time"].(time.Time); ok {
		process.SubmissionTime = submissionTime
	}
	if startTime, ok := fields["start_time"].(time.Time); ok {
		process.StartTime = startTime
	}
	if endTime, ok := fields["end_time"].(time.Time); ok {
		process.EndTime = endTime
	}
	if waitDeadline, ok := fields["wait_deadline"].(time.Time); ok {
		process.WaitDeadline = waitDeadline
	}
	if execDeadline, ok := fields["exec_deadline"].(time.Time); ok {
		process.ExecDeadline = execDeadline
	}
	if retries, ok := fields["retries"].(int); ok {
		process.Retries = retries
	} else if retries, ok := fields["retries"].(float64); ok {
		process.Retries = int(retries)
	}
	if waitForParents, ok := fields["wait_for_parents"].(bool); ok {
		process.WaitForParents = waitForParents
	}
	if parents, ok := fields["parents"].([]string); ok {
		process.Parents = parents
	}
	if children, ok := fields["children"].([]string); ok {
		process.Children = children
	}
	if processGraphID, ok := fields["process_graph_id"].(string); ok {
		process.ProcessGraphID = processGraphID
	}
	if input, ok := fields["input"].([]interface{}); ok {
		process.Input = input
	}
	if output, ok := fields["output"].([]interface{}); ok {
		process.Output = output
	}
	if errors, ok := fields["errors"].([]string); ok {
		process.Errors = errors
	}
	
	// Deserialize FunctionSpec
	if funcSpecStr, ok := fields["function_spec"].(string); ok {
		var functionSpec core.FunctionSpec
		if err := json.Unmarshal([]byte(funcSpecStr), &functionSpec); err == nil {
			process.FunctionSpec = functionSpec
		}
	}
	
	// Deserialize Attributes
	if attributeStr, ok := fields["attributes"].(string); ok {
		var attributes []core.Attribute
		if err := json.Unmarshal([]byte(attributeStr), &attributes); err == nil {
			process.Attributes = attributes
		}
	}
	
	return process, nil
}

func (a *ColonyOSAdapter) getProcessColonyName(process *core.Process) string {
	// Extract colony name from process FunctionSpec
	if process.FunctionSpec.Conditions.ColonyName != "" {
		return process.FunctionSpec.Conditions.ColonyName
	}
	return ""
}