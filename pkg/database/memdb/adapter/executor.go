package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// ExecutorDatabase interface implementation

func (a *ColonyOSAdapter) AddExecutor(executor *core.Executor) error {
	doc := &memdb.VelocityDocument{
		ID:     executor.ID,
		Fields: a.executorToFields(executor),
	}
	
	return a.db.Insert(context.Background(), ExecutorsCollection, doc)
}

func (a *ColonyOSAdapter) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	// Simple implementation - find executor by name and update allocations
	executors, err := a.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, executor := range executors {
		if executor.Name == executorName {
			fields := map[string]interface{}{
				"allocations": a.allocationsToMap(allocations),
			}
			
			_, err = a.db.Update(context.Background(), ExecutorsCollection, executor.ID, fields)
			return err
		}
	}
	
	return fmt.Errorf("executor not found")
}

func (a *ColonyOSAdapter) GetExecutors() ([]*core.Executor, error) {
	result, err := a.db.List(context.Background(), ExecutorsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	executors := make([]*core.Executor, 0, len(result))
	for _, doc := range result {
		executor, err := a.fieldsToExecutor(doc.Fields)
		if err == nil {
			executors = append(executors, executor)
		}
	}
	
	return executors, nil
}

func (a *ColonyOSAdapter) GetExecutorByID(executorID string) (*core.Executor, error) {
	doc, err := a.db.Get(context.Background(), ExecutorsCollection, executorID)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToExecutor(doc.Fields)
}

func (a *ColonyOSAdapter) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	executors, err := a.GetExecutors()
	if err != nil {
		return nil, err
	}
	
	var filtered []*core.Executor
	for _, executor := range executors {
		if executor.ColonyName == colonyName {
			filtered = append(filtered, executor)
		}
	}
	
	return filtered, nil
}

func (a *ColonyOSAdapter) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	executors, err := a.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return nil, err
	}
	
	for _, executor := range executors {
		if executor.Name == executorName {
			return executor, nil
		}
	}
	
	return nil, fmt.Errorf("executor not found")
}

func (a *ColonyOSAdapter) ApproveExecutor(executor *core.Executor) error {
	fields := map[string]interface{}{
		"state": core.APPROVED,
	}
	
	_, err := a.db.Update(context.Background(), ExecutorsCollection, executor.ID, fields)
	return err
}

func (a *ColonyOSAdapter) RejectExecutor(executor *core.Executor) error {
	fields := map[string]interface{}{
		"state": core.REJECTED,
	}
	
	_, err := a.db.Update(context.Background(), ExecutorsCollection, executor.ID, fields)
	return err
}

func (a *ColonyOSAdapter) MarkAlive(executor *core.Executor) error {
	fields := map[string]interface{}{
		"last_heard_from_time": time.Now(),
	}
	
	_, err := a.db.Update(context.Background(), ExecutorsCollection, executor.ID, fields)
	return err
}

func (a *ColonyOSAdapter) RemoveExecutorByName(colonyName string, executorName string) error {
	executor, err := a.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), ExecutorsCollection, executor.ID)
}

func (a *ColonyOSAdapter) RemoveExecutorsByColonyName(colonyName string) error {
	executors, err := a.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, executor := range executors {
		if err := a.db.Delete(context.Background(), ExecutorsCollection, executor.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) CountExecutors() (int, error) {
	result, err := a.db.List(context.Background(), ExecutorsCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

func (a *ColonyOSAdapter) CountExecutorsByColonyName(colonyName string) (int, error) {
	executors, err := a.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return 0, err
	}
	
	return len(executors), nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) executorToFields(executor *core.Executor) map[string]interface{} {
	fields := map[string]interface{}{
		"id":                       executor.ID,
		"type":                     executor.Type,
		"name":                     executor.Name,
		"colony_name":              executor.ColonyName,
		"state":                    executor.State,
		"require_func_reg":         executor.RequireFuncReg,
		"commission_time":          executor.CommissionTime,
		"last_heard_from_time":     executor.LastHeardFromTime,
	}
	
	// Serialize complex fields
	if locationData, err := json.Marshal(executor.Location); err == nil {
		fields["location"] = string(locationData)
	}
	if capabilitiesData, err := json.Marshal(executor.Capabilities); err == nil {
		fields["capabilities"] = string(capabilitiesData)
	}
	if allocationsData, err := json.Marshal(executor.Allocations); err == nil {
		fields["allocations"] = string(allocationsData)
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToExecutor(fields map[string]interface{}) (*core.Executor, error) {
	executor := &core.Executor{}
	
	if id, ok := fields["id"].(string); ok {
		executor.ID = id
	}
	if execType, ok := fields["type"].(string); ok {
		executor.Type = execType
	}
	if name, ok := fields["name"].(string); ok {
		executor.Name = name
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		executor.ColonyName = colonyName
	}
	if state, ok := fields["state"].(int); ok {
		executor.State = state
	} else if state, ok := fields["state"].(float64); ok {
		executor.State = int(state)
	}
	if requireFuncReg, ok := fields["require_func_reg"].(bool); ok {
		executor.RequireFuncReg = requireFuncReg
	}
	if commissionTime, ok := fields["commission_time"].(time.Time); ok {
		executor.CommissionTime = commissionTime
	}
	if lastHeardFromTime, ok := fields["last_heard_from_time"].(time.Time); ok {
		executor.LastHeardFromTime = lastHeardFromTime
	}
	
	// Deserialize complex fields
	if locationStr, ok := fields["location"].(string); ok {
		var location core.Location
		if err := json.Unmarshal([]byte(locationStr), &location); err == nil {
			executor.Location = location
		}
	}
	if capabilitiesStr, ok := fields["capabilities"].(string); ok {
		var capabilities core.Capabilities
		if err := json.Unmarshal([]byte(capabilitiesStr), &capabilities); err == nil {
			executor.Capabilities = capabilities
		}
	}
	if allocationsStr, ok := fields["allocations"].(string); ok {
		var allocations core.Allocations
		if err := json.Unmarshal([]byte(allocationsStr), &allocations); err == nil {
			executor.Allocations = allocations
		}
	}
	
	return executor, nil
}

func (a *ColonyOSAdapter) allocationsToMap(allocations core.Allocations) map[string]interface{} {
	// Simple conversion - in practice would need full implementation
	result := make(map[string]interface{})
	if data, err := json.Marshal(allocations); err == nil {
		json.Unmarshal(data, &result)
	}
	return result
}