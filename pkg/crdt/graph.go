package crdt

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
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
	From         NodeID `json:"from"`
	To           NodeID `json:"to"`
	Label        string `json:"label"`
	Position     int    `json:"position"`
	LSEQPosition []int  `json:"lseqposition"` // LSEQ position
}

type Graph struct {
	Root  *Node
	Nodes map[NodeID]*Node
}

func (g *Graph) CreateAttachedNode(name string, isArray bool, parentID NodeID, clientID ClientID) *Node {
	id := generateRandomNodeID(name)
	node := g.getOrCreateNode(id, isArray, clientID, 1)
	g.AddEdge(parentID, id, "", clientID)
	return node
}

func (g *Graph) CreateNode(name string, isArray bool, clientID ClientID) *Node {
	id := generateRandomNodeID(name)
	node := g.getOrCreateNode(id, isArray, clientID, 1)
	return node
}

func newNodeFromID(id NodeID, isArray bool) *Node {
	node := &Node{
		ID:      id,
		Fields:  make(map[string]VersionedField),
		Edges:   make([]*Edge, 0),
		IsArray: isArray,
	}

	return node
}

func (g *Graph) getOrCreateNode(id NodeID, isArray bool, clientID ClientID, version int) *Node {
	if _, ok := g.Nodes[id]; !ok {
		node := newNodeFromID(id, isArray)
		g.Nodes[id] = node
		node.Clock = make(VectorClock)
		node.Clock[clientID] = version
		node.Owner = clientID
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
		Root:  newNodeFromID(NodeID("root"), false),
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
	currentField := n.Fields[key]

	// Start a fresh clock
	newClock := make(VectorClock)
	newClock[clientID] = version

	// Resolve conflict
	winningClock, _ := resolveConflict(currentField.Clock, newClock, currentField.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) {
		// New value wins
		n.Fields[key] = VersionedField{
			Value: value,
			Clock: newClock,
			Owner: clientID,
		}
		log.WithFields(log.Fields{"NodeID": n.ID, "Key": key, "Value": value}).Debug("Set field")
	} else {
		log.WithFields(log.Fields{"NodeID": n.ID, "Key": key, "CurrentField": currentField, "NewClock": newClock, "WinningClock": winningClock}).Debug("Conflict detected")
	}
}

func (n *Node) RemoveField(key string, clientID ClientID, version int) {
	currentField := n.Fields[key]

	// Start a fresh clock for the removal
	newClock := make(VectorClock)
	newClock[clientID] = version

	// Resolve conflict
	winningClock, _ := resolveConflict(currentField.Clock, newClock, currentField.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) {
		// New clock wins, so remove the field
		delete(n.Fields, key)
		log.WithFields(log.Fields{"NodeID": n.ID, "Key": key}).Debug("Removed field")
	} else {
		log.WithFields(log.Fields{
			"NodeID":       n.ID,
			"Key":          key,
			"CurrentField": currentField,
			"NewClock":     newClock,
			"WinningClock": winningClock,
		}).Debug("RemoveField conflict detected — keeping existing field")
	}
}

func (g *Graph) addEdgeWithVersion(from, to NodeID, label string, clientID ClientID, newVersion int) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot add edge, node not found: " + string(from))
	}

	// Prepare the new clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

	// Resolve clock conflict
	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) && (clientID == winningOwner) {
		edge := &Edge{From: from, To: to, Label: label, Position: -1, LSEQPosition: make([]int, 0)}
		node.Edges = append(node.Edges, edge)
		node.Clock = newClock
		node.Owner = clientID

		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge added")
	} else {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge add ignored due to conflict")
	}

	return nil
}

func (g *Graph) AddEdge(from, to NodeID, label string, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot add edge, node not found: " + string(from))
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.addEdgeWithVersion(from, to, label, clientID, newVersion)
}

