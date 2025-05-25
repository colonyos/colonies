package crdt

import (
	"errors"
	"fmt"
	"sort"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type NodeID string

// TODO: Implement MarshalJSON() / UnmarshalJSON() manually or via a helper struct to prevent raw access to Node stuct fields
type Node struct {
	tree         *TreeCRDT
	ID           NodeID      `json:"id"`
	Edges        []*Edge     `json:"edges"`
	Clock        VectorClock `json:"clock"`
	Owner        ClientID    `json:"owner"`
	IsMap        bool        `json:"ismap"`
	IsArray      bool        `json:"isarray"`
	IsLiteral    bool        `json:"isliteral"`
	LiteralValue interface{} `json:"litteralValue"`
}

type Edge struct {
	From         NodeID `json:"from"`
	To           NodeID `json:"to"`
	Label        string `json:"label"`
	LSEQPosition []int  `json:"lseqposition"` // LSEQ position
}

type TreeCRDT struct {
	Root  *Node
	Nodes map[NodeID]*Node
}

func (c *TreeCRDT) CreateAttachedNode(name string, isArray bool, parentID NodeID, clientID ClientID) *Node {
	id := generateRandomNodeID(name)
	node := c.getOrCreateNode(id, isArray, clientID, 1)
	c.AddEdge(parentID, id, "", clientID)
	return node
}

func (c *TreeCRDT) CreateNode(name string, isArray bool, clientID ClientID) *Node {
	id := generateRandomNodeID(name)
	node := c.getOrCreateNode(id, isArray, clientID, 1)
	return node
}

func newNodeFromID(id NodeID, isArray bool, tree *TreeCRDT) *Node {
	node := &Node{
		ID:      id,
		Edges:   make([]*Edge, 0),
		IsArray: isArray,
		tree:    tree,
	}

	return node
}

func (c *TreeCRDT) getOrCreateNode(id NodeID, isArray bool, clientID ClientID, version int) *Node {
	if _, ok := c.Nodes[id]; !ok {
		node := newNodeFromID(id, isArray, c)
		c.Nodes[id] = node
		node.Clock = make(VectorClock)
		node.Clock[clientID] = version
		node.Owner = clientID
	}
	return c.Nodes[id]
}

func (c *TreeCRDT) GetNode(id NodeID) (*Node, bool) {
	node, ok := c.Nodes[id]
	if !ok {
		return nil, false
	}
	return node, true
}

func NewTreeCRDT() *TreeCRDT {
	rootID := "root"
	root := &Node{
		ID:      NodeID(rootID),
		Edges:   make([]*Edge, 0),
		IsArray: false,
	}
	c := &TreeCRDT{
		Root:  root,
		Nodes: make(map[NodeID]*Node),
	}
	c.Nodes[c.Root.ID] = c.Root
	root.tree = c

	return c
}

func generateRandomNodeID(label string) NodeID {
	id := core.GenerateRandomID()
	id = label + "-" + id
	return NodeID(id)
}

// This functions only appends a new node to the tree, no need for conflict resolution
func (n *Node) CreateMapNode(clientID ClientID) (*Node, error) {
	mapNode := n.tree.CreateNode("map", false, clientID)
	mapNode.IsMap = true
	mapNode.IsArray = false
	if err := n.tree.AddEdge(n.ID, mapNode.ID, "", clientID); err != nil {
		return nil, fmt.Errorf("SetKeyValue: failed to attach map node: %w", err)
	}

	return mapNode, nil
}

func (n *Node) GetNodeForKey(key string) (*Node, bool, error) {
	if !n.IsMap {
		return nil, false, fmt.Errorf("GetKeyValue: node %s is not a map node", n.ID)
	}

	// Search for the key in the edges
	for _, edge := range n.Edges {
		if edge.Label == key {
			valueNodeID := edge.To
			valueNode, exists := n.tree.Nodes[valueNodeID]
			if !exists {
				return nil, false, fmt.Errorf("GetKeyValue: missing node %s", valueNodeID)
			}
			return valueNode, true, nil
		}
	}
	return nil, false, nil
}

func (n *Node) SetKeyValue(key string, value interface{}, clientID ClientID) (NodeID, error) {
	if !n.IsMap {
		return "", fmt.Errorf("SetKeyValue: node %s is not a map node", n.ID)
	}

	// Check if key already exists
	for _, edge := range n.Edges {
		if edge.Label == key {
			valueNodeID := edge.To
			valueNode, exists := n.tree.Nodes[valueNodeID]
			if !exists {
				return "", fmt.Errorf("SetKeyValue: missing node %s", valueNodeID)
			}
			maxVersion := 0
			for _, v := range valueNode.Clock {
				if v > maxVersion {
					maxVersion = v
				}
			}
			version := maxVersion + 1

			err := valueNode.setLiteralWithVersion(value, clientID, version)
			if err != nil {
				log.WithFields(log.Fields{
					"NodeID":         valueNodeID,
					"AttemptedValue": value,
					"ClientID":       clientID,
					"Error":          err,
				}).Error("SetLiteral failed")
			}

			return valueNodeID, err
		}
	}

	// Create new value node
	valueNodeID := generateRandomNodeID("val")
	valueNode := n.tree.getOrCreateNode(valueNodeID, false, clientID, 1)
	if err := valueNode.setLiteralWithVersion(value, clientID, 1); err != nil {
		return "", err
	}

	// Link to map node with key label
	if err := n.tree.AddEdge(n.ID, valueNodeID, key, clientID); err != nil {
		return "", err
	}

	return valueNodeID, nil
}

func (n *Node) RemoveKeyValue(key string, clientID ClientID) error {
	if !n.IsMap {
		return fmt.Errorf("RemoveKeyValue: node %s is not a map node", n.ID)
	}

	for _, edge := range n.Edges {
		if edge.Label == key {
			// Simply unlink the key node by removing the edge
			return n.tree.RemoveEdge(n.ID, edge.To, clientID)
		}
	}

	return fmt.Errorf("RemoveKeyValue: key %s not found", key)
}

func (c *TreeCRDT) addEdgeWithVersion(from, to NodeID, label string, clientID ClientID, newVersion int) error {
	node, ok := c.Nodes[from]
	if !ok {
		return errors.New("Cannot add edge, node not found: " + string(from))
	}

	// Prepare the new clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

	// Resolve clock conflict
	winningClock, winningOwner := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) && (clientID == winningOwner) {
		edge := &Edge{From: from, To: to, Label: label, LSEQPosition: make([]int, 0)}
		node.Edges = append(node.Edges, edge)
		node.Clock = newClock
		node.Owner = clientID

		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge added")
	} else {
		log.WithFields(log.Fields{"NodeID": from, "To": to, "Label": label, "Version": newVersion}).Debug("Edge add ignored due to conflict")
	}

	return nil
}

