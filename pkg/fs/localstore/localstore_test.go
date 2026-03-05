package localstore

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) *LocalObjectStore {
	t.Helper()
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)
	return store
}

func TestPutAndGet(t *testing.T) {
	store := setup(t)
	data := []byte("hello world")
	err := store.Put("colony1", "obj1", bytes.NewReader(data), int64(len(data)))
	assert.NoError(t, err)

	reader, size, err := store.Get("colony1", "obj1")
	assert.NoError(t, err)
	defer reader.Close()

	assert.Equal(t, int64(len(data)), size)
	got, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestPutOverwrite(t *testing.T) {
	store := setup(t)
	data1 := []byte("version1")
	data2 := []byte("version2-longer")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data1), int64(len(data1))))
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data2), int64(len(data2))))

	reader, size, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()

	assert.Equal(t, int64(len(data2)), size)
	got, _ := io.ReadAll(reader)
	assert.Equal(t, data2, got)
}

func TestGetNonExistent(t *testing.T) {
	store := setup(t)
	_, _, err := store.Get("c", "nope")
	assert.Error(t, err)
}

func TestPutEmptyFile(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "empty", bytes.NewReader(nil), 0))

	reader, size, err := store.Get("c", "empty")
	assert.NoError(t, err)
	defer reader.Close()
	assert.Equal(t, int64(0), size)
}

func TestPutLargeFile(t *testing.T) {
	store := setup(t)
	data := make([]byte, 1024*1024) // 1 MiB
	rand.Read(data)
	assert.NoError(t, store.Put("c", "large", bytes.NewReader(data), int64(len(data))))

	reader, size, err := store.Get("c", "large")
	assert.NoError(t, err)
	defer reader.Close()
	assert.Equal(t, int64(len(data)), size)
	got, _ := io.ReadAll(reader)
	assert.Equal(t, data, got)
}

// Chunked upload tests

func TestChunkedUpload(t *testing.T) {
	store := setup(t)

	chunk0 := []byte("aaaa")
	chunk1 := []byte("bbbb")
	chunk2 := []byte("cccc")

	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader(chunk0), int64(len(chunk0))))
	assert.NoError(t, store.PutChunk("c", "obj", 1, bytes.NewReader(chunk1), int64(len(chunk1))))
	assert.NoError(t, store.PutChunk("c", "obj", 2, bytes.NewReader(chunk2), int64(len(chunk2))))

	assert.NoError(t, store.AssembleChunks("c", "obj", 3))

	reader, size, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	assert.Equal(t, int64(12), size)
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("aaaabbbbcccc"), got)
}

func TestChunkedUploadOutOfOrder(t *testing.T) {
	store := setup(t)

	assert.NoError(t, store.PutChunk("c", "obj", 2, bytes.NewReader([]byte("cc")), 2))
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))
	assert.NoError(t, store.PutChunk("c", "obj", 1, bytes.NewReader([]byte("bb")), 2))

	assert.NoError(t, store.AssembleChunks("c", "obj", 3))

	reader, _, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("aabbcc"), got)
}

func TestAssembleMissingChunk(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))
	// Missing chunk 1
	err := store.AssembleChunks("c", "obj", 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing chunk 1")
}

func TestGetChunkStatus(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("a")), 1))
	assert.NoError(t, store.PutChunk("c", "obj", 2, bytes.NewReader([]byte("c")), 1))

	indices, err := store.GetChunkStatus("c", "obj")
	assert.NoError(t, err)
	assert.Equal(t, []int{0, 2}, indices)
}

func TestGetChunkStatusEmpty(t *testing.T) {
	store := setup(t)
	indices, err := store.GetChunkStatus("c", "nope")
	assert.NoError(t, err)
	assert.Empty(t, indices)
}

func TestChunkedUploadCleansStaging(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))
	assert.NoError(t, store.AssembleChunks("c", "obj", 1))

	stagingDir := store.stagingDir("c", "obj")
	_, err := os.Stat(stagingDir)
	assert.True(t, os.IsNotExist(err))
}

