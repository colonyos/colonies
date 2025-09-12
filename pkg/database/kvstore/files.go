package kvstore

import (
	"errors"
	"fmt"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// FileDatabase Interface Implementation
// =====================================

// AddFile adds a file to the database
func (db *KVStoreDatabase) AddFile(file *core.File) error {
	if file == nil {
		return errors.New("file cannot be nil")
	}

	// Store file at /files/{fileID}
	filePath := fmt.Sprintf("/files/%s", file.ID)
	
	// Check if file already exists
	if db.store.Exists(filePath) {
		return fmt.Errorf("file with ID %s already exists", file.ID)
	}

	err := db.store.Put(filePath, file)
	if err != nil {
		return fmt.Errorf("failed to add file %s: %w", file.ID, err)
	}

	return nil
}

// GetFileByID retrieves a file by colony name and file ID
func (db *KVStoreDatabase) GetFileByID(colonyName string, fileID string) (*core.File, error) {
	filePath := fmt.Sprintf("/files/%s", fileID)
	
	if !db.store.Exists(filePath) {
		return nil, fmt.Errorf("file with ID %s not found", fileID)
	}

	fileInterface, err := db.store.Get(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", fileID, err)
	}

	file, ok := fileInterface.(*core.File)
	if !ok {
		return nil, fmt.Errorf("stored object is not a file")
	}

	// Check colony match
	if file.ColonyName != colonyName {
		return nil, fmt.Errorf("file %s does not belong to colony %s", fileID, colonyName)
	}

	return file, nil
}

// GetLatestFileByName retrieves the latest file by colony name, label, and name
func (db *KVStoreDatabase) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	var matchingFiles []*core.File
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if (label == "" || file.Label == label) && (name == "" || file.Name == name) {
				matchingFiles = append(matchingFiles, file)
			}
		}
	}

	// Sort by timestamp and return the latest
	if len(matchingFiles) == 0 {
		return []*core.File{}, nil
	}

	// Sort by timestamp (newest first) and return the latest
	var latestFile *core.File
	for _, file := range matchingFiles {
		if latestFile == nil || file.Added.After(latestFile.Added) {
			latestFile = file
		}
	}

	return []*core.File{latestFile}, nil
}

// GetFileByName retrieves files by colony name, label, and name
func (db *KVStoreDatabase) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	var result []*core.File
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if (label == "" || file.Label == label) && (name == "" || file.Name == name) {
				result = append(result, file)
			}
		}
	}

	return result, nil
}

// GetFilenamesByLabel retrieves filenames by colony name and label
func (db *KVStoreDatabase) GetFilenamesByLabel(colonyName string, label string) ([]string, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	// Use a map to collect unique filenames
	filenameSet := make(map[string]bool)
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if label == "" || file.Label == label {
				filenameSet[file.Name] = true
			}
		}
	}

	// Convert map to slice
	var filenames []string
	for filename := range filenameSet {
		filenames = append(filenames, filename)
	}

	return filenames, nil
}

// GetFileDataByLabel retrieves file data by colony name and label
func (db *KVStoreDatabase) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	var result []*core.FileData
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if label == "" || file.Label == label {
				// Create FileData from File
				fileData := &core.FileData{
					Name:       file.Name,
					Checksum:   file.Checksum,
					Size:       file.Size,
					S3Filename: "", // File doesn't have S3Filename field
				}
				result = append(result, fileData)
			}
		}
	}

	return result, nil
}

// RemoveFileByID removes a file by colony name and file ID
func (db *KVStoreDatabase) RemoveFileByID(colonyName string, fileID string) error {
	// First check if file exists and belongs to colony
	file, err := db.GetFileByID(colonyName, fileID)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("/files/%s", file.ID)
	err = db.store.Delete(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove file %s: %w", fileID, err)
	}

	return nil
}

