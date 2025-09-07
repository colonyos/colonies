package adapter

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// AttributeDatabase interface implementation

func (a *ColonyOSAdapter) AddAttribute(attribute core.Attribute) error {
	doc := &memdb.VelocityDocument{
		ID:     attribute.ID,
		Fields: a.attributeToFields(attribute),
	}
	
	return a.db.Insert(context.Background(), AttributesCollection, doc)
}

func (a *ColonyOSAdapter) AddAttributes(attributes []core.Attribute) error {
	for _, attr := range attributes {
		if err := a.AddAttribute(attr); err != nil {
			return err
		}
	}
	return nil
}

func (a *ColonyOSAdapter) GetAttributeByID(attributeID string) (core.Attribute, error) {
	doc, err := a.db.Get(context.Background(), AttributesCollection, attributeID)
	if err != nil {
		return core.Attribute{}, err
	}
	
	return a.fieldsToAttribute(doc.Fields)
}

func (a *ColonyOSAdapter) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	result, err := a.db.List(context.Background(), AttributesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var attributes []core.Attribute
	for _, doc := range result {
		attr, err := a.fieldsToAttribute(doc.Fields)
		if err == nil && attr.TargetColonyName == colonyName {
			attributes = append(attributes, attr)
		}
	}
	
	return attributes, nil
}

func (a *ColonyOSAdapter) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	result, err := a.db.List(context.Background(), AttributesCollection, 1000, 0)
	if err != nil {
		return core.Attribute{}, err
	}
	
	for _, doc := range result {
		attr, err := a.fieldsToAttribute(doc.Fields)
		if err == nil && attr.TargetID == targetID && attr.Key == key && attr.AttributeType == attributeType {
			return attr, nil
		}
	}
	
	return core.Attribute{}, nil
}

func (a *ColonyOSAdapter) GetAttributes(targetID string) ([]core.Attribute, error) {
	result, err := a.db.List(context.Background(), AttributesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var attributes []core.Attribute
	for _, doc := range result {
		attr, err := a.fieldsToAttribute(doc.Fields)
		if err == nil && attr.TargetID == targetID {
			attributes = append(attributes, attr)
		}
	}
	
	return attributes, nil
}

func (a *ColonyOSAdapter) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	result, err := a.db.List(context.Background(), AttributesCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var attributes []core.Attribute
	for _, doc := range result {
		attr, err := a.fieldsToAttribute(doc.Fields)
		if err == nil && attr.TargetID == targetID && attr.AttributeType == attributeType {
			attributes = append(attributes, attr)
		}
	}
	
	return attributes, nil
}

func (a *ColonyOSAdapter) UpdateAttribute(attribute core.Attribute) error {
	fields := a.attributeToFields(attribute)
	_, err := a.db.Update(context.Background(), AttributesCollection, attribute.ID, fields)
	return err
}

func (a *ColonyOSAdapter) RemoveAttributeByID(attributeID string) error {
	return a.db.Delete(context.Background(), AttributesCollection, attributeID)
}

func (a *ColonyOSAdapter) RemoveAllAttributesByColonyName(colonyName string) error {
	attributes, err := a.GetAttributesByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	attributes, err := a.GetAttributesByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if attr.State == state {
			if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	result, err := a.db.List(context.Background(), AttributesCollection, 1000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		attr, err := a.fieldsToAttribute(doc.Fields)
		if err == nil && attr.TargetProcessGraphID == processGraphID {
			if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	attributes, err := a.GetAttributesByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if attr.TargetProcessGraphID != "" {
			if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	attributes, err := a.GetAttributesByColonyName(colonyName)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if attr.TargetProcessGraphID != "" && attr.State == state {
			if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	attributes, err := a.GetAttributesByType(targetID, attributeType)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributesByTargetID(targetID string) error {
	attributes, err := a.GetAttributes(targetID)
	if err != nil {
		return err
	}
	
	for _, attr := range attributes {
		if err := a.db.Delete(context.Background(), AttributesCollection, attr.ID); err != nil {
			return err
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllAttributes() error {
	result, err := a.db.List(context.Background(), AttributesCollection, 10000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		if err := a.db.Delete(context.Background(), AttributesCollection, doc.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) attributeToFields(attr core.Attribute) map[string]interface{} {
	fields := map[string]interface{}{
		"id":                     attr.ID,
		"target_id":              attr.TargetID,
		"target_colony_name":     attr.TargetColonyName,
		"target_process_graph_id": attr.TargetProcessGraphID,
		"key":                    attr.Key,
		"value":                  attr.Value,
		"attribute_type":         attr.AttributeType,
		"state":                  attr.State,
	}
	
	return fields
}

func (a *ColonyOSAdapter) fieldsToAttribute(fields map[string]interface{}) (core.Attribute, error) {
	attr := core.Attribute{}
	
	if id, ok := fields["id"].(string); ok {
		attr.ID = id
	}
	if targetID, ok := fields["target_id"].(string); ok {
		attr.TargetID = targetID
	}
	if targetColonyName, ok := fields["target_colony_name"].(string); ok {
		attr.TargetColonyName = targetColonyName
	}
	if targetProcessGraphID, ok := fields["target_process_graph_id"].(string); ok {
		attr.TargetProcessGraphID = targetProcessGraphID
	}
	if key, ok := fields["key"].(string); ok {
		attr.Key = key
	}
	if value, ok := fields["value"].(string); ok {
		attr.Value = value
	}
	if attrType, ok := fields["attribute_type"].(int); ok {
		attr.AttributeType = attrType
	} else if attrType, ok := fields["attribute_type"].(float64); ok {
		attr.AttributeType = int(attrType)
	}
	if state, ok := fields["state"].(int); ok {
		attr.State = state
	} else if state, ok := fields["state"].(float64); ok {
		attr.State = int(state)
	}
	
	return attr, nil
}