// Range read tests

func TestGetRange(t *testing.T) {
	store := setup(t)
	data := []byte("0123456789")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 3, 4)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("3456"), got)
}

func TestGetRangeFromStart(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 0, 3)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("abc"), got)
}

func TestGetRangeEnd(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 4, 2)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("ef"), got)
}

func TestGetRangeBeyondEOF(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 4, 100)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("ef"), got)
}

func TestGetRangeOffsetBeyondSize(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	_, err := store.GetRange("c", "obj", 100, 1)
	assert.Error(t, err)
}

// Colony isolation tests

func TestColonyIsolation(t *testing.T) {
	store := setup(t)

	data1 := []byte("colony1-data")
	data2 := []byte("colony2-data")
	assert.NoError(t, store.Put("c1", "obj", bytes.NewReader(data1), int64(len(data1))))
	assert.NoError(t, store.Put("c2", "obj", bytes.NewReader(data2), int64(len(data2))))

	reader1, _, _ := store.Get("c1", "obj")
	got1, _ := io.ReadAll(reader1)
	reader1.Close()

	reader2, _, _ := store.Get("c2", "obj")
	got2, _ := io.ReadAll(reader2)
	reader2.Close()

	assert.Equal(t, data1, got1)
	assert.Equal(t, data2, got2)
}

func TestListCrossColony(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c1", "a", bytes.NewReader([]byte("x")), 1))
	assert.NoError(t, store.Put("c1", "b", bytes.NewReader([]byte("x")), 1))
	assert.NoError(t, store.Put("c2", "c", bytes.NewReader([]byte("x")), 1))

	list1, err := store.List("c1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, list1)

	list2, err := store.List("c2")
	assert.NoError(t, err)
	assert.Equal(t, []string{"c"}, list2)
}

// Exists, Stat, Remove, RemoveAll, DiskUsage

func TestExists(t *testing.T) {
	store := setup(t)
	assert.False(t, store.Exists("c", "obj"))
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("x")), 1))
	assert.True(t, store.Exists("c", "obj"))
}

func TestStat(t *testing.T) {
	store := setup(t)
	data := []byte("hello")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	info, err := store.Stat("c", "obj")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), info.Size)
}

func TestStatNonExistent(t *testing.T) {
	store := setup(t)
	_, err := store.Stat("c", "nope")
	assert.Error(t, err)
}

func TestRemove(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("x")), 1))
	assert.NoError(t, store.Remove("c", "obj"))
	assert.False(t, store.Exists("c", "obj"))
}

func TestRemoveNonExistent(t *testing.T) {
	store := setup(t)
	err := store.Remove("c", "nope")
	assert.NoError(t, err) // idempotent
}

func TestRemoveStaging(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("a")), 1))
	assert.NoError(t, store.RemoveStaging("c", "obj"))

	indices, err := store.GetChunkStatus("c", "obj")
	assert.NoError(t, err)
	assert.Empty(t, indices)
}

func TestRemoveAll(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "a", bytes.NewReader([]byte("x")), 1))
	assert.NoError(t, store.Put("c", "b", bytes.NewReader([]byte("y")), 1))

	assert.NoError(t, store.RemoveAll("c"))

	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Empty(t, list)
}

func TestDiskUsage(t *testing.T) {
	store := setup(t)
	data1 := []byte("hello")
	data2 := []byte("world!")
	assert.NoError(t, store.Put("c", "a", bytes.NewReader(data1), int64(len(data1))))
	assert.NoError(t, store.Put("c", "b", bytes.NewReader(data2), int64(len(data2))))

	usage, err := store.DiskUsage("c")
	assert.NoError(t, err)
	assert.Equal(t, int64(11), usage)
}

func TestDiskUsageEmptyColony(t *testing.T) {
	store := setup(t)
	usage, err := store.DiskUsage("nonexistent")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), usage)
}