func (c *TreeCRDT) AddEdge(from, to NodeID, label string, clientID ClientID) error {
	if c.validAttachment(from, to) != nil {
		return fmt.Errorf("Adding edge would create a cycle: %s -> %s or multiple parents", from, to)
	}

	node, ok := c.Nodes[from]
	if !ok {
		return errors.New("Cannot add edge, node not found: " + string(from))
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return c.addEdgeWithVersion(from, to, label, clientID, newVersion)
}

func (c *TreeCRDT) AppendEdge(from, to NodeID, label string, clientID ClientID) error {
	return c.appendEdge(from, to, label, clientID, false)
}

func (c *TreeCRDT) appendEdge(from, to NodeID, label string, clientID ClientID, ignoreConflicts bool) error {
	if c.validAttachment(from, to) != nil {
		return fmt.Errorf("Adding edge would create a cycle: %s -> %s or multiple parents", from, to)
	}

	node, ok := c.Nodes[from]
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
	return c.insertEdgeWithVersion(from, to, label, lastSibling, false, clientID, newVersion)
}

func (c *TreeCRDT) PrependEdge(from, to NodeID, label string, clientID ClientID) error {
	if c.validAttachment(from, to) != nil {
		return fmt.Errorf("Adding edge would create a cycle: %s -> %s or multiple parents", from, to)
	}

	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("PrependEdge: parent node %s not found", from)
	}

	var firstSibling NodeID
	if len(node.Edges) > 0 {
		// Use the first edge as anchor for left-side insert
		first := node.Edges[0]
		firstSibling = first.To
	} else {
		// No siblings yet, insert at the beginning
		firstSibling = ""
	}

	newVersion := node.Clock[clientID] + 1
	return c.insertEdgeWithVersion(from, to, label, firstSibling, true /* left */, clientID, newVersion)
}

