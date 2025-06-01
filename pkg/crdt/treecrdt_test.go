package crdt

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsDescendant(t *testing.T) {
	tree := newTreeCRDT()
	client := ClientID("test-client")

	// Build structure:
	// root
	// ├── A
	// │   └── B
	// │       └── C
	// └── X
	//     └── Y
	nodeA := tree.CreateAttachedNode("A", Map, tree.Root.ID, client)
	nodeB := tree.CreateAttachedNode("B", Map, nodeA.ID, client)
	nodeC := tree.CreateAttachedNode("C", Map, nodeB.ID, client)

	nodeX := tree.CreateAttachedNode("X", Map, tree.Root.ID, client)
	nodeY := tree.CreateAttachedNode("Y", Map, nodeX.ID, client)

	tests := []struct {
		name     string
		root     NodeID
		target   NodeID
		expected bool
	}{
		{"C is descendant of root", tree.Root.ID, nodeC.ID, true},
		{"B is descendant of A", nodeA.ID, nodeB.ID, true},
		{"A is not descendant of C", nodeC.ID, nodeA.ID, false},
		{"root is descendant of root", tree.Root.ID, tree.Root.ID, true},
		{"unrelated (B is not descendant of Y)", nodeY.ID, nodeB.ID, false},
		{"Y is descendant of X", nodeX.ID, nodeY.ID, true},
		{"X is not descendant of A", nodeA.ID, nodeX.ID, false},
		{"C is not under X", nodeX.ID, nodeC.ID, false},
		{"node not in tree", nodeC.ID, "missing-node", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := tree.isDescendant(test.root, test.target)
			if result != test.expected {
				t.Errorf("IsDescendant(%s, %s) = %v; want %v", test.root, test.target, result, test.expected)
			}
		})
	}
}

func TestTreeCRDTSetFieldArrays(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	json := []byte(`{
	  "a": [
	    {
	      "2": "3"
	    }
	  ]
	}`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(json, clientID)
	assert.NoError(t, err)
}

func TestTreeCRDTSetFieldsConflictLastWriterWins(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)

	c1 := newTreeCRDT()

	clientID1 := ClientID(core.GenerateRandomID())
	clientID2 := ClientID(core.GenerateRandomID())

	rootC1 := c1.Root

	mapNodeC1, err := rootC1.CreateMapNode(clientID1)
	assert.NoError(t, err, "CreateMapNode should not return an error")

	_, err = mapNodeC1.SetKeyValue("key", "value1", clientID1)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	c2, err := c1.Clone()

	assert.NoError(t, err, "Clone should not return an error")
	mapNodeC2, ok := c2.GetNode(mapNodeC1.ID)
	assert.True(t, ok, "Node should exist in cloned graph")

	_, err = mapNodeC2.SetKeyValue("key", "value2", clientID2)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	_, err = mapNodeC2.SetKeyValue("key", "value3", clientID2)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	_, err = mapNodeC1.SetKeyValue("key", "value4", clientID1) // Will be overwritten by c2
	assert.NoError(t, err, "SetKeyValue should not return an error")

	err = c1.Merge(c2)
	assert.NoError(t, err, "Merge should not return an error")

	exportedJSON, err := c1.ExportJSON()
	assert.NoError(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`{"key":"value3"}`)
	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTSetFieldsConflictNodeIDTieBraker(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)

	c1 := newTreeCRDT()

	clientID1 := ClientID(core.GenerateRandomID())
	clientID2 := ClientID(core.GenerateRandomID())

	rootC1 := c1.Root

	mapNodeC1, err := rootC1.CreateMapNode(clientID1)
	assert.NoError(t, err, "CreateMapNode should not return an error")

	_, err = mapNodeC1.SetKeyValue("key", "value1", clientID1)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	c2, err := c1.Clone()

	assert.NoError(t, err, "Clone should not return an error")
	mapNodeC2, ok := c2.GetNode(mapNodeC1.ID)
	assert.True(t, ok, "Node should exist in cloned graph")

	_, err = mapNodeC2.SetKeyValue("key", "value2", clientID2)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	// Conflict, both clients have the same vector clock version
	_, err = mapNodeC1.SetKeyValue("key", "value3", clientID1)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	err = c1.Merge(c2) // Enable conflict resolution with tie-breaker
	assert.NoError(t, err, "Merge should not return an error")

	exportedJSON, err := c1.ExportJSON()
	assert.NoError(t, err, "ExportToJSON should not return an error")

	// This test will result in a conflict resolution where client IDs will be used as tie-breakers.
	if clientID1 < clientID2 {
		expectedJSON := []byte(`{"key":"value3"}`)
		compareJSON(t, expectedJSON, exportedJSON)
	} else {
		expectedJSON := []byte(`{"key":"value2"}`)
		compareJSON(t, expectedJSON, exportedJSON)
	}
}

