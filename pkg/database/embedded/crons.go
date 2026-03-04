package embedded

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func copyCron(c *core.Cron) *core.Cron {
	cp := *c
	return &cp
}

func (db *EmbeddedDatabase) AddCron(cron *core.Cron) error {
	existing, err := db.GetCronByName(cron.ColonyName, cron.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("Cron with name <" + cron.Name + "> in Colony <" + cron.ColonyName + "> already exists")
	}

	cp := copyCron(cron)
	if err := db.crons.Put(cp.ID, cp); err != nil {
		return err
	}

	db.cronsIdx.byColony.Add(cp.ID, cp.ColonyName)
	db.cronsIdx.byName.Add(cp.ID, cp.ColonyName+":"+cp.Name)

	return nil
}

func (db *EmbeddedDatabase) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error {
	c, ok := db.crons.Get(cronID)
	if !ok {
		return nil
	}

	cp := *c
	cp.NextRun = nextRun
	cp.LastRun = lastRun
	cp.PrevProcessGraphID = lastProcessGraphID

	return db.crons.Put(cronID, &cp)
}

func (db *EmbeddedDatabase) GetCronByID(cronID string) (*core.Cron, error) {
	c, ok := db.crons.Get(cronID)
	if !ok {
		return nil, nil
	}
	return copyCron(c), nil
}

func (db *EmbeddedDatabase) GetCronByName(colonyName string, cronName string) (*core.Cron, error) {
	ids := db.cronsIdx.byName.Lookup(colonyName + ":" + cronName)
	if len(ids) == 0 {
		return nil, nil
	}
	c, ok := db.crons.Get(ids[0])
	if !ok {
		return nil, nil
	}
	return copyCron(c), nil
}

func (db *EmbeddedDatabase) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) {
	ids := db.cronsIdx.byColony.Lookup(colonyName)
	var crons []*core.Cron
	for _, id := range ids {
		if len(crons) >= count {
			break
		}
		if c, ok := db.crons.Get(id); ok {
			crons = append(crons, copyCron(c))
		}
	}
	return crons, nil
}

func (db *EmbeddedDatabase) FindAllCrons() ([]*core.Cron, error) {
	all := db.crons.All()
	result := make([]*core.Cron, len(all))
	for i, c := range all {
		result[i] = copyCron(c)
	}
	return result, nil
}

func (db *EmbeddedDatabase) RemoveCronByID(cronID string) error {
	c, ok := db.crons.Get(cronID)
	if !ok {
		return nil
	}

	db.cronsIdx.byColony.Remove(cronID, c.ColonyName)
	db.cronsIdx.byName.Remove(cronID, c.ColonyName+":"+c.Name)

	return db.crons.Delete(cronID)
}

func (db *EmbeddedDatabase) RemoveAllCronsByColonyName(colonyName string) error {
	ids := db.cronsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if c, ok := db.crons.Get(id); ok {
			db.cronsIdx.byName.Remove(id, c.ColonyName+":"+c.Name)
		}
		db.cronsIdx.byColony.Remove(id, colonyName)
		if err := db.crons.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
