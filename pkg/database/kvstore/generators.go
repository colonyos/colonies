package kvstore

import (
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// GeneratorDatabase Interface Implementation
// =====================================

// AddGenerator adds a generator to the database
func (db *KVStoreDatabase) AddGenerator(generator *core.Generator) error {
	if generator == nil {
		return errors.New("generator cannot be nil")
	}

	// Store generator at /generators/{generatorID} (upsert behavior)
	generatorPath := fmt.Sprintf("/generators/%s", generator.ID)
	
	err := db.store.Put(generatorPath, generator)
	if err != nil {
		return fmt.Errorf("failed to add generator %s: %w", generator.ID, err)
	}

	return nil
}

// SetGeneratorLastRun sets the last run time for a generator
func (db *KVStoreDatabase) SetGeneratorLastRun(generatorID string) error {
	generatorPath := fmt.Sprintf("/generators/%s", generatorID)
	
	if !db.store.Exists(generatorPath) {
		return fmt.Errorf("generator with ID %s not found", generatorID)
	}

	generatorInterface, err := db.store.Get(generatorPath)
	if err != nil {
		return fmt.Errorf("failed to get generator %s: %w", generatorID, err)
	}

	storedGenerator, ok := generatorInterface.(*core.Generator)
	if !ok {
		return fmt.Errorf("stored object is not a generator")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	generator := *storedGenerator
	generator.LastRun = time.Now()

	// Store back
	err = db.store.Put(generatorPath, &generator)
	if err != nil {
		return fmt.Errorf("failed to update generator last run: %w", err)
	}

	return nil
}

// SetGeneratorFirstPack sets the first pack flag for a generator
func (db *KVStoreDatabase) SetGeneratorFirstPack(generatorID string) error {
	generatorPath := fmt.Sprintf("/generators/%s", generatorID)
	
	if !db.store.Exists(generatorPath) {
		return fmt.Errorf("generator with ID %s not found", generatorID)
	}

	generatorInterface, err := db.store.Get(generatorPath)
	if err != nil {
		return fmt.Errorf("failed to get generator %s: %w", generatorID, err)
	}

	storedGenerator, ok := generatorInterface.(*core.Generator)
	if !ok {
		return fmt.Errorf("stored object is not a generator")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	generator := *storedGenerator
	generator.FirstPack = time.Now()

	// Store back
	err = db.store.Put(generatorPath, &generator)
	if err != nil {
		return fmt.Errorf("failed to update generator first pack: %w", err)
	}

	return nil
}

// GetGeneratorByID retrieves a generator by ID
func (db *KVStoreDatabase) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	generatorPath := fmt.Sprintf("/generators/%s", generatorID)
	
	if !db.store.Exists(generatorPath) {
		// Return (nil, nil) when generator not found, like PostgreSQL
		return nil, nil
	}

	generatorInterface, err := db.store.Get(generatorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator %s: %w", generatorID, err)
	}

	generator, ok := generatorInterface.(*core.Generator)
	if !ok {
		return nil, fmt.Errorf("stored object is not a generator")
	}

	return generator, nil
}

// GetGeneratorByName retrieves a generator by colony name and name
func (db *KVStoreDatabase) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	// Search for generator by colony name
	generators, err := db.store.FindRecursive("/generators", "colonyname", colonyName)
	if err != nil {
		// Return (nil, nil) when no generators found, like PostgreSQL
		return nil, nil
	}

	for _, searchResult := range generators {
		if generator, ok := searchResult.Value.(*core.Generator); ok && generator.Name == name {
			return generator, nil
		}
	}

	// Return (nil, nil) when generator not found, like PostgreSQL
	return nil, nil
}

// FindGeneratorsByColonyName finds generators by colony name
func (db *KVStoreDatabase) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	// Search for generators by colony name
	generators, err := db.store.FindRecursive("/generators", "colonyname", colonyName)
	if err != nil {
		// Return empty slice when no generators found, like PostgreSQL
		return []*core.Generator{}, nil
	}

	var result []*core.Generator
	for _, searchResult := range generators {
		if generator, ok := searchResult.Value.(*core.Generator); ok {
			result = append(result, generator)
			
			if count > 0 && len(result) >= count {
				break
			}
		}
	}

	return result, nil
}

// FindAllGenerators retrieves all generators
func (db *KVStoreDatabase) FindAllGenerators() ([]*core.Generator, error) {
	// Get all generator IDs by listing children under /generators
	generatorIDs, err := db.store.List("/generators")
	if err != nil {
		// Return empty slice when no generators found, like PostgreSQL
		return []*core.Generator{}, nil
	}

	var result []*core.Generator
	for _, generatorID := range generatorIDs {
		generatorPath := fmt.Sprintf("/generators/%s", generatorID)
		generatorInterface, err := db.store.Get(generatorPath)
		if err != nil {
			continue // Skip if error getting this generator
		}
		
		if generator, ok := generatorInterface.(*core.Generator); ok {
			// Return a copy to prevent race conditions
			generatorCopy := *generator
			result = append(result, &generatorCopy)
		}
	}

	return result, nil
}