func TestTreeCRDTNodeRemoveField(t *testing.T) {
	c := newTreeCRDT()

	clientID := ClientID(core.GenerateRandomID())

	mapNode, err := c.Root.CreateMapNode(clientID)
	assert.NoError(t, err, "CreateMapNode should not return an error")

	_, err = mapNode.SetKeyValue("key1", "value1", clientID)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	_, err = mapNode.SetKeyValue("key2", "value1", clientID)
	assert.NoError(t, err, "SetKeyValue should not return an error")

	valueNode, found, err := mapNode.GetNodeForKey("key1")
	assert.NoError(t, err, "GetValueNode should not return an error")
	assert.NotNil(t, valueNode, "Value node for key1 should exist")
	assert.True(t, found, "Value node for key1 should be found")

	valueNode, found, err = mapNode.GetNodeForKey("key2")
	assert.NoError(t, err, "GetValueNode should not return an error")
	assert.NotNil(t, valueNode, "Value node for key2 should exist")
	assert.True(t, found, "Value node for key2 should be found")

	// Remove key1
	err = mapNode.RemoveKeyValue("key1", clientID)
	assert.NoError(t, err, "RemoveKeyValue should not return an error")

	// Check if key1 is removed
	valueNode, found, err = mapNode.GetNodeForKey("key1")
	assert.NoError(t, err, "GetValueNode should not return an error")
	assert.Nil(t, valueNode, "Value node for key1 should be nil after removal")
	assert.False(t, found, "Value node for key1 should not be found after removal")

	// Check if key2 still exists
	valueNode, found, err = mapNode.GetNodeForKey("key2")
	assert.NoError(t, err, "GetValueNode should not return an error")
	assert.NotNil(t, valueNode, "Value node for key2 should still exist")
	assert.True(t, found, "Value node for key2 should still be found after removal of key1")
}

func TestTreeCRDTAddEdgeWithVersion(t *testing.T) {
	c := newTreeCRDT()

	// To make the test deterministic, we will use fixed client IDs
	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := c.CreateAttachedNode("parent", Map, c.Root.ID, clientID)
	child := c.CreateAttachedNode("child", Map, c.Root.ID, clientID)

	// 1. Add an edge with version 1
	err := c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "AddEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge")
	assert.Equal(t, child.ID, parent.Edges[0].To, "Edge should point to child")
	assert.Equal(t, "link", parent.Edges[0].Label, "Edge label mismatch")

	// 2. Add another edge with higher version (should succeed)
	anotherChild := c.CreateAttachedNode("another_child", Map, c.Root.ID, clientID)
	err = c.addEdgeWithVersion(parent.ID, anotherChild.ID, "link2", clientID, 2)
	assert.Nil(t, err, "AddEdgeWithVersion second time should not return error")

	assert.Equal(t, 2, len(parent.Edges), "Expected 2 edges now")

	// 3. Try to add conflicting edge with lower version (should be ignored)
	fakeChild := c.CreateAttachedNode("fake_child", Map, c.Root.ID, clientID)
	err = c.addEdgeWithVersion(parent.ID, fakeChild.ID, "fake_link", clientID, 1) // lower version
	assert.Nil(t, err, "AddEdgeWithVersion with lower version should not error")

	found := false
	for _, edge := range parent.Edges {
		if edge.To == fakeChild.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Edge with lower version should not overwrite or add")

	// 4. Simulate a tie with another client (new client id)
	tieChild := c.CreateAttachedNode("tie_child", Map, c.Root.ID, otherClientID)
	err = c.addEdgeWithVersion(parent.ID, tieChild.ID, "tie_link", otherClientID, 2) // same version
	assert.Nil(t, err, "AddEdgeWithVersion with same version different client should not error")

	if otherClientID < clientID {
		assert.Equal(t, 3, len(parent.Edges), "Tie-breaker: new client wins")
	} else {
		assert.Equal(t, 2, len(parent.Edges), "Tie-breaker: original client keeps ownership")
	}
}

func TestTreeCRDTRemoveEdgeWithVersion(t *testing.T) {
	c := newTreeCRDT()

	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := c.CreateAttachedNode("parent", Map, c.Root.ID, clientID)
	child := c.CreateAttachedNode("child", Map, c.Root.ID, clientID)

	// Add an edge
	err := c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "addEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge before removal")

	// Remove the edge with higher version (should succeed)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2, false)
	assert.Nil(t, err, "removeEdgeWithVersion should not return error")
	assert.Equal(t, 0, len(parent.Edges), "Expected 0 edges after removal")

	// Re-add it for conflict test
	_ = c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 3)

	// Try to remove with lower version (should be ignored)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2, false)
	assert.NotNil(t, err, "removeEdgeWithVersion with lower version should error")
	assert.Equal(t, 1, len(parent.Edges), "Edge should still exist after invalid removal")

	// Tie-break with other client (lower client ID wins)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, otherClientID, 3, false)
	assert.Nil(t, err, "removeEdgeWithVersion tie-break should not error")

	if otherClientID < clientID {
		assert.Equal(t, 0, len(parent.Edges), "Tie-break: other client removed the edge")
	} else {
		assert.Equal(t, 1, len(parent.Edges), "Tie-break: original client kept the edge")
	}
}

