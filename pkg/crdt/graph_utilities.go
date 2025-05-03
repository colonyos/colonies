package crdt

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/iancoleman/orderedmap"
)

func (g *Graph) ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error) {
	version := 1
	var parent *Node
	if parentID == "" {
		parent = g.Root
	} else {
		parent = g.getOrCreateNode(parentID, isArray, clientID, version)
	}

	var data interface{}
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return "", err
	}

	switch v := data.(type) {
	case map[string]interface{}:
		var id NodeID
		if idVal, ok := v["id"]; ok {
			idStr := fmt.Sprintf("%v", idVal)
			id = generateRandomNodeID(idStr)
		} else {
			id = generateRandomNodeID(string(parent.ID))
		}
		node := g.getOrCreateNode(id, isArray, clientID, version)

		if parent.ID == "" {
			return "", errors.New("parent.ID is empty")
		}

		if isArray {
			g.insertEdge(parent.ID, id, edgeLabel, idx, clientID)
		} else {
			g.AddEdge(parent.ID, id, edgeLabel, clientID)
		}

		for key, val := range v {
			switch child := val.(type) {
			case map[string]interface{}:
				childJSON, _ := json.Marshal(child)
				_, _ = g.ImportJSON(childJSON, id, key, idx, false, clientID)
			case []interface{}:
				for i, item := range child {
					if obj, ok := item.(map[string]interface{}); ok {
						childJSON, _ := json.Marshal(obj)
						_, err := g.ImportJSON(childJSON, id, key, i, true, clientID)
						if err != nil {
							return "", err
						}
					}
				}
			default:
				node.SetField(key, child, clientID, 1)
			}
		}
		return id, nil
	case []interface{}:
		for i, item := range v {
			childJSON, _ := json.Marshal(item)
			_, err := g.ImportJSON(childJSON, parent.ID, "", i, true, clientID)
			if err != nil {
				return "", err
			}
		}
		return parent.ID, nil
	case interface{}:
		if parent.ID == "" {
			return "", errors.New("parent.ID is empty")
		}
		strID := fmt.Sprintf("%v", v)
		id := generateRandomNodeID(strID)
		node := g.getOrCreateNode(id, isArray, clientID, version)
		node.Litteral = true
		node.LitteralValue = v
		if isArray {
			g.insertEdge(parent.ID, id, edgeLabel, idx, clientID)
		} else {
			g.AddEdge(parent.ID, id, edgeLabel, clientID)
		}
		return parent.ID, nil
	default:
		return "", fmt.Errorf("unsupported JSON root type: %T", v)
	}
}

func (g *Graph) Print() {
	for id, node := range g.Nodes {
		fmt.Printf("Node %s:\n", id)
		for k, field := range node.Fields {
			fmt.Printf("  %s: %v (by %s, clock=%v)\n", k, field.Value, field.Owner, field.Clock)
		}

		for _, e := range node.Edges {
			fmt.Printf("  → %s (%s[%d])\n", e.To, e.Label, e.Position)
		}
	}
}

func (g *Graph) exportNodeOrdered(id NodeID, visited map[NodeID]bool) (interface{}, error) {
	if visited[id] {
		return nil, fmt.Errorf("Cycle detected at node %s", id)
	}
	visited[id] = true

	node, ok := g.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("Node %s not found", id)
	}

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
			if e.Position >= 0 {
				isArray = true
				break
			}
		}

		if isArray {
			sort.Slice(edges, func(i, j int) bool {
				return edges[i].Position < edges[j].Position
			})
			var arrayItems []interface{}
			for _, e := range edges {
				child, err := g.exportNodeOrdered(e.To, visited)
				if err != nil {
					return nil, err
				}
				arrayItems = append(arrayItems, child)
			}
			obj.Set(field, arrayItems)
		} else {
			child, err := g.exportNodeOrdered(edges[0].To, visited)
			if err != nil {
				return nil, err
			}
			obj.Set(field, child)
		}
	}

	return obj, nil
}

