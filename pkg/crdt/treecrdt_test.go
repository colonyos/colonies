package crdt

import (
	"fmt"
	"sort"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGraphSetFields(t *testing.T) {
	c := NewTreeCRDT()

	clientID1 := ClientID(core.GenerateRandomID())
	clientID2 := ClientID(core.GenerateRandomID())

	node := c.CreateAttachedNode("obj", false, c.Root.ID, clientID1)

	// 1. Set an initial value
	node.SetField("key", "value1", clientID1, 1)

	// 2. Overwrite with a higher version (should win)
	node.SetField("key", "value2", clientID1, 2)

	field := node.Fields["key"]
	assert.Equal(t, "value2", field.Value, "Expected field to be updated to value2")

	// 3. Lower version from same client (should be ignored)
	node.SetField("key", "value3", clientID1, 1)
	field = node.Fields["key"]
	assert.Equal(t, "value2", field.Value, "Lower version should not overwrite")

	// 4. Simulate a conflict: two different clients, same version
	node.SetField("key", "value4", clientID2, 2)

	field = node.Fields["key"]
	// Tie-breaker: client ID with lowest ID (lexical order)
	var expectedWinner ClientID
	if clientID1 < clientID2 {
		expectedWinner = clientID1
	} else {
		expectedWinner = clientID2
	}
	if field.Owner != expectedWinner {
		t.Errorf("Expected owner %s to win tie-breaker, got %s", expectedWinner, field.Owner)
	}
}

func TestNodeRemoveField(t *testing.T) {
	c := NewTreeCRDT()

	clientID := ClientID(core.GenerateRandomID())

	node := c.CreateAttachedNode("obj", false, c.Root.ID, clientID)

	// 1. Set a field initially
	node.SetField("testkey", "testvalue", clientID, 1)

	field, exists := node.Fields["testkey"]
	assert.True(t, exists, "Field should exist after SetField")
	assert.Equal(t, "testvalue", field.Value, "Field value should match what was set")

	// 2. Remove the field with higher version — should succeed
	node.RemoveField("testkey", clientID, 2)

	_, exists = node.Fields["testkey"]
	assert.False(t, exists, "Field should be deleted after RemoveField with higher version")

	// 3. Set the field again
	node.SetField("testkey", "newvalue", clientID, 3)

	field, exists = node.Fields["testkey"]
	assert.True(t, exists, "Field should exist after being set again")
	assert.Equal(t, "newvalue", field.Value, "New value should be set correctly")

	// 4. Try removing it with lower version — should NOT remove
	node.RemoveField("testkey", clientID, 1)

	field, exists = node.Fields["testkey"]
	assert.True(t, exists, "Field should still exist after failed RemoveField with lower version")
	assert.Equal(t, "newvalue", field.Value, "Field value should remain unchanged after failed remove")
}

func TestGraphAddEdgeWithVersion(t *testing.T) {
	c := NewTreeCRDT()

	// To make the test deterministic, we will use fixed client IDs
	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := c.CreateAttachedNode("parent", false, c.Root.ID, clientID)
	child := c.CreateAttachedNode("child", false, c.Root.ID, clientID)

	// 1. Add an edge with version 1
	err := c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "AddEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge")
	assert.Equal(t, child.ID, parent.Edges[0].To, "Edge should point to child")
	assert.Equal(t, "link", parent.Edges[0].Label, "Edge label mismatch")

	// 2. Add another edge with higher version (should succeed)
	anotherChild := c.CreateAttachedNode("another_child", false, c.Root.ID, clientID)
	err = c.addEdgeWithVersion(parent.ID, anotherChild.ID, "link2", clientID, 2)
	assert.Nil(t, err, "AddEdgeWithVersion second time should not return error")

	assert.Equal(t, 2, len(parent.Edges), "Expected 2 edges now")

	// 3. Try to add conflicting edge with lower version (should be ignored)
	fakeChild := c.CreateAttachedNode("fake_child", false, c.Root.ID, clientID)
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
	tieChild := c.CreateAttachedNode("tie_child", false, c.Root.ID, otherClientID)
	err = c.addEdgeWithVersion(parent.ID, tieChild.ID, "tie_link", otherClientID, 2) // same version
	assert.Nil(t, err, "AddEdgeWithVersion with same version different client should not error")

	if otherClientID < clientID {
		assert.Equal(t, 3, len(parent.Edges), "Tie-breaker: new client wins")
	} else {
		assert.Equal(t, 2, len(parent.Edges), "Tie-breaker: original client keeps ownership")
	}
}