func TestListEmpty(t *testing.T) {
	store := setup(t)
	list, err := store.List("nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, list)
}

// Path security tests

func TestRejectDotDotInColonyName(t *testing.T) {
	store := setup(t)
	err := store.Put("../evil", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "..")
}

func TestRejectDotDotInObjectName(t *testing.T) {
	store := setup(t)
	err := store.Put("c", "../../etc/passwd", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

func TestRejectSlashInColonyName(t *testing.T) {
	store := setup(t)
	err := store.Put("a/b", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path separator")
}

func TestRejectSlashInObjectName(t *testing.T) {
	store := setup(t)
	err := store.Put("c", "a/b", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

func TestRejectBackslash(t *testing.T) {
	store := setup(t)
	err := store.Put("c", "a\\b", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

func TestRejectNullByte(t *testing.T) {
	store := setup(t)
	err := store.Put("c", "obj\x00evil", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null")
}

func TestRejectEmptyName(t *testing.T) {
	store := setup(t)
	err := store.Put("", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)

	err = store.Put("c", "", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

// Test that partial writes leave no corrupt file

func TestPartialWriteNoCorruptFile(t *testing.T) {
	store := setup(t)

	// Put a valid file first
	data := []byte("good data")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	// Try to overwrite with a reader that fails mid-stream
	failReader := &failingReader{data: []byte("bad d"), failAfter: 5}
	err := store.Put("c", "obj", failReader, 100)
	assert.Error(t, err)

	// Original file should still be intact
	reader, _, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, data, got)
}

func TestNewLocalObjectStoreCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	_, err = os.Stat(dir)
	assert.NoError(t, err)
}

func TestNegativeChunkIndex(t *testing.T) {
	store := setup(t)
	err := store.PutChunk("c", "obj", -1, bytes.NewReader([]byte("a")), 1)
	assert.Error(t, err)
}

func TestMultipleObjectsInColony(t *testing.T) {
	store := setup(t)
	for i := 0; i < 100; i++ {
		name := strings.Repeat("x", i+1)
		assert.NoError(t, store.Put("c", name, bytes.NewReader([]byte("data")), 4))
	}
	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Len(t, list, 100)
}

// Validation error paths for every method

func TestPutInvalidColonyName(t *testing.T) {
	store := setup(t)
	r := bytes.NewReader([]byte("x"))
	assert.Error(t, store.Put("", "obj", r, 1))
	assert.Error(t, store.Put("..", "obj", bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.Put("a/b", "obj", bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.Put("a\\b", "obj", bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.Put("a\x00b", "obj", bytes.NewReader([]byte("x")), 1))
}

func TestPutInvalidObjectName(t *testing.T) {
	store := setup(t)
	assert.Error(t, store.Put("c", "", bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.Put("c", "..", bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.Put("c", "a/b", bytes.NewReader([]byte("x")), 1))
}

func TestPutChunkInvalidNames(t *testing.T) {
	store := setup(t)
	r := bytes.NewReader([]byte("x"))
	assert.Error(t, store.PutChunk("", "obj", 0, r, 1))
	assert.Error(t, store.PutChunk("c", "", 0, bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.PutChunk("..", "obj", 0, bytes.NewReader([]byte("x")), 1))
	assert.Error(t, store.PutChunk("c", "..", 0, bytes.NewReader([]byte("x")), 1))
}

func TestAssembleChunksInvalidNames(t *testing.T) {
	store := setup(t)
	assert.Error(t, store.AssembleChunks("", "obj", 1))
	assert.Error(t, store.AssembleChunks("c", "", 1))
	assert.Error(t, store.AssembleChunks("..", "obj", 1))
	assert.Error(t, store.AssembleChunks("c", "..", 1))
}

func TestGetChunkStatusInvalidNames(t *testing.T) {
	store := setup(t)
	_, err := store.GetChunkStatus("", "obj")
	assert.Error(t, err)
	_, err = store.GetChunkStatus("c", "")
	assert.Error(t, err)
}

func TestGetInvalidNames(t *testing.T) {
	store := setup(t)
	_, _, err := store.Get("", "obj")
	assert.Error(t, err)
	_, _, err = store.Get("c", "")
	assert.Error(t, err)
}

func TestGetRangeInvalidNames(t *testing.T) {
	store := setup(t)
	_, err := store.GetRange("", "obj", 0, 1)
	assert.Error(t, err)
	_, err = store.GetRange("c", "", 0, 1)
	assert.Error(t, err)
}

func TestGetRangeNonExistent(t *testing.T) {
	store := setup(t)
	_, err := store.GetRange("c", "nope", 0, 1)
	assert.Error(t, err)
}

func TestRemoveInvalidNames(t *testing.T) {
	store := setup(t)
	assert.Error(t, store.Remove("", "obj"))
	assert.Error(t, store.Remove("c", ""))
	assert.Error(t, store.Remove("..", "obj"))
}

func TestRemoveStagingInvalidNames(t *testing.T) {
	store := setup(t)
	assert.Error(t, store.RemoveStaging("", "obj"))
	assert.Error(t, store.RemoveStaging("c", ""))
}

func TestRemoveStagingNonExistent(t *testing.T) {
	store := setup(t)
	// Should succeed even if nothing to remove
	assert.NoError(t, store.RemoveStaging("c", "nope"))
}

func TestExistsInvalidNames(t *testing.T) {
	store := setup(t)
	assert.False(t, store.Exists("", "obj"))
	assert.False(t, store.Exists("c", ""))
}

func TestStatInvalidNames(t *testing.T) {
	store := setup(t)
	_, err := store.Stat("", "obj")
	assert.Error(t, err)
	_, err = store.Stat("c", "")
	assert.Error(t, err)
}

func TestListInvalidName(t *testing.T) {
	store := setup(t)
	_, err := store.List("")
	assert.Error(t, err)
	_, err = store.List("..")
	assert.Error(t, err)
}

func TestRemoveAllInvalidName(t *testing.T) {
	store := setup(t)
	assert.Error(t, store.RemoveAll(""))
	assert.Error(t, store.RemoveAll(".."))
}

func TestRemoveAllNonExistent(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.RemoveAll("nonexistent"))
}

func TestDiskUsageInvalidName(t *testing.T) {
	store := setup(t)
	_, err := store.DiskUsage("")
	assert.Error(t, err)
	_, err = store.DiskUsage("..")
	assert.Error(t, err)
}

func TestDiskUsageIncludesStaging(t *testing.T) {
	store := setup(t)
	data := []byte("hello")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))
	assert.NoError(t, store.PutChunk("c", "obj2", 0, bytes.NewReader(data), int64(len(data))))

	usage, err := store.DiskUsage("c")
	assert.NoError(t, err)
	assert.True(t, usage >= int64(len(data)*2))
}

// Permission denied paths (atomicWrite, Put to read-only dir)

func TestPutToReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Create colony dir then make it read-only
	colonyDir := filepath.Join(dir, "c", "objects")
	assert.NoError(t, os.MkdirAll(colonyDir, 0755))
	assert.NoError(t, os.Chmod(colonyDir, 0555))
	defer os.Chmod(colonyDir, 0755)

	err = store.Put("c", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

func TestPutChunkToReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Create staging dir then make it read-only
	stagingDir := filepath.Join(dir, "c", ".staging", "obj")
	assert.NoError(t, os.MkdirAll(stagingDir, 0755))
	assert.NoError(t, os.Chmod(stagingDir, 0555))
	defer os.Chmod(stagingDir, 0755)

	err = store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
}

func TestAssembleChunksToReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Upload a chunk normally
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))

	// Make objects dir read-only so assemble can't write
	objDir := filepath.Join(dir, "c", "objects")
	assert.NoError(t, os.MkdirAll(objDir, 0755))
	assert.NoError(t, os.Chmod(objDir, 0555))
	defer os.Chmod(objDir, 0755)

	err = store.AssembleChunks("c", "obj", 1)
	assert.Error(t, err)
}

func TestNewLocalObjectStoreReadOnlyParent(t *testing.T) {
	dir := t.TempDir()
	readOnly := filepath.Join(dir, "ro")
	assert.NoError(t, os.MkdirAll(readOnly, 0555))
	defer os.Chmod(readOnly, 0755)

	_, err := NewLocalObjectStore(filepath.Join(readOnly, "store"))
	assert.Error(t, err)
}

// Symlink escape detection

func TestVerifyWithinBaseRejectsSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(filepath.Join(dir, "store"))
	assert.NoError(t, err)

	// Create a colony dir, then replace it with a symlink pointing outside
	colonyObjDir := filepath.Join(dir, "store", "evil", "objects")
	assert.NoError(t, os.MkdirAll(colonyObjDir, 0755))
	assert.NoError(t, os.RemoveAll(colonyObjDir))
	outsideDir := filepath.Join(dir, "outside")
	assert.NoError(t, os.MkdirAll(outsideDir, 0755))
	assert.NoError(t, os.Symlink(outsideDir, colonyObjDir))

	err = store.Put("evil", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escapes base directory")
}

// Edge cases for chunked uploads

func TestAssembleChunksSingleChunk(t *testing.T) {
	store := setup(t)
	data := []byte("only-chunk")
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader(data), int64(len(data))))
	assert.NoError(t, store.AssembleChunks("c", "obj", 1))

	reader, size, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	assert.Equal(t, int64(len(data)), size)
	got, _ := io.ReadAll(reader)
	assert.Equal(t, data, got)
}

func TestChunkOverwrite(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("old")), 3))
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("new")), 3))
	assert.NoError(t, store.AssembleChunks("c", "obj", 1))

	reader, _, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("new"), got)
}

