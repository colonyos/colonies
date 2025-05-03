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
	From     NodeID `json:"from"`
	To       NodeID `json:"to"`
	Label    string `json:"label"`
	Position int    `json:"position"`
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
		edge := &Edge{From: from, To: to, Label: label, Position: -1}
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

// func (g *Graph) insertEdgeWithVersionOld(from, to NodeID, label string, position int, clientID ClientID, newVersion int) error {
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
// 	winningClock, _ := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)
//
// 	if clocksEqual(winningClock, newClock) {
// 		// New clock wins -> insert edge
// 		for _, edge := range node.Edges {
// 			if edge.Position >= position {
// 				edge.Position++
// 			}
// 		}
//
// 		newEdge := &Edge{From: from, To: to, Label: label, Position: position}
// 		node.Edges = append(node.Edges, newEdge)
// 		node.Clock = newClock
// 		node.Owner = clientID
//
// 		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Position": position, "Version": newVersion}).Debug("Edge inserted")
// 	} else {
// 		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Position": position, "Version": newVersion}).Debug("Edge insert ignored due to conflict")
// 	}
//
// 	return nil
// }
//

func (g *Graph) insertEdgeWithVersion(from, to NodeID, label string, position int, clientID ClientID, newVersion int) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot insert edge, node not found: " + string(from))
	}

	if position < 0 {
		return errors.New("Cannot insert edge, position must be non-negative")
	} else if position > len(node.Edges) {
		return errors.New("Cannot insert edge, position out of bounds")
	}

	// Prepare the new clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

	// Resolve clock conflict
	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
		// New clock wins -> insert edge
		for _, edge := range node.Edges {
			if edge.Position >= position {
				edge.Position++
			}
		}

		newEdge := &Edge{
			From:     from,
			To:       to,
			Label:    label,
			Position: position,
		}
		node.Edges = append(node.Edges, newEdge)

		node.Clock = newClock
		node.Owner = winningOwner

		log.WithFields(log.Fields{
			"NodeID":   from,
			"To":       to,
			"Label":    label,
			"Position": position,
			"Version":  newVersion,
		}).Debug("Edge inserted")
	} else {
		log.WithFields(log.Fields{
			"NodeID":   from,
			"To":       to,
			"Label":    label,
			"Position": position,
			"Version":  newVersion,
		}).Debug("Edge insert ignored due to conflict")
	}

	return nil
}

func (g *Graph) insertEdge(from, to NodeID, label string, position int, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot insert edge, node not found: " + string(from))
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.insertEdgeWithVersion(from, to, label, position, clientID, newVersion)
}

// We cannot use the position in the array directly as it may change when merging
// nodes. Instead, we need to find the sibling edge and insert before or after it.
func (g *Graph) InsertEdge(from, to, sibling NodeID, label string, left bool, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot insert edge, node not found: " + string(from))
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	if len(node.Edges) == 0 {
		return g.insertEdgeWithVersion(from, to, label, 0, clientID, newVersion)
	}

	// Find the singling edge
	var siblingEdge *Edge
	for _, edge := range node.Edges {
		if edge.To == sibling {
			siblingEdge = edge
			break
		}
	}

	if siblingEdge == nil {
		return errors.New("Cannot insert edge, sibling edge not found")
	}

	if left {
		// Insert before the sibling edge
		return g.insertEdgeWithVersion(from, to, label, siblingEdge.Position, clientID, newVersion)
	} else {
		// Insert after the sibling edge
		return g.insertEdgeWithVersion(from, to, label, siblingEdge.Position+1, clientID, newVersion)
	}
}

