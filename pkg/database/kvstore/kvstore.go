package kvstore

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// NodeType represents the type of node
type NodeType int

const (
	NodeTypeMap NodeType = iota
	NodeTypeArray
)

// Node interface
type Node[T any] interface {
	Type() NodeType
	GetChild(key string) (Node[T], bool)
	SetChild(key string, node Node[T]) error
	GetValue(key string) (T, bool)
	SetValue(key string, value T) error
}

// MapNode for storing values of type T
type MapNode[T any] struct {
	children map[string]Node[T]
	values   map[string]T
}

func NewMapNode[T any]() *MapNode[T] {
	return &MapNode[T]{
		children: make(map[string]Node[T]),
		values:   make(map[string]T),
	}
}

func (m *MapNode[T]) Type() NodeType { return NodeTypeMap }

func (m *MapNode[T]) GetChild(key string) (Node[T], bool) {
	child, exists := m.children[key]
	return child, exists
}

func (m *MapNode[T]) SetChild(key string, node Node[T]) error {
	m.children[key] = node
	return nil
}

func (m *MapNode[T]) GetValue(key string) (T, bool) {
	value, exists := m.values[key]
	return value, exists
}

func (m *MapNode[T]) SetValue(key string, value T) error {
	m.values[key] = value
	return nil
}

// ArrayNode for arrays of type T
type ArrayNode[T any] struct {
	items []Node[T]
}

func NewArrayNode[T any]() *ArrayNode[T] {
	return &ArrayNode[T]{
		items: make([]Node[T], 0),
	}
}

func (a *ArrayNode[T]) Type() NodeType { return NodeTypeArray }

func (a *ArrayNode[T]) GetChild(key string) (Node[T], bool) {
	// Try to parse key as numeric index
	index, err := strconv.Atoi(key)
	if err != nil {
		var zero Node[T]
		return zero, false // Not a valid numeric index
	}
	
	if index < 0 || index >= len(a.items) {
		var zero Node[T]
		return zero, false // Index out of bounds
	}
	
	return a.items[index], true
}

func (a *ArrayNode[T]) SetChild(key string, node Node[T]) error {
	// Try to parse key as numeric index
	index, err := strconv.Atoi(key)
	if err != nil {
		return errors.New("array keys must be numeric indices")
	}
	
	// Extend array if necessary
	for len(a.items) <= index {
		a.items = append(a.items, NewMapNode[T]())
	}
	
	a.items[index] = node
	return nil
}

func (a *ArrayNode[T]) GetValue(key string) (T, bool) {
	var zero T
	return zero, false // Arrays don't store direct values
}

func (a *ArrayNode[T]) SetValue(key string, value T) error {
	return errors.New("arrays don't support direct value assignment")
}

func (a *ArrayNode[T]) Append(node Node[T]) {
	a.items = append(a.items, node)
}

func (a *ArrayNode[T]) Get(index int) (Node[T], bool) {
	if index < 0 || index >= len(a.items) {
		var zero Node[T]
		return zero, false
	}
	return a.items[index], true
}

func (a *ArrayNode[T]) Length() int {
	return len(a.items)
}

// KVStore
type KVStore[T any] struct {
	root Node[T]
	mu   sync.RWMutex
}

func NewKVStore[T any]() *KVStore[T] {
	return &KVStore[T]{
		root: NewMapNode[T](),
	}
}

// navigateToParentForRead navigates to parent without creating intermediate nodes (for read operations)
func (kv *KVStore[T]) navigateToParentForRead(path string) (Node[T], string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		var zero Node[T]
		return zero, "", errors.New("invalid path")
	}

	current := kv.root
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		child, exists := current.GetChild(part)
		if !exists {
			var zero Node[T]
			return zero, "", errors.New("path not found")
		}
		current = child
	}

	return current, parts[len(parts)-1], nil
}

// navigateToParent navigates to the parent node of the given path
func (kv *KVStore[T]) navigateToParent(path string) (Node[T], string, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		var zero Node[T]
		return zero, "", errors.New("invalid path")
	}

	current := kv.root
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		child, exists := current.GetChild(part)
		if !exists {
			// Create intermediate map nodes
			child = NewMapNode[T]()
			current.SetChild(part, child)
		}
		current = child
	}

	return current, parts[len(parts)-1], nil
}

