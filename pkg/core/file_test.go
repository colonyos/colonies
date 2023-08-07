package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestFile() *File {
	s3Object := S3Object{
		Server:        "test_server",
		Port:          1111,
		TLS:           true,
		AccessKey:     "test_accesskey",
		SecretKey:     "test_secretkey",
		Region:        "test_region",
		EncryptionKey: "test_encrytionkey",
		EncryptionAlg: "test_encrytionalg",
		Object:        "test_object",
		Bucket:        "test_bucket",
	}
	ref := Reference{Protocol: "s3", S3Object: s3Object}
	file := File{
		ID:          "test_id",
		ColonyID:    "test_colonyid",
		Prefix:      "test_prefix",
		Name:        "test_name",
		Size:        1111,
		Checksum:    "test_checksum",
		ChecksumAlg: "test_checksumalg",
		Reference:   ref,
		Added:       time.Time{}}

	return &file
}

func TestIsFileEquals(t *testing.T) {
	file1 := createTestFile()
	file2 := createTestFile()
	assert.True(t, file1.Equals(file2))
	file1.Name = "changed_name"
	assert.False(t, file1.Equals(file2))
}

func TestFileToJSON(t *testing.T) {
	file1 := createTestFile()
	jsonStr, err := file1.ToJSON()
	assert.Nil(t, err)

	file2, err := ConvertJSONToFile(jsonStr)
	assert.Nil(t, err)
	assert.True(t, file1.Equals(file2))
}

func TestIsFileArraysEquals(t *testing.T) {
	file1 := createTestFile()
	file2 := createTestFile()
	file3 := createTestFile()
	file4 := createTestFile()
	file4.Name = "chan"
	files1 := []*File{file1, file2}
	files2 := []*File{file3, file4}
	assert.True(t, IsFileArraysEqual(files1, files1))
	assert.False(t, IsFileArraysEqual(files1, files2))
}

func TestFileArrayToJSON(t *testing.T) {
	file1 := createTestFile()
	file2 := createTestFile()
	files1 := []*File{file1, file2}

	jsonStr, err := ConvertFileArrayToJSON(files1)
	assert.Nil(t, err)

	files2, err := ConvertJSONToFileArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsFileArraysEqual(files1, files2))
}