func TestTreeCRDTRemoveIndexInArray(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "C"]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	// Find the node with ID "B"
	arrNodeID := c.Root.Edges[0].To
	arrNode, ok := c.GetNode(arrNodeID)
	assert.True(t, ok, "Array node should exist")
	edges := arrNode.Edges
	for _, edge := range edges {
		node, ok := c.GetNode(edge.To)
		assert.True(t, ok, "Node should exist in the array")
		if node.LiteralValue.(string) == "B" {
			// Remove the edge with ID "B"
			err = c.RemoveEdge(arrNodeID, node.ID, clientID)
			assert.Nil(t, err, "removeEdgeWithVersion should not return an error")
			break
		}
	}

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	// Correct expected JSON
	expectedJSON := []byte(`[
		"A",
		"C"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTTidy(t *testing.T) {
	c := newTreeCRDT()

	clientID := ClientID("client")

	c.CreateAttachedNode("parent", Map, c.Root.ID, clientID)
	c.CreateAttachedNode("child", Map, c.Root.ID, clientID)

	// Create an orphan node manually (NOT attached)
	orphanID := generateRandomNodeID("orphan")
	orphan := c.getOrCreateNode(orphanID, Map, clientID, 1)

	assert.Equal(t, 4, len(c.Nodes), "Expected 4 nodes before purge (root, parent, child, orphan)")

	c.Tidy() // Remove orphan nodes

	// Should only have root, parent, and child left
	_, orphanExists := c.Nodes[orphan.ID]
	assert.False(t, orphanExists, "Orphan should have been purged")
	assert.Equal(t, 3, len(c.Nodes), "Expected 3 nodes after purge (root, parent, child)")
}

func TestTreeCRDTNodeSetLiteral(t *testing.T) {
	c := newTreeCRDT()

	clientID1 := ClientID("client1")
	clientID2 := ClientID("client2")

	node := c.CreateAttachedNode("literalNode", Literal, c.Root.ID, clientID1)

	// 1. Set an initial literal value
	node.setLiteralWithVersion("hello", clientID1, 1)

	assert.True(t, node.IsLiteral, "Expected node to be marked as literal")
	assert.Equal(t, "hello", node.LiteralValue, "Expected literal value to be 'hello'")

	// 2. Set a higher version value (should overwrite)
	node.setLiteralWithVersion("world", clientID1, 2)

	assert.Equal(t, "world", node.LiteralValue, "Expected literal value to be updated to 'world'")

	// 3. Attempt to set with a lower version (should be ignored)
	node.setLiteralWithVersion("ignored", clientID1, 1)

	assert.Equal(t, "world", node.LiteralValue, "Lower version should not overwrite the value")

	// 4. Simulate conflict: different client, same version
	node.setLiteralWithVersion("conflict", clientID2, 2)

	// Resolve which client should win
	expectedWinner := clientID1
	if clientID2 < clientID1 {
		expectedWinner = clientID2
	}

	expectedValue := "world"
	if expectedWinner == clientID2 {
		expectedValue = "conflict"
	}

	assert.Equal(t, expectedWinner, node.Owner, fmt.Sprintf("Expected owner %s to win tie-breaker, got %s", expectedWinner, node.Owner))
	assert.Equal(t, expectedValue, node.LiteralValue, fmt.Sprintf("Expected literal value %s after conflict resolution, got %s", expectedValue, node.LiteralValue))
}

func TestTreeCRDTValidation(t *testing.T) {
	client := ClientID("clientA")

	t.Run("Valid tree passes validation", func(t *testing.T) {
		c := newTreeCRDT()
		nodeA := c.CreateAttachedNode("A", Map, c.Root.ID, client)
		nodeB := c.CreateAttachedNode("B", Map, nodeA.ID, client)
		c.CreateAttachedNode("C", Map, nodeB.ID, client)

		err := c.ValidateTree()
		assert.NoError(t, err, "Valid tree structure should pass validation")
	})

	t.Run("Multiple parents detected", func(t *testing.T) {
		c := newTreeCRDT()
		nodeA := c.CreateAttachedNode("A", Map, c.Root.ID, client)
		nodeB := c.CreateAttachedNode("B", Map, nodeA.ID, client)
		nodeC := c.CreateAttachedNode("C", Map, nodeB.ID, client)

		// Add second parent (invalid) through API
		err := c.AddEdge(nodeA.ID, nodeC.ID, "", client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multiple parents")

		// Simulate corruption manually
		c.Nodes[nodeA.ID].Edges = append(c.Nodes[nodeA.ID].Edges, &EdgeCRDT{
			From:         nodeA.ID,
			To:           nodeC.ID,
			Label:        "",
			LSEQPosition: []int{42},
		})

		err = c.ValidateTree()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "multiple parents")
	})

	t.Run("Cycle detection", func(t *testing.T) {
		c := newTreeCRDT()
		nodeA := c.CreateAttachedNode("A", Map, c.Root.ID, client)
		nodeB := c.CreateAttachedNode("B", Map, nodeA.ID, client)
		nodeC := c.CreateAttachedNode("C", Map, nodeB.ID, client)

		// Create a cycle: C -> A
		c.Nodes[nodeC.ID].Edges = append(c.Nodes[nodeC.ID].Edges, &EdgeCRDT{
			From:         nodeC.ID,
			To:           nodeA.ID,
			Label:        "",
			LSEQPosition: []int{99},
		})

		err := c.validAttachment(nodeC.ID, nodeA.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "would create a cycle")

		err = c.ValidateTree()
		assert.Error(t, err)
	})

	t.Run("Literal node with children fails validation", func(t *testing.T) {
		c := newTreeCRDT()
		lit := c.CreateAttachedNode("Literal", Literal, c.Root.ID, client)
		c.CreateAttachedNode("Child", Map, lit.ID, client)

		err := c.ValidateTree()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must not have children")
	})

	t.Run("Node with multiple types fails validation", func(t *testing.T) {
		c := newTreeCRDT()
		node := c.CreateAttachedNode("BadNode", Map, c.Root.ID, client)
		node.IsArray = true // Invalid: now both map and array

		err := c.ValidateTree()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have exactly one type")
	})

	t.Run("Unreachable node fails validation", func(t *testing.T) {
		c := newTreeCRDT()
		_ = c.CreateAttachedNode("A", Map, c.Root.ID, client)

		// Add isolated node
		isolated := &NodeCRDT{
			ID:        NodeID("isolated"),
			IsMap:     true,
			IsRoot:    false,
			Owner:     client,
			tree:      c,
			Clock:     VectorClock{},
			Nounce:    "iso",
			Signature: "sig",
		}
		c.Nodes[isolated.ID] = isolated

		err := c.ValidateTree()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unreachable node")
	})
}

// // Test case:
// // 1. Create two graphs with shared nodes
// // 2. Set different literal values on the same node in both graphs
// // 3. Merge the graphs
// // 4. The merged graph should be an array of literals since n1 + n2 → [n1, n2] sorted by node ID
func TestTreeCRDTMergeLitterals(t *testing.T) {
	c1 := newTreeCRDT()
	c2 := newTreeCRDT()

	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Create shared nodes in both graphs
	node1 := c1.CreateAttachedNode("sharedA", Literal, c1.Root.ID, clientA)
	node2 := c2.CreateAttachedNode("sharedB", Literal, c2.Root.ID, clientB)
	err := node1.SetLiteral("A-literal", clientA)
	assert.Nil(t, err, "SetLiteral should not return an error")
	err = node2.SetLiteral("B-literal", clientB)
	assert.Nil(t, err, "SetLiteral should not return an error")

	c1Copy, err := c1.Clone()
	c2Copy, err := c2.Clone()

	// Perform merge
	err = c1.Merge(c2)
	assert.Nil(t, err, "Merge should not return an error")
	err = c2Copy.Merge(c1Copy)
	assert.Nil(t, err, "Merge should not return an error")

	// Check that all nodes exist
	_, ok1 := c1.GetNode(node1.ID)
	_, ok2 := c1.GetNode(node2.ID)
	assert.True(t, ok1, "Node1 should exist after merge")
	assert.True(t, ok2, "Node2 should exist after merge")

	json, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	json2, err := c2Copy.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	compareJSON(t, json, json2)

	if node1.ID < node2.ID {
		expectedJSON := []byte(`["A-literal", "B-literal"]`)
		compareJSON(t, expectedJSON, json)
	} else {
		expectedJSON := []byte(`["B-literal", "A-literal"]`)
		compareJSON(t, expectedJSON, json)
	}
}

func TestTreeCRDTMergeLists(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	initialJSON := []byte(`[1, 2, 4]`)

	c1 := newTreeCRDT()
	_, err := c1.ImportJSON(initialJSON, ClientID(clientA))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	c2 := newTreeCRDT()
	c2.Load(rawJSON)
	assert.Nil(t, err, "ImportRawJSON should not return an error")

	rawJSONBefore, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	err = c1.Merge(c2)
	assert.Nil(t, err, "Merge should not return an error")

	rawJSONAfter, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	// Trees should be identical before and after merge
	assert.Equal(t, rawJSONBefore, rawJSONAfter, "Trees should be identical before and after merge")
	assert.True(t, c1.Equal(c2), "Trees should be equal after merge")

	// Let's do some modifications on the graph independently
	// Original    :    [1, 2, 4]
	// G1(A):        [0, 1, 2, 4]
	// G2(B):           [1, 2, 3, 4]
	// G1 + G2:      [0, 1, 2, 3, 4] <- 4 is added to G1, owner of root is B
	// G2 + G1:      [0, 1, 2, 3, 4] <- 0 is added to G2, owner of root is A

	// 1. Create a new node in c1
	node0 := c1.CreateNode("0", Literal, clientA)
	node0.SetLiteral(0, clientA)

	// First child is the array
	assert.Len(t, c1.Root.Edges, 1, "Root should have one edge")
	c1ArrayNodeID := c1.Root.Edges[0].To

	// Find the node with id "0"
	sibling, err := c1.GetSibling(c1ArrayNodeID, 0)
	assert.Nil(t, err, "GetSiblingNode should not return an error")
	err = c1.InsertEdgeLeft(c1ArrayNodeID, node0.ID, "", sibling.ID, clientA)
	assert.Nil(t, err, "InsertEdge should not return an error")
	// G1: [0, 1, 2, 4]  <-- 0 added

	// 2. Create a new node in c2
	node3 := c2.CreateNode("3", Literal, clientA)
	node3.SetLiteral(3, clientA)
	// node3.IsLiteral = true
	// node3.LiteralValue = 3.0

	// First child is the array
	assert.Len(t, c2.Root.Edges, 1, "Root should have one edge")
	c2ArrayNodeID := c2.Root.Edges[0].To
	sibling, err = c2.GetSibling(c2ArrayNodeID, 1)
	assert.Nil(t, err, "GetSiblingNode should not return an error")
	err = c2.InsertEdgeRight(c2ArrayNodeID, node3.ID, "", sibling.ID, clientB)
	assert.Nil(t, err, "InsertEdge should not return an error")
	//  G2: [1, 2, 3, 4]   <-- 3 added

	c1Clone, err := c1.Clone()
	assert.Nil(t, err, "Clone should not return an error")

	// set debug level to see the merge process
	logrus.SetLevel(logrus.ErrorLevel)

	// 3. Merge the graphs
	err = c1.Merge(c2)
	assert.Nil(t, err, "Merge should not return an error")
	err = c2.Merge(c1Clone)
	assert.Nil(t, err, "Merge should not return an error")

	json, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	expectedJSON := []byte(`[0, 1, 2, 3, 4]`)
	compareJSON(t, expectedJSON, json)

	json2, err := c2.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	expectedJSON2 := []byte(`[0, 1, 2, 3, 4]`)
	compareJSON(t, expectedJSON2, json2)

	// Turn on warning log
	logrus.SetLevel(logrus.WarnLevel)

	// C2 == C1
	assert.True(t, c1.Equal(c2), "Graphs should be equal after merge")
	assert.True(t, c1.Root.Owner == c2.Root.Owner, "Owners should be equal after merge")
}

func TestTreeCRDTMergeListsConflicts(t *testing.T) {
	clientA := ClientID("A")
	clientB := ClientID("B")

	initialJSON := []byte(`[2, 3, 4]`)

	c1 := newTreeCRDT()
	_, err := c1.ImportJSON(initialJSON, ClientID(clientA))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	c2, err := c1.Clone()
	assert.Nil(t, err, "Clone should not return an error")

	// C1 prepares nodes
	node := c1.CreateNode("1", Literal, clientA)
	node.IsLiteral = true
	node.LiteralValue = 1
	err = c1.PrependEdge(c1.Root.ID, node.ID, "", clientA)
	assert.Nil(t, err, "PrependEdge should not return an error")

	node = c1.CreateNode("0", Literal, clientA)
	node.IsLiteral = true
	node.LiteralValue = 0
	err = c1.PrependEdge(c1.Root.ID, node.ID, "", clientA)
	assert.Nil(t, err, "PrependEdge should not return an error")

	// C2 appends nodes
	node = c2.CreateNode("5", Literal, clientB)
	node.IsLiteral = true
	node.LiteralValue = 5
	err = c2.AppendEdge(c2.Root.ID, node.ID, "", clientB)
	assert.Nil(t, err, "AppendEdge should not return an error")

	node = c2.CreateNode("6", Literal, clientB)
	node.IsLiteral = true
	node.LiteralValue = 6
	err = c2.AppendEdge(c2.Root.ID, node.ID, "", clientB)
	assert.Nil(t, err, "AppendEdge should not return an error")

	//logrus.SetLevel(logrus.DebugLevel)

	err = c2.Merge(c1)
	assert.Nil(t, err, "Merge should not return an error")
	_, err = c2.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
}

func TestTreeCRDTMergeKVListsWithConflicts(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	initialJSON := []byte(`[
		{"id": "A", "value": "1"}
	]`)

	c1 := newTreeCRDT()
	_, err := c1.ImportJSON(initialJSON, ClientID(clientA))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	c2, err := c1.Clone()
	assert.Nil(t, err, "Clone should not return an error")

	arrNodeID := c1.Root.Edges[0].To
	arrNode, ok := c1.GetNode(arrNodeID)
	assert.True(t, ok, "Array node should exist in c1")
	mapNodeID := arrNode.Edges[0].To
	mapNode, ok := c1.GetNode(mapNodeID)
	assert.True(t, ok, "Map node should exist in c1")

	assert.True(t, ok, "Array node should exist in c1")

	_, err = mapNode.SetKeyValue("value", "11", clientA)
	_, err = mapNode.SetKeyValue("value", "22", clientA)

	json, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	arrNodeID2 := c2.Root.Edges[0].To
	arrNode2, ok := c2.GetNode(arrNodeID2)
	assert.True(t, ok, "Array node should exist in c2")
	mapNodeID2 := arrNode2.Edges[0].To
	mapNode2, ok := c2.GetNode(mapNodeID2)
	assert.True(t, ok, "Map node should exist in c2")

	// Set a different value in c2
	_, err = mapNode2.SetKeyValue("value", "33", clientB) // <- Should we overwriting, according to last writer wins policy
	assert.Nil(t, err, "SetKeyValue should not return an error")

	err = c1.Merge(c2) // Enable conflict resolution with last writer wins
	assert.Nil(t, err, "Merge should not return an error")
	err = c2.Merge(c1)
	assert.Nil(t, err, "Merge should not return an error")
	json, err = c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	json2, err := c2.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`[
	 	{"id": "A", "value": "22"}
	]`)

	compareJSON(t, expectedJSON, json)
	compareJSON(t, expectedJSON, json2)
}

func TestTreeCRDTMergeJSON1(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	json1 := []byte(`{
	  "1": [
	    {
	      "2": "3"
	    },
	    {
	      "4": [
	        {
	          "5": "6"
	        }
	      ]
	    }
	  ]
	}`)

	expectedJSON := []byte(`[
	  {
	    "1": [
	      {
	        "2": "3"
	      },
	      {
	        "4": [
	          {
	            "5": "6"
	          }
	        ]
	      }
	    ]
	  },
	  {
	    "1": [
	      {
	        "2": "3"
	      },
	      {
	        "4": [
	          {
	            "5": "6"
	          }
	        ]
	      }
	    ]
	  }
	]`)

	// The order depends on how Node IDs are generated
	expectedJSONAlt := []byte(`[
	  {
	    "1": [
	      {
	        "2": "3"
	      },
	      {
	        "4": [
	          {
	            "5": "6"
	          }
	        ]
	      }
	    ]
	  },
	  {
	    "1": [
	      {
	        "2": "3"
	      },
	      {
	        "4": [
	          {
	            "5": "6"
	          }
	        ]
	      }
	    ]
	  }
	]`)

	// Build and merge CRDTs
	c1 := newTreeCRDT()
	_, err := c1.ImportJSON(json1, clientID)
	assert.NoError(t, err)

	c2 := newTreeCRDT()
	_, err = c2.ImportJSON(json1, clientID)
	assert.NoError(t, err)

	// Since, the node IDs are generated randomly, it imported json will duplicated in an array
	err = c1.Merge(c2) // Enable conflict resolution with last writer wins
	assert.NoError(t, err, "Merge should not return an error")

	exportedJSON, err := c1.ExportJSON()
	assert.NoError(t, err)

	exportedEqualsExpected := isJSONEqual(t, exportedJSON, expectedJSON) || isJSONEqual(t, exportedJSON, expectedJSONAlt)
	assert.True(t, exportedEqualsExpected, "Exported JSON should match expected JSON")

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTMergeHelloWorld(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Step 1: Start from empty CRDT
	c1 := newTreeCRDT()
	_, err := c1.ImportJSON([]byte(`[]`), clientA)
	assert.Nil(t, err)

	// Step 2: Insert "Hello" in c1
	charsA := []string{"H", "e", "l", "l", "o"}
	parentNode := c1.Root.Edges[0].To
	var leftID NodeID
	for _, ch := range charsA {
		n := c1.CreateNode(ch, Literal, clientA)
		n.SetLiteral(ch, clientA)
		err := c1.InsertEdgeRight(parentNode, n.ID, "", leftID, clientA)
		assert.Nil(t, err)
		leftID = n.ID
	}
	lastAID := leftID

	// Step 3: Clone c1 into c2 (simulate clientB syncing)
	raw, err := c1.Save()
	assert.Nil(t, err)

	c2 := newTreeCRDT()
	err = c2.Load(raw)
	assert.Nil(t, err)

	// Step 4: Insert " world!" in c2 after last "o"
	charsB := []string{" ", "w", "o", "r", "l", "d", "!"}
	parentNode = c2.Root.Edges[0].To
	leftID = lastAID
	for _, ch := range charsB {
		n := c2.CreateNode(ch, Literal, clientB)
		n.SetLiteral(ch, clientB)
		err := c2.InsertEdgeRight(parentNode, n.ID, "", leftID, clientB)
		assert.Nil(t, err)
		leftID = n.ID
	}

	// Step 5: Merge back both ways
	err = c1.Merge(c2)
	assert.Nil(t, err, "Merge c1 with c2 should not return an error")
	err = c2.Merge(c1)
	assert.Nil(t, err, "Merge c2 with c1 should not return an error")

	// Step 6: Export and verify
	json1, err := c1.ExportJSON()
	assert.Nil(t, err)
	json2, err := c2.ExportJSON()
	assert.Nil(t, err)

	expected := []byte(`["H","e","l","l","o"," ","w","o","r","l","d","!"]`)
	compareJSON(t, expected, json1)
	compareJSON(t, expected, json2)

	assert.True(t, c1.Equal(c2), "Graphs should be equal after merge")
	assert.Equal(t, c1.Root.Owner, c2.Root.Owner, "Root owners should match after merge")
}

func TestTreeCRDTSingleTreeTwoClientsHelloWorld(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Step 1: Initialize TreeCRDT with an empty array
	tree := newTreeCRDT()
	_, err := tree.ImportJSON([]byte(`[]`), clientA)
	assert.Nil(t, err, "ImportJSON should not return an error")

	parentNode := tree.Root.Edges[0].To
	var leftID NodeID

	// Step 2: Client A inserts "Hello"
	charsA := []string{"H", "e", "l", "l", "o"}
	for _, ch := range charsA {
		n := tree.CreateNode(ch, Literal, clientA)
		n.SetLiteral(ch, clientA)
		err := tree.InsertEdgeRight(parentNode, n.ID, "", leftID, clientA)
		assert.Nil(t, err, "InsertEdgeRight (clientA) should not return an error")
		leftID = n.ID
	}

	// Step 3: Client B inserts " world!"
	charsB := []string{" ", "w", "o", "r", "l", "d", "!"}
	for _, ch := range charsB {
		n := tree.CreateNode(ch, Literal, clientB)
		n.SetLiteral(ch, clientB)
		err := tree.InsertEdgeRight(parentNode, n.ID, "", leftID, clientB)
		assert.Nil(t, err, "InsertEdgeRight (clientB) should not return an error")
		leftID = n.ID
	}

	// Step 4: Export final tree and validate JSON
	json, err := tree.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")

	expected := []byte(`["H","e","l","l","o"," ","w","o","r","l","d","!"]`)
	compareJSON(t, expected, json)
}

func TestTreeCRDTSingleTreeInterleavedClientsHelloWorld(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Step 1: Initialize shared TreeCRDT with an empty array
	tree := newTreeCRDT()
	_, err := tree.ImportJSON([]byte(`[]`), clientA)
	assert.Nil(t, err, "ImportJSON should not return an error")

	parentNode := tree.Root.Edges[0].To
	var leftID NodeID

	// Step 2: Interleave clients while inserting "Hello world!"
	chars := []string{"H", "e", "l", "l", "o", " ", "w", "o", "r", "l", "d", "!"}
	clients := []ClientID{clientA, clientB} // Alternating clients

	for i, ch := range chars {
		client := clients[i%2]
		n := tree.CreateNode(ch, Literal, client)
		n.SetLiteral(ch, client)

		err := tree.InsertEdgeRight(parentNode, n.ID, "", leftID, client)
		assert.Nil(t, err, "InsertEdgeRight (interleaved) should not return an error")
		leftID = n.ID
	}

	// Step 3: Export final document and validate JSON structure
	json, err := tree.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")

	expected := []byte(`["H","e","l","l","o"," ","w","o","r","l","d","!"]`)
	compareJSON(t, expected, json)
}

func TestTreeCRTDSync(t *testing.T) {
	clientID1 := ClientID(core.GenerateRandomID())
	clientID2 := ClientID(core.GenerateRandomID())
	clientID3 := ClientID(core.GenerateRandomID())

	json1 := []byte(`[
		{
		  "uid": "user_1",
		   "name": "Bob"
		},
		{
		   "uid": "user_2",
		   "name": "Charlie",
		   "friends": [
			  {  
				"uid": "user_3",
				"name": "Dana"
			  }
			]
		}
	]`)

	json2 := []byte(`[
		{
		  "uid": "user_3",
		   "name": "Dana"
		},
		{
		   "uid": "user_4",
		   "name": "Charlie",
		   "friends": [
			  {  
				"uid": "user_1",
				"name": "Bob"
			  }
			]
		}
	]`)

	c1 := newTreeCRDT()
	_, err := c1.ImportJSON(json1, clientID1)
	assert.NoError(t, err, "ImportJSON should not return an error")

	c2 := newTreeCRDT()
	_, err = c2.ImportJSON(json2, clientID2)
	assert.NoError(t, err, "ImportJSON should not return an error")

	err = c1.Sync(c2)
	assert.NoError(t, err, "Sync should not return an error")

	exportedJSON1, err := c1.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")

	exportedJSON2, err := c2.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")
	compareJSON(t, exportedJSON1, exportedJSON2)

	c3 := newTreeCRDT()
	_, err = c3.ImportJSON(exportedJSON1, clientID3)
	assert.NoError(t, err, "Clone should not return an error")
	err = c3.Sync(c2)
	assert.NoError(t, err, "Sync should not return an error")

	exportedJSON3, err := c3.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")

	exportedJSON2, err = c2.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")
	compareJSON(t, exportedJSON2, exportedJSON3)

	err = c3.Sync(c3)
	assert.NoError(t, err, "Sync with itself should not return an error")

	exportedJSON3, err = c3.ExportJSON()
	compareJSON(t, exportedJSON3, exportedJSON2)
}

func TestTreeCRDTMarkDeletedArray(t *testing.T) {
	clientID := ClientID("clientA")

	initialJSON := []byte(`[2, 3, 4]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, clientID)
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	node3, err := c.GetNodeByPath("/1")
	assert.NoError(t, err, "GetNodeByPath should not return an error")

	// Mark deleted
	err = node3.MarkDeleted(clientID)
	assert.NoError(t, err, "SetDeleted should not return an error")

	// Check if the node is marked as deleted
	assert.True(t, node3.IsDeleted, "Node should be marked as deleted")

	exportedJSON, err := c.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")

	expectedJSON := []byte(`[2, 4]`)
	compareJSON(t, expectedJSON, exportedJSON)

	arrayNodeID := c.Root.Edges[0].To
	arrayNode, ok := c.GetNode(arrayNodeID)
	assert.True(t, ok, "Array node should exist in the tree")

	// List number of edges and nodes
	assert.Equal(t, 3, len(arrayNode.Edges), "Deleted node should have no edges")
	assert.Equal(t, 5, len(c.Nodes), "Tree should still have the root node after deletion")

	// Tidy the tree
	c.Tidy()

	assert.Equal(t, 2, len(arrayNode.Edges), "Deleted node should have no edges")
	assert.Equal(t, 4, len(c.Nodes), "Tree should still have the root node after deletion")

}