// navigateTo navigates to the node at the given path
func (kv *KVStore[T]) navigateTo(path string) (Node[T], error) {
	if path == "" || path == "/" {
		return kv.root, nil
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := kv.root

	for _, part := range parts {
		if part == "" {
			continue
		}

		child, exists := current.GetChild(part)
		if !exists {
			var zero Node[T]
			return zero, errors.New("path not found")
		}
		current = child
	}

	return current, nil
}

// CreateArray creates an array at the specified path
func (kv *KVStore[T]) CreateArray(path string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	parent, key, err := kv.navigateToParent(path)
	if err != nil {
		return err
	}

	arrayNode := NewArrayNode[T]()
	return parent.SetChild(key, arrayNode)
}

// Put stores a value at the specified path
func (kv *KVStore[T]) Put(path string, value T) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	parent, key, err := kv.navigateToParent(path)
	if err != nil {
		return err
	}

	return parent.SetValue(key, value)
}

// Get retrieves a value from the specified path
func (kv *KVStore[T]) Get(path string) (T, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	parent, key, err := kv.navigateToParentForRead(path)
	if err != nil {
		var zero T
		return zero, err
	}

	value, exists := parent.GetValue(key)
	if !exists {
		var zero T
		return zero, errors.New("key not found")
	}

	return value, nil
}

// GetNode retrieves a node at the specified path
func (kv *KVStore[T]) GetNode(path string) (Node[T], error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	return kv.navigateTo(path)
}

// AppendToArray appends a value to an array at the specified path
func (kv *KVStore[T]) AppendToArray(path string, value T) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	node, err := kv.navigateTo(path)
	if err != nil {
		return err
	}

	arrayNode, ok := node.(*ArrayNode[T])
	if !ok {
		return errors.New("node is not an array")
	}

	mapNode := NewMapNode[T]()
	mapNode.SetValue("value", value)
	
	arrayNode.Append(mapNode)
	return nil
}

// AppendMapToArray appends a map of values to an array (like original interface{} version)
func (kv *KVStore[T]) AppendMapToArray(path string, mapData map[string]T) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	node, err := kv.navigateTo(path)
	if err != nil {
		return err
	}

	arrayNode, ok := node.(*ArrayNode[T])
	if !ok {
		return errors.New("node is not an array")
	}

	mapNode := NewMapNode[T]()
	for key, value := range mapData {
		mapNode.SetValue(key, value)
	}

	arrayNode.Append(mapNode)
	return nil
}

// SearchResult
type SearchResult[T any] struct {
	Path  string
	Value T
}

// FindInArray searches for values in an array by JSON field name and value  
func (kv *KVStore[T]) FindInArray(arrayPath string, jsonFieldName string, searchValue interface{}) ([]SearchResult[T], error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	node, err := kv.navigateTo(arrayPath)
	if err != nil {
		return nil, err
	}
	
	arrayNode, ok := node.(*ArrayNode[T])
	if !ok {
		return nil, errors.New("path does not point to an array")
	}
	
	var results []SearchResult[T]
	
	for i := 0; i < arrayNode.Length(); i++ {
		itemNode, exists := arrayNode.Get(i)
		if !exists {
			continue
		}
		
		mapNode, ok := itemNode.(*MapNode[T])
		if !ok {
			continue
		}
		
		// Check if this map contains a value with the field we're looking for
		for key, mapValue := range mapNode.values {
			fieldValue, found := extractFieldValue(mapValue, jsonFieldName)
			if found && reflect.DeepEqual(fieldValue, searchValue) {
				resultPath := arrayPath + "/" + strconv.Itoa(i) + "/" + key
				results = append(results, SearchResult[T]{
					Path:  resultPath,
					Value: mapValue,
				})
			}
		}
	}
	
	return results, nil
}

