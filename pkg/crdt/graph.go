package crdt

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/iancoleman/orderedmap"
)

type NodeID string

type VersionedField struct {
	Value interface{} `json:"value"`
	Clock VectorClock `json:"clock"`
	Owner ClientID    `json:"owner"`
}

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

type Edge struct {
	From     NodeID `json:"from"`
	To       NodeID `json:"to"`
	Label    string `json:"label"`
	Position int    `json:"position"`
}

type Graph struct {
	Root  *Node
	Nodes map[NodeID]*Node
}

func NewNodeFromID(id NodeID, isArray bool) *Node {
	return &Node{
		ID:     id,
		Fields: make(map[string]VersionedField),
		Edges:  make([]*Edge, 0),
	}
}

func (g *Graph) GetOrCreateNode(id NodeID, isArray bool) *Node {
	if _, ok := g.Nodes[id]; !ok {
		node := NewNodeFromID(id, isArray)
		g.Nodes[id] = node
	}
	return g.Nodes[id]
}

func (g *Graph) GetNode(id NodeID) (*Node, bool) {
	node, ok := g.Nodes[id]
	if !ok {
		return nil, false
	}
	return node, true
}

func NewGraph() *Graph {
	g := &Graph{
		Root:  NewNodeFromID(NodeID("root"), false),
		Nodes: make(map[NodeID]*Node),
	}
	g.Nodes[g.Root.ID] = g.Root

	return g
}

func generateRandomNodeID(label string) NodeID {
	id := core.GenerateRandomID()
	id = label + "-" + id
	return NodeID(id)
}

func (n *Node) SetField(key string, value interface{}, clientID ClientID, version int) {
	newClock := copyClock(n.Fields[key].Clock)
	newClock[clientID] = version

	existing := n.Fields[key]
	cmp := compareClocks(newClock, existing.Clock)

	if cmp == 1 || (cmp == 0 && clientID < existing.Owner) {
		n.Fields[key] = VersionedField{
			Value: value,
			Clock: newClock,
			Owner: clientID,
		}
	}
}

func (g *Graph) AddEdge(from, to NodeID, label string) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot add edge, node not found: " + string(from))
	}

	edge := &Edge{From: from, To: to, Label: label, Position: -1}
	node.Edges = append(node.Edges, edge)

	return nil
}

func (g *Graph) InsertEdge(from, to NodeID, label string, position int) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot insert edge, node not found: " + string(from))
	}

	if position < 0 {
		return errors.New("Cannot insert edge, position must be non-negative")
	} else if position > len(node.Edges) {
		return errors.New("Cannot insert edge, position out of bounds")
	}

	// Loop through the edges to find the correct position and update all edges after it
	if edges := node.Edges; ok {
		for _, edge := range edges {
			if edge.Position >= position {
				// Update the position of the edge
				edge.Position++
			}
		}
	}

	edge := &Edge{From: from, To: to, Label: label, Position: position}
	node.Edges = append(node.Edges, edge)

	return nil
}

func (g *Graph) ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error) {
	var parent *Node
	if parentID == "" {
		parent = g.Root
	} else {
		parent = g.GetOrCreateNode(parentID, isArray)
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
		node := g.GetOrCreateNode(id, isArray)

		if parent.ID == "" {
			return "", errors.New("parent.ID is empty")
		}

		if isArray {
			g.InsertEdge(parent.ID, id, edgeLabel, idx)
		} else {
			g.AddEdge(parent.ID, id, edgeLabel)
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
		node := g.GetOrCreateNode(id, parent.IsArray)
		node.Litteral = true
		node.LitteralValue = v
		if isArray {
			g.InsertEdge(parent.ID, id, edgeLabel, idx)
		} else {
			g.AddEdge(parent.ID, id, edgeLabel)
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
		return nil, fmt.Errorf("cycle detected at node %s", id)
	}
	visited[id] = true

	node, ok := g.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
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

func (g *Graph) ExportRaw() (map[string]interface{}, error) {
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

func (g *Graph) ExportRawToJSON() ([]byte, error) {
	exported, err := g.ExportRaw()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(exported, "", "  ")
}

func (g *Graph) ImportRaw(data map[string]interface{}) error {
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
		node := NewNodeFromID(id, false)

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

func (g *Graph) ImportRawJSON(data []byte) error {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}
	return g.ImportRaw(rawData)
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

	allEdgesEmptyLabel := true
	for _, e := range g.Root.Edges {
		if e.Label != "" {
			allEdgesEmptyLabel = false
			break
		}
	}

	if allEdgesEmptyLabel {
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
