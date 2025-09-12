package kvstore

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestFileClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	now := time.Now()
	file := utils.CreateTestFileWithID("test_file_id", "test_colony", now)
	file.Label = "test_label"
	file.Name = "test_name"
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddFile(file)
	assert.Nil(t, err)

	_, err = db.GetFileByID("test_colony", "invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.GetLatestFileByName("test_colony", "test_label", "test_name")
	assert.Nil(t, err) // Returns nil for non-existing

	_, err = db.GetFileByName("test_colony", "test_label", "test_name")
	assert.Nil(t, err)

	_, err = db.GetFilenamesByLabel("test_colony", "test_label")
	assert.Nil(t, err)

	_, err = db.GetFileDataByLabel("test_colony", "test_label")
	assert.Nil(t, err)

	err = db.RemoveFileByID("test_colony", "invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveFileByName("test_colony", "test_label", "invalid_name")
	assert.Nil(t, err)

	_, err = db.GetFileLabels("test_colony")
	assert.Nil(t, err)

	_, err = db.GetFileLabelsByName("test_colony", "test_name", true)
	assert.Nil(t, err)

	err = db.RemoveFilesByLabel("test_colony", "invalid_label")
	assert.Nil(t, err)

	err = db.RemoveFilesByColonyName("invalid_colony")
	assert.Nil(t, err)

	err = db.RemoveAllFiles()
	assert.Nil(t, err)
}

func TestAddGetFileByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	now := time.Now()
	file := utils.CreateTestFileWithID("test_file_id", "test_colony", now)
	file.Label = "test_label"
	file.Name = "test_name"

	err = db.AddFile(file)
	assert.Nil(t, err)

	// Get by ID
	fileFromDB, err := db.GetFileByID("test_colony", file.ID)
	assert.Nil(t, err)
	assert.True(t, file.Equals(fileFromDB))

	// Test non-existing ID
	_, err = db.GetFileByID("test_colony", "non_existing_id")
	assert.NotNil(t, err)

	// Test non-existing colony
	_, err = db.GetFileByID("invalid_colony", file.ID)
	assert.NotNil(t, err)
}

func TestGetLatestFileByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"
	fileName := "test_file"

	now := time.Now()

	// Add multiple revisions of the same file
	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = label
	file1.Name = fileName
	file1.SequenceNumber = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now.Add(time.Minute))
	file2.Label = label
	file2.Name = fileName
	file2.SequenceNumber = 2
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFileWithID("file3_id", colonyName, now.Add(2*time.Minute))
	file3.Label = label
	file3.Name = fileName
	file3.SequenceNumber = 3
	err = db.AddFile(file3)
	assert.Nil(t, err)

	// Get latest should return file3
	latestFiles, err := db.GetLatestFileByName(colonyName, label, fileName)
	assert.Nil(t, err)
	assert.Len(t, latestFiles, 1)
	assert.Equal(t, latestFiles[0].ID, "file3_id")
	assert.Equal(t, latestFiles[0].SequenceNumber, int64(3))

	// Test non-existing file
	nonExisting, err := db.GetLatestFileByName(colonyName, label, "non_existing_file")
	assert.Nil(t, err)
	assert.Empty(t, nonExisting)
}

func TestGetFileByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"

	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = label
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = label
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	files, err := db.GetFileByName(colonyName, label, "file1")
	assert.Nil(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, files[0].ID, "file1_id")

	// Test non-existing file
	emptyFiles, err := db.GetFileByName(colonyName, label, "non_existing")
	assert.Nil(t, err)
	assert.Empty(t, emptyFiles)
}

func TestGetFilenamesByLabel(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"

	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = label
	file1.Name = "unique_file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = label
	file2.Name = "unique_file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Add duplicate name to test uniqueness
	file3 := utils.CreateTestFileWithID("file3_id", colonyName, now)
	file3.Label = label
	file3.Name = "unique_file1"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	filenames, err := db.GetFilenamesByLabel(colonyName, label)
	assert.Nil(t, err)
	assert.Len(t, filenames, 2) // Should be unique names
	assert.Contains(t, filenames, "unique_file1")
	assert.Contains(t, filenames, "unique_file2")

	// Test non-existing label
	emptyNames, err := db.GetFilenamesByLabel(colonyName, "non_existing_label")
	assert.Nil(t, err)
	assert.Empty(t, emptyNames)
}

