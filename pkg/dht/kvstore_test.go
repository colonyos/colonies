package dht

import (
	"reflect"
	"sort"
	"testing"
)

func TestPutGetSimple(t *testing.T) {
	kv := NewKVStore()
	key := "/prefix/key"
	value := "value"

	err := kv.Put(key, value)
	if err != nil {
		t.Fatalf("Put() error = %v, wantErr %v", err, false)
	}

	got, err := kv.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}
}

func TestGetNonExistent(t *testing.T) {
	kv := NewKVStore()
	key := "/prefix/nonexistent"

	_, err := kv.Get(key)
	if err == nil {
		t.Errorf("Expected an error when getting a nonexistent key, but got nil")
	}
}

func TestPutGetNested(t *testing.T) {
	kv := NewKVStore()
	key := "/prefix/path1/path2/key"
	value := "deepValue"

	err := kv.Put(key, value)
	if err != nil {
		t.Fatalf("Put() error = %v, wantErr %v", err, false)
	}

	got, err := kv.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}
}

func TestPutOverwrite(t *testing.T) {
	kv := NewKVStore()
	key := "/prefix/key"
	firstValue := "firstValue"
	secondValue := "secondValue"

	kv.Put(key, firstValue)

	err := kv.Put(key, secondValue)
	if err != nil {
		t.Fatalf("Put() error = %v, wantErr %v", err, false)
	}

	got, err := kv.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != secondValue {
		t.Errorf("Get() = %v, want %v", got, secondValue)
	}
}

func TestGetAllValuesWithSimplePrefix(t *testing.T) {
	kv := NewKVStore()
	kv.Put("/prefix/key1", "value1")
	kv.Put("/prefix/key2", "value2")

	expected := []string{"value1", "value2"}

	got, err := kv.GetAllValuesWithPrefix("/prefix")
	if err != nil {
		t.Fatalf("GetAllValuesWithPrefix() error = %v", err)
	}

	sort.Strings(got)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetAllValuesWithPrefix() = %v, want %v", got, expected)
	}
}

func TestGetAllValuesWithNestedPrefix(t *testing.T) {
	kv := NewKVStore()
	kv.Put("/prefix/key1/nested1", "nestedValue1")
	kv.Put("/prefix/key1/nested2", "nestedValue2")

	expected := []string{"nestedValue1", "nestedValue2"}

	got, err := kv.GetAllValuesWithPrefix("/prefix/key1")
	if err != nil {
		t.Fatalf("GetAllValuesWithPrefix() error = %v", err)
	}

	sort.Strings(got)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetAllValuesWithPrefix() = %v, want %v", got, expected)
	}
}

func TestGetAllValuesWithNonExistentPrefix(t *testing.T) {
	kv := NewKVStore()

	_, err := kv.GetAllValuesWithPrefix("/nonexistent")
	if err == nil {
		t.Errorf("GetAllValuesWithPrefix() expected an error, but got nil")
	}
}

func TestGetAllValuesWithComplexTree(t *testing.T) {
	kv := NewKVStore()
	kv.Put("/prefix/key1", "value1")
	kv.Put("/prefix/key1/nested1", "nestedValue1")
	kv.Put("/prefix/key1/nested2", "nestedValue2")
	kv.Put("/prefix/key2", "value2")

	expectedPrefix := []string{"value1", "value2", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedPrefix)

	expectedKey1 := []string{"value1", "nestedValue1", "nestedValue2"}
	sort.Strings(expectedKey1)

	gotPrefix, err := kv.GetAllValuesWithPrefix("/prefix")
	if err != nil {
		t.Fatalf("GetAllValuesWithPrefix('/prefix') error = %v", err)
	}

	gotKey1, err := kv.GetAllValuesWithPrefix("/prefix/key1")
	if err != nil {
		t.Fatalf("GetAllValuesWithPrefix('/prefix/key1') error = %v", err)
	}

	sort.Strings(gotPrefix)
	sort.Strings(gotKey1)

	if !reflect.DeepEqual(gotPrefix, expectedPrefix) {
		t.Errorf("GetAllValuesWithPrefix('/prefix') = %v, want %v", gotPrefix, expectedPrefix)
	}

	if !reflect.DeepEqual(gotKey1, expectedKey1) {
		t.Errorf("GetAllValuesWithPrefix('/prefix/key1') = %v, want %v", gotKey1, expectedKey1)
	}
}
