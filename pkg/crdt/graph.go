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
	node := g.GetOrCreateNode(id, isArray)
	g.AddEdge(parentID, id, "", clientID)
	return node
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
	winningClock, _ := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) {
		// New clock wins -> insert edge
		for _, edge := range node.Edges {
			if edge.Position >= position {
				edge.Position++
			}
		}

		newEdge := &Edge{From: from, To: to, Label: label, Position: position}
		node.Edges = append(node.Edges, newEdge)

		node.Clock = newClock
		node.Owner = clientID

		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Position": position, "Version": newVersion}).Debug("Edge inserted")
	} else {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Position": position, "Version": newVersion}).Debug("Edge insert ignored due to conflict")
	}

	return nil
}

func (g *Graph) InsertEdge(from, to NodeID, label string, position int, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return errors.New("Cannot insert edge, node not found: " + string(from))
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.insertEdgeWithVersion(from, to, label, position, clientID, newVersion)
}

func (g *Graph) removeEdgeWithVersion(from, to NodeID, label string, clientID ClientID, newVersion int) error {
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
		newEdges := []*Edge{}
		for _, edge := range node.Edges {
			if !(edge.To == to && edge.Label == label) {
				newEdges = append(newEdges, edge)
			}
		}
		node.Edges = newEdges
		node.Clock = newClock
		node.Owner = clientID

		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge removed")
	} else {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge remove ignored due to conflict")
	}

	return nil
}

func (g *Graph) RemoveEdge(from, to NodeID, label string, clientID ClientID) error {
	node, ok := g.Nodes[from]
	if !ok {
		return fmt.Errorf("Cannot remove edge, node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return g.removeEdgeWithVersion(from, to, label, clientID, newVersion)
}

func (n *Node) SetLiteral(value interface{}, clientID ClientID, version int) {
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
	}
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
