package embedded

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func copyColony(c *core.Colony) *core.Colony {
	cp := *c
	return &cp
}

func (db *EmbeddedDatabase) AddColony(colony *core.Colony) error {
	if colony == nil {
		return errors.New("Colony is nil")
	}

	existing, _ := db.GetColonyByName(colony.Name)
	if existing != nil {
		return errors.New("Colony with name <" + colony.Name + "> already exists")
	}

	cp := copyColony(colony)
	if err := db.colonies.Put(cp.Name, cp); err != nil {
		return err
	}

	db.coloniesIdx.byID.Add(cp.Name, cp.ID)

	return nil
}

func (db *EmbeddedDatabase) GetColonies() ([]*core.Colony, error) {
	all := db.colonies.All()
	result := make([]*core.Colony, len(all))
	for i, c := range all {
		result[i] = copyColony(c)
	}
	return result, nil
}

func (db *EmbeddedDatabase) GetColonyByID(id string) (*core.Colony, error) {
	names := db.coloniesIdx.byID.Lookup(id)
	if len(names) == 0 {
		return nil, nil
	}

	colony, ok := db.colonies.Get(names[0])
	if !ok {
		return nil, nil
	}

	return copyColony(colony), nil
}

func (db *EmbeddedDatabase) GetColonyByName(name string) (*core.Colony, error) {
	colony, ok := db.colonies.Get(name)
	if !ok {
		return nil, nil
	}

	return copyColony(colony), nil
}

func (db *EmbeddedDatabase) RenameColony(colonyName string, newName string) error {
	colony, ok := db.colonies.Get(colonyName)
	if !ok {
		return errors.New("Colony does not exist")
	}

	// Remove old entry
	db.coloniesIdx.byID.Remove(colonyName, colony.ID)
	if err := db.colonies.Delete(colonyName); err != nil {
		return err
	}

	// Add with new name
	cp := copyColony(colony)
	cp.Name = newName
	if err := db.colonies.Put(newName, cp); err != nil {
		return err
	}
	db.coloniesIdx.byID.Add(newName, cp.ID)

	return nil
}

func (db *EmbeddedDatabase) RemoveColonyByName(colonyName string) error {
	colony, ok := db.colonies.Get(colonyName)
	if !ok {
		return errors.New("Colony does not exist")
	}

	// Cascade delete all dependent entities
	if err := db.RemoveUsersByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveExecutorsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveLocationsByColonyName(colonyName); err != nil {
		return err
	}

	// Remove colony itself
	db.coloniesIdx.byID.Remove(colonyName, colony.ID)
	if err := db.colonies.Delete(colonyName); err != nil {
		return err
	}

	if err := db.RemoveAllProcessesByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveAllProcessGraphsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveAllGeneratorsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveAllCronsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveFunctionsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveLogsByColonyName(colonyName); err != nil {
		return err
	}

	if err := db.RemoveSnapshotsByColonyName(colonyName); err != nil {
		return err
	}

	return nil
}

func (db *EmbeddedDatabase) CountColonies() (int, error) {
	return db.colonies.Len(), nil
}
