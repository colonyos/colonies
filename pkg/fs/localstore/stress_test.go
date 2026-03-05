package localstore

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentPutDifferentObjects(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup
	count := 50

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("obj-%d", idx)
			data := []byte(fmt.Sprintf("data-%d", idx))
			err := store.Put("c", name, bytes.NewReader(data), int64(len(data)))
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Len(t, list, count)
}

func TestConcurrentPutSameObject(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup
	count := 20

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			data := []byte(fmt.Sprintf("data-%04d", idx))
			store.Put("c", "obj", bytes.NewReader(data), int64(len(data)))
		}(i)
	}
	wg.Wait()

	// File should exist and not be corrupt
	reader, size, err := store.Get("c", "obj")
	assert.NoError(t, err)
	defer reader.Close()
	assert.True(t, size > 0)
	got, _ := io.ReadAll(reader)
	assert.Len(t, got, int(size))
}

func TestConcurrentGetWhilePut(t *testing.T) {
	store := setup(t)
	data := []byte("initial data")
	assert.NoError(t, store.Put("c", "obj", bytes.NewReader(data), int64(len(data))))

	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			d := []byte(fmt.Sprintf("update-%04d", idx))
			store.Put("c", "obj", bytes.NewReader(d), int64(len(d)))
		}(i)
	}

	// Readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reader, size, err := store.Get("c", "obj")
			if err != nil {
				return
			}
			defer reader.Close()
			got, _ := io.ReadAll(reader)
			assert.True(t, len(got) > 0)
			assert.True(t, size > 0)
		}()
	}

	wg.Wait()
}

func TestConcurrentPutAndRemove(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("obj-%d", idx%10)
			if idx%3 == 0 {
				store.Remove("c", name)
			} else {
				data := []byte(fmt.Sprintf("data-%d", idx))
				store.Put("c", name, bytes.NewReader(data), int64(len(data)))
			}
		}(i)
	}
	wg.Wait()

	_, err := store.List("c")
	assert.NoError(t, err)
}

func TestConcurrentChunkedUploads(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup
	objects := 5
	chunksPerObject := 4

	for i := 0; i < objects; i++ {
		wg.Add(1)
		go func(objIdx int) {
			defer wg.Done()
			name := fmt.Sprintf("obj-%d", objIdx)
			for c := 0; c < chunksPerObject; c++ {
				data := make([]byte, 1024)
				rand.Read(data)
				store.PutChunk("c", name, c, bytes.NewReader(data), int64(len(data)))
			}
			store.AssembleChunks("c", name, chunksPerObject)
		}(i)
	}
	wg.Wait()

	list, err := store.List("c")
	assert.NoError(t, err)
	assert.Len(t, list, objects)

	for _, name := range list {
		info, err := store.Stat("c", name)
		assert.NoError(t, err)
		assert.Equal(t, int64(chunksPerObject*1024), info.Size)
	}
}

func TestConcurrentMultipleColonies(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup
	colonies := 5
	objectsPerColony := 10

	for c := 0; c < colonies; c++ {
		wg.Add(1)
		go func(colIdx int) {
			defer wg.Done()
			colony := fmt.Sprintf("colony-%d", colIdx)
			for i := 0; i < objectsPerColony; i++ {
				name := fmt.Sprintf("obj-%d", i)
				data := []byte(fmt.Sprintf("data-%d-%d", colIdx, i))
				store.Put(colony, name, bytes.NewReader(data), int64(len(data)))
			}
		}(c)
	}
	wg.Wait()

	for c := 0; c < colonies; c++ {
		colony := fmt.Sprintf("colony-%d", c)
		list, err := store.List(colony)
		assert.NoError(t, err)
		assert.Len(t, list, objectsPerColony)
	}
}

func TestConcurrentDiskUsage(t *testing.T) {
	store := setup(t)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("obj-%d", idx)
			data := make([]byte, 1000)
			store.Put("c", name, bytes.NewReader(data), int64(len(data)))
		}(i)
	}
	wg.Wait()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			usage, err := store.DiskUsage("c")
			assert.NoError(t, err)
			assert.True(t, usage > 0)
		}()
	}
	wg.Wait()
}

func TestConcurrentRemoveAll(t *testing.T) {
	store := setup(t)

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("obj-%d", i)
		assert.NoError(t, store.Put("c", name, bytes.NewReader([]byte("data")), 4))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		store.RemoveAll("c")
	}()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.List("c")
		}()
	}

	wg.Wait()
}