// FindAllInArray searches for all values in an array that contain a specific JSON field
func (kv *KVStore[T]) FindAllInArray(arrayPath string, jsonFieldName string) ([]SearchResult[T], error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	node, err := kv.navigateTo(arrayPath)
	if err != nil {
		return nil, err
	}
	
	arrayNode, ok := node.(*ArrayNode[T])
	if !ok {
		return nil, errors.New("path does not point to an array")
	}
	
	var results []SearchResult[T]
	
	for i := 0; i < arrayNode.Length(); i++ {
		itemNode, exists := arrayNode.Get(i)
		if !exists {
			continue
		}
		
		mapNode, ok := itemNode.(*MapNode[T])
		if !ok {
			continue
		}
		
		// Check if this map contains a value with the field we're looking for
		for key, mapValue := range mapNode.values {
			if _, found := extractFieldValue(mapValue, jsonFieldName); found {
				resultPath := arrayPath + "/" + strconv.Itoa(i) + "/" + key
				results = append(results, SearchResult[T]{
					Path:  resultPath,
					Value: mapValue,
				})
			}
		}
	}
	
	return results, nil
}

// GetFieldValue gets a field value from a stored value using JSON field name
func (kv *KVStore[T]) GetFieldValue(objectPath string, jsonFieldName string) (interface{}, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	parent, key, err := kv.navigateToParentForRead(objectPath)
	if err != nil {
		return nil, err
	}
	
	obj, exists := parent.GetValue(key)
	if !exists {
		return nil, errors.New("object not found")
	}
	
	fieldValue, found := extractFieldValue(obj, jsonFieldName)
	if !found {
		return nil, errors.New("field not found")
	}
	
	return fieldValue, nil
}

// Delete removes a value or entire subtree at the specified path
func (kv *KVStore[T]) Delete(path string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	parent, key, err := kv.navigateToParentForRead(path)
	if err != nil {
		return err
	}

	// Try to remove as a child node first
	if mapNode, ok := parent.(*MapNode[T]); ok {
		// Remove from children if it exists
		if _, exists := mapNode.children[key]; exists {
			delete(mapNode.children, key)
			return nil
		}
		// Remove from values if it exists
		if _, exists := mapNode.values[key]; exists {
			delete(mapNode.values, key)
			return nil
		}
	}

	return errors.New("path not found")
}

// DeleteRecursive removes an entire subtree at the specified path
func (kv *KVStore[T]) DeleteRecursive(path string) error {
	return kv.Delete(path) // Same as Delete since our Delete already removes subtrees
}

// RemoveFromArray removes an item from an array at the specified index
func (kv *KVStore[T]) RemoveFromArray(arrayPath string, index int) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	node, err := kv.navigateTo(arrayPath)
	if err != nil {
		return err
	}

	arrayNode, ok := node.(*ArrayNode[T])
	if !ok {
		return errors.New("path does not point to an array")
	}

	if index < 0 || index >= len(arrayNode.items) {
		return errors.New("index out of bounds")
	}

	// Remove the item by slicing
	arrayNode.items = append(arrayNode.items[:index], arrayNode.items[index+1:]...)
	return nil
}

// Clear removes all data from the store
func (kv *KVStore[T]) Clear() error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.root = NewMapNode[T]()
	return nil
}

// Exists checks if a path exists in the store (either as a node or value)
func (kv *KVStore[T]) Exists(path string) bool {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	// First try to navigate to the path as a node
	_, err := kv.navigateTo(path)
	if err == nil {
		return true
	}

	// If that fails, check if it exists as a value
	parent, key, err := kv.navigateToParentForRead(path)
	if err != nil {
		return false
	}

	// Check if it exists as a value in the parent
	_, exists := parent.GetValue(key)
	return exists
}

// Count returns the number of items in an array
func (kv *KVStore[T]) Count(arrayPath string) (int, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	node, err := kv.navigateTo(arrayPath)
	if err != nil {
		return 0, err
	}

	if arrayNode, ok := node.(*ArrayNode[T]); ok {
		return arrayNode.Length(), nil
	}

	return 0, errors.New("path does not point to an array")
}

// List returns all child keys/indices at the specified path
func (kv *KVStore[T]) List(path string) ([]string, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	node, err := kv.navigateTo(path)
	if err != nil {
		return nil, err
	}

	var keys []string

	switch n := node.(type) {
	case *MapNode[T]:
		// Add all child keys
		for key := range n.children {
			keys = append(keys, key)
		}
		// Add all value keys
		for key := range n.values {
			keys = append(keys, key)
		}
	case *ArrayNode[T]:
		// Add array indices as strings
		for i := 0; i < len(n.items); i++ {
			keys = append(keys, strconv.Itoa(i))
		}
	}

	return keys, nil
}

