package adapter

import (
	"context"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// GeneratorDatabase interface implementation

func (a *ColonyOSAdapter) AddGenerator(generator *core.Generator) error {
	doc := &memdb.VelocityDocument{
		ID:     generator.ID,
		Fields: a.generatorToFields(generator),
	}
	
	return a.db.Insert(context.Background(), GeneratorsCollection, doc)
}

func (a *ColonyOSAdapter) SetGeneratorLastRun(generatorID string) error {
	fields := map[string]interface{}{
		"last_run": "updated", // Simplified - would use actual timestamp
	}
	
	_, err := a.db.Update(context.Background(), GeneratorsCollection, generatorID, fields)
	return err
}

func (a *ColonyOSAdapter) SetGeneratorFirstPack(generatorID string) error {
	fields := map[string]interface{}{
		"first_pack": true,
	}
	
	_, err := a.db.Update(context.Background(), GeneratorsCollection, generatorID, fields)
	return err
}

func (a *ColonyOSAdapter) GetGenerators(colonyName string) ([]*core.Generator, error) {
	result, err := a.db.List(context.Background(), GeneratorsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var generators []*core.Generator
	for _, doc := range result {
		generator, err := a.fieldsToGenerator(doc.Fields)
		if err == nil && generator.ColonyName == colonyName {
			generators = append(generators, generator)
		}
	}
	
	return generators, nil
}

func (a *ColonyOSAdapter) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	doc, err := a.db.Get(context.Background(), GeneratorsCollection, generatorID)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToGenerator(doc.Fields)
}

func (a *ColonyOSAdapter) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	generators, err := a.GetGenerators(colonyName)
	if err != nil {
		return nil, err
	}
	
	for _, generator := range generators {
		if generator.Name == name {
			return generator, nil
		}
	}
	
	return nil, fmt.Errorf("generator not found")
}

func (a *ColonyOSAdapter) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	generators, err := a.GetGenerators(colonyName)
	if err != nil {
		return nil, err
	}
	
	if count > 0 && len(generators) > count {
		generators = generators[:count]
	}
	
	return generators, nil
}

func (a *ColonyOSAdapter) FindAllGenerators() ([]*core.Generator, error) {
	result, err := a.db.List(context.Background(), GeneratorsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var generators []*core.Generator
	for _, doc := range result {
		generator, err := a.fieldsToGenerator(doc.Fields)
		if err == nil {
			generators = append(generators, generator)
		}
	}
	
	return generators, nil
}

func (a *ColonyOSAdapter) UpdateGenerator(generator *core.Generator) error {
	fields := a.generatorToFields(generator)
	_, err := a.db.Update(context.Background(), GeneratorsCollection, generator.ID, fields)
	return err
}

func (a *ColonyOSAdapter) RemoveGeneratorByID(generatorID string) error {
	return a.db.Delete(context.Background(), GeneratorsCollection, generatorID)
}

func (a *ColonyOSAdapter) RemoveGeneratorByName(colonyName, generatorName string) error {
	generator, err := a.GetGeneratorByName(colonyName, generatorName)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), GeneratorsCollection, generator.ID)
}

func (a *ColonyOSAdapter) RemoveAllGeneratorsByColonyName(colonyName string) error {
	generators, err := a.GetGenerators(colonyName)
	if err != nil {
		return err
	}
	
	for _, generator := range generators {
		if err := a.db.Delete(context.Background(), GeneratorsCollection, generator.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) CountGeneratorsByColonyName(colonyName string) (int, error) {
	generators, err := a.GetGenerators(colonyName)
	if err != nil {
		return 0, err
	}
	
	return len(generators), nil
}

// Generator Args - simplified implementation
func (a *ColonyOSAdapter) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	// Simplified - would need actual GeneratorArg collection
	return nil
}

func (a *ColonyOSAdapter) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	// Simplified - would need actual GeneratorArg collection
	return []*core.GeneratorArg{}, nil
}

func (a *ColonyOSAdapter) CountGeneratorArgs(generatorID string) (int, error) {
	// Simplified - would need actual GeneratorArg collection
	return 0, nil
}

func (a *ColonyOSAdapter) RemoveGeneratorArgByID(generatorArgsID string) error {
	// Simplified - would need actual GeneratorArg collection
	return nil
}

func (a *ColonyOSAdapter) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error {
	// Simplified - would need actual GeneratorArg collection
	return nil
}

func (a *ColonyOSAdapter) RemoveAllGeneratorArgsByColonyName(generatorID string) error {
	// Simplified - would need actual GeneratorArg collection
	return nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) generatorToFields(generator *core.Generator) map[string]interface{} {
	fields := map[string]interface{}{
		"id":             generator.ID,
		"initiator_id":   generator.InitiatorID,
		"initiator_name": generator.InitiatorName,
		"colony_name":    generator.ColonyName,
		"name":           generator.Name,
		"workflow_spec":  generator.WorkflowSpec,
		"trigger":        generator.Trigger,
		"timeout":        generator.Timeout,
		"first_pack":     generator.FirstPack,
		"last_run":       generator.LastRun,
		"queue_size":     generator.QueueSize,
		"checker_period": generator.CheckerPeriod,
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToGenerator(fields map[string]interface{}) (*core.Generator, error) {
	generator := &core.Generator{}
	
	if id, ok := fields["id"].(string); ok {
		generator.ID = id
	}
	if initiatorID, ok := fields["initiator_id"].(string); ok {
		generator.InitiatorID = initiatorID
	}
	if initiatorName, ok := fields["initiator_name"].(string); ok {
		generator.InitiatorName = initiatorName
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		generator.ColonyName = colonyName
	}
	if name, ok := fields["name"].(string); ok {
		generator.Name = name
	}
	if workflowSpec, ok := fields["workflow_spec"].(string); ok {
		generator.WorkflowSpec = workflowSpec
	}
	if trigger, ok := fields["trigger"].(int); ok {
		generator.Trigger = trigger
	} else if trigger, ok := fields["trigger"].(float64); ok {
		generator.Trigger = int(trigger)
	}
	if timeout, ok := fields["timeout"].(int); ok {
		generator.Timeout = timeout
	} else if timeout, ok := fields["timeout"].(float64); ok {
		generator.Timeout = int(timeout)
	}
	if queueSize, ok := fields["queue_size"].(int); ok {
		generator.QueueSize = queueSize
	} else if queueSize, ok := fields["queue_size"].(float64); ok {
		generator.QueueSize = int(queueSize)
	}
	if checkerPeriod, ok := fields["checker_period"].(int); ok {
		generator.CheckerPeriod = checkerPeriod
	} else if checkerPeriod, ok := fields["checker_period"].(float64); ok {
		generator.CheckerPeriod = int(checkerPeriod)
	}
	// Simplified time handling - would properly parse timestamps
	
	return generator, nil
}