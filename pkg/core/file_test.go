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
		ColonyName:  "test_colony",
		Label:       "test_label",
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

func TestFileEqualsAllFields(t *testing.T) {
	// Test each field comparison in the Equals function
	tests := []struct {
		name   string
		modify func(f *File)
	}{
		{"Server", func(f *File) { f.Reference.S3Object.Server = "different" }},
		{"Port", func(f *File) { f.Reference.S3Object.Port = 9999 }},
		{"TLS", func(f *File) { f.Reference.S3Object.TLS = false }},
		{"AccessKey", func(f *File) { f.Reference.S3Object.AccessKey = "different" }},
		{"SecretKey", func(f *File) { f.Reference.S3Object.SecretKey = "different" }},
		{"Region", func(f *File) { f.Reference.S3Object.Region = "different" }},
		{"EncryptionKey", func(f *File) { f.Reference.S3Object.EncryptionKey = "different" }},
		{"EncryptionAlg", func(f *File) { f.Reference.S3Object.EncryptionAlg = "different" }},
		{"Object", func(f *File) { f.Reference.S3Object.Object = "different" }},
		{"Bucket", func(f *File) { f.Reference.S3Object.Bucket = "different" }},
		{"Protocol", func(f *File) { f.Reference.Protocol = "different" }},
		{"ID", func(f *File) { f.ID = "different" }},
		{"ColonyName", func(f *File) { f.ColonyName = "different" }},
		{"Label", func(f *File) { f.Label = "different" }},
		{"Size", func(f *File) { f.Size = 9999 }},
		{"SequenceNumber", func(f *File) { f.SequenceNumber = 9999 }},
		{"Checksum", func(f *File) { f.Checksum = "different" }},
		{"ChecksumAlg", func(f *File) { f.ChecksumAlg = "different" }},
		{"Added", func(f *File) { f.Added = time.Now() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file1 := createTestFile()
			file2 := createTestFile()
			tt.modify(file2)
			assert.False(t, file1.Equals(file2), "Expected files to be not equal when %s differs", tt.name)
		})
	}
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

func TestIsFileArraysEqualEdgeCases(t *testing.T) {
	file1 := createTestFile()
	file2 := createTestFile()

	// Test nil arrays
	assert.False(t, IsFileArraysEqual(nil, []*File{file1}))
	assert.False(t, IsFileArraysEqual([]*File{file1}, nil))
	assert.False(t, IsFileArraysEqual(nil, nil))

	// Test different lengths
	assert.False(t, IsFileArraysEqual([]*File{file1}, []*File{file1, file2}))
}

func TestConvertJSONToFileError(t *testing.T) {
	_, err := ConvertJSONToFile("invalid json")
	assert.NotNil(t, err)
}

func TestConvertJSONToFileArrayError(t *testing.T) {
	_, err := ConvertJSONToFileArray("invalid json")
	assert.NotNil(t, err)
}