func (g *Graph) exportRaw() (map[string]interface{}, error) {
	nodes := make(map[string]interface{})

	for id, node := range g.Nodes {
		nodeData := map[string]interface{}{
			"id":            string(node.ID),
			"isArray":       node.IsArray,
			"litteral":      node.Litteral,
			"litteralValue": node.LitteralValue,
			"owner":         string(node.Owner),
			"clock":         node.Clock,
			"fields":        map[string]interface{}{},
			"edges":         []map[string]interface{}{},
		}

		for key, field := range node.Fields {
			nodeData["fields"].(map[string]interface{})[key] = map[string]interface{}{
				"value": field.Value,
				"clock": field.Clock,
				"owner": string(field.Owner),
			}
		}

		for _, edge := range node.Edges {
			nodeData["edges"] = append(nodeData["edges"].([]map[string]interface{}), map[string]interface{}{
				"label":    edge.Label,
				"position": edge.Position,
				"to":       string(edge.To),
				"from":     string(edge.From),
			})
		}

		nodes[string(id)] = nodeData
	}

	return map[string]interface{}{
		"root":  string(g.Root.ID),
		"nodes": nodes,
	}, nil
}

func (g *Graph) Save() ([]byte, error) {
	exported, err := g.exportRaw()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(exported, "", "  ")
}

func (g *Graph) importRaw(data map[string]interface{}) error {
	nodesData, ok := data["nodes"].(map[string]interface{})
	if !ok {
		return errors.New("invalid raw data: nodes missing")
	}
	g.Nodes = make(map[NodeID]*Node)

	// First pass: create nodes
	for idStr, nodeRaw := range nodesData {
		nodeMap, ok := nodeRaw.(map[string]interface{})
		if !ok {
			return errors.New("invalid node data")
		}

		id := NodeID(idStr)
		node := newNodeFromID(id, false)

		if isArray, ok := nodeMap["isArray"].(bool); ok {
			node.IsArray = isArray
		}
		if litteral, ok := nodeMap["litteral"].(bool); ok {
			node.Litteral = litteral
		}
		node.LitteralValue = nodeMap["litteralValue"]

		if ownerStr, ok := nodeMap["owner"].(string); ok {
			node.Owner = ClientID(ownerStr)
		}

		if clockMap, ok := nodeMap["clock"].(map[string]interface{}); ok {
			node.Clock = make(VectorClock)
			for cid, v := range clockMap {
				if vInt, ok := v.(float64); ok {
					node.Clock[ClientID(cid)] = int(vInt)
				}
			}
		}

		if fields, ok := nodeMap["fields"].(map[string]interface{}); ok {
			for key, f := range fields {
				if fMap, ok := f.(map[string]interface{}); ok {
					field := VersionedField{
						Value: fMap["value"],
						Clock: make(VectorClock),
						Owner: ClientID(fmt.Sprintf("%v", fMap["owner"])),
					}
					if clockMap, ok := fMap["clock"].(map[string]interface{}); ok {
						for cid, v := range clockMap {
							if vInt, ok := v.(float64); ok {
								field.Clock[ClientID(cid)] = int(vInt)
							}
						}
					}
					node.Fields[key] = field
				}
			}
		}

		g.Nodes[id] = node
	}

	// Second pass: recreate edges
	for idStr, nodeRaw := range nodesData {
		nodeMap, _ := nodeRaw.(map[string]interface{})
		id := NodeID(idStr)
		node := g.Nodes[id]

		if edges, ok := nodeMap["edges"].([]interface{}); ok {
			for _, e := range edges {
				edgeMap, ok := e.(map[string]interface{})
				if !ok {
					continue
				}
				edge := &Edge{}
				if toStr, ok := edgeMap["to"].(string); ok {
					edge.To = NodeID(toStr)
				}
				if fromStr, ok := edgeMap["from"].(string); ok {
					edge.From = NodeID(fromStr)
				}
				if labelStr, ok := edgeMap["label"].(string); ok {
					edge.Label = labelStr
				}
				if pos, ok := edgeMap["position"].(float64); ok {
					edge.Position = int(pos)
				}
				edge.From = id
				node.Edges = append(node.Edges, edge)
			}
		}
	}

	rootID, ok := data["root"].(string)
	if !ok {
		return errors.New("invalid raw data: root missing")
	}
	rootNode, ok := g.Nodes[NodeID(rootID)]
	if !ok {
		return errors.New("root node not found")
	}
	g.Root = rootNode

	return nil
}

