package kvstore

import (
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// SnapshotDatabase Interface Implementation
// =====================================

// CreateSnapshot creates a new snapshot
func (db *KVStoreDatabase) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) {
	// Check if snapshot with same name already exists
	existingSnapshot, _ := db.GetSnapshotByName(colonyName, name)
	if existingSnapshot != nil {
		return nil, fmt.Errorf("snapshot with name %s in colony %s already exists", name, colonyName)
	}

	// Generate snapshot ID
	snapshotID := core.GenerateRandomID()
	now := time.Now().UTC()

	// Get all files with the specified label
	var fileIDs []string
	
	// Find files that match the label
	filenames, err := db.GetFilenamesByLabel(colonyName, label)
	if err != nil {
		return nil, fmt.Errorf("failed to get filenames for label %s: %w", label, err)
	}

	// For each filename, get the latest version
	for _, filename := range filenames {
		latestFiles, err := db.GetLatestFileByName(colonyName, label, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest file %s: %w", filename, err)
		}
		
		if len(latestFiles) == 1 {
			fileIDs = append(fileIDs, latestFiles[0].ID)
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

	// Store snapshot at /snapshots/{snapshotID}
	snapshotPath := fmt.Sprintf("/snapshots/%s", snapshotID)
	
	err = db.store.Put(snapshotPath, snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot %s: %w", snapshotID, err)
	}

	return snapshot, nil
}

// GetSnapshotByID retrieves a snapshot by colony name and snapshot ID
func (db *KVStoreDatabase) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) {
	snapshotPath := fmt.Sprintf("/snapshots/%s", snapshotID)
	
	if !db.store.Exists(snapshotPath) {
		return nil, fmt.Errorf("snapshot with ID %s not found", snapshotID)
	}

	snapshotInterface, err := db.store.Get(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot %s: %w", snapshotID, err)
	}

	snapshot, ok := snapshotInterface.(*core.Snapshot)
	if !ok {
		return nil, fmt.Errorf("stored object is not a snapshot")
	}

	// Check colony match
	if snapshot.ColonyName != colonyName {
		return nil, fmt.Errorf("snapshot %s does not belong to colony %s", snapshotID, colonyName)
	}

	return snapshot, nil
}

// GetSnapshotsByColonyName retrieves all snapshots for a colony
func (db *KVStoreDatabase) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) {
	// Search for snapshots by colony name
	snapshots, err := db.store.FindRecursive("/snapshots", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find snapshots for colony %s: %w", colonyName, err)
	}

	var result []*core.Snapshot
	for _, searchResult := range snapshots {
		if snapshot, ok := searchResult.Value.(*core.Snapshot); ok {
			result = append(result, snapshot)
		}
	}

	return result, nil
}

// RemoveSnapshotByID removes a snapshot by colony name and snapshot ID
func (db *KVStoreDatabase) RemoveSnapshotByID(colonyName string, snapshotID string) error {
	// First check if snapshot exists and belongs to colony
	snapshot, err := db.GetSnapshotByID(colonyName, snapshotID)
	if err != nil {
		return err
	}

	snapshotPath := fmt.Sprintf("/snapshots/%s", snapshot.ID)
	err = db.store.Delete(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %w", snapshotID, err)
	}

	return nil
}

// GetSnapshotByName retrieves a snapshot by colony name and name
func (db *KVStoreDatabase) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) {
	// Search for snapshots by colony name
	snapshots, err := db.store.FindRecursive("/snapshots", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find snapshots for colony %s: %w", colonyName, err)
	}

	for _, searchResult := range snapshots {
		if snapshot, ok := searchResult.Value.(*core.Snapshot); ok {
			if snapshot.Name == name {
				return snapshot, nil
			}
		}
	}

	return nil, fmt.Errorf("snapshot with name %s not found in colony %s", name, colonyName)
}

// RemoveSnapshotByName removes a snapshot by colony name and name
func (db *KVStoreDatabase) RemoveSnapshotByName(colonyName string, name string) error {
	// Find snapshot by name
	snapshot, err := db.GetSnapshotByName(colonyName, name)
	if err != nil {
		return err
	}

	snapshotPath := fmt.Sprintf("/snapshots/%s", snapshot.ID)
	err = db.store.Delete(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %w", snapshot.ID, err)
	}

	return nil
}

// RemoveSnapshotsByColonyName removes all snapshots for a colony
func (db *KVStoreDatabase) RemoveSnapshotsByColonyName(colonyName string) error {
	// Find all snapshots for the colony
	snapshots, err := db.store.FindRecursive("/snapshots", "colonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find snapshots for colony %s: %w", colonyName, err)
	}

	// Remove each snapshot
	for _, searchResult := range snapshots {
		if snapshot, ok := searchResult.Value.(*core.Snapshot); ok {
			snapshotPath := fmt.Sprintf("/snapshots/%s", snapshot.ID)
			err := db.store.Delete(snapshotPath)
			if err != nil {
				return fmt.Errorf("failed to remove snapshot %s: %w", snapshot.ID, err)
			}
		}
	}

	return nil
}