// UpdateField updates a specific field in an object (for objects with JSON tags)
func (kv *KVStore[T]) UpdateField(objectPath string, fieldName string, newValue interface{}) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	parent, key, err := kv.navigateToParentForRead(objectPath)
	if err != nil {
		return err
	}

	currentValue, exists := parent.GetValue(key)
	if !exists {
		return errors.New("object not found")
	}

	// Use reflection to update the field
	objValue := reflect.ValueOf(currentValue)
	if objValue.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			return errors.New("object is nil")
		}
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return errors.New("object is not a struct")
	}

	objType := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		jsonFieldName := getJSONFieldName(field)

		if jsonFieldName == fieldName {
			fieldValue := objValue.Field(i)
			if !fieldValue.CanSet() {
				return errors.New("field cannot be set")
			}

			newVal := reflect.ValueOf(newValue)
			if !newVal.Type().AssignableTo(fieldValue.Type()) {
				return errors.New("new value type is not assignable to field")
			}

			fieldValue.Set(newVal)
			return nil
		}
	}

	return errors.New("field not found")
}

// Copy copies data from source path to destination path
func (kv *KVStore[T]) Copy(srcPath, dstPath string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	// Get the source value
	srcParent, srcKey, err := kv.navigateToParentForRead(srcPath)
	if err != nil {
		return err
	}

	var srcValue T
	var srcExists bool

	// Try to get from values first
	srcValue, srcExists = srcParent.GetValue(srcKey)
	if !srcExists {
		// Try to get from children (this is more complex as we need to deep copy nodes)
		return errors.New("copying nodes not yet implemented, only values supported")
	}

	// Set at destination
	dstParent, dstKey, err := kv.navigateToParent(dstPath)
	if err != nil {
		return err
	}

	return dstParent.SetValue(dstKey, srcValue)
}

// Move moves data from source path to destination path
func (kv *KVStore[T]) Move(srcPath, dstPath string) error {
	err := kv.Copy(srcPath, dstPath)
	if err != nil {
		return err
	}

	return kv.Delete(srcPath)
}

// FindRecursive searches recursively from root (or specified path) for objects with JSON field matching value
func (kv *KVStore[T]) FindRecursive(startPath string, jsonFieldName string, searchValue interface{}) ([]SearchResult[T], error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	startNode := kv.root
	if startPath != "" && startPath != "/" {
		node, err := kv.navigateTo(startPath)
		if err != nil {
			return nil, err
		}
		startNode = node
	}
	
	var results []SearchResult[T]
	kv.searchNodeRecursively(startNode, startPath, jsonFieldName, searchValue, &results)
	
	return results, nil
}

// FindAllRecursive searches recursively from root (or specified path) for all objects with JSON field
func (kv *KVStore[T]) FindAllRecursive(startPath string, jsonFieldName string) ([]SearchResult[T], error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	
	startNode := kv.root
	if startPath != "" && startPath != "/" {
		node, err := kv.navigateTo(startPath)
		if err != nil {
			return nil, err
		}
		startNode = node
	}
	
	var results []SearchResult[T]
	kv.searchAllNodeRecursively(startNode, startPath, jsonFieldName, &results)
	
	return results, nil
}

// searchNodeRecursively performs recursive search for matching field values
func (kv *KVStore[T]) searchNodeRecursively(node Node[T], currentPath string, jsonFieldName string, searchValue interface{}, results *[]SearchResult[T]) {
	switch n := node.(type) {
	case *MapNode[T]:
		// Check values in this map node
		for key, value := range n.values {
			fullPath := currentPath
			if fullPath == "" || fullPath == "/" {
				fullPath = key
			} else {
				fullPath = currentPath + "/" + key
			}
			
			fieldValue, found := extractFieldValue(value, jsonFieldName)
			if found && reflect.DeepEqual(fieldValue, searchValue) {
				*results = append(*results, SearchResult[T]{
					Path:  fullPath,
					Value: value,
				})
			}
		}
		
		// Recursively search child nodes
		for key, childNode := range n.children {
			childPath := currentPath
			if childPath == "" || childPath == "/" {
				childPath = key
			} else {
				childPath = currentPath + "/" + key
			}
			kv.searchNodeRecursively(childNode, childPath, jsonFieldName, searchValue, results)
		}
		
	case *ArrayNode[T]:
		// Search through array items
		for i, item := range n.items {
			itemPath := currentPath
			if itemPath == "" || itemPath == "/" {
				itemPath = strconv.Itoa(i)
			} else {
				itemPath = currentPath + "/" + strconv.Itoa(i)
			}
			kv.searchNodeRecursively(item, itemPath, jsonFieldName, searchValue, results)
		}
	}
}

