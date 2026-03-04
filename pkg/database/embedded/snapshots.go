package embedded

import (
	"errors"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func copySnapshot(s *core.Snapshot) *core.Snapshot {
	cp := *s
	if s.FileIDs != nil {
		cp.FileIDs = make([]string, len(s.FileIDs))
		copy(cp.FileIDs, s.FileIDs)
	}
	return &cp
}

func (db *EmbeddedDatabase) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) {
	existingSnapshot, _ := db.GetSnapshotByName(colonyName, name)
	if existingSnapshot != nil {
		return nil, errors.New("Snapshot with name <" + name + "> in Colony <" + colonyName + "> already exists")
	}

	label = strings.TrimSuffix(label, "/")

	allLabels, err := db.GetFileLabelsByName(colonyName, label, true)
	if err != nil {
		return nil, err
	}

	snapshotID := core.GenerateRandomID()
	now := time.Now().UTC()

	var fileIDs []string
	for _, l := range allLabels {
		filenames, err := db.GetFilenamesByLabel(colonyName, l.Name)
		if err != nil {
			return nil, err
		}

		for _, filename := range filenames {
			file, err := db.GetLatestFileByName(colonyName, l.Name, filename)
			if err != nil {
				return nil, err
			}
			if len(file) != 1 {
				return nil, errors.New("failed to get file info, len>1")
			}
			fileIDs = append(fileIDs, file[0].ID)
		}
	}

	snapshot := &core.Snapshot{
		ID:         snapshotID,
		ColonyName: colonyName,
		Label:      label,
		Name:       name,
		FileIDs:    fileIDs,
		Added:      now,
	}

	if err := db.snapshots.Put(snapshotID, snapshot); err != nil {
		return nil, err
	}

	db.snapshotsIdx.byColony.Add(snapshotID, colonyName)
	db.snapshotsIdx.byName.Add(snapshotID, colonyName+":"+name)

	return copySnapshot(snapshot), nil
}

func (db *EmbeddedDatabase) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) {
	s, ok := db.snapshots.Get(snapshotID)
	if !ok {
		return nil, errors.New("Snapshot not found with Id <" + snapshotID + "> in Colony <" + colonyName + "> does not exists")
	}
	if s.ColonyName != colonyName {
		return nil, errors.New("Snapshot not found with Id <" + snapshotID + "> in Colony <" + colonyName + "> does not exists")
	}
	return copySnapshot(s), nil
}

func (db *EmbeddedDatabase) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) {
	ids := db.snapshotsIdx.byColony.Lookup(colonyName)
	var snapshots []*core.Snapshot
	for _, id := range ids {
		if s, ok := db.snapshots.Get(id); ok {
			snapshots = append(snapshots, copySnapshot(s))
		}
	}
	return snapshots, nil
}

func (db *EmbeddedDatabase) RemoveSnapshotByID(colonyName string, snapshotID string) error {
	s, ok := db.snapshots.Get(snapshotID)
	if !ok {
		return nil
	}
	if s.ColonyName != colonyName {
		return nil
	}

	db.snapshotsIdx.byColony.Remove(snapshotID, s.ColonyName)
	db.snapshotsIdx.byName.Remove(snapshotID, s.ColonyName+":"+s.Name)

	return db.snapshots.Delete(snapshotID)
}

func (db *EmbeddedDatabase) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) {
	ids := db.snapshotsIdx.byName.Lookup(colonyName + ":" + name)
	if len(ids) == 0 {
		return nil, errors.New("Snapshot not found with name <" + name + "> in Colony <" + colonyName + "> does not exists")
	}
	s, ok := db.snapshots.Get(ids[0])
	if !ok {
		return nil, errors.New("Snapshot not found with name <" + name + "> in Colony <" + colonyName + "> does not exists")
	}
	return copySnapshot(s), nil
}

func (db *EmbeddedDatabase) RemoveSnapshotByName(colonyName string, name string) error {
	ids := db.snapshotsIdx.byName.Lookup(colonyName + ":" + name)
	for _, id := range ids {
		if s, ok := db.snapshots.Get(id); ok {
			if s.ColonyName == colonyName {
				db.snapshotsIdx.byColony.Remove(id, s.ColonyName)
				db.snapshotsIdx.byName.Remove(id, s.ColonyName+":"+s.Name)
				if err := db.snapshots.Delete(id); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (db *EmbeddedDatabase) RemoveSnapshotsByColonyName(colonyName string) error {
	ids := db.snapshotsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if s, ok := db.snapshots.Get(id); ok {
			db.snapshotsIdx.byName.Remove(id, s.ColonyName+":"+s.Name)
		}
		db.snapshotsIdx.byColony.Remove(id, colonyName)
		if err := db.snapshots.Delete(id); err != nil {
			return err
		}
	}
	return nil
}
