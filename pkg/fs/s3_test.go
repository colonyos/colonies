package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateS3Client(t *testing.T) {
	_, err := CreateS3Client()
	assert.Nil(t, err)
}

func TestS3Upload(t *testing.T) {
	s3Client, err := CreateS3Client()
	assert.Nil(t, err)

	// Create a local file
	tmpDir, err := ioutil.TempDir("/tmp/", "test")
	assert.Nil(t, err)
	f, err := ioutil.TempFile(tmpDir, "test")
	assert.Nil(t, err)
	filename := filepath.Base(f.Name())
	data := "testdata"
	_, err = f.Write([]byte(data))
	assert.Nil(t, err)

	// Upload file to S3
	err = s3Client.Upload(tmpDir, filename, filename, int64(len(data)))
	assert.Nil(t, err)

	// Clean up
	err = os.RemoveAll(tmpDir)
	assert.Nil(t, err)
	err = s3Client.Remove(filename)
	assert.Nil(t, err)
}

func TestS3Download(t *testing.T) {
	s3Client, err := CreateS3Client()
	assert.Nil(t, err)

	// Create a local file
	srcTmpDir, err := ioutil.TempDir("/tmp/", "src")
	assert.Nil(t, err)
	f, err := ioutil.TempFile(srcTmpDir, "test")
	assert.Nil(t, err)
	filename := filepath.Base(f.Name())
	data := "testdata"
	_, err = f.Write([]byte(data))
	assert.Nil(t, err)

	// Create download dir
	dstTmpDir, err := ioutil.TempDir("/tmp/", "dst")
	assert.Nil(t, err)

	// Upload file to S3
	err = s3Client.Upload(srcTmpDir, filename, filename, int64(len(data)))
	assert.Nil(t, err)

	// Download file back to client
	err = s3Client.Download(filename, filename, dstTmpDir)
	assert.Nil(t, err)

	fileContent, err := os.ReadFile(dstTmpDir + "/" + filename)
	assert.Nil(t, err)
	assert.Equal(t, data, (string(fileContent)))

	// Clean up
	err = os.RemoveAll(srcTmpDir)
	assert.Nil(t, err)
	err = os.RemoveAll(dstTmpDir)
	assert.Nil(t, err)
	err = s3Client.Remove(filename)
	assert.Nil(t, err)
}

func TestS3Remove(t *testing.T) {
	s3Client, err := CreateS3Client()
	assert.Nil(t, err)

	// Create a local file
	tmpDir, err := ioutil.TempDir("/tmp/", "test")
	assert.Nil(t, err)
	f, err := ioutil.TempFile(tmpDir, "test")
	assert.Nil(t, err)
	filename := filepath.Base(f.Name())
	assert.Nil(t, err)
	data := "testdata"
	_, err = f.Write([]byte(data))
	assert.Nil(t, err)

	// Upload file to S3
	err = s3Client.Upload(tmpDir, filename, filename, int64(len(data)))
	assert.Nil(t, err)
	assert.True(t, s3Client.Exists(filename))

	// Remove file
	err = s3Client.Remove(filename)
	assert.Nil(t, err)
	assert.False(t, s3Client.Exists(filename))

	// Clean up
	err = os.RemoveAll(tmpDir)
	assert.Nil(t, err)
	err = s3Client.Remove(filename)
	assert.Nil(t, err)
}
