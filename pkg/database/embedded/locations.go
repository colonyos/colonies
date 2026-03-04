package embedded

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func copyLocation(l *core.Location) *core.Location {
	cp := *l
	return &cp
}

func (db *EmbeddedDatabase) AddLocation(location *core.Location) error {
	if location == nil {
		return errors.New("Location is nil")
	}

	existing, err := db.GetLocationByName(location.ColonyName, location.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("Location with name <" + location.Name + "> already exists in Colony with name <" + location.ColonyName + ">")
	}

	cp := copyLocation(location)
	if err := db.locations.Put(cp.ID, cp); err != nil {
		return err
	}

	db.locationsIdx.byColony.Add(location.ID, location.ColonyName)
	db.locationsIdx.byName.Add(location.ID, location.ColonyName+":"+location.Name)

	return nil
}

func (db *EmbeddedDatabase) GetLocationsByColonyName(colonyName string) ([]*core.Location, error) {
	ids := db.locationsIdx.byColony.Lookup(colonyName)
	var locations []*core.Location
	for _, id := range ids {
		if l, ok := db.locations.Get(id); ok {
			locations = append(locations, copyLocation(l))
		}
	}
	return locations, nil
}

func (db *EmbeddedDatabase) GetLocationByID(locationID string) (*core.Location, error) {
	l, ok := db.locations.Get(locationID)
	if !ok {
		return nil, nil
	}
	return copyLocation(l), nil
}

func (db *EmbeddedDatabase) GetLocationByName(colonyName string, name string) (*core.Location, error) {
	ids := db.locationsIdx.byName.Lookup(colonyName + ":" + name)
	if len(ids) == 0 {
		return nil, nil
	}
	l, ok := db.locations.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyLocation(l), nil
}

func (db *EmbeddedDatabase) RemoveLocationByID(locationID string) error {
	l, ok := db.locations.Get(locationID)
	if !ok {
		return nil
	}

	db.locationsIdx.byColony.Remove(locationID, l.ColonyName)
	db.locationsIdx.byName.Remove(locationID, l.ColonyName+":"+l.Name)
	return db.locations.Delete(locationID)
}

func (db *EmbeddedDatabase) RemoveLocationByName(colonyName string, name string) error {
	ids := db.locationsIdx.byName.Lookup(colonyName + ":" + name)
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		if l, ok := db.locations.Get(id); ok {
			db.locationsIdx.byColony.Remove(id, l.ColonyName)
			db.locationsIdx.byName.Remove(id, l.ColonyName+":"+l.Name)
			if err := db.locations.Delete(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveLocationsByColonyName(colonyName string) error {
	ids := db.locationsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if l, ok := db.locations.Get(id); ok {
			db.locationsIdx.byName.Remove(id, l.ColonyName+":"+l.Name)
		}
		db.locationsIdx.byColony.Remove(id, colonyName)
		if err := db.locations.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
