package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestFileData() *FileData {
	fileData := &FileData{Name: "test_name", Checksum: "test_checksum", Size: 1, S3Filename: "test_s3_filename"}
	return fileData
}

func TestIsFileDataEquals(t *testing.T) {
	fileData1 := createTestFileData()
	fileData2 := createTestFileData()
	assert.True(t, fileData1.Equals(fileData2))
	fileData1.Name = "changed_name"
	assert.False(t, fileData1.Equals(fileData2))
}

func TestFileDataToJSON(t *testing.T) {
	fileData1 := createTestFileData()
	jsonStr, err := fileData1.ToJSON()
	assert.Nil(t, err)

	fileData2, err := ConvertJSONToFileData(jsonStr)
	assert.Nil(t, err)
	assert.True(t, fileData1.Equals(fileData2))
}

func TestIsFileDataArraysEquals(t *testing.T) {
	fileData1 := createTestFileData()
	fileData2 := createTestFileData()
	fileData3 := createTestFileData()
	fileData4 := createTestFileData()
	fileData4.Name = "chan"
	fileDataArr1 := []*FileData{fileData1, fileData2}
	fileDataArr2 := []*FileData{fileData3, fileData4}
	assert.True(t, IsFileDataArraysEqual(fileDataArr1, fileDataArr1))
	assert.False(t, IsFileDataArraysEqual(fileDataArr1, fileDataArr2))
}

func TestFileDataArrayToJSON(t *testing.T) {
	fileData1 := createTestFileData()
	fileData2 := createTestFileData()
	fileDataArr1 := []*FileData{fileData1, fileData2}

	jsonStr, err := ConvertFileDataArrayToJSON(fileDataArr1)
	assert.Nil(t, err)

	fileDataArr2, err := ConvertJSONToFileDataArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsFileDataArraysEqual(fileDataArr1, fileDataArr2))
}
