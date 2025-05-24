package crdt

import (
	"encoding/json"
	"fmt"

	"github.com/iancoleman/orderedmap"
)

type Node struct {
	ID            NodeID                    `json:"id"`
	Fields        map[string]VersionedField `json:"fields"`
	Edges         []*Edge                   `json:"edges"`
	Clock         VectorClock               `json:"clock"`
	Owner         ClientID                  `json:"owner"`
	IsArray       bool                      `json:"isarray"`
	Litteral      bool                      `json:"litteral"`
	LitteralValue interface{}               `json:"litteralValue"`
}

func (n *Node) ExportJSON(crdt *TreeCRDT) ([]byte, error) {
	visited := make(map[NodeID]bool)
	result, err := exportNodeRecursive(n, crdt, visited)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(result, "", "  ")
}

func exportNodeRecursive(node *Node, crdt *TreeCRDT, visited map[NodeID]bool) (interface{}, error) {
	if visited[node.ID] {
		return nil, fmt.Errorf("cycle detected at node %s", node.ID)
	}
	visited[node.ID] = true

	if node.Litteral {
		return node.LitteralValue, nil
	}

	obj := orderedmap.New()

	for k, f := range node.Fields {
		obj.Set(k, f.Value)
	}

	edgeGroups := make(map[string][]*Edge)
	for _, edge := range node.Edges {
		edgeGroups[edge.Label] = append(edgeGroups[edge.Label], edge)
	}

	for field, edges := range edgeGroups {
		isArray := false
		for _, e := range edges {
			if child, ok := crdt.Nodes[e.To]; ok && child.IsArray {
				isArray = true
				break
			}
		}

		if isArray {
			var arrayItems []interface{}
			for _, e := range edges {
				child := crdt.Nodes[e.To]
				exportedChild, err := exportNodeRecursive(child, crdt, visited)
				if err != nil {
					return nil, err
				}
				arrayItems = append(arrayItems, exportedChild)
			}
			obj.Set(field, arrayItems)
		} else {
			child := crdt.Nodes[edges[0].To]
			exportedChild, err := exportNodeRecursive(child, crdt, visited)
			if err != nil {
				return nil, err
			}
			obj.Set(field, exportedChild)
		}
	}

	return obj, nil
}
