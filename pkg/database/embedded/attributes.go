package embedded

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *EmbeddedDatabase) AddAttribute(attribute core.Attribute) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.addAttribute(attribute)
}

func (db *EmbeddedDatabase) addAttribute(attribute core.Attribute) error {
	if err := db.attributes.Put(attribute.ID, &attribute); err != nil {
		return err
	}
	db.attributesIdx.byTarget.Add(attribute.ID, attribute.TargetID)
	db.attributesIdx.byColony.Add(attribute.ID, attribute.TargetColonyName)
	if attribute.TargetProcessGraphID != "" {
		db.attributesIdx.byGraph.Add(attribute.ID, attribute.TargetProcessGraphID)
	}
	return nil
}

func (db *EmbeddedDatabase) AddAttributes(attributes []core.Attribute) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.addAttributes(attributes)
}

func (db *EmbeddedDatabase) addAttributes(attributes []core.Attribute) error {
	for _, attr := range attributes {
		if err := db.addAttribute(attr); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) GetAttributeByID(attributeID string) (core.Attribute, error) {
	a, ok := db.attributes.Get(attributeID)
	if !ok {
		return core.Attribute{}, errors.New("Attribute does not exist")
	}
	return *a, nil
}

func (db *EmbeddedDatabase) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	ids := db.attributesIdx.byColony.Lookup(colonyName)
	result := make([]core.Attribute, 0, len(ids))
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (db *EmbeddedDatabase) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.Key == key && a.AttributeType == attributeType {
				return *a, nil
			}
		}
	}
	return core.Attribute{}, errors.New("Attribute does not exist")
}

func (db *EmbeddedDatabase) GetAttributes(targetID string) ([]core.Attribute, error) {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	result := make([]core.Attribute, 0, len(ids))
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (db *EmbeddedDatabase) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	var result []core.Attribute
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.AttributeType == attributeType {
				result = append(result, *a)
			}
		}
	}
	if result == nil {
		result = make([]core.Attribute, 0)
	}
	return result, nil
}

func (db *EmbeddedDatabase) UpdateAttribute(attribute core.Attribute) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	existing, ok := db.attributes.Get(attribute.ID)
	if !ok {
		return errors.New("Attribute does not exist")
	}
	cp := *existing
	cp.Value = attribute.Value
	return db.attributes.Put(attribute.ID, &cp)
}

func (db *EmbeddedDatabase) RemoveAttributeByID(attributeID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	a, ok := db.attributes.Get(attributeID)
	if !ok {
		return nil
	}
	db.removeAttributeFromIndexes(attributeID, a)
	return db.attributes.Delete(attributeID)
}

func (db *EmbeddedDatabase) RemoveAllAttributesByColonyName(colonyName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesByColonyName(colonyName)
}

func (db *EmbeddedDatabase) removeAllAttributesByColonyName(colonyName string) error {
	ids := db.attributesIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			db.removeAttributeFromIndexes(id, a)
			db.attributes.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesByColonyNameWithState(colonyName, state)
}

func (db *EmbeddedDatabase) removeAllAttributesByColonyNameWithState(colonyName string, state int) error {
	ids := db.attributesIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.TargetProcessGraphID == "" && a.State == state {
				db.removeAttributeFromIndexes(id, a)
				db.attributes.Delete(id)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesByProcessGraphID(processGraphID)
}

func (db *EmbeddedDatabase) removeAllAttributesByProcessGraphID(processGraphID string) error {
	ids := db.attributesIdx.byGraph.Lookup(processGraphID)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			db.removeAttributeFromIndexes(id, a)
			db.attributes.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesInProcessGraphsByColonyName(colonyName)
}

func (db *EmbeddedDatabase) removeAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	ids := db.attributesIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.TargetProcessGraphID != "" {
				db.removeAttributeFromIndexes(id, a)
				db.attributes.Delete(id)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesInProcessGraphsByColonyNameWithState(colonyName, state)
}

func (db *EmbeddedDatabase) removeAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	ids := db.attributesIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.TargetProcessGraphID != "" && a.State == state {
				db.removeAttributeFromIndexes(id, a)
				db.attributes.Delete(id)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAttributesByTargetID(targetID, attributeType)
}

func (db *EmbeddedDatabase) removeAttributesByTargetID(targetID string, attributeType int) error {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			if a.AttributeType == attributeType {
				db.removeAttributeFromIndexes(id, a)
				db.attributes.Delete(id)
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributesByTargetID(targetID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributesByTargetID(targetID)
}

func (db *EmbeddedDatabase) removeAllAttributesByTargetID(targetID string) error {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			db.removeAttributeFromIndexes(id, a)
			db.attributes.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveAllAttributes() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.removeAllAttributes()
}

func (db *EmbeddedDatabase) removeAllAttributes() error {
	for _, a := range db.attributes.All() {
		db.attributes.Delete(a.ID)
	}
	db.attributesIdx.byTarget.Clear()
	db.attributesIdx.byColony.Clear()
	db.attributesIdx.byGraph.Clear()
	return nil
}

// setAttributeState updates the state of all attributes for the given targetID.
func (db *EmbeddedDatabase) setAttributeState(targetID string, state int) {
	ids := db.attributesIdx.byTarget.Lookup(targetID)
	for _, id := range ids {
		if a, ok := db.attributes.Get(id); ok {
			cp := *a
			cp.State = state
			db.attributes.Put(id, &cp)
		}
	}
}

// removeAttributeFromIndexes removes a single attribute from all indexes.
func (db *EmbeddedDatabase) removeAttributeFromIndexes(id string, a *core.Attribute) {
	db.attributesIdx.byTarget.Remove(id, a.TargetID)
	db.attributesIdx.byColony.Remove(id, a.TargetColonyName)
	if a.TargetProcessGraphID != "" {
		db.attributesIdx.byGraph.Remove(id, a.TargetProcessGraphID)
	}
}
