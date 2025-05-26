package crdt

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *TreeCRDT) GetNodeByPath(path string) (*NodeCRDT, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with a slash: %s", path)
	}

	parts := strings.Split(path, "/")[1:]
	if len(parts) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	node := c.Root

	// Automatically descend into single child if root is wrapper
	for len(node.Edges) == 1 && node == c.Root {
		childID := node.Edges[0].To
		child, exists := c.Nodes[childID]
		if !exists {
			return nil, fmt.Errorf("invalid CRDT: root child %s not found", childID)
		}
		node = child
	}

	for _, part := range parts {
		if node.IsArray {
			index, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array index at '%s': %v", part, err)
			}

			edges := node.Edges
			if index < 0 || index >= len(edges) {
				return nil, fmt.Errorf("array index out of bounds at '%s'", part)
			}

			childID := edges[index].To
			child, exists := c.Nodes[childID]
			if !exists {
				return nil, fmt.Errorf("node %s at array index not found", childID)
			}
			node = child

		} else {
			found := false
			for _, edge := range node.Edges {
				if edge.Label == part {
					child, exists := c.Nodes[edge.To]
					if !exists {
						return nil, fmt.Errorf("node %s for key '%s' not found", edge.To, part)
					}
					node = child
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("path not found at '%s'", part)
			}
		}
	}

	return node, nil
}

func (c *TreeCRDT) GetValueByPath(path string) (interface{}, error) {
	node, err := c.GetNodeByPath(path)
	if err != nil {
		return nil, err
	}
	if !node.IsLiteral {
		return nil, fmt.Errorf("node at path '%s' is not a literal", path)
	}
	if node.LiteralValue == nil {
		return nil, fmt.Errorf("node at path '%s' has no value", path)
	}
	return node.LiteralValue, nil
}

func (c *TreeCRDT) GetStringValueByPath(path string) (string, error) {
	value, err := c.GetValueByPath(path)
	if err != nil {
		return "", err
	}
	if strValue, ok := value.(string); ok {
		return strValue, nil
	}
	return "", fmt.Errorf("value at path '%s' is not a string", path)
}
