package crdt

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/iancoleman/orderedmap"
	log "github.com/sirupsen/logrus"
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
			// ) InsertEdge(from, to NodeID, label string, leftOf NodeID, clientID ClientID)
			g.AppendEdge(parent.ID, id, edgeLabel, clientID)
			//	g.insertEdge(parent.ID, id, edgeLabel, idx, clientID)
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
			g.AppendEdge(parent.ID, id, edgeLabel, clientID)
			//g.insertEdge(parent.ID, id, edgeLabel, idx, clientID)
		} else {
			g.AddEdge(parent.ID, id, edgeLabel, clientID)
		}
		return parent.ID, nil
	default:
		return "", fmt.Errorf("Unsupported JSON root type: %T", v)
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
			node := g.Nodes[e.To]
			if node != nil {
				if node.IsArray {
					isArray = true
					break
				}
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
				"label":        edge.Label,
				"lseqposition": edge.LSEQPosition,
				"to":           string(edge.To),
				"from":         string(edge.From),
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
				if lseqPos, ok := edgeMap["lseqposition"].([]interface{}); ok {
					for _, v := range lseqPos {
						if intVal, ok := v.(float64); ok {
							edge.LSEQPosition = append(edge.LSEQPosition, int(intVal))
						}
					}
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
		return nil, fmt.Errorf("Root node has no edges")
	}

	if len(g.Root.Edges) == 1 {
		childID := g.Root.Edges[0].To
		return g.exportNodeOrdered(childID, visited)
	}

	isArray := false
	for _, e := range g.Root.Edges {
		node := g.Nodes[e.To]
		if node.IsArray {
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
		log.WithFields(log.Fields{"Nodes1": len(g.Nodes), "Nodes2": len(other.Nodes)}).Warning("Node counts not equal")
		return false
	}

	for id, node := range g.Nodes {
		otherNode, ok := other.Nodes[id]
		if !ok || !nodesSemanticallyEqual(node, otherNode) {
			log.WithFields(log.Fields{"NodeID": id, "Node1": node, "Node2": otherNode}).Warning("Nodes not equal")
			return false
		}
	}

	return true
}

func nodesSemanticallyEqual(n1, n2 *Node) bool {
	if n1.IsArray != n2.IsArray || n1.Litteral != n2.Litteral {
		log.WithFields(log.Fields{"IsArray1": n1.IsArray, "IsArray2": n2.IsArray, "Litteral1": n1.Litteral, "Litteral2": n2.Litteral}).Warning("Node types not equal")
		return false
	}

	switch v1 := n1.LitteralValue.(type) {
	case float64:
		switch v2 := n2.LitteralValue.(type) {
		case float64:
			if v1 != v2 {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (float64)")
				return false
			}
		case int:
			if v1 != float64(v2) {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (float64 vs int)")
				return false
			}
		default:
			log.Warning("Type mismatch: float64 vs unsupported type")
			return false
		}

	case int:
		switch v2 := n2.LitteralValue.(type) {
		case int:
			if v1 != v2 {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (int)")
				return false
			}
		case float64:
			if float64(v1) != v2 {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (int vs float64)")
				return false
			}
		default:
			log.Warning("Type mismatch: int vs unsupported type")
			return false
		}

	case string:
		if v2, ok := n2.LitteralValue.(string); ok {
			if v1 != v2 {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (string)")
				return false
			}
		} else {
			log.Warning("Type mismatch: string vs non-string")
			return false
		}

	case bool:
		if v2, ok := n2.LitteralValue.(bool); ok {
			if v1 != v2 {
				log.WithFields(log.Fields{"LitteralValue1": v1, "LitteralValue2": v2}).Warning("Litteral values not equal (bool)")
				return false
			}
		} else {
			log.Warning("Type mismatch: bool vs non-bool")
			return false
		}

	case nil:
		if n2.LitteralValue != nil {
			log.Warning("Litteral values not equal: nil vs non-nil")
			return false
		}

	default:
		// Final fallback: reflect (last resort for complex or unknown types)
		if !reflect.DeepEqual(n1.LitteralValue, n2.LitteralValue) {
			log.WithFields(log.Fields{
				"LitteralValue1": n1.LitteralValue,
				"LitteralValue2": n2.LitteralValue,
			}).Warning("Litteral values not equal (fallback reflect)")
			return false
		}
	}

	if len(n1.Fields) != len(n2.Fields) {
		log.WithFields(log.Fields{"Fields1": len(n1.Fields), "Fields2": len(n2.Fields)}).Warning("Field counts not equal")
		return false
	}
	for key, f1 := range n1.Fields {
		f2, ok := n2.Fields[key]
		if !ok || !reflect.DeepEqual(f1.Value, f2.Value) {
			log.WithFields(log.Fields{"Key": key, "Value1": f1.Value, "Value2": f2.Value}).Warning("Field values not equal")
			return false
		}
	}

	if len(n1.Edges) != len(n2.Edges) {
		log.WithFields(log.Fields{"Edges1": len(n1.Edges), "Edges2": len(n2.Edges)}).Warning("Edge counts not equal")
		return false
	}

	// NOTE: Each graph will be different due to random nature of the LSEQ algorithm
	if n1.IsArray {
		// Compare edges by LSEQ order
		sorted1 := make([]*Edge, len(n1.Edges))
		sorted2 := make([]*Edge, len(n2.Edges))
		copy(sorted1, n1.Edges)
		copy(sorted2, n2.Edges)
		sortEdgesByLSEQ(sorted1)
		sortEdgesByLSEQ(sorted2)

		for i := range sorted1 {
			if sorted1[i].To != sorted2[i].To || sorted1[i].Label != sorted2[i].Label {
				log.WithFields(log.Fields{"Edge1": sorted1[i], "Edge2": sorted2[i]}).Warning("Edges not equal")
				return false
			}
		}
	} else {
		// Compare edges as unordered field entries
		labelMap1 := map[string]NodeID{}
		labelMap2 := map[string]NodeID{}
		for _, e := range n1.Edges {
			labelMap1[e.Label] = e.To
		}
		for _, e := range n2.Edges {
			labelMap2[e.Label] = e.To
		}
		if !reflect.DeepEqual(labelMap1, labelMap2) {
			log.WithFields(log.Fields{"LabelMap1": labelMap1, "LabelMap2": labelMap2}).Warning("Edge labels not equal")
			return false
		}
	}

	return true
}

func sortEdgesByToStable(edges []*Edge) []*Edge {
	copied := make([]*Edge, len(edges))
	copy(copied, edges)
	sort.SliceStable(copied, func(i, j int) bool {
		// Deterministic sort using LSEQPosition
		p1 := copied[i].LSEQPosition
		p2 := copied[j].LSEQPosition
		for k := 0; k < len(p1) && k < len(p2); k++ {
			if p1[k] != p2[k] {
				return p1[k] < p2[k]
			}
		}
		if len(p1) != len(p2) {
			return len(p1) < len(p2)
		}
		// Fallback to .To string as tie-breaker
		return copied[i].To < copied[j].To
	})
	return copied
}

//
// func (g *Graph) Equal(other *Graph) bool {
// 	// Compare node sets
// 	if len(g.Nodes) != len(other.Nodes) {
// 		return false
// 	}
//
// 	for id, node := range g.Nodes {
// 		otherNode, ok := other.Nodes[id]
// 		if !ok || !nodesEqual(node, otherNode) {
// 			return false
// 		}
// 	}
//
// 	// Check for extra nodes in 'other'
// 	for id := range other.Nodes {
// 		if _, ok := g.Nodes[id]; !ok {
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// func nodesEqual(n1, n2 *Node) bool {
// 	if n1.Owner != n2.Owner || !clocksEqual(n1.Clock, n2.Clock) {
// 		return false
// 	}
// 	if n1.IsArray != n2.IsArray || n1.Litteral != n2.Litteral {
// 		return false
// 	}
// 	if !reflect.DeepEqual(n1.LitteralValue, n2.LitteralValue) {
// 		return false
// 	}
//
// 	if len(n1.Fields) != len(n2.Fields) {
// 		return false
// 	}
// 	for key, f1 := range n1.Fields {
// 		f2, ok := n2.Fields[key]
// 		if !ok || !reflect.DeepEqual(f1.Value, f2.Value) || f1.Owner != f2.Owner || !clocksEqual(f1.Clock, f2.Clock) {
// 			return false
// 		}
// 	}
//
// 	if len(n1.Edges) != len(n2.Edges) {
// 		return false
// 	}
// 	for i := range n1.Edges {
// 		e1 := n1.Edges[i]
// 		e2 := n2.Edges[i]
// 		if e1.From != e2.From || e1.To != e2.To || e1.Label != e2.Label {
// 			return false
// 		}
// 		if !reflect.DeepEqual(e1.LSEQPosition, e2.LSEQPosition) {
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// func (g *Graph) Equal(other *Graph) bool {
// 	if len(g.Nodes) != len(other.Nodes) {
// 		return false
// 	}
//
// 	for id, node := range g.Nodes {
// 		otherNode, ok := other.Nodes[id]
// 		if !ok {
// 			return false
// 		}
//
// 		if node.Owner != otherNode.Owner || !clocksEqual(node.Clock, otherNode.Clock) {
// 			return false
// 		}
//
// 		if node.IsArray != otherNode.IsArray || node.Litteral != otherNode.Litteral {
// 			return false
// 		}
//
// 		if fmt.Sprintf("%v", node.LitteralValue) != fmt.Sprintf("%v", otherNode.LitteralValue) {
// 			return false
// 		}
//
// 		if len(node.Fields) != len(otherNode.Fields) {
// 			return false
// 		}
//
// 		for key, field := range node.Fields {
// 			otherField, ok := otherNode.Fields[key]
// 			if !ok {
// 				return false
// 			}
// 			if field.Value != otherField.Value || field.Owner != otherField.Owner || !clocksEqual(field.Clock, otherField.Clock) {
// 				return false
// 			}
// 		}
//
// 		if len(node.Edges) != len(otherNode.Edges) {
// 			return false
// 		}
//
// 		for i, edge := range node.Edges {
// 			otherEdge := otherNode.Edges[i]
// 			if edge.From != otherEdge.From || edge.To != otherEdge.To || edge.Label != otherEdge.Label || edge.Position != otherEdge.Position {
// 				return false
// 			}
// 		}
// 	}
//
// 	return true
// }

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