func TestTreeCRDTMarkDeletedMap(t *testing.T) {
	clientID := ClientID("clientA")

	initialJSON := []byte(`{"A": 1, "B": 2, "C": 3}`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, clientID)
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	nodeB, err := c.GetNodeByPath("/B")
	assert.NoError(t, err, "GetNodeByPath should not return an error")

	// Mark deleted
	err = nodeB.MarkDeleted(clientID)
	assert.NoError(t, err, "SetDeleted should not return an error")

	// Check if the node is marked as deleted
	assert.True(t, nodeB.IsDeleted, "Node should be marked as deleted")

	exportedJSON, err := c.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")

	expectedJSON := []byte(`{"A": 1, "C": 3}`)
	compareJSON(t, expectedJSON, exportedJSON)

	arrayNodeID := c.Root.Edges[0].To
	arrayNode, ok := c.GetNode(arrayNodeID)
	assert.True(t, ok, "Array node should exist in the tree")

	// List number of edges and nodes
	assert.Equal(t, 3, len(arrayNode.Edges), "Deleted node should have no edges")
	assert.Equal(t, 5, len(c.Nodes), "Tree should still have the root node after deletion")

	// Tidy the tree
	c.Tidy()

	assert.Equal(t, 2, len(arrayNode.Edges), "Deleted node should have no edges")
	assert.Equal(t, 4, len(c.Nodes), "Tree should still have the root node after deletion")
}

func TestTreeCRDTMarkDeletedLitteral(t *testing.T) {
	clientID := ClientID("clientA")

	initialJSON := []byte(`"A"`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, clientID)
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	nodeAID := c.Root.Edges[0].To
	assert.NoError(t, err, "GetNodeByPath should not return an error")
	nodeA, ok := c.GetNode(nodeAID)
	assert.True(t, ok, "Node A should exist in the tree")

	// Mark deleted
	err = nodeA.MarkDeleted(clientID)
	assert.NoError(t, err, "SetDeleted should not return an error")

	// Check if the node is marked as deleted
	assert.True(t, nodeA.IsDeleted, "Node should be marked as deleted")

	exportedJSON, err := c.ExportJSON()
	assert.NoError(t, err, "ExportJSON should not return an error")

	expectedJSON := []byte(`null`)
	compareJSON(t, expectedJSON, exportedJSON)
}