// searchAllNodeRecursively performs recursive search for all objects with field
func (kv *KVStore[T]) searchAllNodeRecursively(node Node[T], currentPath string, jsonFieldName string, results *[]SearchResult[T]) {
	switch n := node.(type) {
	case *MapNode[T]:
		// Check values in this map node
		for key, value := range n.values {
			fullPath := currentPath
			if fullPath == "" || fullPath == "/" {
				fullPath = key
			} else {
				fullPath = currentPath + "/" + key
			}
			
			if _, found := extractFieldValue(value, jsonFieldName); found {
				*results = append(*results, SearchResult[T]{
					Path:  fullPath,
					Value: value,
				})
			}
		}
		
		// Recursively search child nodes
		for key, childNode := range n.children {
			childPath := currentPath
			if childPath == "" || childPath == "/" {
				childPath = key
			} else {
				childPath = currentPath + "/" + key
			}
			kv.searchAllNodeRecursively(childNode, childPath, jsonFieldName, results)
		}
		
	case *ArrayNode[T]:
		// Search through array items
		for i, item := range n.items {
			itemPath := currentPath
			if itemPath == "" || itemPath == "/" {
				itemPath = strconv.Itoa(i)
			} else {
				itemPath = currentPath + "/" + strconv.Itoa(i)
			}
			kv.searchAllNodeRecursively(item, itemPath, jsonFieldName, results)
		}
	}
}

// getJSONFieldName extracts the JSON tag name from a struct field
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name // Fall back to field name if no JSON tag
	}
	
	// Handle cases like `json:"name,omitempty"`
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// extractFieldValue gets a field value from an object using JSON tag name
func extractFieldValue(obj interface{}, jsonFieldName string) (interface{}, bool) {
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)
	
	// Handle pointers
	if objValue.Kind() == reflect.Ptr {
		if objValue.IsNil() {
			return nil, false
		}
		objValue = objValue.Elem()
		objType = objType.Elem()
	}
	
	if objValue.Kind() != reflect.Struct {
		return nil, false
	}
	
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		fieldJSONName := getJSONFieldName(field)
		
		if fieldJSONName == jsonFieldName {
			fieldValue := objValue.Field(i)
			if fieldValue.CanInterface() {
				return fieldValue.Interface(), true
			}
		}
	}
	
	return nil, false
}

// MixedKVStore wraps KVStore[interface{}] for original mixed-type functionality
type MixedKVStore struct {
	*KVStore[interface{}]
}

// NewMixedKVStore creates a new mixed-type KVStore using interface{} (original functionality)
func NewMixedKVStore() *MixedKVStore {
	return &MixedKVStore{
		KVStore: NewKVStore[interface{}](),
	}
}

// AppendToArrayMixed appends mixed-type map data to an array (restores original functionality)
func (kv *MixedKVStore) AppendToArrayMixed(path string, mapData map[string]interface{}) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	node, err := kv.navigateTo(path)
	if err != nil {
		return err
	}

	arrayNode, ok := node.(*ArrayNode[interface{}])
	if !ok {
		return errors.New("node is not an array")
	}

	mapNode := NewMapNode[interface{}]()
	for key, value := range mapData {
		// If value is a nested map, create a sub-node structure
		if nestedMap, isMap := value.(map[string]interface{}); isMap {
			subMapNode := NewMapNode[interface{}]()
			for subKey, subValue := range nestedMap {
				subMapNode.SetValue(subKey, subValue)
			}
			mapNode.SetChild(key, subMapNode)
		} else {
			mapNode.SetValue(key, value)
		}
	}

	arrayNode.Append(mapNode)
	return nil
}