func TestGraphRemoveEdgeWithVersion(t *testing.T) {
	c := NewTreeCRDT()

	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := c.CreateAttachedNode("parent", false, c.Root.ID, clientID)
	child := c.CreateAttachedNode("child", false, c.Root.ID, clientID)

	// Add an edge
	err := c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "addEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge before removal")

	// Remove the edge with higher version (should succeed)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2)
	assert.Nil(t, err, "removeEdgeWithVersion should not return error")
	assert.Equal(t, 0, len(parent.Edges), "Expected 0 edges after removal")

	// Re-add it for conflict test
	_ = c.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 3)

	// Try to remove with lower version (should be ignored)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2)
	assert.NotNil(t, err, "removeEdgeWithVersion with lower version should error")
	assert.Equal(t, 1, len(parent.Edges), "Edge should still exist after invalid removal")

	// Tie-break with other client (lower client ID wins)
	err = c.removeEdgeWithVersion(parent.ID, child.ID, otherClientID, 3)
	assert.Nil(t, err, "removeEdgeWithVersion tie-break should not error")

	if otherClientID < clientID {
		assert.Equal(t, 0, len(parent.Edges), "Tie-break: other client removed the edge")
	} else {
		assert.Equal(t, 1, len(parent.Edges), "Tie-break: original client kept the edge")
	}
}