func TestAssembleChunksOverwritesExistingObject(t *testing.T) {
	store := setup(t)
	// Put an existing object
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("old")), 3))

	// Upload chunks and assemble to overwrite
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("new-data")), 8))
	assert.NoError(t, store.AssembleChunks("c", "obj", 1))

	reader, _, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("new-data"), got)
}

func TestGetRangeZeroLength(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	// length=0 means "to end of file" in our GetRange (clamped)
	reader, err := store.GetRange("c", "obj", 0, 0)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte(""), got)
}

func TestGetRangeExactSize(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 0, int64(len(data)))
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, data, got)
}

func TestGetRangeLastByte(t *testing.T) {
	store := setup(t)
	data := []byte("abcdef")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	reader, err := store.GetRange("c", "obj", 5, 1)
	assert.NoError(t, err)
	defer reader.Close()
	got, _ := io.ReadAll(reader)
	assert.Equal(t, []byte("f"), got)
}

func TestRemoveAfterPutThenGet(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("data")), 4))
	assert.NoError(t, store.Remove("c", "obj"))
	_, _, err := store.Get("c", "obj")
	assert.Error(t, err)
	assert.False(t, store.Exists("c", "obj"))
}

func TestListAfterRemoveAll(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "a", bytes.NewReader([]byte("x")), 1))
	assert.NoError(t, store.RemoveAll("c"))
	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Empty(t, list)
	assert.False(t, store.Exists("c", "a"))
}

