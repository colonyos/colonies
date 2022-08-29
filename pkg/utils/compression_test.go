package utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompression(t *testing.T) {
	// Create an empty directory
	bundleTestDir, err := ioutil.TempDir("/tmp/", "bundletest")
	assert.Nil(t, err)

	// Add a test file to the directory
	dummyFile, err := ioutil.TempFile(bundleTestDir, "dummydata")
	assert.Nil(t, err)
	_, err = dummyFile.Write([]byte("testdata"))
	assert.Nil(t, err)

	// Compress the directory to byte buffer
	var buf bytes.Buffer
	err = Compress("/tmp", bundleTestDir, &buf)
	assert.Nil(t, err)

	// Save the byte buffer to a tar-ball
	bundleFile, err := os.OpenFile("/tmp/bundle.tar.gz", os.O_RDWR|os.O_CREATE, 0600)
	assert.Nil(t, err)
	byteArr := buf.Bytes()
	_, err = io.Copy(bundleFile, &buf)
	assert.Nil(t, err)
	bundleFile.Close()

	// Uncompress the tar-ball
	reader := bytes.NewReader(byteArr)
	bundleTestUncompressDir, err := ioutil.TempDir("/tmp/", "bundletestuncompress")
	assert.Nil(t, err)
	Decompress(reader, bundleTestUncompressDir)

	// Read the dummy file and check content
	content, err := os.ReadFile(bundleTestUncompressDir + "/" + filepath.Base(bundleTestDir) + "/" + filepath.Base(dummyFile.Name()))
	assert.Nil(t, err)
	assert.Equal(t, string(content), "testdata")

	// Clean up
	err = os.RemoveAll(bundleTestDir)
	assert.Nil(t, err)
	err = os.RemoveAll(bundleTestUncompressDir)
	assert.Nil(t, err)
	err = os.Remove("/tmp/bundle.tar.gz")
	assert.Nil(t, err)
}
