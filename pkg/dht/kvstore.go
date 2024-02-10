package dht

import (
	"fmt"
	"strings"
)

type node struct {
	children map[string]*node
	value    string
	isValue  bool
}

func newNode() *node {
	return &node{children: make(map[string]*node)}
}

type kvStore struct {
	root *node
}

func createKVStore() *kvStore {
	return &kvStore{root: &node{children: make(map[string]*node)}}
}

func (kvs *kvStore) put(key string, value string) error {
	if len(key) == 0 || key[0] != '/' {
		return fmt.Errorf("Invalid key, must start with /")
	}

	if len(value) == 0 {
		return fmt.Errorf("Invalid value, cannot be empty")
	}

	parts := strings.Split(key, "/")[1:]

	current := kvs.root
	for _, part := range parts[:len(parts)-1] {
		if _, ok := current.children[part]; !ok {
			current.children[part] = newNode()
		}
		current = current.children[part]
	}

	lastPart := parts[len(parts)-1]
	if current.children[lastPart] == nil {
		current.children[lastPart] = newNode()
	}
	current.children[lastPart].value = value
	current.children[lastPart].isValue = true

	return nil
}

func (kvs *kvStore) get(key string) (string, error) {
	parts := strings.Split(key, "/")[1:]

	current := kvs.root
	for _, part := range parts {
		if _, ok := current.children[part]; !ok {
			return "", fmt.Errorf("Key not found")
		}
		current = current.children[part]
	}

	if current.isValue {
		return current.value, nil
	} else {
		return "", fmt.Errorf("Key not found")
	}
}

func (kvs *kvStore) getAllValuesWithPrefix(prefix string) ([]string, error) {
	parts := strings.Split(prefix, "/")[1:]
	current := kvs.root

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return nil, fmt.Errorf("Prefix not found")
		}
	}

	var values []string
	kvs.collectValues(current, &values)
	return values, nil
}

func (kvs *kvStore) collectValues(node *node, values *[]string) {
	if node.isValue {
		*values = append(*values, node.value)
	}
	for _, child := range node.children {
		kvs.collectValues(child, values)
	}
}

func (kvs *kvStore) removeKey(key string) error {
	if len(key) == 0 || key[0] != '/' {
		return fmt.Errorf("Invalid key, must start with /")
	}

	parts := strings.Split(key, "/")[1:]
	if len(parts) == 0 {
		return fmt.Errorf("Invalid key, cannot be the root")
	}

	current := kvs.root
	for i, part := range parts {
		if _, ok := current.children[part]; !ok {
			return fmt.Errorf("Key not found")
		}
		if i == len(parts)-1 {
			if len(current.children[part].children) > 0 {
				current.children[part].value = ""
				current.children[part].isValue = false
			} else {
				delete(current.children, part)
			}
		} else {
			current = current.children[part]
		}
	}

	return nil
}

func (kvs *kvStore) cleanupParents(node *node, parts []string) {
	if len(parts) == 0 || node == nil {
		return
	}

	parent := node
	for _, part := range parts[:len(parts)-1] {
		parent = parent.children[part]
	}

	lastPart := parts[len(parts)-1]
	child, ok := parent.children[lastPart]
	if !ok {
		return
	}

	if len(child.children) == 0 && !child.isValue {
		delete(parent.children, lastPart)
		kvs.cleanupParents(kvs.root, parts[:len(parts)-1])
	}
}