// GetChunkStatus filtering: dirs, .tmp- files, non-numeric entries

func TestGetChunkStatusIgnoresDirsAndTmpFiles(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("a")), 1))
	assert.NoError(t, store.PutChunk("c", "obj", 2, bytes.NewReader([]byte("c")), 1))

	// Manually create noise inside the staging dir
	stagingDir := store.stagingDir("c", "obj")
	os.MkdirAll(filepath.Join(stagingDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(stagingDir, ".tmp-partial"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(stagingDir, "notanumber"), []byte("x"), 0644)

	indices, err := store.GetChunkStatus("c", "obj")
	assert.NoError(t, err)
	assert.Equal(t, []int{0, 2}, indices)
}

func TestGetChunkStatusReadDirError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Create staging dir then make it unreadable
	stagingDir := filepath.Join(dir, "c", ".staging", "obj")
	assert.NoError(t, os.MkdirAll(stagingDir, 0755))
	assert.NoError(t, os.Chmod(stagingDir, 0000))
	defer os.Chmod(stagingDir, 0755)

	_, err = store.GetChunkStatus("c", "obj")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read staging dir")
}

// List filtering: dirs and .tmp- files in object dir

func TestListIgnoresDirsAndTmpFiles(t *testing.T) {
	store := setup(t)
	assert.NoError(t, store.Put("c", "real-obj", bytes.NewReader([]byte("data")), 4))

	// Manually create noise inside the objects dir
	objDir := store.objectDir("c")
	os.MkdirAll(filepath.Join(objDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(objDir, ".tmp-leftover"), []byte("x"), 0644)

	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Equal(t, []string{"real-obj"}, list)
}

func TestListReadDirError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Create objects dir then make it unreadable
	objDir := filepath.Join(dir, "c", "objects")
	assert.NoError(t, os.MkdirAll(objDir, 0755))
	assert.NoError(t, os.Chmod(objDir, 0000))
	defer os.Chmod(objDir, 0755)

	_, err = store.List("c")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read object dir")
}