// RemoveGeneratorByID removes a generator by ID (with cascade delete of generator args)
func (db *KVStoreDatabase) RemoveGeneratorByID(generatorID string) error {
	generatorPath := fmt.Sprintf("/generators/%s", generatorID)
	
	if !db.store.Exists(generatorPath) {
		return fmt.Errorf("generator with ID %s not found", generatorID)
	}

	// First remove all generator args for this generator (cascade delete)
	err := db.RemoveAllGeneratorArgsByGeneratorID(generatorID)
	if err != nil {
		return fmt.Errorf("failed to remove generator args for generator %s: %w", generatorID, err)
	}

	// Then remove the generator itself
	err = db.store.Delete(generatorPath)
	if err != nil {
		return fmt.Errorf("failed to remove generator %s: %w", generatorID, err)
	}

	return nil
}

// RemoveAllGeneratorsByColonyName removes all generators for a colony
func (db *KVStoreDatabase) RemoveAllGeneratorsByColonyName(colonyName string) error {
	// Find all generators for the colony
	generators, err := db.store.FindRecursive("/generators", "colonyname", colonyName)
	if err != nil {
		// No generators found to remove, that's okay
		return nil
	}

	// Remove each generator (with cascade delete of generator args)
	for _, searchResult := range generators {
		if generator, ok := searchResult.Value.(*core.Generator); ok {
			// First remove all generator args for this generator (cascade delete)
			err := db.RemoveAllGeneratorArgsByGeneratorID(generator.ID)
			if err != nil {
				return fmt.Errorf("failed to remove generator args for generator %s: %w", generator.ID, err)
			}

			// Then remove the generator itself
			generatorPath := fmt.Sprintf("/generators/%s", generator.ID)
			err = db.store.Delete(generatorPath)
			if err != nil {
				return fmt.Errorf("failed to remove generator %s: %w", generator.ID, err)
			}
		}
	}

	return nil
}

// AddGeneratorArg adds a generator argument to the database
func (db *KVStoreDatabase) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	if generatorArg == nil {
		return errors.New("generator arg cannot be nil")
	}

	// Store generator arg at /generatorargs/{generatorArgID} (upsert behavior)
	generatorArgPath := fmt.Sprintf("/generatorargs/%s", generatorArg.ID)
	
	err := db.store.Put(generatorArgPath, generatorArg)
	if err != nil {
		return fmt.Errorf("failed to add generator arg %s: %w", generatorArg.ID, err)
	}

	return nil
}

// GetGeneratorArgs retrieves generator arguments by generator ID
func (db *KVStoreDatabase) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	// Search for generator args by generator ID
	generatorArgs, err := db.store.FindRecursive("/generatorargs", "generatorid", generatorID)
	if err != nil {
		// Return empty slice when no generator args found, like PostgreSQL
		return []*core.GeneratorArg{}, nil
	}

	var result []*core.GeneratorArg
	for _, searchResult := range generatorArgs {
		if generatorArg, ok := searchResult.Value.(*core.GeneratorArg); ok {
			result = append(result, generatorArg)
			
			if count > 0 && len(result) >= count {
				break
			}
		}
	}

	return result, nil
}

// CountGeneratorArgs counts generator arguments by generator ID
func (db *KVStoreDatabase) CountGeneratorArgs(generatorID string) (int, error) {
	generatorArgs, err := db.GetGeneratorArgs(generatorID, -1)
	if err != nil {
		return 0, err
	}

	return len(generatorArgs), nil
}

// RemoveGeneratorArgByID removes a generator argument by ID
func (db *KVStoreDatabase) RemoveGeneratorArgByID(generatorArgsID string) error {
	generatorArgPath := fmt.Sprintf("/generatorargs/%s", generatorArgsID)
	
	if !db.store.Exists(generatorArgPath) {
		return fmt.Errorf("generator arg with ID %s not found", generatorArgsID)
	}

	err := db.store.Delete(generatorArgPath)
	if err != nil {
		return fmt.Errorf("failed to remove generator arg %s: %w", generatorArgsID, err)
	}

	return nil
}

// RemoveAllGeneratorArgsByGeneratorID removes all generator arguments for a generator
func (db *KVStoreDatabase) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error {
	// Find all generator args for the generator
	generatorArgs, err := db.store.FindRecursive("/generatorargs", "generatorid", generatorID)
	if err != nil {
		// No generator args found to remove, that's okay
		return nil
	}

	// Remove each generator arg
	for _, searchResult := range generatorArgs {
		if generatorArg, ok := searchResult.Value.(*core.GeneratorArg); ok {
			generatorArgPath := fmt.Sprintf("/generatorargs/%s", generatorArg.ID)
			err := db.store.Delete(generatorArgPath)
			if err != nil {
				return fmt.Errorf("failed to remove generator arg %s: %w", generatorArg.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllGeneratorArgsByColonyName removes all generator arguments for a colony
func (db *KVStoreDatabase) RemoveAllGeneratorArgsByColonyName(colonyName string) error {
	// Find all generator args for the colony
	generatorArgs, err := db.store.FindRecursive("/generatorargs", "colonyname", colonyName)
	if err != nil {
		// No generator args found to remove, that's okay
		return nil
	}

	// Remove each generator arg
	for _, searchResult := range generatorArgs {
		if generatorArg, ok := searchResult.Value.(*core.GeneratorArg); ok {
			generatorArgPath := fmt.Sprintf("/generatorargs/%s", generatorArg.ID)
			err := db.store.Delete(generatorArgPath)
			if err != nil {
				return fmt.Errorf("failed to remove generator arg %s: %w", generatorArg.ID, err)
			}
		}
	}

	return nil
}