func (g *Graph) AppendEdge(from, to NodeID, label string, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("AppendEdge: parent node %s not found", from)
	}

	var lastSibling NodeID
	if len(node.Edges) > 0 {
		// Use the last edge as anchor for right-side insert
		last := node.Edges[len(node.Edges)-1]
		lastSibling = last.To
	} else {
		// No siblings yet, insert at the beginning
		lastSibling = ""
	}

	newVersion := node.Clock[clientID] + 1
	return g.insertEdgeWithVersion(from, to, label, lastSibling, false, clientID, newVersion)
}

func (g *Graph) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("InsertEdge: parent node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.insertEdgeWithVersion(from, to, label, sibling, true, clientID, newVersion)
}

func (g *Graph) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("InsertEdge: parent node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.insertEdgeWithVersion(from, to, label, sibling, false, clientID, newVersion)
}

func (g *Graph) insertEdgeWithVersion(from, to NodeID, label string, sibling NodeID, left bool, clientID ClientID, newVersion int) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("insertWithVersion: parent node %s not found", from)
	}

	// Prepare clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion
	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)
	if !clocksEqual(winningClock, newClock) || winningOwner != clientID {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Version": newVersion}).Error("insertWithVersion ignored due to conflict")
		return nil
	}

	// Sort edges for position lookup
	sorted := make([]*Edge, len(node.Edges))
	copy(sorted, node.Edges)
	sortEdgesByLSEQ(sorted)

	var leftPos, rightPos Position
	found := false

	if sibling == "" || len(sorted) == 0 {
		// Insert at beginning
		leftPos = []int{}
		rightPos = []int{Base}
	} else {
		for i, e := range sorted {
			if e.To == sibling {
				found = true
				if left {
					// Insert to the left of sibling
					if i > 0 {
						leftPos = sorted[i-1].LSEQPosition
					} else {
						leftPos = []int{}
					}
					rightPos = e.LSEQPosition
				} else {
					// Insert to the right of sibling
					leftPos = e.LSEQPosition
					if i+1 < len(sorted) {
						rightPos = sorted[i+1].LSEQPosition
					} else {
						rightPos = []int{Base}
					}
				}
				break
			}
		}
		if !found {
			leftPos = []int{}
			rightPos = []int{Base}
		}
	}

	newPos := generatePositionBetweenLSEQ(leftPos, rightPos)

	edge := &Edge{
		From:         from,
		To:           to,
		Label:        label,
		LSEQPosition: newPos,
		Position:     -1, // Deprecated
	}
	node.Edges = append(node.Edges, edge)
	sortEdgesByLSEQ(node.Edges)

	node.Clock = newClock
	node.Owner = clientID

	log.WithFields(log.Fields{
		"NodeID":       from,
		"To":           to,
		"Sibling":      sibling,
		"Left":         left,
		"LSEQPosition": newPos,
		"Version":      newVersion,
	}).Debug("InsertEdge succeeded")

	return nil
}