func (g *Graph) Load(data []byte) error {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}
	return g.importRaw(rawData)
}

func (g *Graph) ExportToJSON() ([]byte, error) {
	exported, err := g.Export()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(exported, "", "  ")
}

func (g *Graph) Export() (interface{}, error) {
	visited := make(map[NodeID]bool)

	if len(g.Root.Edges) == 0 {
		return nil, fmt.Errorf("root node has no edges")
	}

	if len(g.Root.Edges) == 1 {
		childID := g.Root.Edges[0].To
		return g.exportNodeOrdered(childID, visited)
	}

	isArray := false
	for _, e := range g.Root.Edges {
		if e.Position >= 0 {
			isArray = true
			break
		}
	}

	if isArray {
		// Root points to an array
		sort.Slice(g.Root.Edges, func(i, j int) bool {
			return g.Root.Edges[i].Position < g.Root.Edges[j].Position
		})
		var arrayItems []interface{}
		for _, e := range g.Root.Edges {
			child, err := g.exportNodeOrdered(e.To, visited)
			if err != nil {
				return nil, err
			}
			arrayItems = append(arrayItems, child)
		}
		return arrayItems, nil
	} else {
		// Root points to a map
		result := orderedmap.New()
		for _, e := range g.Root.Edges {
			child, err := g.exportNodeOrdered(e.To, visited)
			if err != nil {
				return nil, err
			}
			result.Set(e.Label, child)
		}
		return result, nil
	}
}

func (g *Graph) Clone() (*Graph, error) {
	safeCopy, err := g.Save()
	if err != nil {
		return nil, err
	}
	newGraph := NewGraph()
	if err := newGraph.Load(safeCopy); err != nil {
		return nil, err
	}
	return newGraph, nil
}

func (g *Graph) Equal(other *Graph) bool {
	if len(g.Nodes) != len(other.Nodes) {
		return false
	}

	for id, node := range g.Nodes {
		otherNode, ok := other.Nodes[id]
		if !ok {
			return false
		}

		if node.Owner != otherNode.Owner || !clocksEqual(node.Clock, otherNode.Clock) {
			return false
		}

		if node.IsArray != otherNode.IsArray || node.Litteral != otherNode.Litteral {
			return false
		}

		if fmt.Sprintf("%v", node.LitteralValue) != fmt.Sprintf("%v", otherNode.LitteralValue) {
			return false
		}

		if len(node.Fields) != len(otherNode.Fields) {
			return false
		}

		for key, field := range node.Fields {
			otherField, ok := otherNode.Fields[key]
			if !ok {
				return false
			}
			if field.Value != otherField.Value || field.Owner != otherField.Owner || !clocksEqual(field.Clock, otherField.Clock) {
				return false
			}
		}

		if len(node.Edges) != len(otherNode.Edges) {
			return false
		}

		for i, edge := range node.Edges {
			otherEdge := otherNode.Edges[i]
			if edge.From != otherEdge.From || edge.To != otherEdge.To || edge.Label != otherEdge.Label || edge.Position != otherEdge.Position {
				return false
			}
		}
	}

	return true
}

func copyFields(original map[string]VersionedField) map[string]VersionedField {
	newFields := make(map[string]VersionedField)
	for k, v := range original {
		newFields[k] = v
	}
	return newFields
}

func edgeExists(edges []*Edge, candidate *Edge) bool {
	for _, e := range edges {
		if e.From == candidate.From && e.To == candidate.To && e.Label == candidate.Label && e.Position == candidate.Position {
			return true
		}
	}
	return false
}

func mergeClocks(a, b VectorClock) VectorClock {
	merged := make(VectorClock)
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		if mv, ok := merged[k]; !ok || v > mv {
			merged[k] = v
		}
	}
	return merged
}

func lowestClientID(a, b ClientID) ClientID {
	if a < b {
		return a
	}
	return b
}

func sortEdgesByNodeID(edges []*Edge) {
	sort.SliceStable(edges, func(i, j int) bool {
		return edges[i].To < edges[j].To
	})
	for i, edge := range edges {
		edge.Position = i
	}
}
