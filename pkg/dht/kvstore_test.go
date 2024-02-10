package dht

import (
	"sort"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestPutGetSimple(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix/key"
	value := "value"
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got.Value, value)
	assert.Equal(t, got.Sig, sig)
}

func TestPutGetSimple2(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix"
	value := "value"
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got.Value, value)
	assert.Equal(t, got.Sig, sig)
}

func TestPutGetSimple3(t *testing.T) {
	kvs := createKVStore()

	key := "/"
	value := "value"
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got.Value, value)
	assert.Equal(t, got.Sig, sig)
}

func TestPutGetSimple4(t *testing.T) {
	kvs := createKVStore()

	key := ""
	value := "value"
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.NotNil(t, err)
}

func TestPutGetSimple5(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix"
	value := ""
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.NotNil(t, err)
}

func TestGetNonExistent(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix/nonexistent"

	_, err := kvs.get(key)
	assert.NotNil(t, err)
}

func TestPutGetNested(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix/path1/path2/key"
	value := "deepValue"
	sig := core.GenerateRandomID()

	err := kvs.put(key, value, sig)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got.Value, value)
	assert.Equal(t, got.Sig, sig)
}

func TestPutOverwrite(t *testing.T) {
	kvs := createKVStore()

	key := "/prefix/key"
	firstValue := "firstValue"
	secondValue := "secondValue"
	sig := core.GenerateRandomID()

	err := kvs.put(key, firstValue, sig)
	assert.Nil(t, err)

	err = kvs.put(key, secondValue, sig)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got.Value, secondValue)
	assert.Equal(t, got.Sig, sig)
}

func TestGetAllValuesWithSimplePrefix(t *testing.T) {
	kvs := createKVStore()

	sig := core.GenerateRandomID()

	kvs.put("/prefix/key1", "value1", sig)
	kvs.put("/prefix/key2", "value2", sig)

	got, err := kvs.getAllValuesWithPrefix("/prefix")
	assert.Nil(t, err)
	assert.Equal(t, len(got), 2)

	count := 0
	for _, kv := range got {
		if kv.Value == "value1" || kv.Value == "value2" {
			count++
		}
	}
	assert.Equal(t, count, 2)
}

func TestGetAllValuesWithNestedPrefix(t *testing.T) {
	kvs := createKVStore()

	sig := core.GenerateRandomID()
	kvs.put("/prefix/key1/nested1", "nestedValue1", sig)
	kvs.put("/prefix/key1/nested2", "nestedValue2", sig)

	got, err := kvs.getAllValuesWithPrefix("/prefix/key1")
	assert.Nil(t, err)

	count := 0
	for _, kv := range got {
		if kv.Value == "nestedValue1" || kv.Value == "nestedValue2" {
			count++
		}

	}
	assert.Equal(t, count, 2)
}

func TestGetAllValuesWithNonExistentPrefix(t *testing.T) {
	kvs := createKVStore()

	_, err := kvs.getAllValuesWithPrefix("/nonexistent")
	assert.NotNil(t, err)
}

func TestGetAllValuesWithComplexTree(t *testing.T) {
	kvs := createKVStore()

	sig := core.GenerateRandomID()

	kvs.put("/prefix/key1", "value1", sig)
	kvs.put("/prefix/key1/nested1", "nestedValue1", sig)
	kvs.put("/prefix/key1/nested2", "nestedValue2", sig)
	kvs.put("/prefix/key2", "value2", sig)

	expectedPrefix := []string{"value1", "value2", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedPrefix)

	expectedKey1 := []string{"value1", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedKey1)

	gotPrefix, err := kvs.getAllValuesWithPrefix("/prefix")
	assert.Nil(t, err)

	gotKey1, err := kvs.getAllValuesWithPrefix("/prefix/key1")
	assert.Nil(t, err)

	count := 0
	for _, kv := range gotPrefix {
		if kv.Value == "value1" || kv.Value == "value2" || kv.Value == "nestedValue1" || kv.Value == "nestedValue2" {
			count++
		}
	}
	assert.Equal(t, count, 4)

	count = 0
	for _, kv := range gotKey1 {
		if kv.Value == "value1" || kv.Value == "nestedValue1" || kv.Value == "nestedValue2" {
			count++
		}
	}
}

func TestRemoveKey(t *testing.T) {
	kvs := createKVStore()

	sig := core.GenerateRandomID()

	err := kvs.put("/a/b/c", "value1", sig)
	assert.Nil(t, err)
	err = kvs.put("/a/b/d", "value2", sig)
	assert.Nil(t, err)
	err = kvs.put("/a/b", "value3", sig)
	assert.Nil(t, err)

	if err := kvs.removeKey("/a/b/c"); err != nil {
		t.Fatalf("Failed to remove key /a/b/c: %v", err)
	}
	if _, err := kvs.get("/a/b/c"); err == nil {
		t.Errorf("Expected /a/b/c to be removed, but it was found")
	}

	if err := kvs.removeKey("/a/b"); err != nil {
		t.Fatalf("Failed to remove key /a/b: %v", err)
	}
	if _, err := kvs.get("/a/b"); err == nil {
		t.Errorf("Expected /a/b to be removed, but it was found")
	}
	if _, err := kvs.get("/a/b/d"); err != nil {
		t.Errorf("Expected /a/b/d to exist, but it was not found")
	}

	if err := kvs.removeKey("/non/existing"); err == nil {
		t.Errorf("Expected error when removing non-existing key, but got nil")
	}

	if err := kvs.removeKey("/"); err == nil {
		t.Errorf("Expected error when attempting to remove the root, but got nil")
	}
}