// AssembleChunks: chunk becomes unreadable between verify and copy

func TestAssembleChunksUnreadableChunk(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))
	assert.NoError(t, store.PutChunk("c", "obj", 1, bytes.NewReader([]byte("bb")), 2))

	// Make chunk 1 unreadable after it was written
	chunkPath := filepath.Join(dir, "c", ".staging", "obj", "1")
	assert.NoError(t, os.Chmod(chunkPath, 0000))
	defer os.Chmod(chunkPath, 0644)

	err = store.AssembleChunks("c", "obj", 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open chunk 1")
}

// Remove with permission denied (non-IsNotExist error)

func TestRemovePermissionDenied(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("data")), 4))

	// Make objects dir read-only so Remove fails with permission denied
	objDir := filepath.Join(dir, "c", "objects")
	assert.NoError(t, os.Chmod(objDir, 0555))
	defer os.Chmod(objDir, 0755)

	err = store.Remove("c", "obj")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove object")
}

// RemoveStaging with permission denied

func TestRemoveStagingPermissionDenied(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("a")), 1))

	// Make the .staging dir non-removable by removing write on parent
	parentDir := filepath.Join(dir, "c", ".staging")
	chunkDir := filepath.Join(parentDir, "obj")
	assert.NoError(t, os.Chmod(chunkDir, 0000))
	assert.NoError(t, os.Chmod(parentDir, 0555))
	defer func() {
		os.Chmod(parentDir, 0755)
		os.Chmod(chunkDir, 0755)
	}()

	err = store.RemoveStaging("c", "obj")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove staging dir")
}

// RemoveAll with permission denied

func TestRemoveAllPermissionDenied(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("data")), 4))

	// Make the colony dir non-removable
	colonyDir := filepath.Join(dir, "c")
	objDir := filepath.Join(colonyDir, "objects")
	assert.NoError(t, os.Chmod(objDir, 0000))
	assert.NoError(t, os.Chmod(colonyDir, 0555))
	defer func() {
		os.Chmod(colonyDir, 0755)
		os.Chmod(objDir, 0755)
	}()

	err = store.RemoveAll("c")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove colony dir")
}

// DiskUsage with Walk encountering permission error

func TestDiskUsageWalkError(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.Put("c", "obj", bytes.NewReader([]byte("data")), 4))

	// Make objects dir unreadable to cause Walk error
	objDir := filepath.Join(dir, "c", "objects")
	assert.NoError(t, os.Chmod(objDir, 0000))
	defer os.Chmod(objDir, 0755)

	_, err = store.DiskUsage("c")
	assert.Error(t, err)
}

// Put where MkdirAll fails (colony parent dir is read-only)

