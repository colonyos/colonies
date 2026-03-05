package embedded

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
)

// BlueprintDefinition methods

func copyBlueprintDefinition(sd *core.BlueprintDefinition) *core.BlueprintDefinition {
	cp := *sd

	// Deep copy ShortNames slice
	if sd.Spec.Names.ShortNames != nil {
		cp.Spec.Names.ShortNames = make([]string, len(sd.Spec.Names.ShortNames))
		copy(cp.Spec.Names.ShortNames, sd.Spec.Names.ShortNames)
	}

	// Deep copy Schema pointer and its contents
	if sd.Spec.Schema != nil {
		schemaCopy := *sd.Spec.Schema
		if sd.Spec.Schema.Required != nil {
			schemaCopy.Required = make([]string, len(sd.Spec.Schema.Required))
			copy(schemaCopy.Required, sd.Spec.Schema.Required)
		}
		if sd.Spec.Schema.Properties != nil {
			schemaCopy.Properties = make(map[string]core.SchemaProperty, len(sd.Spec.Schema.Properties))
			for k, v := range sd.Spec.Schema.Properties {
				propCopy := v
				if v.Enum != nil {
					propCopy.Enum = make([]interface{}, len(v.Enum))
					copy(propCopy.Enum, v.Enum)
				}
				if v.Properties != nil {
					propCopy.Properties = make(map[string]core.SchemaProperty, len(v.Properties))
					for pk, pv := range v.Properties {
						propCopy.Properties[pk] = pv
					}
				}
				if v.Items != nil {
					itemsCopy := *v.Items
					propCopy.Items = &itemsCopy
				}
				schemaCopy.Properties[k] = propCopy
			}
		}
		cp.Spec.Schema = &schemaCopy
	}

	return &cp
}

