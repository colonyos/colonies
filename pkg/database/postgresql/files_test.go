package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddGetFile(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file := utils.CreateTestFile("test_id", "test_colonyid", now)
	err = db.AddFile(file)
	assert.Nil(t, err)

	fileFromDB, err := db.GetFileByID("test_colonyid", file.ID)
	assert.Nil(t, err)

	// Set SequenceNumber and Added timestamp to same to make comparison possible
	fileFromDB.SequenceNumber = 1
	fileFromDB.Added = time.Time{}
	file.SequenceNumber = 1
	file.Added = time.Time{}

	assert.True(t, file.Equals(fileFromDB))
}

func TestGetFileByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file1.Prefix = "/testpath"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testpath"
	file2.Name = "test_file.txt"
	file2.Size = 2 // NOTE we changed the size to 2
	err = db.AddFile(file2)
	assert.Nil(t, err)

	fileFromDB, err := db.GetLatestFileByName("test_colonyid", file1.Prefix, file1.Name)
	assert.Nil(t, err)
	assert.Len(t, fileFromDB, 1)
	assert.Equal(t, fileFromDB[0].Size, int64(2))

	filesFromDB, err := db.GetFileByName("test_colonyid", file1.Prefix, file1.Name)
	assert.Nil(t, err)
	assert.Len(t, filesFromDB, 2)
}

func TestGetFileNamesByPrefix(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file1.ID = core.GenerateRandomID()
	file1.Prefix = "/testpath"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testdir"
	file2.Name = "test_file.txt"
	file2.Size = 1
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file3.ID = core.GenerateRandomID()
	file3.Prefix = "/testdir"
	file3.Name = "test_file2.txt"
	file3.Size = 1
	err = db.AddFile(file3)
	assert.Nil(t, err)

	file4 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file4.ID = core.GenerateRandomID()
	file4.Prefix = "/testdir2"
	file4.Name = "test_file.txt"
	file4.Size = 1
	err = db.AddFile(file4)
	assert.Nil(t, err)

	filesnames, err := db.GetFileNamesByPrefix("test_colonyid", "/testdir")
	assert.Nil(t, err)
	assert.Len(t, filesnames, 2)

	filesnames, err = db.GetFileNamesByPrefix("test_colonyid", "/testdir2")
	assert.Nil(t, err)
	assert.Len(t, filesnames, 1)
}

func TestDeleteFileByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file1.ID = core.GenerateRandomID()
	file1.Prefix = "/testdir"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testdir"
	file2.Name = "test_file2.txt"
	file2.Size = 1
	err = db.AddFile(file2)
	assert.Nil(t, err)

	filesnames, err := db.GetFileNamesByPrefix("test_colonyid", "/testdir")
	assert.Nil(t, err)
	assert.Len(t, filesnames, 2)

	file1FromDB, err := db.GetFileByID("test_colonyid", file2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, file1FromDB)

	err = db.DeleteFileByID("test_colonyid", file2.ID)
	assert.Nil(t, err)

	filesnames, err = db.GetFileNamesByPrefix("test_colonyid", "/testdir")
	assert.Nil(t, err)
	assert.Len(t, filesnames, 1)

	file1FromDB, err = db.GetFileByID("test_colonyid", file2.ID)
	assert.Nil(t, err)
	assert.Nil(t, file1FromDB)
}

func TestDeleteFilesByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid1", now)
	file1.ID = core.GenerateRandomID()
	file1.Prefix = "/testdir"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid2", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testdir"
	file2.Name = "test_file2.txt"
	file2.Size = 1
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFile("test_id", "test_colonyid2", now)
	file3.ID = core.GenerateRandomID()
	file3.Prefix = "/testdir"
	file3.Name = "test_file3.txt"
	file3.Size = 1
	err = db.AddFile(file3)
	assert.Nil(t, err)

	files, err := db.CountFiles("test_colonyid1")
	assert.Nil(t, err)
	assert.Equal(t, files, 1)

	files, err = db.CountFiles("test_colonyid2")
	assert.Nil(t, err)
	assert.Equal(t, files, 2)

	err = db.DeleteFilesByColonyID("test_colonyid2")
	assert.Nil(t, err)

	files, err = db.CountFiles("test_colonyid1")
	assert.Nil(t, err)
	assert.Equal(t, files, 1)

	files, err = db.CountFiles("test_colonyid2")
	assert.Nil(t, err)
	assert.Equal(t, files, 0)
}

func TestDeleteFileByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file1.ID = core.GenerateRandomID()
	file1.Prefix = "/testdir"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testdir"
	file2.Name = "test_file2.txt"
	file2.Size = 1
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file3.ID = core.GenerateRandomID()
	file3.Prefix = "/testdir"
	file3.Name = "test_file2.txt"
	file3.Size = 1
	err = db.AddFile(file3)
	assert.Nil(t, err)

	file4 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file4.ID = core.GenerateRandomID()
	file4.Prefix = "/testdir"
	file4.Name = "test_file2.txt"
	file4.Size = 1
	err = db.AddFile(file4)
	assert.Nil(t, err)

	files, err := db.GetFileByName("test_colonyid", file4.Prefix, file4.Name)
	assert.Nil(t, err)
	assert.Len(t, files, 3)

	err = db.DeleteFileByID("test_colonyid", file4.ID)
	assert.Nil(t, err)

	files, err = db.GetFileByName("test_colonyid", file4.Prefix, file4.Name)
	assert.Nil(t, err)
	assert.Len(t, files, 2)

	err = db.DeleteFileByName("test_colonyid", file4.Prefix, file4.Name)
	assert.Nil(t, err)

	files, err = db.GetFileByName("test_colonyid", file4.Prefix, file4.Name)
	assert.Nil(t, err)
	assert.Len(t, files, 0)

	fileFromDB, err := db.GetFileByID("test_colonyid", file4.ID)
	assert.Nil(t, err)
	assert.Nil(t, fileFromDB)

	fileFromDB, err = db.GetFileByID("test_colonyid", file1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, fileFromDB)
}

func TestGetFilePrefixes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	now := time.Now()
	file1 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file1.ID = core.GenerateRandomID()
	file1.Prefix = "/testdir1"
	file1.Name = "test_file.txt"
	file1.Size = 1
	err = db.AddFile(file1)
	assert.Nil(t, err)

	file2 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file2.ID = core.GenerateRandomID()
	file2.Prefix = "/testdir2"
	file2.Name = "test_file2.txt"
	file2.Size = 1
	err = db.AddFile(file2)
	assert.Nil(t, err)

	file3 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file3.ID = core.GenerateRandomID()
	file3.Prefix = "/testdir3"
	file3.Name = "test_file3.txt"
	file3.Size = 1
	err = db.AddFile(file3)
	assert.Nil(t, err)

	file4 := utils.CreateTestFile("test_id", "test_colonyid", now)
	file4.ID = core.GenerateRandomID()
	file4.Prefix = "/testdir3"
	file4.Name = "test_file4.txt"
	file4.Size = 1
	err = db.AddFile(file4)
	assert.Nil(t, err)

	prefixes, err := db.GetFilePrefixes("test_colonyid")
	assert.Nil(t, err)
	assert.Len(t, prefixes, 3)
}
