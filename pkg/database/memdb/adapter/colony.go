package adapter

import (
	"context"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// ColonyDatabase interface implementation

func (a *ColonyOSAdapter) AddColony(colony *core.Colony) error {
	doc := &memdb.VelocityDocument{
		ID:     colony.ID,
		Fields: a.colonyToFields(colony),
	}
	
	return a.db.Insert(context.Background(), ColoniesCollection, doc)
}

func (a *ColonyOSAdapter) GetColonies() ([]*core.Colony, error) {
	result, err := a.db.List(context.Background(), ColoniesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	colonies := make([]*core.Colony, 0, len(result))
	for _, doc := range result {
		colony, err := a.fieldsToColony(doc.Fields)
		if err == nil {
			colonies = append(colonies, colony)
		}
	}
	
	return colonies, nil
}

func (a *ColonyOSAdapter) GetColonyByID(id string) (*core.Colony, error) {
	doc, err := a.db.Get(context.Background(), ColoniesCollection, id)
	if err != nil {
		return nil, err
	}
	
	return a.fieldsToColony(doc.Fields)
}

func (a *ColonyOSAdapter) GetColonyByName(name string) (*core.Colony, error) {
	// Simple linear search through all colonies to find by name
	colonies, err := a.GetColonies()
	if err != nil {
		return nil, err
	}
	
	for _, colony := range colonies {
		if colony.Name == name {
			return colony, nil
		}
	}
	
	return nil, fmt.Errorf("colony not found")
}

func (a *ColonyOSAdapter) RenameColony(colonyName string, newColonyName string) error {
	colony, err := a.GetColonyByName(colonyName)
	if err != nil {
		return err
	}
	
	colony.Name = newColonyName
	fields := a.colonyToFields(colony)
	
	_, err = a.db.Update(context.Background(), ColoniesCollection, colony.ID, fields)
	return err
}

func (a *ColonyOSAdapter) RemoveColonyByName(colonyName string) error {
	colony, err := a.GetColonyByName(colonyName)
	if err != nil {
		return err
	}
	
	return a.db.Delete(context.Background(), ColoniesCollection, colony.ID)
}

func (a *ColonyOSAdapter) CountColonies() (int, error) {
	result, err := a.db.List(context.Background(), ColoniesCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) colonyToFields(colony *core.Colony) map[string]interface{} {
	return map[string]interface{}{
		"id":   colony.ID,
		"name": colony.Name,
	}
}

func (a *ColonyOSAdapter) fieldsToColony(fields map[string]interface{}) (*core.Colony, error) {
	colony := &core.Colony{}
	
	if id, ok := fields["id"].(string); ok {
		colony.ID = id
	}
	if name, ok := fields["name"].(string); ok {
		colony.Name = name
	}
	
	return colony, nil
}