// RemoveFileByName removes a file by colony name, label, and name
func (db *KVStoreDatabase) RemoveFileByName(colonyName string, label string, name string) error {
	// Find files by name and label
	files, err := db.GetFileByName(colonyName, label, name)
	if err != nil {
		return err
	}

	// Remove all matching files
	for _, file := range files {
		filePath := fmt.Sprintf("/files/%s", file.ID)
		err := db.store.Delete(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove file %s: %w", file.ID, err)
		}
	}

	return nil
}

// GetFileLabels retrieves file labels for a colony
func (db *KVStoreDatabase) GetFileLabels(colonyName string) ([]*core.Label, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	// Collect unique labels
	labelMap := make(map[string]*core.Label)
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if file.Label != "" {
				if _, exists := labelMap[file.Label]; !exists {
					labelMap[file.Label] = &core.Label{
						Name:  file.Label,
						Files: 0, // Count would need to be calculated
					}
				}
			}
		}
	}

	// Convert map to slice
	var result []*core.Label
	for _, label := range labelMap {
		result = append(result, label)
	}

	return result, nil
}

// GetFileLabelsByName retrieves file labels by name pattern
func (db *KVStoreDatabase) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	// Collect unique labels from files whose names match the criteria
	labelMap := make(map[string]*core.Label)
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if file.Label != "" {
				match := false
				if exact {
					match = file.Name == name
				} else {
					match = strings.Contains(strings.ToLower(file.Name), strings.ToLower(name))
				}
				
				if match {
					if _, exists := labelMap[file.Label]; !exists {
						labelMap[file.Label] = &core.Label{
							Name:  file.Label,
							Files: 0, // Count would need to be calculated
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	var result []*core.Label
	for _, label := range labelMap {
		result = append(result, label)
	}

	return result, nil
}

// CountFilesWithLabel counts files with a specific label
func (db *KVStoreDatabase) CountFilesWithLabel(colonyName string, label string) (int, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return 0, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	count := 0
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if file.Label == label {
				count++
			}
		}
	}

	return count, nil
}

// CountFiles counts all files for a colony
func (db *KVStoreDatabase) CountFiles(colonyName string) (int, error) {
	// Search for files by colony name
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return 0, fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	count := 0
	for _, searchResult := range files {
		if _, ok := searchResult.Value.(*core.File); ok {
			count++
		}
	}

	return count, nil
}

// RemoveFilesByLabel removes all files with a specific label from a colony
func (db *KVStoreDatabase) RemoveFilesByLabel(colonyName string, label string) error {
	// Find all files in the colony
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	// Remove all matching files
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			if file.Label == label {
				filePath := fmt.Sprintf("/files/%s", file.ID)
				err := db.store.Delete(filePath)
				if err != nil {
					return fmt.Errorf("failed to remove file %s: %w", file.ID, err)
				}
			}
		}
	}

	return nil
}

// RemoveFilesByColonyName removes all files from a colony
func (db *KVStoreDatabase) RemoveFilesByColonyName(colonyName string) error {
	// Find all files in the colony
	files, err := db.store.FindRecursive("/files", "colonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find files for colony %s: %w", colonyName, err)
	}

	// Remove all files
	for _, searchResult := range files {
		if file, ok := searchResult.Value.(*core.File); ok {
			filePath := fmt.Sprintf("/files/%s", file.ID)
			err := db.store.Delete(filePath)
			if err != nil {
				return fmt.Errorf("failed to remove file %s: %w", file.ID, err)
			}
		}
	}

	return nil
}

// RemoveAllFiles removes all files from the database
func (db *KVStoreDatabase) RemoveAllFiles() error {
	filesPath := "/files"
	
	if !db.store.Exists(filesPath) {
		return nil // No files to remove
	}

	err := db.store.Delete(filesPath)
	if err != nil {
		return fmt.Errorf("failed to remove all files: %w", err)
	}

	// Recreate the files structure
	err = db.store.CreateArray("/files")
	if err != nil {
		return fmt.Errorf("failed to recreate files structure: %w", err)
	}

	return nil
}