// func (g *Graph) insertEdgeWithVersion(from, to NodeID, label string, leftOf NodeID, clientID ClientID, newVersion int) error {
// 	node, ok := g.Nodes[from]
// 	if !ok {
// 		return fmt.Errorf("InsertEdge: parent node %s not found", from)
// 	}
//
// 	// Prepare clock
// 	newClock := copyClock(node.Clock)
// 	newClock[clientID] = newVersion
// 	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)
// 	if !clocksEqual(winningClock, newClock) || winningOwner != clientID {
// 		log.WithFields(log.Fields{"NodeID": from, "To": to, "Version": newVersion}).Debug("InsertEdgeLSEQ ignored due to conflict")
// 		return nil
// 	}
//
// 	// Find left and right LSEQ positions
// 	var leftPos, rightPos Position
//
// 	if leftOf == "" || len(node.Edges) == 0 {
// 		// No left sibling, insert at the beginning
// 		leftPos = []int{}
// 		rightPos = []int{Base}
// 	} else {
// 		// Find LSEQ positions surrounding the insertion point
// 		found := false
// 		for i, e := range node.Edges {
// 			if e.To == leftOf {
// 				leftPos = e.LSEQPosition
// 				if i+1 < len(node.Edges) {
// 					rightPos = node.Edges[i+1].LSEQPosition
// 				} else {
// 					rightPos = []int{Base}
// 				}
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			// Fallback to inserting at the beginning if sibling not found
// 			leftPos = []int{}
// 			rightPos = []int{Base}
// 		}
// 	}
//
// 	// Generate LSEQ position
// 	newPos := generatePositionBetweenLSEQ(leftPos, rightPos)
//
// 	// Insert edge
// 	edge := &Edge{
// 		From:         from,
// 		To:           to,
// 		Label:        label,
// 		LSEQPosition: newPos,
// 		Position:     -1, // Deprecated
// 	}
// 	node.Edges = append(node.Edges, edge)
// 	sortEdgesByLSEQ(node.Edges)
//
// 	node.Clock = newClock
// 	node.Owner = clientID
//
// 	log.WithFields(log.Fields{"NodeID": from, "To": to, "LSEQPosition": newPos, "Version": newVersion}).Debug("InsertEdge succeeded")
// 	return nil
// }

//
// func (g *Graph) insertEdgeWithVersion(from, to NodeID, label string, position int, clientID ClientID, newVersion int) error {
// 	node, ok := g.Nodes[from]
// 	if !ok {
// 		return errors.New("Cannot insert edge, node not found: " + string(from))
// 	}
//
// 	if position < 0 {
// 		return errors.New("Cannot insert edge, position must be non-negative")
// 	} else if position > len(node.Edges) {
// 		return errors.New("Cannot insert edge, position out of bounds")
// 	}
//
// 	// Prepare the new clock
// 	newClock := copyClock(node.Clock)
// 	newClock[clientID] = newVersion
//
// 	// Resolve clock conflict
// 	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)
//
// 	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
// 		// New clock wins -> insert edge
// 		for _, edge := range node.Edges {
// 			if edge.Position >= position {
// 				edge.Position++
// 			}
// 		}
//
// 		newEdge := &Edge{
// 			From:     from,
// 			To:       to,
// 			Label:    label,
// 			Position: position,
// 		}
// 		node.Edges = append(node.Edges, newEdge)
//
// 		node.Clock = newClock
// 		node.Owner = winningOwner
//
// 		log.WithFields(log.Fields{
// 			"NodeID":   from,
// 			"To":       to,
// 			"Label":    label,
// 			"Position": position,
// 			"Version":  newVersion,
// 		}).Debug("Edge inserted")
// 	} else {
// 		log.WithFields(log.Fields{
// 			"NodeID":   from,
// 			"To":       to,
// 			"Label":    label,
// 			"Position": position,
// 			"Version":  newVersion,
// 		}).Debug("Edge insert ignored due to conflict")
// 	}
//
// 	return nil
// }
//
// func (g *Graph) insertEdge(from, to NodeID, label string, position int, clientID ClientID) error {
// 	node, ok := g.Nodes[from]
// 	if !ok {
// 		return errors.New("Cannot insert edge, node not found: " + string(from))
// 	}
// 	latestVersion := node.Clock[clientID]
// 	newVersion := latestVersion + 1
//
// 	return g.insertEdgeWithVersion(from, to, label, position, clientID, newVersion)
// }
//
// // We cannot use the position in the array directly as it may change when merging
// // nodes. Instead, we need to find the sibling edge and insert before or after it.
// func (g *Graph) InsertEdge(from, to, sibling NodeID, label string, left bool, clientID ClientID) error {
// 	node, ok := g.Nodes[from]
// 	if !ok {
// 		return errors.New("Cannot insert edge, node not found: " + string(from))
// 	}
// 	latestVersion := node.Clock[clientID]
// 	newVersion := latestVersion + 1
//
// 	if len(node.Edges) == 0 {
// 		return g.insertEdgeWithVersion(from, to, label, 0, clientID, newVersion)
// 	}
//
// 	// Find the singling edge
// 	var siblingEdge *Edge
// 	for _, edge := range node.Edges {
// 		if edge.To == sibling {
// 			siblingEdge = edge
// 			break
// 		}
// 	}
//
// 	if siblingEdge == nil {
// 		return errors.New("Cannot insert edge, sibling edge not found")
// 	}
//
// 	if left {
// 		// Insert before the sibling edge
// 		return g.insertEdgeWithVersion(from, to, label, siblingEdge.Position, clientID, newVersion)
// 	} else {
// 		// Insert after the sibling edge
// 		return g.insertEdgeWithVersion(from, to, label, siblingEdge.Position+1, clientID, newVersion)
// 	}
// }

