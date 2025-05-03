package crdt

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGraphSetFields(t *testing.T) {
	g := NewGraph()

	clientID1 := ClientID(core.GenerateRandomID())
	clientID2 := ClientID(core.GenerateRandomID())

	node := g.CreateAttachedNode("obj", false, g.Root.ID, clientID1)

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
	g := NewGraph()

	clientID := ClientID(core.GenerateRandomID())

	node := g.CreateAttachedNode("obj", false, g.Root.ID, clientID)

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
	g := NewGraph()

	// To make the test deterministic, we will use fixed client IDs
	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := g.CreateAttachedNode("parent", false, g.Root.ID, clientID)
	child := g.CreateAttachedNode("child", false, g.Root.ID, clientID)

	// 1. Add an edge with version 1
	err := g.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "AddEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge")
	assert.Equal(t, child.ID, parent.Edges[0].To, "Edge should point to child")
	assert.Equal(t, "link", parent.Edges[0].Label, "Edge label mismatch")

	// 2. Add another edge with higher version (should succeed)
	anotherChild := g.CreateAttachedNode("another_child", false, g.Root.ID, clientID)
	err = g.addEdgeWithVersion(parent.ID, anotherChild.ID, "link2", clientID, 2)
	assert.Nil(t, err, "AddEdgeWithVersion second time should not return error")

	assert.Equal(t, 2, len(parent.Edges), "Expected 2 edges now")

	// 3. Try to add conflicting edge with lower version (should be ignored)
	fakeChild := g.CreateAttachedNode("fake_child", false, g.Root.ID, clientID)
	err = g.addEdgeWithVersion(parent.ID, fakeChild.ID, "fake_link", clientID, 1) // lower version
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
	tieChild := g.CreateAttachedNode("tie_child", false, g.Root.ID, otherClientID)
	err = g.addEdgeWithVersion(parent.ID, tieChild.ID, "tie_link", otherClientID, 2) // same version
	assert.Nil(t, err, "AddEdgeWithVersion with same version different client should not error")

	if otherClientID < clientID {
		assert.Equal(t, 3, len(parent.Edges), "Tie-breaker: new client wins")
	} else {
		assert.Equal(t, 2, len(parent.Edges), "Tie-breaker: original client keeps ownership")
	}
}

