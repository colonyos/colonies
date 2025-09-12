package kvstore

import (
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// CronDatabase Interface Implementation
// =====================================

// AddCron adds a cron job to the database
func (db *KVStoreDatabase) AddCron(cron *core.Cron) error {
	if cron == nil {
		return errors.New("cron cannot be nil")
	}

	// Store cron at /crons/{cronID}
	cronPath := fmt.Sprintf("/crons/%s", cron.ID)
	
	// Check if cron already exists
	if db.store.Exists(cronPath) {
		return fmt.Errorf("cron with ID %s already exists", cron.ID)
	}

	err := db.store.Put(cronPath, cron)
	if err != nil {
		return fmt.Errorf("failed to add cron %s: %w", cron.ID, err)
	}

	return nil
}

// UpdateCron updates a cron job's run times and process graph ID
func (db *KVStoreDatabase) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error {
	cronPath := fmt.Sprintf("/crons/%s", cronID)
	
	if !db.store.Exists(cronPath) {
		return fmt.Errorf("cron with ID %s not found", cronID)
	}

	cronInterface, err := db.store.Get(cronPath)
	if err != nil {
		return fmt.Errorf("failed to get cron %s: %w", cronID, err)
	}

	storedCron, ok := cronInterface.(*core.Cron)
	if !ok {
		return fmt.Errorf("stored object is not a cron")
	}

	// Create a copy to avoid modifying the original
	updatedCron := *storedCron

	// Update the cron fields on the copy
	updatedCron.NextRun = nextRun
	updatedCron.LastRun = lastRun
	updatedCron.PrevProcessGraphID = lastProcessGraphID

	err = db.store.Put(cronPath, &updatedCron)
	if err != nil {
		return fmt.Errorf("failed to update cron %s: %w", cronID, err)
	}

	return nil
}

// GetCronByID retrieves a cron job by ID
func (db *KVStoreDatabase) GetCronByID(cronID string) (*core.Cron, error) {
	cronPath := fmt.Sprintf("/crons/%s", cronID)
	
	if !db.store.Exists(cronPath) {
		// Return (nil, nil) when cron not found, like PostgreSQL
		return nil, nil
	}

	cronInterface, err := db.store.Get(cronPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron %s: %w", cronID, err)
	}

	storedCron, ok := cronInterface.(*core.Cron)
	if !ok {
		return nil, fmt.Errorf("stored object is not a cron")
	}

	// Return a copy to prevent race conditions when multiple goroutines access the same object
	cronCopy := *storedCron
	return &cronCopy, nil
}

// GetCronByName retrieves a cron job by colony and cron name
func (db *KVStoreDatabase) GetCronByName(colonyName string, cronName string) (*core.Cron, error) {
	// Get all cron IDs by listing children under /crons
	cronIDs, err := db.store.List("/crons")
	if err != nil {
		// Return (nil, nil) when no crons found, like PostgreSQL
		return nil, nil
	}

	for _, cronID := range cronIDs {
		cronPath := fmt.Sprintf("/crons/%s", cronID)
		cronInterface, err := db.store.Get(cronPath)
		if err != nil {
			continue // Skip if error getting this cron
		}
		
		if cron, ok := cronInterface.(*core.Cron); ok {
			if cron.ColonyName == colonyName && cron.Name == cronName {
				return cron, nil
			}
		}
	}

	// Return (nil, nil) when cron not found, like PostgreSQL
	return nil, nil
}

// FindCronsByColonyName finds cron jobs by colony name
func (db *KVStoreDatabase) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) {
	// Get all cron IDs by listing children under /crons
	cronIDs, err := db.store.List("/crons")
	if err != nil {
		// Return empty slice when no crons found, like PostgreSQL
		return []*core.Cron{}, nil
	}

	var result []*core.Cron
	for _, cronID := range cronIDs {
		cronPath := fmt.Sprintf("/crons/%s", cronID)
		cronInterface, err := db.store.Get(cronPath)
		if err != nil {
			continue // Skip if error getting this cron
		}
		
		if cron, ok := cronInterface.(*core.Cron); ok {
			if cron.ColonyName == colonyName {
				result = append(result, cron)
				if count > 0 && len(result) >= count {
					break
				}
			}
		}
	}

	return result, nil
}

// FindAllCrons retrieves all cron jobs
func (db *KVStoreDatabase) FindAllCrons() ([]*core.Cron, error) {
	// Get all cron IDs by listing children under /crons
	cronIDs, err := db.store.List("/crons")
	if err != nil {
		// Return empty slice when no crons found, like PostgreSQL
		return []*core.Cron{}, nil
	}

	var result []*core.Cron
	for _, cronID := range cronIDs {
		cronPath := fmt.Sprintf("/crons/%s", cronID)
		cronInterface, err := db.store.Get(cronPath)
		if err != nil {
			continue // Skip if error getting this cron
		}
		
		if cron, ok := cronInterface.(*core.Cron); ok {
			// Return a copy to prevent race conditions
			cronCopy := *cron
			result = append(result, &cronCopy)
		}
	}

	return result, nil
}

// RemoveCronByID removes a cron job by ID
func (db *KVStoreDatabase) RemoveCronByID(cronID string) error {
	cronPath := fmt.Sprintf("/crons/%s", cronID)
	
	if !db.store.Exists(cronPath) {
		return fmt.Errorf("cron with ID %s not found", cronID)
	}

	err := db.store.Delete(cronPath)
	if err != nil {
		return fmt.Errorf("failed to remove cron %s: %w", cronID, err)
	}

	return nil
}

// RemoveAllCronsByColonyName removes all cron jobs for a colony
func (db *KVStoreDatabase) RemoveAllCronsByColonyName(colonyName string) error {
	// Get all cron IDs by listing children under /crons
	cronIDs, err := db.store.List("/crons")
	if err != nil {
		// No crons found to remove, that's okay
		return nil
	}

	for _, cronID := range cronIDs {
		cronPath := fmt.Sprintf("/crons/%s", cronID)
		cronInterface, err := db.store.Get(cronPath)
		if err != nil {
			continue // Skip if error getting this cron
		}
		
		if cron, ok := cronInterface.(*core.Cron); ok {
			if cron.ColonyName == colonyName {
				err := db.store.Delete(cronPath)
				if err != nil {
					return fmt.Errorf("failed to remove cron %s: %w", cron.ID, err)
				}
			}
		}
	}

	return nil
}