// func (g *Graph) GetSiblingNode(nodeID NodeID, pos int) (*Node, error) {
// 	node, ok := g.Nodes[nodeID]
// 	if !ok {
// 		return nil, errors.New("Cannot find node: " + string(nodeID))
// 	}
// 	if len(node.Edges) == 0 {
// 		return nil, errors.New("Cannot find sibling node, no edges")
// 	}
//
// 	if pos < 0 || pos >= len(node.Edges) {
// 		return nil, errors.New("Cannot find sibling node, position out of bounds")
// 	}
// 	for _, edge := range node.Edges {
// 		if edge.Position == pos {
// 			return g.Nodes[edge.To], nil
// 		}
// 	}
// 	return nil, errors.New("Cannot find sibling node, position not found")
// }

func (g *Graph) GetSibling(nodeID NodeID, index int) (*Node, error) {
	node, ok := g.Nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("Cannot find node: %s", nodeID)
	}

	if len(node.Edges) == 0 {
		return nil, fmt.Errorf("Cannot find sibling node, no edges")
	}

	// Sort edges by LSEQ
	sorted := make([]*Edge, len(node.Edges))
	copy(sorted, node.Edges)
	sortEdgesByLSEQ(sorted)

	if index < 0 || index >= len(sorted) {
		return nil, fmt.Errorf("Sibling index %d out of bounds", index)
	}

	siblingID := sorted[index].To
	sibling, exists := g.Nodes[siblingID]
	if !exists {
		return nil, fmt.Errorf("Sibling node %s not found in graph", siblingID)
	}

	return sibling, nil
}

func (g *Graph) insertEdgeLSEQ(from, to NodeID, label string, leftOf NodeID, clientID ClientID, version int) error {
	parent, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("insertEdgeLSEQ: parent node %s not found", from)
	}

	var leftPos, rightPos Position
	found := false

	// Find LSEQ positions surrounding the insertion point
	for i, e := range parent.Edges {
		if e.To == leftOf {
			found = true
			leftPos = e.LSEQPosition
			if i+1 < len(parent.Edges) {
				rightPos = parent.Edges[i+1].LSEQPosition
			} else {
				rightPos = []int{Base}
			}
			break
		}
	}

	if !found {
		leftPos = []int{}
		rightPos = []int{Base}
	}

	newPos := generatePositionBetweenLSEQ(leftPos, rightPos)

	newEdge := &Edge{
		From:         from,
		To:           to,
		Label:        label,
		LSEQPosition: newPos,
	}
	parent.Edges = append(parent.Edges, newEdge)
	sortEdgesByLSEQ(parent.Edges)

	return nil
}

func (g *Graph) removeEdgeWithVersion(from, to NodeID, clientID ClientID, newVersion int) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("Cannot remove edge, node %s not found", from)
	}

	// Prepare the new clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

	// Resolve clock conflict
	winningClock, _ := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) {
		// New clock wins -> allow edge removal
		index := -1
		newEdges := []*Edge{}
		for _, edge := range node.Edges {
			if !(edge.To == to) {
				newEdges = append(newEdges, edge)
			}
			if edge.To == to {
				index = edge.Position
			}
		}
		node.Edges = newEdges
		node.Clock = newClock
		node.Owner = clientID

		for _, edge := range newEdges {
			if edge.Position > index {
				edge.Position--
			}
		}

		log.WithFields(log.Fields{"NodeID": from, "To": to, "Version": newVersion}).Debug("Edge removed")
	} else {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Version": newVersion}).Debug("Edge remove ignored due to conflict")
		return fmt.Errorf("Cannot remove edge, conflict detected: %s", from)
	}

	return nil
}