func (db *EmbeddedDatabase) AddBlueprintDefinition(sd *core.BlueprintDefinition) error {
	if sd == nil {
		return errors.New("BlueprintDefinition is nil")
	}

	// Enforce unique (ColonyName, Name) constraint
	key := sd.Metadata.ColonyName + ":" + sd.Metadata.Name
	existing := db.blueprintDefsIdx.byName.Lookup(key)
	if len(existing) > 0 {
		return fmt.Errorf("BlueprintDefinition with name %s already exists in colony %s", sd.Metadata.Name, sd.Metadata.ColonyName)
	}

	cp := copyBlueprintDefinition(sd)
	if err := db.blueprintDefs.Put(cp.ID, cp); err != nil {
		return err
	}

	db.blueprintDefsIdx.byNamespace.Add(sd.ID, sd.Metadata.ColonyName)
	db.blueprintDefsIdx.byName.Add(sd.ID, sd.Metadata.ColonyName+":"+sd.Metadata.Name)
	db.blueprintDefsIdx.byKind.Add(sd.ID, sd.Spec.Names.Kind)

	return nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitionByID(id string) (*core.BlueprintDefinition, error) {
	sd, ok := db.blueprintDefs.Get(id)
	if !ok {
		return nil, nil
	}
	return copyBlueprintDefinition(sd), nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitionByName(namespace, name string) (*core.BlueprintDefinition, error) {
	ids := db.blueprintDefsIdx.byName.Lookup(namespace + ":" + name)
	if len(ids) == 0 {
		return nil, nil
	}
	sd, ok := db.blueprintDefs.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyBlueprintDefinition(sd), nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitions() ([]*core.BlueprintDefinition, error) {
	result := db.blueprintDefs.All()
	for i, sd := range result {
		result[i] = copyBlueprintDefinition(sd)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitionsByNamespace(namespace string) ([]*core.BlueprintDefinition, error) {
	ids := db.blueprintDefsIdx.byNamespace.Lookup(namespace)
	var result []*core.BlueprintDefinition
	for _, id := range ids {
		if sd, ok := db.blueprintDefs.Get(id); ok {
			result = append(result, copyBlueprintDefinition(sd))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitionsByGroup(group string) ([]*core.BlueprintDefinition, error) {
	all := db.blueprintDefs.All()
	var result []*core.BlueprintDefinition
	for _, sd := range all {
		if sd.Spec.Group == group {
			result = append(result, copyBlueprintDefinition(sd))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintDefinitionByKind(kind string) (*core.BlueprintDefinition, error) {
	ids := db.blueprintDefsIdx.byKind.Lookup(kind)
	if len(ids) == 0 {
		return nil, nil
	}
	sd, ok := db.blueprintDefs.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyBlueprintDefinition(sd), nil
}

func (db *EmbeddedDatabase) UpdateBlueprintDefinition(sd *core.BlueprintDefinition) error {
	if sd == nil {
		return errors.New("BlueprintDefinition is nil")
	}

	old, ok := db.blueprintDefs.Get(sd.ID)
	if ok {
		db.blueprintDefsIdx.byNamespace.Remove(sd.ID, old.Metadata.ColonyName)
		db.blueprintDefsIdx.byName.Remove(sd.ID, old.Metadata.ColonyName+":"+old.Metadata.Name)
		db.blueprintDefsIdx.byKind.Remove(sd.ID, old.Spec.Names.Kind)
	}

	cp := copyBlueprintDefinition(sd)
	if err := db.blueprintDefs.Put(cp.ID, cp); err != nil {
		return err
	}

	db.blueprintDefsIdx.byNamespace.Add(sd.ID, sd.Metadata.ColonyName)
	db.blueprintDefsIdx.byName.Add(sd.ID, sd.Metadata.ColonyName+":"+sd.Metadata.Name)
	db.blueprintDefsIdx.byKind.Add(sd.ID, sd.Spec.Names.Kind)

	return nil
}

func (db *EmbeddedDatabase) RemoveBlueprintDefinitionByID(id string) error {
	sd, ok := db.blueprintDefs.Get(id)
	if !ok {
		return nil
	}

	db.blueprintDefsIdx.byNamespace.Remove(id, sd.Metadata.ColonyName)
	db.blueprintDefsIdx.byName.Remove(id, sd.Metadata.ColonyName+":"+sd.Metadata.Name)
	db.blueprintDefsIdx.byKind.Remove(id, sd.Spec.Names.Kind)

	return db.blueprintDefs.Delete(id)
}

func (db *EmbeddedDatabase) RemoveBlueprintDefinitionByName(namespace, name string) error {
	ids := db.blueprintDefsIdx.byName.Lookup(namespace + ":" + name)
	for _, id := range ids {
		if err := db.RemoveBlueprintDefinitionByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) CountBlueprintDefinitions() (int, error) {
	return db.blueprintDefs.Len(), nil
}

// Blueprint methods

func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func deepCopyInterface(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return copyInterfaceMap(val)
	case []interface{}:
		cp := make([]interface{}, len(val))
		for i, item := range val {
			cp[i] = deepCopyInterface(item)
		}
		return cp
	default:
		return v
	}
}

func copyInterfaceMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	cp := make(map[string]interface{}, len(m))
	for k, v := range m {
		cp[k] = deepCopyInterface(v)
	}
	return cp
}

func copyBlueprint(b *core.Blueprint) *core.Blueprint {
	cp := *b
	cp.Spec = copyInterfaceMap(b.Spec)
	cp.Status = copyInterfaceMap(b.Status)
	cp.Metadata.Labels = copyStringMap(b.Metadata.Labels)
	cp.Metadata.Annotations = copyStringMap(b.Metadata.Annotations)
	return &cp
}

func (db *EmbeddedDatabase) AddBlueprint(blueprint *core.Blueprint) error {
	if blueprint == nil {
		return errors.New("Blueprint is nil")
	}

	existingBlueprint, err := db.GetBlueprintByName(blueprint.Metadata.ColonyName, blueprint.Metadata.Name)
	if err != nil {
		return err
	}

	if existingBlueprint != nil {
		return errors.New("Blueprint with name <" + blueprint.Metadata.Name + "> in namespace <" + blueprint.Metadata.ColonyName + "> already exists")
	}

	cp := copyBlueprint(blueprint)
	if err := db.blueprints.Put(cp.ID, cp); err != nil {
		return err
	}

	db.blueprintsIdx.byNamespace.Add(blueprint.ID, blueprint.Metadata.ColonyName)
	db.blueprintsIdx.byName.Add(blueprint.ID, blueprint.Metadata.ColonyName+":"+blueprint.Metadata.Name)
	db.blueprintsIdx.byKind.Add(blueprint.ID, blueprint.Kind)

	return nil
}

func (db *EmbeddedDatabase) GetBlueprintByID(id string) (*core.Blueprint, error) {
	b, ok := db.blueprints.Get(id)
	if !ok {
		return nil, nil
	}
	return copyBlueprint(b), nil
}

func (db *EmbeddedDatabase) GetBlueprintByName(namespace, name string) (*core.Blueprint, error) {
	ids := db.blueprintsIdx.byName.Lookup(namespace + ":" + name)
	if len(ids) == 0 {
		return nil, nil
	}
	b, ok := db.blueprints.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyBlueprint(b), nil
}

func (db *EmbeddedDatabase) GetBlueprints() ([]*core.Blueprint, error) {
	result := db.blueprints.All()
	for i, b := range result {
		result[i] = copyBlueprint(b)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.ColonyName != result[j].Metadata.ColonyName {
			return result[i].Metadata.ColonyName < result[j].Metadata.ColonyName
		}
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintsByNamespace(namespace string) ([]*core.Blueprint, error) {
	ids := db.blueprintsIdx.byNamespace.Lookup(namespace)
	var result []*core.Blueprint
	for _, id := range ids {
		if b, ok := db.blueprints.Get(id); ok {
			result = append(result, copyBlueprint(b))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintsByKind(kind string) ([]*core.Blueprint, error) {
	ids := db.blueprintsIdx.byKind.Lookup(kind)
	var result []*core.Blueprint
	for _, id := range ids {
		if b, ok := db.blueprints.Get(id); ok {
			result = append(result, copyBlueprint(b))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.ColonyName != result[j].Metadata.ColonyName {
			return result[i].Metadata.ColonyName < result[j].Metadata.ColonyName
		}
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintsByNamespaceAndKind(namespace, kind string) ([]*core.Blueprint, error) {
	ids := db.blueprintsIdx.byNamespace.Lookup(namespace)
	var result []*core.Blueprint
	for _, id := range ids {
		if b, ok := db.blueprints.Get(id); ok {
			if b.Kind == kind {
				result = append(result, copyBlueprint(b))
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintsByNamespaceKindAndLocation(namespace, kind, locationName string) ([]*core.Blueprint, error) {
	if locationName == "" {
		return db.GetBlueprintsByNamespaceAndKind(namespace, kind)
	}

	ids := db.blueprintsIdx.byNamespace.Lookup(namespace)
	var result []*core.Blueprint
	for _, id := range ids {
		if b, ok := db.blueprints.Get(id); ok {
			if b.Kind == kind && strings.EqualFold(b.Metadata.LocationName, locationName) {
				result = append(result, copyBlueprint(b))
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})
	return result, nil
}

func (db *EmbeddedDatabase) UpdateBlueprint(blueprint *core.Blueprint) error {
	if blueprint == nil {
		return errors.New("Blueprint is nil")
	}

	old, ok := db.blueprints.Get(blueprint.ID)
	if ok {
		db.blueprintsIdx.byNamespace.Remove(blueprint.ID, old.Metadata.ColonyName)
		db.blueprintsIdx.byName.Remove(blueprint.ID, old.Metadata.ColonyName+":"+old.Metadata.Name)
		db.blueprintsIdx.byKind.Remove(blueprint.ID, old.Kind)
	}

	cp := copyBlueprint(blueprint)
	if err := db.blueprints.Put(cp.ID, cp); err != nil {
		return err
	}

	db.blueprintsIdx.byNamespace.Add(blueprint.ID, blueprint.Metadata.ColonyName)
	db.blueprintsIdx.byName.Add(blueprint.ID, blueprint.Metadata.ColonyName+":"+blueprint.Metadata.Name)
	db.blueprintsIdx.byKind.Add(blueprint.ID, blueprint.Kind)

	return nil
}

func (db *EmbeddedDatabase) UpdateBlueprintStatus(id string, status map[string]interface{}) error {
	b, ok := db.blueprints.Get(id)
	if !ok {
		return errors.New("Blueprint not found")
	}

	cp := copyBlueprint(b)

	// Marshal/unmarshal status to ensure clean copy (matching PQ atomicity)
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return err
	}
	var newStatus map[string]interface{}
	if err := json.Unmarshal(statusJSON, &newStatus); err != nil {
		return err
	}

	cp.Status = newStatus

	return db.blueprints.Put(id, cp)
}

func (db *EmbeddedDatabase) RemoveBlueprintByID(id string) error {
	b, ok := db.blueprints.Get(id)
	if !ok {
		return nil
	}

	db.blueprintsIdx.byNamespace.Remove(id, b.Metadata.ColonyName)
	db.blueprintsIdx.byName.Remove(id, b.Metadata.ColonyName+":"+b.Metadata.Name)
	db.blueprintsIdx.byKind.Remove(id, b.Kind)

	return db.blueprints.Delete(id)
}

func (db *EmbeddedDatabase) RemoveBlueprintByName(namespace, name string) error {
	ids := db.blueprintsIdx.byName.Lookup(namespace + ":" + name)
	for _, id := range ids {
		if err := db.RemoveBlueprintByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveBlueprintsByNamespace(namespace string) error {
	ids := db.blueprintsIdx.byNamespace.Lookup(namespace)
	for _, id := range ids {
		if err := db.RemoveBlueprintByID(id); err != nil {
			return err
		}
	}
	return nil
}

func (db *EmbeddedDatabase) CountBlueprints() (int, error) {
	return db.blueprints.Len(), nil
}

func (db *EmbeddedDatabase) CountBlueprintsByNamespace(namespace string) (int, error) {
	return db.blueprintsIdx.byNamespace.Count(namespace), nil
}

// BlueprintHistory methods

func copyBlueprintHistory(h *core.BlueprintHistory) *core.BlueprintHistory {
	cp := *h
	cp.Spec = copyInterfaceMap(h.Spec)
	cp.Status = copyInterfaceMap(h.Status)
	return &cp
}

func (db *EmbeddedDatabase) AddBlueprintHistory(history *core.BlueprintHistory) error {
	cp := copyBlueprintHistory(history)
	if err := db.blueprintHistory.Put(cp.ID, cp); err != nil {
		return err
	}

	db.blueprintHistoryIdx.byBlueprint.Add(history.ID, history.BlueprintID)

	return nil
}

func (db *EmbeddedDatabase) GetBlueprintHistory(blueprintID string, limit int) ([]*core.BlueprintHistory, error) {
	ids := db.blueprintHistoryIdx.byBlueprint.Lookup(blueprintID)
	var result []*core.BlueprintHistory
	for _, id := range ids {
		if h, ok := db.blueprintHistory.Get(id); ok {
			result = append(result, copyBlueprintHistory(h))
		}
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (db *EmbeddedDatabase) GetBlueprintHistoryByGeneration(blueprintID string, generation int64) (*core.BlueprintHistory, error) {
	ids := db.blueprintHistoryIdx.byBlueprint.Lookup(blueprintID)
	for _, id := range ids {
		if h, ok := db.blueprintHistory.Get(id); ok {
			if h.Generation == generation {
				return copyBlueprintHistory(h), nil
			}
		}
	}
	return nil, nil
}

func (db *EmbeddedDatabase) RemoveBlueprintHistory(blueprintID string) error {
	ids := db.blueprintHistoryIdx.byBlueprint.Lookup(blueprintID)
	for _, id := range ids {
		db.blueprintHistoryIdx.byBlueprint.Remove(id, blueprintID)
		if err := db.blueprintHistory.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
