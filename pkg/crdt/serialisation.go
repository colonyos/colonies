package crdt

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/iancoleman/orderedmap"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

func (c *TreeCRDT) ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error) {
	version := 1
	var parent *NodeCRDT
	if parentID == "" {
		parent = c.Root
	} else {
		parent = c.getOrCreateNode(parentID, isArray, clientID, version)
	}

	var data interface{}
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return "", err
	}

	return c.importRecursive(data, parent, edgeLabel, idx, isArray, clientID)
}

func (c *TreeCRDT) importRecursive(v interface{}, parent *NodeCRDT, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error) {
	version := 1

	switch val := v.(type) {

	case map[string]interface{}:
		// Map node
		mapNodeID := generateRandomNodeID("map")
		mapNode := c.getOrCreateNode(mapNodeID, false, clientID, version)
		mapNode.IsMap = true

		if parent != nil {
			if isArray {
				err := c.AppendEdge(parent.ID, mapNodeID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			} else {
				err := c.AddEdge(parent.ID, mapNodeID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			}
		}

		for key, child := range val {
			_, err := c.importRecursive(child, mapNode, key, -1, false, clientID)
			if err != nil {
				return "", err
			}
		}

		return mapNodeID, nil

	case []interface{}:
		// Array node
		arrayNodeID := generateRandomNodeID("arr")
		arrayNode := c.getOrCreateNode(arrayNodeID, true, clientID, version)
		arrayNode.IsArray = true

		if parent != nil {
			if isArray {
				err := c.AppendEdge(parent.ID, arrayNodeID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			} else {
				err := c.AddEdge(parent.ID, arrayNodeID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			}
		}

		for i, item := range val {
			_, err := c.importRecursive(item, arrayNode, "", i, true, clientID)
			if err != nil {
				return "", err
			}
		}

		return arrayNodeID, nil

	default:
		// Literal node
		literalID := generateRandomNodeID("lit")
		literalNode := c.getOrCreateNode(literalID, false, clientID, version)
		err := literalNode.setLiteralWithVersion(val, clientID, version)
		if err != nil {
			return "", err
		}
		if parent != nil {
			if isArray {
				err := c.AppendEdge(parent.ID, literalID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			} else {
				err := c.AddEdge(parent.ID, literalID, edgeLabel, clientID)
				if err != nil {
					return "", err
				}
			}
		}
		return literalID, nil
	}
}

func (c *TreeCRDT) exportRaw() (map[string]interface{}, error) {
	nodes := make(map[string]interface{})

	for id, node := range c.Nodes {
		nodeData := map[string]interface{}{
			"id":           string(node.ID),
			"isArray":      node.IsArray,
			"isMap":        node.IsMap,
			"isLiteral":    node.IsLiteral,
			"literalValue": node.LiteralValue,
			"owner":        string(node.Owner),
			"clock":        node.Clock,
			"edges":        []map[string]interface{}{},
		}

		// Sort edges by LSEQ before exporting
		sortedEdges := make([]*EdgeCRDT, len(node.Edges))
		copy(sortedEdges, node.Edges)
		sortEdgesByLSEQ(sortedEdges)

		for _, edge := range sortedEdges {
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
		"root":  string(c.Root.ID),
		"nodes": nodes,
	}, nil
}

func (c *TreeCRDT) Save() ([]byte, error) {
	exported, err := c.exportRaw()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(exported, "", "  ")
}

func (c *TreeCRDT) importRaw(data map[string]interface{}) error {
	nodesData, ok := data["nodes"].(map[string]interface{})
	if !ok {
		return errors.New("invalid raw data: nodes missing")
	}
	c.Nodes = make(map[NodeID]*NodeCRDT)

	// First pass: create nodes
	for idStr, nodeRaw := range nodesData {
		nodeMap, ok := nodeRaw.(map[string]interface{})
		if !ok {
			return errors.New("invalid node data")
		}

		id := NodeID(idStr)
		node := newNodeFromID(id, false, c)

		if isArray, ok := nodeMap["isArray"].(bool); ok {
			node.IsArray = isArray
		}
		if isMap, ok := nodeMap["isMap"].(bool); ok {
			node.IsMap = isMap
		}
		if isLiteral, ok := nodeMap["isLiteral"].(bool); ok {
			node.IsLiteral = isLiteral
		}
		node.LiteralValue = nodeMap["literalValue"]

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

		c.Nodes[id] = node
	}

	// Second pass: recreate edges
	for idStr, nodeRaw := range nodesData {
		nodeMap := nodeRaw.(map[string]interface{})
		id := NodeID(idStr)
		node := c.Nodes[id]

		if edges, ok := nodeMap["edges"].([]interface{}); ok {
			for _, e := range edges {
				edgeMap, ok := e.(map[string]interface{})
				if !ok {
					continue
				}
				edge := &EdgeCRDT{}
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
				edge.From = id // override to ensure consistency
				node.Edges = append(node.Edges, edge)
			}
		}
	}

	rootID, ok := data["root"].(string)
	if !ok {
		return errors.New("invalid raw data: root missing")
	}
	rootNode, ok := c.Nodes[NodeID(rootID)]
	if !ok {
		return errors.New("root node not found")
	}
	c.Root = rootNode

	return nil
}

func (c *TreeCRDT) Load(data []byte) error {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}
	return c.importRaw(rawData)
}

func (c *TreeCRDT) ExportJSON() ([]byte, error) {
	exported, err := c.export()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(exported, "", "  ")
}

func (c *TreeCRDT) export() (interface{}, error) {
	visited := make(map[NodeID]bool)

	if len(c.Root.Edges) == 0 {
		return nil, fmt.Errorf("Root node has no edges")
	}

	if len(c.Root.Edges) == 1 {
		childID := c.Root.Edges[0].To
		return c.exportNodeOrdered(childID, visited)
	}

	isArray := false
	for _, e := range c.Root.Edges {
		node := c.Nodes[e.To]
		if node.IsArray {
			isArray = true
			break
		}
	}

	if isArray {
		sortEdgesByLSEQ(c.Root.Edges)

		var arrayItems []interface{}
		for _, e := range c.Root.Edges {
			child, err := c.exportNodeOrdered(e.To, visited)
			if err != nil {
				return nil, err
			}
			arrayItems = append(arrayItems, child)
		}
		return arrayItems, nil
	} else {
		// Root points to a map — order by label (in edge order)
		result := orderedmap.New()
		for _, e := range c.Root.Edges {
			child, err := c.exportNodeOrdered(e.To, visited)
			if err != nil {
				return nil, err
			}
			result.Set(e.Label, child)
		}
		return result, nil
	}
}

func (n *NodeCRDT) ExportJSON(crdt *TreeCRDT) ([]byte, error) {
	visited := make(map[NodeID]bool)
	result, err := exportNodeRecursive(n, crdt, visited)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(result, "", "  ")
}

func exportNodeRecursive(node *NodeCRDT, crdt *TreeCRDT, visited map[NodeID]bool) (interface{}, error) {
	if visited[node.ID] {
		return nil, fmt.Errorf("cycle detected at node %s", node.ID)
	}
	visited[node.ID] = true

	if node.IsLiteral {
		return node.LiteralValue, nil
	}

	obj := orderedmap.New()

	// Group edges by label (like keys in a JSON object)
	edgeGroups := make(map[string][]*EdgeCRDT)
	for _, edge := range node.Edges {
		edgeGroups[edge.Label] = append(edgeGroups[edge.Label], edge)
	}

	for field, edges := range edgeGroups {
		if len(edges) == 0 {
			continue
		}

		isArray := false
		for _, e := range edges {
			if child, ok := crdt.Nodes[e.To]; ok && child.IsArray {
				isArray = true
				break
			}
		}

		if isArray {
			// Export as array
			var arrayItems []interface{}
			for _, e := range edges {
				child, ok := crdt.Nodes[e.To]
				if !ok {
					return nil, fmt.Errorf("missing child node %s", e.To)
				}
				exportedChild, err := exportNodeRecursive(child, crdt, visited)
				if err != nil {
					return nil, err
				}
				arrayItems = append(arrayItems, exportedChild)
			}
			obj.Set(field, arrayItems)
		} else {
			// Export as single value
			child, ok := crdt.Nodes[edges[0].To]
			if !ok {
				return nil, fmt.Errorf("missing child node %s", edges[0].To)
			}
			exportedChild, err := exportNodeRecursive(child, crdt, visited)
			if err != nil {
				return nil, err
			}
			obj.Set(field, exportedChild)
		}
	}

	return obj, nil
}

func (c *TreeCRDT) exportNodeOrdered(id NodeID, visited map[NodeID]bool) (interface{}, error) {
	if visited[id] {
		return nil, fmt.Errorf("cycle detected at node %s", id)
	}
	visited[id] = true

	node, ok := c.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}

	// Literal node
	if node.IsLiteral {
		return node.LiteralValue, nil
	}

	// Array node
	if node.IsArray {
		sortEdgesByLSEQ(node.Edges)
		var arrayItems []interface{}
		for _, edge := range node.Edges {
			child, err := c.exportNodeOrdered(edge.To, visited)
			if err != nil {
				return nil, err
			}
			arrayItems = append(arrayItems, child)
		}
		return arrayItems, nil
	}

	// Map node
	if node.IsMap {
		result := orderedmap.New()
		for _, edge := range node.Edges {
			child, err := c.exportNodeOrdered(edge.To, visited)
			if err != nil {
				return nil, err
			}
			result.Set(edge.Label, child)
		}
		return result, nil
	}

	return nil, fmt.Errorf("node %s is neither literal, array, nor map", id)
}

func (c *TreeCRDT) Clone() (*TreeCRDT, error) {
	safeCopy, err := c.Save()
	if err != nil {
		return nil, err
	}
	newTreeCRDT := NewTreeCRDT()
	if err := newTreeCRDT.Load(safeCopy); err != nil {
		return nil, err
	}
	return newTreeCRDT, nil
}

func (c *TreeCRDT) Equal(other *TreeCRDT) bool {
	if len(c.Nodes) != len(other.Nodes) {
		log.WithFields(log.Fields{"Nodes1": len(c.Nodes), "Nodes2": len(other.Nodes)}).Warning("Node counts not equal")
		return false
	}

	for id, node := range c.Nodes {
		otherNode, ok := other.Nodes[id]
		if !ok || !nodesSemanticallyEqual(node, otherNode) {
			log.WithFields(log.Fields{"NodeID": id, "Node1": node, "Node2": otherNode}).Warning("Nodes not equal")
			return false
		}
	}

	return true
}

func nodesSemanticallyEqual(n1, n2 *NodeCRDT) bool {
	if n1.IsArray != n2.IsArray || n1.IsLiteral != n2.IsLiteral || n1.IsMap != n2.IsMap {
		log.WithFields(log.Fields{
			"IsArray1": n1.IsArray, "IsArray2": n2.IsArray,
			"IsLiteral1": n1.IsLiteral, "IsLiteral2": n2.IsLiteral,
			"IsMap1": n1.IsMap, "IsMap2": n2.IsMap,
		}).Warning("Node type flags not equal")
		return false
	}

	if n1.IsLiteral {
		if !reflect.DeepEqual(n1.LiteralValue, n2.LiteralValue) {
			log.WithFields(log.Fields{
				"NodeID1":    n1.ID,
				"NodeID2":    n2.ID,
				"Value1":     fmt.Sprintf("%#v", n1.LiteralValue),
				"Type1":      fmt.Sprintf("%T", n1.LiteralValue),
				"Value2":     fmt.Sprintf("%#v", n2.LiteralValue),
				"Type2":      fmt.Sprintf("%T", n2.LiteralValue),
				"Owner1":     n1.Owner,
				"Owner2":     n2.Owner,
				"Clock1":     n1.Clock,
				"Clock2":     n2.Clock,
				"IsArray1":   n1.IsArray,
				"IsArray2":   n2.IsArray,
				"IsMap1":     n1.IsMap,
				"IsMap2":     n2.IsMap,
				"EdgeCount1": len(n1.Edges),
				"EdgeCount2": len(n2.Edges),
			}).Warning("Literal values not equal")
			return false
		}

		// if n1.IsLiteral {
		// 	if !reflect.DeepEqual(n1.LiteralValue, n2.LiteralValue) {
		// 		log.WithFields(log.Fields{
		// 			"LiteralValue1": n1.LiteralValue,
		// 			"LiteralValue2": n2.LiteralValue,
		// 		}).Warning("Literal values not equal")
		// 		return false
		// 	}
	}

	if len(n1.Edges) != len(n2.Edges) {
		log.WithFields(log.Fields{"Edges1": len(n1.Edges), "Edges2": len(n2.Edges)}).Warning("Edge counts not equal")
		return false
	}

	if n1.IsArray {
		// Compare edges by LSEQ order
		sorted1 := make([]*EdgeCRDT, len(n1.Edges))
		sorted2 := make([]*EdgeCRDT, len(n2.Edges))
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
		// Compare edges as unordered field entries (map-like)
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

func (c *TreeCRDT) SemanticVersion() (string, error) {
	exported, err := c.export()
	if err != nil {
		return "", err
	}

	// Marshal to canonical JSON using orderedmap and sorted arrays
	bytes, err := json.Marshal(exported)
	if err != nil {
		return "", err
	}

	hash := sha3.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}