func (g *Graph) RemoveEdge(from, to NodeID, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("Cannot remove edge, node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.removeEdgeWithVersion(from, to, clientID, newVersion)
}

func (n *Node) SetLiteral(value interface{}, clientID ClientID, version int) error {
	currentClock := n.Clock
	newClock := make(VectorClock)
	newClock[clientID] = version

	winningClock, winningOwner := resolveConflict(currentClock, newClock, n.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
		n.Litteral = true
		n.LitteralValue = value
		n.Clock = newClock
		n.Owner = clientID
		log.WithFields(log.Fields{"NodeID": n.ID, "LiteralValue": value}).Debug("Set literal value")
	} else {
		log.WithFields(log.Fields{"NodeID": n.ID, "AttemptedLiteralValue": value, "ExistingOwner": n.Owner, "WinningOwner": winningOwner}).Debug("Literal set ignored due to conflict")
		return fmt.Errorf("Cannot set literal value, conflict detected: %s", n.ID)
	}

	return nil
}

// Tidy removes all nodes that are not referenced by any edges.
//
// WARNING:
// - This function should NOT be called automatically after every change.
// - In CRDTs, a node that looks "orphaned" now may be referenced later by concurrent operations.
//
// Recommended usage:
//   - Call Tidy() manually after a batch of operations is complete,
//     when the graph is known to be stable.
//   - Optionally call Tidy() periodically (e.g., background maintenance) or before persisting to disk.
//
// This helps keep the graph compact without risking consistency.
func (g *Graph) Tidy() {
	referenced := make(map[NodeID]bool)

	// Mark all referenced nodes (target of edges)
	for _, node := range g.Nodes {
		for _, edge := range node.Edges {
			referenced[edge.To] = true
		}
	}

	// Always preserve the root node
	referenced[g.Root.ID] = true

	// Now delete all nodes that are unreferenced
	for id := range g.Nodes {
		if !referenced[id] {
			delete(g.Nodes, id)
			log.WithFields(log.Fields{"NodeID": id}).Debug("Purged unreferenced node")
		}
	}
}

// Rules for merging graphs containing arrays:
//
// 1. When two nodes with the same parent and label are concurrently added,
//    they are treated as elements of an array (since JSON has no native set type).
//    This ensures deterministic, append-only behavior.
//
//    Example:
//      n1 + n2 → [n1, n2] (becomes an array even if originally a single value)
//    The nodes must be sorted by ther node ID to ensure deterministic order.
//
// 2. When two arrays are merged, their items are combined while preserving order
//    and deduplicating based on NodeID. Relative sibling order determines final position.
//
//    Example:
//      [n1, n2, n3] + [n0, n1] → [n0, n1, n2, n3]
//
// 3. If a sibling (anchor) is not found for a node, the node is inserted at the start.
//
//    Example:
//      [n1, n2] + [n3] → [n3, n1, n2]
//
// 4. If multiple arrays are merged concurrently, insertion order is resolved by
//    using the earliest common sibling (or none, default to front).
//
//    Example:
//      [n1, n2] + [n2, n3, n4] + [n1, n5, n6] → [n1, n5, n6, n2, n3, n4]
//      [n1, n2] + [n1, n5, n6] + [n2, n3, n4] → [n1, n5, n6, n2, n3, n4]
//
// 5. Concurrent insertions after same anchor.
//    Considering the following merge operations leading to inconsistent order:
//      [n1] + [n1, n3, n4] + [n1, n5, n6] → [n1, n5, n6, n3, n4]
//      [n1] + [n1, n5, n6] + [n1, n3, n4] → [n1, n3, n4, n5, n6]
//
//    Solution: User must specify a start and end position for the insertion.
//    merge([n1], [n1, n3, n4], start=n1, end="")  -> [n1, n3, n4]
//    When merging second time, we can detect that there are already two nodes added between start and stop, then
//    we need to sort them by their node ID to ensure deterministic order.
//    merge([n1, n3, n4], [n1, n5, n6], start=n1, end="")  -> [n1, sort([n3, n4, n5, n6])] -> [n1, n3, n4, n5, n6]

func (g *Graph) Merge(g2 *Graph) {
	for id, remote := range g2.Nodes {
		log.WithField("NodeID", id).Debug("Merging node")

		local, exists := g.Nodes[id]
		if !exists {
			log.WithField("NodeID", id).Debug("Node does not exist in local graph, cloning from remote")
			cloned := newNodeFromID(id, remote.IsArray)
			cloned.Fields = make(map[string]VersionedField)
			for k, v := range remote.Fields {
				cloned.Fields[k] = v
			}
			cloned.Litteral = remote.Litteral
			cloned.LitteralValue = remote.LitteralValue
			cloned.Clock = copyClock(remote.Clock)
			cloned.Owner = remote.Owner
			g.Nodes[id] = cloned
			continue
		}

		// Precompute merged values before CRDT ops can override them
		mergedClock := mergeClocks(local.Clock, remote.Clock)
		mergedOwner := lowestClientID(local.Owner, remote.Owner)

		// Merge fields
		for k, remoteField := range remote.Fields {
			local.SetField(k, remoteField.Value, remoteField.Owner, remoteField.Clock[remoteField.Owner])
		}

		// Merge literal
		if remote.Litteral {
			_ = local.SetLiteral(remote.LitteralValue, remote.Owner, remote.Clock[remote.Owner])
		}

		// Merge edges
		for _, re := range remote.Edges {
			alreadyExists := false
			for _, le := range local.Edges {
				if le.To == re.To {
					alreadyExists = true
					break
				}
			}
			if alreadyExists {
				continue
			}

			// Clone target node if missing
			if _, exists := g.Nodes[re.To]; !exists {
				if remoteNode, ok := g2.Nodes[re.To]; ok {
					cloned := newNodeFromID(re.To, true)
					cloned.Fields = make(map[string]VersionedField)
					for k, v := range remoteNode.Fields {
						cloned.Fields[k] = v
					}
					cloned.Litteral = remoteNode.Litteral
					cloned.LitteralValue = remoteNode.LitteralValue
					cloned.Clock = copyClock(remoteNode.Clock)
					cloned.Owner = remoteNode.Owner
					g.Nodes[re.To] = cloned
				}
			}

			// Determine sibling node (if any)
			var sibling *Node
			if re.Position > 0 {
				if s, err := g2.GetSibling(re.From, re.Position-1); err == nil {
					sibling = s
				}
			}

			// Insert edge respecting sibling, or default to beginning
			if sibling != nil {
				//InsertEdge(from, to NodeID, label string, leftOf NodeID, clientID ClientID) error {
				if err := g.InsertEdgeLeft(re.From, re.To, re.Label, sibling.ID, remote.Owner); err != nil {
					log.WithField("NodeID", re.From).Error("InsertEdge failed")
				}
			} else {
				err := g.InsertEdgeLeft(re.From, re.To, re.Label, "", remote.Owner)
				if err != nil {
					log.WithField("NodeID", re.From).Error("InsertEdge failed")
				}
			}
		}

		// Apply merged values after mutation logic
		local.Clock = mergedClock
		local.Owner = mergedOwner
	}

	g.normalize()
}

func (g *Graph) normalize() {
	for _, node := range g.Nodes {
		sortEdgesByNodeID(node.Edges)
	}
}