func TestGraphRemoveIndexInArray(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "C"]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── D

	c := NewTreeCRDT()
	_, err := c.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	// Find the node with ID "B"
	edges := c.Root.Edges
	for _, edge := range edges {
		node := c.Nodes[edge.To]
		if node.LitteralValue.(string) == "B" {
			// Remove the edge with ID "B"
			err = c.removeEdgeWithVersion(c.Root.ID, node.ID, clientID, 3)
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

func TestGraphTidy(t *testing.T) {
	c := NewTreeCRDT()

	clientID := ClientID("client")

	parent := c.CreateAttachedNode("parent", false, c.Root.ID, clientID)
	child := c.CreateAttachedNode("child", false, c.Root.ID, clientID)

	// Create an orphan node manually (NOT attached)
	orphanID := generateRandomNodeID("orphan")
	orphan := c.getOrCreateNode(orphanID, false, clientID, 1)

	// Connect parent → child
	err := c.AddEdge(parent.ID, child.ID, "link", clientID)
	assert.Nil(t, err, "AddEdge should not fail")

	assert.Equal(t, 4, len(c.Nodes), "Expected 4 nodes before purge (root, parent, child, orphan)")

	c.Tidy() // Remove orphan nodes

	// Should only have root, parent, and child left
	_, orphanExists := c.Nodes[orphan.ID]
	assert.False(t, orphanExists, "Orphan should have been purged")
	assert.Equal(t, 3, len(c.Nodes), "Expected 3 nodes after purge (root, parent, child)")
}

func TestNodeSetLiteral(t *testing.T) {
	c := NewTreeCRDT()

	clientID1 := ClientID("client1")
	clientID2 := ClientID("client2")

	node := c.CreateAttachedNode("literalNode", false, c.Root.ID, clientID1)

	// 1. Set an initial literal value
	node.SetLiteral("hello", clientID1, 1)

	assert.True(t, node.Litteral, "Expected node to be marked as literal")
	assert.Equal(t, "hello", node.LitteralValue, "Expected literal value to be 'hello'")

	// 2. Set a higher version value (should overwrite)
	node.SetLiteral("world", clientID1, 2)

	assert.Equal(t, "world", node.LitteralValue, "Expected literal value to be updated to 'world'")

	// 3. Attempt to set with a lower version (should be ignored)
	node.SetLiteral("ignored", clientID1, 1)

	assert.Equal(t, "world", node.LitteralValue, "Lower version should not overwrite the value")

	// 4. Simulate conflict: different client, same version
	node.SetLiteral("conflict", clientID2, 2)

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
	assert.Equal(t, expectedValue, node.LitteralValue, fmt.Sprintf("Expected literal value %s after conflict resolution, got %s", expectedValue, node.LitteralValue))
}

// // Test case:
// // 1. Create two graphs with shared nodes
// // 2. Set different literal values on the same node in both graphs
// // 3. Merge the graphs
// // 4. The merged graph should be an array of literals since n1 + n2 → [n1, n2] sorted by node ID
func TestGraphMergeLitterals(t *testing.T) {
	c1 := NewTreeCRDT()
	c2 := NewTreeCRDT()

	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Create shared nodes in both graphs
	node1 := c1.CreateAttachedNode("sharedA", false, c1.Root.ID, clientA)
	node2 := c2.CreateAttachedNode("sharedB", false, c2.Root.ID, clientB)
	err := node1.SetLiteral("A-literal", clientA, 1)
	assert.Nil(t, err, "SetLiteral should not return an error")
	err = node2.SetLiteral("B-literal", clientB, 1)
	assert.Nil(t, err, "SetLiteral should not return an error")

	// Perform merge
	c1.Merge(c2)

	// Check that all nodes exist
	_, ok1 := c1.GetNode(node1.ID)
	_, ok2 := c1.GetNode(node2.ID)
	assert.True(t, ok1, "Node1 should exist after merge")
	assert.True(t, ok2, "Node2 should exist after merge")

	// Check if literal value is merged into array and sorted by NodeID
	root := c1.Root
	assert.GreaterOrEqual(t, len(root.Edges), 2, "Expected at least two edges from root")

	ids := make([]string, 0)
	for _, edge := range root.Edges {
		ids = append(ids, string(edge.To))
	}

	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)

	json, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	// Since B-literal did not have any siblings, it will be prepended to the array, that why is it first
	expectedJSON := []byte(`["B-literal", "A-literal"]`)
	compareJSON(t, expectedJSON, json)
}

func TestGraphMergeLists(t *testing.T) {
	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	initialJSON := []byte(`[1, 2, 4]`)

	// We will create graph that looks like this:
	// Root
	// ├── 1
	// ├── 2
	// └── 4

	c1 := NewTreeCRDT()
	_, err := c1.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientA))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	c2 := NewTreeCRDT()
	c2.Load(rawJSON)
	assert.Nil(t, err, "ImportRawJSON should not return an error")

	rawJSONBefore, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	c1.Merge(c2)

	rawJSONAfter, err := c1.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	// Graph should be identical before and after merge
	assert.Equal(t, rawJSONBefore, rawJSONAfter, "Graph should be identical before and after merge")
	assert.True(t, c1.Equal(c2), "Graphs should be equal after merge")

	// Let's do some modifications on the graph independently
	// Original    :    [1, 2, 4]
	// G1(A):        [0, 1, 2, 4]
	// G2(B):           [1, 2, 3, 4]
	// G1 + G2:      [0, 1, 2, 3, 4] <- 4 is added to G1, owner of root is B
	// G2 + G1:      [0, 1, 2, 3, 4] <- 0 is added to G2, owner of root is A

	// 1. Create a new node in g1
	node0 := c1.CreateNode("0", true, clientA)
	node0.Litteral = true
	node0.LitteralValue = 0

	// Find the node with id "0"
	sibling, err := c1.GetSibling(c1.Root.ID, 0)
	assert.Nil(t, err, "GetSiblingNode should not return an error")
	err = c1.InsertEdgeLeft(c1.Root.ID, node0.ID, "", sibling.ID, clientA)
	assert.Nil(t, err, "InsertEdge should not return an error")
	// G1: [0, 1, 2, 4]  <-- 0 added

	// 2. Create a new node in g2
	node3 := c2.CreateNode("3", true, clientA)
	node3.Litteral = true
	node3.LitteralValue = 3
	sibling, err = c2.GetSibling(c2.Root.ID, 1)
	assert.Nil(t, err, "GetSiblingNode should not return an error")
	err = c2.InsertEdgeRight(c2.Root.ID, node3.ID, "", sibling.ID, clientB)
	assert.Nil(t, err, "InsertEdge should not return an error")
	// G2: [1, 2, 3, 4]   <-- 3 added

	c1Clone, err := c1.Clone()
	assert.Nil(t, err, "Clone should not return an error")

	// 3. Merge the graphs
	c1.Merge(c2)
	c2.Merge(c1Clone)

	jsom, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	expectedJSON := []byte(`[0, 1, 2, 3, 4]`)
	compareJSON(t, expectedJSON, jsom)

	json2, err := c2.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	expectedJSON2 := []byte(`[0, 1, 2, 3, 4]`)
	compareJSON(t, expectedJSON2, json2)

	// C2 == C1
	assert.True(t, c1.Equal(c2), "Graphs should be equal after merge")
	assert.True(t, c1.Root.Owner == c2.Root.Owner, "Owners should be equal after merge")
}
