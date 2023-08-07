package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsFileEquals(t *testing.T) {
	now := time.Now()

	s3Object1 := S3Object{
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
	fileRef1 := FileReference{Protocol: "s3", S3Object: s3Object1}
	file1 := File{
		ID:             "test_id",
		ColonyID:       "test_colonyid",
		Prefix:         "test_prefix",
		Name:           "test_name",
		Size:           1111,
		SequenceNumber: 1,
		Checksum:       "test_checksum",
		ChecksumAlg:    "test_checksumalg",
		FileReference:  fileRef1,
		Added:          now}

	s3Object2 := S3Object{
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
	fileRef2 := FileReference{Protocol: "s3", S3Object: s3Object2}
	file2 := File{
		ID:             "test_id",
		ColonyID:       "test_colonyid",
		Prefix:         "test_prefix",
		Name:           "test_name",
		Size:           1111,
		SequenceNumber: 1,
		Checksum:       "test_checksum",
		ChecksumAlg:    "test_checksumalg",
		FileReference:  fileRef2,
		Added:          now}

	assert.True(t, file1.Equals(file2))
	file1.Name = "changed_name"
	assert.False(t, file1.Equals(file2))
}

func TestFileToJSON(t *testing.T) {
	now := time.Now()

	s3Object1 := S3Object{
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
	fileRef1 := FileReference{Protocol: "s3", S3Object: s3Object1}
	file1 := File{
		ID:             "test_id",
		ColonyID:       "test_colonyid",
		Prefix:         "test_prefix",
		Name:           "test_name",
		Size:           1111,
		SequenceNumber: 1,
		Checksum:       "test_checksum",
		ChecksumAlg:    "test_checksumalg",
		FileReference:  fileRef1,
		Added:          now}

	jsonStr, err := file1.ToJSON()
	assert.Nil(t, err)

	file2, err := ConvertJSONToFile(jsonStr)
	assert.Nil(t, err)
	assert.True(t, file1.Equals(file2))
}