func TestPutMkdirAllFails(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Make base dir read-only so MkdirAll for colony/objects fails
	assert.NoError(t, os.Chmod(dir, 0555))
	defer os.Chmod(dir, 0755)

	err = store.Put("newcolony", "obj", bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create object dir")
}

// PutChunk where MkdirAll fails

func TestPutChunkMkdirAllFails(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Make base dir read-only so MkdirAll for staging dir fails
	assert.NoError(t, os.Chmod(dir, 0555))
	defer os.Chmod(dir, 0755)

	err = store.PutChunk("newcolony", "obj", 0, bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create staging dir")
}

// AssembleChunks where MkdirAll for object dir fails

func TestAssembleChunksMkdirAllFails(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	assert.NoError(t, store.PutChunk("c", "obj", 0, bytes.NewReader([]byte("aa")), 2))

	// Make base dir read-only so creating c/objects fails
	// First remove the objects dir if it exists
	os.RemoveAll(filepath.Join(dir, "c", "objects"))
	assert.NoError(t, os.Chmod(filepath.Join(dir, "c"), 0555))
	defer os.Chmod(filepath.Join(dir, "c"), 0755)

	err = store.AssembleChunks("c", "obj", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create object dir")
}

// verifyWithinBase: grandparent fallback and both-fail paths

func TestVerifyWithinBaseParentMissing(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Path where parent doesn't exist but grandparent does and is within base
	// baseDir/colony/objects/objectName -- objects/ doesn't exist
	path := filepath.Join(dir, "colony", "objects", "obj")
	os.MkdirAll(filepath.Join(dir, "colony"), 0755) // grandparent exists
	err = store.verifyWithinBase(path)
	assert.NoError(t, err) // should succeed via grandparent fallback
}

func TestVerifyWithinBaseBothParentsFail(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(dir)
	assert.NoError(t, err)

	// Path where neither parent nor grandparent exist
	path := filepath.Join(dir, "no", "such", "deep", "path", "obj")
	err = store.verifyWithinBase(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path verification failed")
}

func TestVerifyWithinBaseGrandparentEscapes(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(filepath.Join(dir, "store"))
	assert.NoError(t, err)

	// Grandparent exists but resolves outside base
	outsideDir := filepath.Join(dir, "outside")
	os.MkdirAll(outsideDir, 0755)
	// Path: outsideDir/nonexistent/obj -> parent doesn't exist, grandparent is outsideDir
	path := filepath.Join(outsideDir, "nonexistent", "obj")
	err = store.verifyWithinBase(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escapes base directory")
}

// PutChunk and AssembleChunks with symlink escape

func TestPutChunkSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(filepath.Join(dir, "store"))
	assert.NoError(t, err)

	// Create staging dir structure, then replace staging/obj with symlink
	stagingDir := filepath.Join(dir, "store", "evil", ".staging", "obj")
	os.MkdirAll(stagingDir, 0755)
	os.RemoveAll(stagingDir)
	outsideDir := filepath.Join(dir, "outside")
	os.MkdirAll(outsideDir, 0755)
	os.Symlink(outsideDir, stagingDir)

	err = store.PutChunk("evil", "obj", 0, bytes.NewReader([]byte("x")), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escapes base directory")
}

func TestAssembleChunksSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalObjectStore(filepath.Join(dir, "store"))
	assert.NoError(t, err)

	// Upload chunks normally
	assert.NoError(t, store.PutChunk("evil", "obj", 0, bytes.NewReader([]byte("aa")), 2))

	// Replace objects dir with symlink pointing outside
	objDir := filepath.Join(dir, "store", "evil", "objects")
	os.MkdirAll(objDir, 0755)
	os.RemoveAll(objDir)
	outsideDir := filepath.Join(dir, "outside")
	os.MkdirAll(outsideDir, 0755)
	os.Symlink(outsideDir, objDir)

	err = store.AssembleChunks("evil", "obj", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "escapes base directory")
}

// failingReader returns data up to failAfter bytes then returns an error
type failingReader struct {
	data      []byte
	failAfter int
	pos       int
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.pos >= r.failAfter {
		return 0, io.ErrUnexpectedEOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= r.failAfter {
		return n, io.ErrUnexpectedEOF
	}
	return n, nil
}
