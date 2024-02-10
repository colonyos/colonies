package dht

import (
	"fmt"
	"strings"
)

type node struct {
	children map[string]*node
	kv       KV
	isValue  bool
}

func newNode() *node {
	return &node{children: make(map[string]*node)}
}

type kvStore struct {
	root *node
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Sig   string `json:"sig"`
}

func (kv *KV) String() string {
	return fmt.Sprintf("%s:%s", kv.Key, kv.Value)
}

func createKVStore() *kvStore {
	return &kvStore{root: &node{children: make(map[string]*node)}}
}

func (kvStore *kvStore) put(key string, value string, sig string) error {
	if len(key) == 0 || key[0] != '/' {
		return fmt.Errorf("Invalid key, must start with /")
	}

	if len(value) == 0 {
		return fmt.Errorf("Invalid value, cannot be empty")
	}

	parts := strings.Split(key, "/")[1:]

	current := kvStore.root
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
	current.children[lastPart].kv = KV{Key: key, Value: value, Sig: sig}
	current.children[lastPart].isValue = true

	return nil
}

func (kvStore *kvStore) get(key string) (KV, error) {
	parts := strings.Split(key, "/")[1:]

	current := kvStore.root
	for _, part := range parts {
		if _, ok := current.children[part]; !ok {
			return KV{}, fmt.Errorf("Key not found")
		}
		current = current.children[part]
	}

	if current.isValue {
		return current.kv, nil
	} else {
		return KV{}, fmt.Errorf("Key not found")
	}
}

func (kvStore *kvStore) getAllValuesWithPrefix(prefix string) ([]KV, error) {
	parts := strings.Split(prefix, "/")[1:]
	current := kvStore.root

	for _, part := range parts {
		if child, ok := current.children[part]; ok {
			current = child
		} else {
			return nil, fmt.Errorf("Prefix not found")
		}
	}

	var coll []KV
	kvStore.collectValues(current, &coll)
	return coll, nil
}

func (kvStore *kvStore) collectValues(node *node, coll *[]KV) {
	if node.isValue {
		*coll = append(*coll, node.kv)
	}
	for _, child := range node.children {
		kvStore.collectValues(child, coll)
	}
}

func (kvStore *kvStore) removeKey(key string) error {
	if len(key) == 0 || key[0] != '/' {
		return fmt.Errorf("Invalid key, must start with /")
	}

	parts := strings.Split(key, "/")[1:]
	if len(parts) == 0 {
		return fmt.Errorf("Invalid key, cannot be the root")
	}

	current := kvStore.root
	for i, part := range parts {
		if _, ok := current.children[part]; !ok {
			return fmt.Errorf("Key not found")
		}
		if i == len(parts)-1 {
			if len(current.children[part].children) > 0 {
				current.children[part].kv = KV{}
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