func TestGraphInsertEdgeWithVersion(t *testing.T) {
	g := NewGraph()

	clientID := ClientID("bbbb") // Lexicographically after "aaaa"
	otherClientID := ClientID("aaaa")

	parent := g.CreateAttachedNode("parent", false, g.Root.ID, clientID)
	child := g.CreateAttachedNode("child", false, g.Root.ID, clientID)

	// Insert at position 0
	err := g.insertEdgeWithVersion(parent.ID, child.ID, "link", 0, clientID, 1)
	assert.Nil(t, err, "insertEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge after insert")
	assert.Equal(t, child.ID, parent.Edges[0].To, "Edge should point to child")
	assert.Equal(t, "link", parent.Edges[0].Label, "Edge label mismatch")
	assert.Equal(t, 0, parent.Edges[0].Position, "Edge position mismatch")

	// Insert another child at position 1
	anotherChild := g.CreateAttachedNode("another_child", false, g.Root.ID, clientID)
	err = g.insertEdgeWithVersion(parent.ID, anotherChild.ID, "link2", 1, clientID, 2)
	assert.Nil(t, err, "Second insertEdgeWithVersion should not error")

	assert.Equal(t, 2, len(parent.Edges), "Expected 2 edges after second insert")

	// Try inserting lower version (should be ignored)
	fakeChild := g.CreateAttachedNode("fake_child", false, g.Root.ID, clientID)
	err = g.insertEdgeWithVersion(parent.ID, fakeChild.ID, "fake_link", 0, clientID, 1) // Lower version
	assert.Nil(t, err, "insertEdgeWithVersion with lower version should not error")

	found := false
	for _, edge := range parent.Edges {
		if edge.To == fakeChild.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Edge with lower version should not be added")

	// Simulate tie: different client, same version
	tieChild := g.CreateAttachedNode("tie_child", false, g.Root.ID, otherClientID)
	err = g.insertEdgeWithVersion(parent.ID, tieChild.ID, "tie_link", 2, otherClientID, 2)
	assert.Nil(t, err, "insertEdgeWithVersion with tie should not error")

	if otherClientID < clientID {
		assert.Equal(t, 3, len(parent.Edges), "Tie-breaker: other client wins and edge inserted")
	} else {
		assert.Equal(t, 2, len(parent.Edges), "Tie-breaker: original client keeps ownership")
	}
}

func TestGraphRemoveEdgeWithVersion(t *testing.T) {
	g := NewGraph()

	clientID := ClientID("bbbb")
	otherClientID := ClientID("aaaa")

	parent := g.CreateAttachedNode("parent", false, g.Root.ID, clientID)
	child := g.CreateAttachedNode("child", false, g.Root.ID, clientID)

	// Add an edge
	err := g.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 1)
	assert.Nil(t, err, "addEdgeWithVersion should not return error")

	assert.Equal(t, 1, len(parent.Edges), "Expected 1 edge before removal")

	// Remove the edge with higher version (should succeed)
	err = g.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2)
	assert.Nil(t, err, "removeEdgeWithVersion should not return error")
	assert.Equal(t, 0, len(parent.Edges), "Expected 0 edges after removal")

	// Re-add it for conflict test
	_ = g.addEdgeWithVersion(parent.ID, child.ID, "link", clientID, 3)

	// Try to remove with lower version (should be ignored)
	err = g.removeEdgeWithVersion(parent.ID, child.ID, clientID, 2)
	assert.NotNil(t, err, "removeEdgeWithVersion with lower version should error")
	assert.Equal(t, 1, len(parent.Edges), "Edge should still exist after invalid removal")

	// Tie-break with other client (lower client ID wins)
	err = g.removeEdgeWithVersion(parent.ID, child.ID, otherClientID, 3)
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

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	// Find the node with ID "B"
	edges := g.Root.Edges
	for _, edge := range edges {
		node := g.Nodes[edge.To]
		if node.LitteralValue.(string) == "B" {
			// Remove the edge with ID "B"
			err = g.removeEdgeWithVersion(g.Root.ID, node.ID, clientID, 3)
			assert.Nil(t, err, "removeEdgeWithVersion should not return an error")
			break
		}
	}

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	// Correct expected JSON
	expectedJSON := []byte(`[
		"A",
		"C"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)

	for _, edge := range g.Root.Edges {
		node := g.Nodes[edge.To]
		if node.LitteralValue.(string) == "A" {
			assert.Equal(t, 0, edge.Position, "Edge position should be 0")
		} else if node.LitteralValue.(string) == "C" {
			assert.Equal(t, 1, edge.Position, "Edge position should be 1")
		}
	}
}

func TestGraphTidy(t *testing.T) {
	g := NewGraph()

	clientID := ClientID("client")

	parent := g.CreateAttachedNode("parent", false, g.Root.ID, clientID)
	child := g.CreateAttachedNode("child", false, g.Root.ID, clientID)

	// Create an orphan node manually (NOT attached)
	orphanID := generateRandomNodeID("orphan")
	orphan := g.GetOrCreateNode(orphanID, false)

	// Connect parent → child
	err := g.AddEdge(parent.ID, child.ID, "link", clientID)
	assert.Nil(t, err, "AddEdge should not fail")

	assert.Equal(t, 4, len(g.Nodes), "Expected 4 nodes before purge (root, parent, child, orphan)")

	g.Tidy() // Remove orphan nodes

	// Should only have root, parent, and child left
	_, orphanExists := g.Nodes[orphan.ID]
	assert.False(t, orphanExists, "Orphan should have been purged")

	assert.Equal(t, 3, len(g.Nodes), "Expected 3 nodes after purge (root, parent, child)")
}

func TestNodeSetLiteral(t *testing.T) {
	g := NewGraph()

	clientID1 := ClientID("client1")
	clientID2 := ClientID("client2")

	node := g.CreateAttachedNode("literalNode", false, g.Root.ID, clientID1)

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

// TODO: it should not be possible to setField and use litteral at the same time

// Test case:
// 1. Create two graphs with shared nodes
// 2. Set different literal values on the same node in both graphs
// 3. Merge the graphs
// 4. The merged graph should be an array of literals since n1 + n2 → [n1, n2] sorted by node ID
func TestGraphMergeLitterals(t *testing.T) {
	g1 := NewGraph()
	g2 := NewGraph()

	clientA := ClientID("clientA")
	clientB := ClientID("clientB")

	// Create shared nodes in both graphs
	node1 := g1.CreateAttachedNode("sharedA", false, g1.Root.ID, clientA)
	node2 := g2.CreateAttachedNode("sharedB", false, g2.Root.ID, clientB)
	err := node2.SetLiteral("B-literal", clientB, 1)
	assert.Nil(t, err, "SetLiteral should not return an error")
	err = node1.SetLiteral("A-literal", clientA, 1)
	assert.Nil(t, err, "SetLiteral should not return an error")

	// Perform merge
	g1.Merge(g2)

	// Export and log results
	jsonOutput, err := g1.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	log.Printf("JSON after sync: %s", string(jsonOutput))

	rawJSON, err := g1.ExportRawToJSON()
	assert.Nil(t, err, "ExportRawToJSON should not return an error")
	log.Printf("Raw JSON after sync: %s", string(rawJSON))

	// Check that all nodes exist
	_, ok1 := g1.GetNode(node1.ID)
	_, ok2 := g1.GetNode(node2.ID)
	assert.True(t, ok1, "Node1 should exist after merge")
	assert.True(t, ok2, "Node2 should exist after merge")

	// Check if literal value is merged into array and sorted by NodeID
	root := g1.Root
	assert.GreaterOrEqual(t, len(root.Edges), 2, "Expected at least two edges from root")

	ids := make([]string, 0)
	for _, edge := range root.Edges {
		ids = append(ids, string(edge.To))
	}

	sorted := make([]string, len(ids))
	copy(sorted, ids)
	sort.Strings(sorted)

	assert.Equal(t, sorted, ids, "Edges from root should be sorted by NodeID after merge")
}

func TestGraphMergeLists(t *testing.T) {
	clientA := ClientID("clientA")
	//clientB := ClientID("clientB")

	initialJSON := []byte(`[1, 2, 4]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── D

	g1 := NewGraph()
	_, err := g1.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientA))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := g1.ExportRawToJSON()
	assert.Nil(t, err, "ExportToRaw should not return an error")
	t.Logf("Exported Graph Raw JSON:\n%s", string(rawJSON))

	g2 := NewGraph()
	g2.ImportRawJSON(rawJSON)
	assert.Nil(t, err, "ImportRawJSON should not return an error")

	//g1.Merge(g2)

	json, err := g1.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON after merge:\n%s", string(json))

	rawJSON, err = g1.ExportRawToJSON()
	assert.Nil(t, err, "ExportToRaw should not return an error")
	t.Logf("Exported Graph Raw JSON after merge:\n%s", string(rawJSON))

	// nodeIDC := generateRandomNodeID(string("C"))
	// nodeC := g.GetOrCreateNode(nodeIDC, false)
	// nodeC.Litteral = true
	// nodeC.LitteralValue = 3
	//
	// err = g.InsertEdge(g.Root.ID, nodeIDC, "", 10, clientID) // Invalid position
	// assert.NotNil(t, err, "InsertEdge should return an error for invalid position")
	// err = g.InsertEdge(g.Root.ID, nodeIDC, "", -1, clientID) // Invalid position
	// assert.NotNil(t, err, "InsertEdge should return an error for invalid position")
	// err = g.InsertEdge(g.Root.ID, nodeIDC, "", 2, clientID) // Position 2
	// assert.Nil(t, err, "InsertEdge should not return an error")
	//
	// exportedJSON, err := g.ExportToJSON()
	// assert.Nil(t, err, "ExportToJSON should not return an error")
	// t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))
	//
	// // Correct expected JSON
	// expectedJSON := []byte(`[
	// 	1,
	// 	2,
	// 	3,
	// 	4
	// ]`)
	//
	// compareJSON(t, expectedJSON, exportedJSON)
}
