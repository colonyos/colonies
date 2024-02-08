package dht

import (
	"fmt"
	"strings"
)

type Node struct {
	Children map[string]*Node
	Value    string
	IsValue  bool
}

func NewNode() *Node {
	return &Node{Children: make(map[string]*Node)}
}

type KVStore struct {
	root *Node
}

func NewKVStore() *KVStore {
	return &KVStore{root: &Node{Children: make(map[string]*Node)}}
}

func (kv *KVStore) Put(key string, value string) error {
	parts := strings.Split(key, "/")[1:]

	current := kv.root
	for _, part := range parts[:len(parts)-1] {
		if _, ok := current.Children[part]; !ok {
			current.Children[part] = NewNode()
		}
		current = current.Children[part]
	}

	lastPart := parts[len(parts)-1]
	if current.Children[lastPart] == nil {
		current.Children[lastPart] = NewNode()
	}
	current.Children[lastPart].Value = value
	current.Children[lastPart].IsValue = true

	return nil
}

func (kv *KVStore) Get(key string) (string, error) {
	parts := strings.Split(key, "/")[1:]

	current := kv.root
	for _, part := range parts {
		if _, ok := current.Children[part]; !ok {
			return "", fmt.Errorf("key not found")
		}
		current = current.Children[part]
	}

	if current.IsValue {
		return current.Value, nil
	} else {
		return "", fmt.Errorf("key not found")
	}
}

func (kv *KVStore) GetAllValuesWithPrefix(prefix string) ([]string, error) {
	parts := strings.Split(prefix, "/")[1:]
	current := kv.root

	for _, part := range parts {
		if child, ok := current.Children[part]; ok {
			current = child
		} else {
			return nil, fmt.Errorf("prefix not found")
		}
	}

	var values []string
	kv.collectValues(current, &values)
	return values, nil
}

func (kv *KVStore) collectValues(node *Node, values *[]string) {
	if node.IsValue {
		*values = append(*values, node.Value)
	}
	for _, child := range node.Children {
		kv.collectValues(child, values)
	}
}