func TestGetFileDataByLabel(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"

	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = label
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = label
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	fileData, err := db.GetFileDataByLabel(colonyName, label)
	assert.Nil(t, err)
	assert.Len(t, fileData, 2)

	// Verify files are returned - FileData has Name, not ID
	foundFile1, foundFile2 := false, false
	for _, file := range fileData {
		if file.Name == "file1" {
			foundFile1 = true
		}
		if file.Name == "file2" {
			foundFile2 = true
		}
	}
	assert.True(t, foundFile1 && foundFile2)

	// Test non-existing label
	emptyData, err := db.GetFileDataByLabel(colonyName, "non_existing_label")
	assert.Nil(t, err)
	assert.Empty(t, emptyData)
}

func TestRemoveFileByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = "test_label"
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = "test_label"
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Remove file1
	err = db.RemoveFileByID(colonyName, "file1_id")
	assert.Nil(t, err)

	// Verify file1 is gone
	_, err = db.GetFileByID(colonyName, "file1_id")
	assert.NotNil(t, err)

	// Verify file2 still exists
	_, err = db.GetFileByID(colonyName, "file2_id")
	assert.Nil(t, err)

	// Test removing non-existing file
	err = db.RemoveFileByID(colonyName, "non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveFileByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = label
	file1.Name = "target_file"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = label
	file2.Name = "other_file"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Remove by name
	err = db.RemoveFileByName(colonyName, label, "target_file")
	assert.Nil(t, err)

	// Verify target_file is gone
	files, err := db.GetFileByName(colonyName, label, "target_file")
	assert.Nil(t, err)
	assert.Empty(t, files)

	// Verify other_file still exists
	files, err = db.GetFileByName(colonyName, label, "other_file")
	assert.Nil(t, err)
	assert.Len(t, files, 1)

	// Test removing non-existing file - should not error
	err = db.RemoveFileByName(colonyName, label, "non_existing_file")
	assert.Nil(t, err)
}

func TestGetFileLabels(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = "label1"
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = "label2"
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Add duplicate label
	file3 := utils.CreateTestFileWithID("file3_id", colonyName, now)
	file3.Label = "label1"
	file3.Name = "file3"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	labels, err := db.GetFileLabels(colonyName)
	assert.Nil(t, err)
	assert.Len(t, labels, 2) // Should be unique
	
	// Check that we have the expected label names
	labelNames := make([]string, len(labels))
	for i, label := range labels {
		labelNames[i] = label.Name
	}
	assert.Contains(t, labelNames, "label1")
	assert.Contains(t, labelNames, "label2")

	// Test non-existing colony
	emptyLabels, err := db.GetFileLabels("non_existing_colony")
	assert.Nil(t, err)
	assert.Empty(t, emptyLabels)
}

func TestGetFileLabelsByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	fileName := "shared_name"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = "label1"
	file1.Name = fileName
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = "label2"
	file2.Name = fileName
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFileWithID("file3_id", colonyName, now)
	file3.Label = "label3"
	file3.Name = "different_name"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	// Get labels for files with shared_name
	labels, err := db.GetFileLabelsByName(colonyName, fileName, true)
	assert.Nil(t, err)
	assert.Len(t, labels, 2)
	
	// Check that we have the expected label names
	labelNames := make([]string, len(labels))
	for i, label := range labels {
		labelNames[i] = label.Name
	}
	assert.Contains(t, labelNames, "label1")
	assert.Contains(t, labelNames, "label2")
	assert.NotContains(t, labelNames, "label3")

	// Test non-existing file name
	emptyLabels, err := db.GetFileLabelsByName(colonyName, "non_existing_name", true)
	assert.Nil(t, err)
	assert.Empty(t, emptyLabels)
}

func TestRemoveFilesByLabel(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colonyName, now)
	file1.Label = "target_label"
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colonyName, now)
	file2.Label = "target_label"
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFileWithID("file3_id", colonyName, now)
	file3.Label = "other_label"
	file3.Name = "file3"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	// Remove files by label
	err = db.RemoveFilesByLabel(colonyName, "target_label")
	assert.Nil(t, err)

	// Verify target_label files are gone
	files, err := db.GetFileDataByLabel(colonyName, "target_label")
	assert.Nil(t, err)
	assert.Empty(t, files)

	// Verify other_label file still exists
	files, err = db.GetFileDataByLabel(colonyName, "other_label")
	assert.Nil(t, err)
	assert.Len(t, files, 1)

	// Test removing non-existing label - should not error
	err = db.RemoveFilesByLabel(colonyName, "non_existing_label")
	assert.Nil(t, err)
}

func TestRemoveFilesByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := "colony1"
	colony2 := "colony2"
	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", colony1, now)
	file1.Label = "label1"
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", colony1, now)
	file2.Label = "label2"
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFileWithID("file3_id", colony2, now)
	file3.Label = "label3"
	file3.Name = "file3"
	err = db.AddFile(file3)
	assert.Nil(t, err)

	// Remove all files from colony1
	err = db.RemoveFilesByColonyName(colony1)
	assert.Nil(t, err)

	// Verify colony1 files are gone
	labels, err := db.GetFileLabels(colony1)
	assert.Nil(t, err)
	assert.Empty(t, labels)

	// Verify colony2 files still exist
	labels, err = db.GetFileLabels(colony2)
	assert.Nil(t, err)
	assert.Len(t, labels, 1)

	// Test removing non-existing colony - should not error
	err = db.RemoveFilesByColonyName("non_existing_colony")
	assert.Nil(t, err)
}

func TestRemoveAllFiles(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	now := time.Now()

	file1 := utils.CreateTestFileWithID("file1_id", "colony1", now)
	file1.Label = "label1"
	file1.Name = "file1"
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFileWithID("file2_id", "colony2", now)
	file2.Label = "label2"
	file2.Name = "file2"
	err = db.AddFile(file2)
	assert.Nil(t, err)

	// Remove all files
	err = db.RemoveAllFiles()
	assert.Nil(t, err)

	// Verify all files are gone
	labels1, err := db.GetFileLabels("colony1")
	assert.Nil(t, err)
	assert.Empty(t, labels1)

	labels2, err := db.GetFileLabels("colony2")
	assert.Nil(t, err)
	assert.Empty(t, labels2)
}

func TestFileComplexScenarios(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := "test_colony"
	label := "test_label"
	now := time.Now()

	// Test file with complex metadata
	complexFile := utils.CreateTestFileWithID("complex_id", colonyName, now)
	complexFile.Label = label
	complexFile.Name = "complex_file"
	complexFile.Size = 1024 * 1024 // 1MB
	complexFile.Checksum = "sha256:abcdef123456"
	complexFile.ChecksumAlg = "sha256"
	complexFile.SequenceNumber = 42

	err = db.AddFile(complexFile)
	assert.Nil(t, err)

	// Retrieve and verify all metadata
	retrieved, err := db.GetFileByID(colonyName, "complex_id")
	assert.Nil(t, err)
	assert.True(t, complexFile.Equals(retrieved))
	assert.Equal(t, retrieved.Size, int64(1024*1024))
	assert.Equal(t, retrieved.Checksum, "sha256:abcdef123456")
	assert.Equal(t, retrieved.ChecksumAlg, "sha256")
	assert.Equal(t, retrieved.SequenceNumber, int64(42))

	// Test file versioning scenario
	baseTime := time.Now()
	
	v1 := utils.CreateTestFileWithID("v1_id", colonyName, baseTime)
	v1.Label = "versioned"
	v1.Name = "document.txt"
	v1.SequenceNumber = 1
	err = db.AddFile(v1)
	assert.Nil(t, err)

	v2 := utils.CreateTestFileWithID("v2_id", colonyName, baseTime.Add(time.Hour))
	v2.Label = "versioned"
	v2.Name = "document.txt"
	v2.SequenceNumber = 2
	err = db.AddFile(v2)
	assert.Nil(t, err)

	v3 := utils.CreateTestFileWithID("v3_id", colonyName, baseTime.Add(2*time.Hour))
	v3.Label = "versioned"
	v3.Name = "document.txt"
	v3.SequenceNumber = 3
	err = db.AddFile(v3)
	assert.Nil(t, err)

	// Get latest version
	latestVersions, err := db.GetLatestFileByName(colonyName, "versioned", "document.txt")
	assert.Nil(t, err)
	assert.Len(t, latestVersions, 1)
	assert.Equal(t, latestVersions[0].ID, "v3_id")
	assert.Equal(t, latestVersions[0].SequenceNumber, int64(3))

	// Get all versions
	allVersions, err := db.GetFileByName(colonyName, "versioned", "document.txt")
	assert.Nil(t, err)
	assert.Len(t, allVersions, 3)
}