func (c *TreeCRDT) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, clientID ClientID) error {
	if c.validAttachment(from, to) != nil {
		return fmt.Errorf("Adding edge would create a cycle: %s -> %s or multiple parents", from, to)
	}

	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("InsertEdge: parent node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return c.insertEdgeWithVersion(from, to, label, sibling, true, clientID, newVersion)
}

func (c *TreeCRDT) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, clientID ClientID) error {
	if c.validAttachment(from, to) != nil {
		return fmt.Errorf("Adding edge would create a cycle: %s -> %s or multiple parents", from, to)
	}

	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("InsertEdge: parent node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return c.insertEdgeWithVersion(from, to, label, sibling, false, clientID, newVersion)
}

func (c *TreeCRDT) insertEdgeWithVersion(from, to NodeID, label string, sibling NodeID, left bool, clientID ClientID, newVersion int) error {
	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("insertWithVersion: parent node %s not found", from)
	}

	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

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

func (c *TreeCRDT) GetSibling(parentNodeID NodeID, index int) (*Node, error) {
	node, ok := c.Nodes[parentNodeID]
	if !ok {
		return nil, fmt.Errorf("Cannot find node: %s", parentNodeID)
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
	sibling, exists := c.Nodes[siblingID]
	if !exists {
		return nil, fmt.Errorf("Sibling node %s not found in CRDT tree", siblingID)
	}

	return sibling, nil
}

func (c *TreeCRDT) removeEdgeWithVersion(from, to NodeID, clientID ClientID, newVersion int, ignoreConflicts bool) error {
	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("Cannot remove edge, node %s not found", from)
	}

	// Prepare the new clock
	newClock := copyClock(node.Clock)
	newClock[clientID] = newVersion

	// Resolve clock conflict
	winningClock, _ := resolveConflict(node.Clock, newClock, node.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) || ignoreConflicts {
		// New clock wins -> allow edge removal
		newEdges := []*Edge{}
		for _, edge := range node.Edges {
			if !(edge.To == to) {
				newEdges = append(newEdges, edge)
			}
		}
		node.Edges = newEdges
		node.Clock = newClock
		node.Owner = clientID

		log.WithFields(log.Fields{
			"NodeID":  from,
			"To":      to,
			"Version": newVersion}).Debug("Edge removed")
	} else {
		log.WithFields(log.Fields{
			"NodeID":    from,
			"To":        to,
			"NodeClock": node.Clock,
			"NewClock":  newClock,
			"Version":   newVersion}).Error("Edge remove ignored due to conflict")
		return fmt.Errorf("Cannot remove edge, conflict detected: %s", from)
	}

	return nil
}

func (c *TreeCRDT) RemoveEdge(from, to NodeID, clientID ClientID) error {
	node, ok := c.Nodes[from]
	if !ok {
		return fmt.Errorf("Cannot remove edge, node %s not found", from)
	}
	latestVersion := node.Clock[clientID]
	newVersion := latestVersion + 1

	return c.removeEdgeWithVersion(from, to, clientID, newVersion, false)
}

func (n *Node) GetLiteral() (interface{}, error) {
	if !n.IsLiteral {
		return nil, fmt.Errorf("GetLiteral: node %s is not a literal", n.ID)
	}
	return n.LiteralValue, nil
}

func (n *Node) SetLiteral(value interface{}, clientID ClientID) error {
	// Find max version for this client
	maxVersion := 0
	for _, v := range n.Clock {
		if v > maxVersion {
			maxVersion = v
		}
	}
	version := maxVersion + 1

	return n.setLiteralWithVersion(value, clientID, version)
}

func (n *Node) setLiteralWithVersion(value interface{}, clientID ClientID, version int) error {
	value = normalizeNumber(value) // If value is a number, normalize it to float64 since JS uses float64 for all numbers
	currentClock := n.Clock
	newClock := make(VectorClock)
	newClock[clientID] = version

	winningClock, winningOwner := resolveConflict(currentClock, newClock, n.Owner, clientID, false)

	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
		n.IsLiteral = true
		n.LiteralValue = value
		n.Clock = newClock
		n.Owner = clientID
		log.WithFields(log.Fields{
			"NodeID":       n.ID,
			"NodeClock":    currentClock,
			"NewClock":     newClock,
			"WinningClock": winningClock,
			"WinningOwner": winningOwner,
			"ClientID":     clientID,
			"LiteralValue": value}).Debug("Set literal value")
	} else {
		log.WithFields(log.Fields{"NodeID": n.ID,
			"AttemptedLiteralValue": value,
			"ClientID":              clientID,
			"NodeClock":             currentClock,
			"NewClock":              newClock,
			"WinningClock":          winningClock,
			"ExistingOwner":         n.Owner,
			"WinningOwner":          winningOwner}).Debug("Literal set ignored due to conflict")
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
//     when the CRDT tree is known to be stable.
//   - Optionally call Tidy() periodically (e.g., background maintenance) or before persisting to disk.
//
// This helps keep the CRDT tree compact without risking consistency.
func (c *TreeCRDT) Tidy() {
	referenced := make(map[NodeID]bool)

	// Mark all referenced nodes (target of edges)
	for _, node := range c.Nodes {
		for _, edge := range node.Edges {
			referenced[edge.To] = true
		}
	}

	// Always preserve the root node
	referenced[c.Root.ID] = true

	// Now delete all nodes that are unreferenced
	for id := range c.Nodes {
		if !referenced[id] {
			delete(c.Nodes, id)
			log.WithFields(log.Fields{"NodeID": id}).Debug("Purged unreferenced node")
		}
	}
}

func (c *TreeCRDT) Merge(c2 *TreeCRDT) {
	promotions := make(map[NodeID]NodeID) // fromNodeID -> arrayNodeID

	for id, remote := range c2.Nodes {
		local, exists := c.Nodes[id]
		if !exists {
			cloned := newNodeFromID(id, remote.IsArray, c)
			cloned.IsLiteral = remote.IsLiteral
			cloned.IsMap = remote.IsMap
			cloned.IsArray = remote.IsArray
			cloned.LiteralValue = remote.LiteralValue
			cloned.Clock = copyClock(remote.Clock)
			cloned.Owner = remote.Owner
			c.Nodes[id] = cloned
			local = cloned
		}

		mergedClock := mergeClocks(local.Clock, remote.Clock)
		mergedOwner := lowestClientID(local.Owner, remote.Owner)

		if remote.IsLiteral {
			err := local.setLiteralWithVersion(remote.LiteralValue, remote.Owner, remote.Clock[remote.Owner])
			if err != nil {
				log.WithFields(log.Fields{
					"NodeID": remote.ID,
					"Error":  err,
				}).Debug("Failed to set literal value during merge")
				continue
			}
		}

		for _, re := range remote.Edges {
			if _, exists := c.Nodes[re.From]; !exists {
				c.cloneNodeFromRemote(c2, re.From)
			}
			if _, exists := c.Nodes[re.To]; !exists {
				c.cloneNodeFromRemote(c2, re.To)
			}

			fromNode := c.Nodes[re.From]
			toNode := c.Nodes[re.To]

			if c.edgeExists(fromNode, re.To) {
				continue
			}

			// Promote to array if single child and not already array or map
			if len(fromNode.Edges) == 1 && !fromNode.IsArray && !fromNode.IsMap {
				existingEdge := fromNode.Edges[0]
				existingChild := c.Nodes[existingEdge.To]

				arrayNode := c.CreateNode("arr", true, fromNode.Owner)
				arrayNode.IsArray = true

				_ = c.AddEdge(fromNode.ID, arrayNode.ID, "", fromNode.Owner)
				_ = c.removeEdgeWithVersion(fromNode.ID, existingChild.ID, existingChild.Owner, existingChild.Clock[existingChild.Owner], true)

				// Insert both existing and new child sorted by NodeID
				children := []*Node{existingChild, toNode}
				sort.Slice(children, func(i, j int) bool {
					return children[i].ID < children[j].ID
				})
				for _, child := range children {
					_ = c.AppendEdge(arrayNode.ID, child.ID, "", fromNode.Owner)
				}

				promotions[fromNode.ID] = arrayNode.ID
				continue
			}

			if arrayNodeID, promoted := promotions[re.From]; promoted {
				// Prevent duplicate
				if c.edgeExists(c.Nodes[arrayNodeID], re.To) {
					continue
				}

				// Ensure deterministic order using NodeID
				arrayNode := c.Nodes[arrayNodeID]
				existingChildren := make([]*Edge, len(arrayNode.Edges))
				copy(existingChildren, arrayNode.Edges)
				sort.SliceStable(existingChildren, func(i, j int) bool {
					return existingChildren[i].To < existingChildren[j].To
				})

				inserted := false
				for i, edge := range existingChildren {
					if re.To < edge.To {
						var leftSiblingID NodeID
						if i > 0 {
							leftSiblingID = existingChildren[i-1].To
							_ = c.InsertEdgeRight(arrayNodeID, re.To, re.Label, leftSiblingID, remote.Owner)
						} else {
							_ = c.PrependEdge(arrayNodeID, re.To, re.Label, remote.Owner)
						}
						inserted = true
						break
					}
				}
				if !inserted {
					_ = c.AppendEdge(arrayNodeID, re.To, re.Label, remote.Owner)
				}
				continue
			}

			if fromNode.IsArray {
				// Sort remote parent's edges to find left sibling
				remoteParent := c2.Nodes[re.From]
				sortEdgesByLSEQ(remoteParent.Edges)

				var siblingID NodeID
				var sibling *Node = nil

				for i, edge := range remoteParent.Edges {
					if edge.To == re.To && i > 0 {
						siblingID = remoteParent.Edges[i-1].To
						break
					}
				}

				if siblingID != "" {
					var exists bool
					sibling, exists = c.Nodes[siblingID]
					if !exists {
						sibling = nil
					}
				}

				if sibling == nil {
					log.WithFields(log.Fields{
						"From":     re.From,
						"To":       re.To,
						"Label":    re.Label,
						"ClientID": remote.Owner,
					}).Debug("Appending edge to array (no left sibling found in local CRDT tree)")
					err := c.PrependEdge(re.From, re.To, re.Label, remote.Owner)
					if err != nil {
						log.WithFields(log.Fields{
							"NodeID": re.From,
							"To":     re.To,
							"Label":  re.Label,
							"Error":  err,
						}).Error("AppendEdge failed")
					}
				} else {
					log.WithFields(log.Fields{
						"From":      re.From,
						"To":        re.To,
						"Label":     re.Label,
						"SiblingID": sibling.ID,
						"ClientID":  remote.Owner,
					}).Debug("Inserting edge to array (right of sibling from remote CRDT tree)")
					err := c.InsertEdgeRight(re.From, re.To, re.Label, sibling.ID, remote.Owner)
					if err != nil {
						log.WithFields(log.Fields{
							"NodeID": re.From,
							"To":     re.To,
							"Label":  re.Label,
							"Error":  err,
						}).Error("InsertEdgeLeft failed")
					}
				}

			} else {
				if !c.edgeExists(fromNode, re.To) {
					version := fromNode.Clock[remote.Owner] + 1
					err := c.addEdgeWithVersion(fromNode.ID, re.To, re.Label, remote.Owner, version)
					if err != nil {
						log.WithFields(log.Fields{
							"NodeID": re.From,
							"To":     re.To,
							"Label":  re.Label,
							"Error":  err,
						}).Error("AddEdgeWithVersion failed")
						continue
					}
				} else {
					log.WithFields(log.Fields{
						"From":     re.From,
						"To":       re.To,
						"Label":    re.Label,
						"ClientID": remote.Owner,
					}).Debug("Edge already exists, skipping")
					continue
				}
				_ = c.AddEdge(fromNode.ID, re.To, re.Label, remote.Owner)
			}
		}

		local.Clock = mergedClock
		local.Owner = mergedOwner
	}

	c.normalize()
}

func (c *TreeCRDT) cloneNodeFromRemote(c2 *TreeCRDT, id NodeID) {
	remote := c2.Nodes[id]
	cloned := newNodeFromID(id, remote.IsArray, c)
	cloned.IsLiteral = remote.IsLiteral
	cloned.IsMap = remote.IsMap
	cloned.IsArray = remote.IsArray
	cloned.LiteralValue = remote.LiteralValue
	cloned.Clock = copyClock(remote.Clock)
	cloned.Owner = remote.Owner
	c.Nodes[id] = cloned
}

func (c *TreeCRDT) edgeExists(node *Node, to NodeID) bool {
	for _, e := range node.Edges {
		if e.To == to {
			return true
		}
	}
	return false
}

func cloneNodeWithoutEdges(n *Node, crdt *TreeCRDT) *Node {
	cloned := newNodeFromID(n.ID, n.IsArray, crdt)
	cloned.IsLiteral = n.IsLiteral
	cloned.LiteralValue = n.LiteralValue
	cloned.Clock = copyClock(n.Clock)
	cloned.Owner = n.Owner
	return cloned
}

func (c *TreeCRDT) normalize() {
	log.Debug("Normalizing CRDT tree")
	sortEdgesByLSEQ(c.Root.Edges)
	for _, node := range c.Nodes {
		sortEdgesByLSEQ(node.Edges)
	}
}

func (c *TreeCRDT) validAttachment(from, to NodeID) error {
	if from == to {
		return fmt.Errorf("cannot attach node %s to itself", from)
	}

	// 1. Check for cycle
	visited := make(map[NodeID]bool)
	var dfs func(NodeID) bool
	dfs = func(id NodeID) bool {
		if id == from {
			return true
		}
		visited[id] = true
		node := c.Nodes[id]
		for _, edge := range node.Edges {
			if !visited[edge.To] && dfs(edge.To) {
				return true
			}
		}
		return false
	}
	if dfs(to) {
		return fmt.Errorf("adding edge from %s to %s would create a cycle", from, to)
	}

	// 2. Check if `to` already has a parent
	for _, parent := range c.Nodes {
		for _, edge := range parent.Edges {
			if edge.To == to {
				return fmt.Errorf("node %s already has a parent", to)
			}
		}
	}

	return nil
}

func (c *TreeCRDT) ValidateTree() error {
	parentMap := make(map[NodeID]NodeID)
	visited := make(map[NodeID]bool)

	var dfs func(current NodeID, ancestors map[NodeID]bool) error
	dfs = func(current NodeID, ancestors map[NodeID]bool) error {
		if ancestors[current] {
			log.WithField("NodeID", current).Debug("Cycle detected")
			return fmt.Errorf("Cycle detected at node %s", current)
		}
		if visited[current] {
			return nil // Already validated
		}
		visited[current] = true

		node, exists := c.Nodes[current]
		if !exists {
			log.WithField("NodeID", current).Debug("Node not found in tree")
			return fmt.Errorf("Node %s not found in tree", current)
		}

		ancestors[current] = true
		for _, edge := range node.Edges {
			childID := edge.To

			if _, ok := c.Nodes[childID]; !ok {
				log.WithField("ChildID", childID).Debug("Edge to non-existent node")
				return fmt.Errorf("Edge to non-existent node: %s", childID)
			}

			if existingParent, ok := parentMap[childID]; ok && existingParent != current {
				log.WithFields(log.Fields{
					"ChildID":        childID,
					"ExistingParent": existingParent,
					"CurrentParent":  current,
				}).Debug("Multiple parents detected")

				return fmt.Errorf("Node %s has multiple parents: %s and %s", childID, existingParent, current)
			}
			parentMap[childID] = current

			if err := dfs(childID, ancestors); err != nil {
				return err
			}
		}
		delete(ancestors, current)
		return nil
	}

	// Begin DFS from root
	if err := dfs(c.Root.ID, make(map[NodeID]bool)); err != nil {
		return err
	}

	// Ensure all nodes are reachable
	for id := range c.Nodes {
		if !visited[id] {
			log.WithField("NodeID", id).Debug("Unreachable node detected")
			// Check if the node is not the root
			if id != c.Root.ID {
				log.WithFields(log.Fields{
					"NodeID": id,
					"Reason": "Unreachable node",
				}).Debug("Unreachable node detected")

				return fmt.Errorf("Unreachable node found: %s", id)
			}
		}
	}

	return nil
}
