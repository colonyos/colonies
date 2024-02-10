package dht

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutGetSimple(t *testing.T) {
	kvs := createKVStore()
	key := "/prefix/key"
	value := "value"

	err := kvs.put(key, value)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got, value)
}

func TestPutGetSimple2(t *testing.T) {
	kvs := createKVStore()
	key := "/prefix"
	value := "value"

	err := kvs.put(key, value)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got, value)
}

func TestPutGetSimple3(t *testing.T) {
	kvs := createKVStore()
	key := "/"
	value := "value"

	err := kvs.put(key, value)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got, value)
}

func TestPutGetSimple4(t *testing.T) {
	kvs := createKVStore()
	key := ""
	value := "value"

	err := kvs.put(key, value)
	assert.NotNil(t, err)
}

func TestPutGetSimple5(t *testing.T) {
	kvs := createKVStore()
	key := "/prefix"
	value := ""

	err := kvs.put(key, value)
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

	err := kvs.put(key, value)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got, value)
}

func TestPutOverwrite(t *testing.T) {
	kvs := createKVStore()
	key := "/prefix/key"
	firstValue := "firstValue"
	secondValue := "secondValue"

	kvs.put(key, firstValue)

	err := kvs.put(key, secondValue)
	assert.Nil(t, err)

	got, err := kvs.get(key)
	assert.Nil(t, err)
	assert.Equal(t, got, secondValue)
}

func TestGetAllValuesWithSimplePrefix(t *testing.T) {
	kvs := createKVStore()
	kvs.put("/prefix/key1", "value1")
	kvs.put("/prefix/key2", "value2")

	expected := []string{"value1", "value2"}

	got, err := kvs.getAllValuesWithPrefix("/prefix")
	assert.Nil(t, err)

	sort.Strings(got)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetAllValuesWithPrefix() = %v, want %v", got, expected)
	}
}

func TestGetAllValuesWithNestedPrefix(t *testing.T) {
	kvs := createKVStore()
	kvs.put("/prefix/key1/nested1", "nestedValue1")
	kvs.put("/prefix/key1/nested2", "nestedValue2")

	expected := []string{"nestedValue1", "nestedValue2"}

	got, err := kvs.getAllValuesWithPrefix("/prefix/key1")
	assert.Nil(t, err)

	sort.Strings(got)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetAllValuesWithPrefix() = %v, want %v", got, expected)
	}
}

func TestGetAllValuesWithNonExistentPrefix(t *testing.T) {
	kvs := createKVStore()

	_, err := kvs.getAllValuesWithPrefix("/nonexistent")
	assert.NotNil(t, err)
}

func TestGetAllValuesWithComplexTree(t *testing.T) {
	kvs := createKVStore()
	kvs.put("/prefix/key1", "value1")
	kvs.put("/prefix/key1/nested1", "nestedValue1")
	kvs.put("/prefix/key1/nested2", "nestedValue2")
	kvs.put("/prefix/key2", "value2")

	expectedPrefix := []string{"value1", "value2", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedPrefix)

	expectedKey1 := []string{"value1", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedKey1)

	gotPrefix, err := kvs.getAllValuesWithPrefix("/prefix")
	assert.Nil(t, err)

	gotKey1, err := kvs.getAllValuesWithPrefix("/prefix/key1")
	assert.Nil(t, err)

	sort.Strings(gotPrefix)
	sort.Strings(gotKey1)

	if !reflect.DeepEqual(gotPrefix, expectedPrefix) {
		t.Errorf("GetAllValuesWithPrefix('/prefix') = %v, want %v", gotPrefix, expectedPrefix)
	}

	if !reflect.DeepEqual(gotKey1, expectedKey1) {
		t.Errorf("GetAllValuesWithPrefix('/prefix/key1') = %v, want %v", gotKey1, expectedKey1)
	}
}

func TestRemoveKey(t *testing.T) {
	kvs := createKVStore()

	kvs.put("/a/b/c", "value1")
	kvs.put("/a/b/d", "value2")
	kvs.put("/a/b", "value3")

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
