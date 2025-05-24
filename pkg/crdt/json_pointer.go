package crdt

import (
	"fmt"
	"strings"
)

func (c *TreeCRDT) GetNodeByPath(path string) (*Node, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with a slash: %s", path)
	}
	parts := strings.Split(path, "/")[1:]
	if len(parts) == 0 {
		return nil, fmt.Errorf("path is empty")
	}

	node := c.Root

	// Skip to first child if root has a single child
	if len(node.Edges) == 1 {
		childID := node.Edges[0].To
		child, ok := c.Nodes[childID]
		if !ok {
			return nil, fmt.Errorf("child node not found")
		}
		node = child
	}

	for _, part := range parts {
		if node.IsArray {
			index, err := parseIndex(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array index at '%s': %v", part, err)
			}
			sorted := sortEdgesByToStable(node.Edges)
			if index < 0 || index >= len(sorted) {
				return nil, fmt.Errorf("index out of bounds at '%s'", part)
			}
			node = c.Nodes[sorted[index].To]
		} else {
			found := false
			for _, edge := range node.Edges {
				if edge.Label == part {
					node = c.Nodes[edge.To]
					found = true
					break
				}
			}

			if !found {
				// Fallback: check if it's a field in the current node
				_, ok := node.Fields[part]
				if ok {
					return node, nil
				}

				return nil, fmt.Errorf("path not found at '%s'", part)
			}
		}
	}

	return node, nil
}

func (c *TreeCRDT) GetStringValueByPath(path string) (string, bool, error) {
	value, ok, err := c.GetValueByPath(path)
	if err != nil {
		return "", false, err
	}

	if ok {
		if strValue, ok := value.(string); ok {
			return strValue, true, nil
		}
	}

	return "", false, nil
}

func (c *TreeCRDT) GetValueByPath(path string) (interface{}, bool, error) {
	node, err := c.GetNodeByPath(path)
	if err != nil {
		return nil, false, err
	}

	if node.IsArray {
		return nil, false, fmt.Errorf("path points to an array node: %s", path)
	}

	if node.Litteral {
		return nil, false, fmt.Errorf("path points to a literal node: %s", path)
	}

	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	value, ok := node.GetValue(path)

	return value, ok, nil
}

func parseIndex(s string) (int, error) {
	var idx int
	_, err := fmt.Sscanf(s, "%d", &idx)
	return idx, err
}