func (g *Graph) GetSiblingNode(nodeID NodeID, pos int) (*Node, error) {
	node, ok := g.Nodes[nodeID]
	if !ok {
		return nil, errors.New("Cannot find node: " + string(nodeID))
	}
	if len(node.Edges) == 0 {
		return nil, errors.New("Cannot find sibling node, no edges")
	}

	if pos < 0 || pos >= len(node.Edges) {
		return nil, errors.New("Cannot find sibling node, position out of bounds")
	}
	for _, edge := range node.Edges {
		if edge.Position == pos {
			return g.Nodes[edge.To], nil
		}
	}
	return nil, errors.New("Cannot find sibling node, position not found")
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

// // Note when merging two nodes with the same ID, the following rules apply:
// // n1 + n2 → [n1, n2]
// // [n1, n2] + [n2, n3] → [n1, n2, n3] (with deduplication and order preservation)
// func (g *Graph) MergeOld(g2 *Graph) {
// 	//log.SetLevel(log.DebugLevel)
//
// 	for id, remote := range g2.Nodes {
// 		log.WithField("NodeID", id).Debug("Merging node")
//
// 		local, exists := g.Nodes[id]
// 		if !exists {
// 			log.WithField("NodeID", id).Debug("Node does not exist in local graph, cloning from remote")
// 			cloned := newNodeFromID(id, remote.IsArray)
// 			cloned.Fields = make(map[string]VersionedField)
// 			for k, v := range remote.Fields {
// 				cloned.Fields[k] = v
// 			}
// 			cloned.Litteral = remote.Litteral
// 			cloned.LitteralValue = remote.LitteralValue
// 			cloned.Clock = copyClock(remote.Clock)
// 			cloned.Owner = remote.Owner
// 			g.Nodes[id] = cloned
// 			continue
// 		}
//
// 		// Merge fields
// 		for k, remoteField := range remote.Fields {
// 			g.Nodes[id].SetField(k, remoteField.Value, remoteField.Owner, remoteField.Clock[remoteField.Owner])
// 		}
//
// 		// Merge literal
// 		if remote.Litteral {
// 			// Ignore error
// 			err := local.SetLiteral(remote.LitteralValue, remote.Owner, remote.Clock[remote.Owner])
// 			if err != nil {
// 				log.WithField("NodeID", id).Error("Failed to set literal value")
// 			}
// 		}
//
// 		// Merge edges
// 		for _, re := range remote.Edges {
// 			exists := false
// 			for _, le := range local.Edges {
// 				if le.To == re.To {
// 					exists = true
// 					break
// 				}
// 			}
// 			if !exists {
// 				var err error
// 				if _, exists := g.Nodes[re.To]; !exists {
// 					if remoteNode, ok := g2.Nodes[re.To]; ok {
// 						cloned := newNodeFromID(re.To, true) // Always become an array when merging nodes
// 						cloned.Fields = make(map[string]VersionedField)
// 						for k, v := range remoteNode.Fields {
// 							cloned.Fields[k] = v
// 						}
// 						cloned.Litteral = remoteNode.Litteral
// 						cloned.LitteralValue = remoteNode.LitteralValue
// 						cloned.Clock = copyClock(remoteNode.Clock)
// 						cloned.Owner = remoteNode.Owner
// 						g.Nodes[re.To] = cloned
// 					}
// 				}
// 				var sibling *Node
// 				sibling, err = g2.GetSiblingNode(re.From, re.Position-1)
// 				if err != nil {
// 					log.WithField("NodeID", re.From).Error("Failed to find sibling node")
// 				}
// 				if re.Position < 0 {
// 					// Special case: We are merging two non-array nodes
// 					// We need to find the sibling node and insert before or after it
// 					err := g.insertEdge(re.From, re.To, re.Label, 0, remote.Owner)
// 					if err != nil {
// 						log.WithField("NodeID", re.From).Error("Failed to insert edge")
// 					} else {
// 						// Sort edges deterministically after insert
// 						if parent, ok := g.Nodes[re.From]; ok {
// 							sortEdgesByNodeID(parent.Edges)
// 						}
// 					}
// 					continue
// 				}
//
// 				if sibling == nil {
// 					err := g.insertEdge(re.From, re.To, re.Label, 0, remote.Owner)
// 					if err != nil {
// 						log.WithField("NodeID", re.From).Debug("Failed to insert edge")
// 					}
// 				} else {
// 					fmt.Println("Sibloing to remote node", re.From, "is ", sibling.ID)
// 					fmt.Println("Sibling found", sibling.ID)
// 					// Insert edge before or after the sibling node
// 					err := g.InsertEdge(re.From, re.To, sibling.ID, re.Label, false, remote.Owner)
// 					if err != nil {
// 						log.WithField("NodeID", re.From).Error("Failed to insert edge")
// 					}
// 				}
// 			}
// 		}
// 	}
// }

// Note when merging two nodes with the same ID, the following rules apply:
// n1 + n2 → [n1, n2]
// [n1, n2] + [n2, n3] → [n1, n2, n3] (with deduplication and order preservation)
// Note when merging two nodes with the same ID, the following rules apply:
// n1 + n2 → [n1, n2]
// [n1, n2] + [n2, n3] → [n1, n2, n3] (with deduplication and order preservation)
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

		// 🔒 Precompute merged values before CRDT ops can override them
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
				if s, err := g2.GetSiblingNode(re.From, re.Position-1); err == nil {
					sibling = s
				}
			}

			// Insert edge respecting sibling, or default to beginning
			if sibling != nil {
				if err := g.InsertEdge(re.From, re.To, sibling.ID, re.Label, false, remote.Owner); err != nil {
					log.WithField("NodeID", re.From).Error("InsertEdge failed")
				}
			} else {
				if err := g.insertEdge(re.From, re.To, re.Label, 0, remote.Owner); err != nil {
					log.WithField("NodeID", re.From).Error("InsertEdge fallback failed")
				} else {
					if parent, ok := g.Nodes[re.From]; ok {
						sortEdgesByNodeID(parent.Edges)
					}
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
