package crdt

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func compareJSON(t *testing.T, expectedJSON, exportedJSON []byte) {
	var expected, actual interface{}
	err := json.Unmarshal(expectedJSON, &expected)
	assert.Nil(t, err, "Failed to unmarshal expected JSON: %v", err)
	err = json.Unmarshal(exportedJSON, &actual)
	assert.Nil(t, err, "Failed to unmarshal exported JSON: %v", err)
	assert.True(t, reflect.DeepEqual(expected, actual), "Exported JSON does not match expected.\nExpected:\n%v\n\nGot:\n%v\n", expected, actual)
}

func isJSONEqual(t *testing.T, expectedJSON, exportedJSON []byte) bool {
	var expected, actual interface{}
	err := json.Unmarshal(expectedJSON, &expected)
	assert.Nil(t, err, "Failed to unmarshal expected JSON: %v", err)
	err = json.Unmarshal(exportedJSON, &actual)
	assert.Nil(t, err, "Failed to unmarshal exported JSON: %v", err)
	return reflect.DeepEqual(expected, actual)
}

func TestTreeCRDTImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	originalJSON := []byte(`{
		"uid": "user_1",
		"name": "Alice",
		"friends": [
			{
				"uid": "user_2",
				"name": "Bob"
			},
			{
				"uid": "user_3",
				"name": "Charlie",
				"friends": [
					{
						"uid": "user_4",
						"name": "Dana"
					}
				]
			}
		]
	}`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(originalJSON, clientID)
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	compareJSON(t, originalJSON, exportedJSON)
}

func TestTreeCRDTImportToArray(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	originalJSON := []byte(`{
		"uid": "user_1",
		"name": "Alice",
		"friends": [
			{
				"uid": "user_2",
				"name": "Bob"
			},
			{
				"uid": "user_3",
				"name": "Charlie",
				"friends": [
					{
						"uid": "user_4",
						"name": "Dana"
					}
				]
			}
		]
	}`)

	childJSON := []byte(`{
		"uid": "user_4",
		"name": "Bob2"
	}`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(originalJSON, clientID)
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	friendsNode, err := c.GetNodeByPath("/friends")
	assert.Nil(t, err, "GetNodeByPath should not return an error")

	_, err = c.ImportJSONToArray(childJSON, friendsNode.ID, clientID)
	assert.Nil(t, err, "ImportJSON should not return an error")

	str, err := c.GetStringValueByPath("/friends/0/uid")
	assert.Nil(t, err, "GetNodeByPath should not return an error")
	assert.Equal(t, "user_4", str, "Expected uid to be 'user_4'")
}

func TestTreeCRDTImportToMap(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	originalJSON := []byte(`{
		"uid": "user_1",
		"name": "Alice",
		"friends": [
			{
				"uid": "user_2",
				"name": "Bob"
			},
			{
				"uid": "user_3",
				"name": "Charlie",
				"friends": [
					{
						"uid": "user_4",
						"name": "Dana"
					}
				]
			}
		]
	}`)

	childJSON := []byte(`{
		"uid": "user_4",
		"name": "Bob"
	}`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(originalJSON, clientID)
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	mapNode, err := c.GetNodeByPath("/friends/0")
	assert.Nil(t, err, "GetNodeByPath should not return an error")

	_, err = c.ImportJSONToMap(childJSON, mapNode.ID, "testkey", clientID)
	assert.Nil(t, err, "ImportJSON should not return an error")

	str, err := c.GetStringValueByPath("/friends/0/testkey/name")
	assert.Nil(t, err, "GetNodeByPath should not return an error")
	assert.Equal(t, "Bob", str, "Expected name to be 'Bob'")
}

func TestTreeCRDTSetKeyValueAfterImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`{
        "A": "1",
        "B": "2",
        "D": "2"
    }`)

	c := newTreeCRDT()

	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	// Find the root child
	firstObjID := c.Root.Edges[0].To
	childNode, ok := c.GetNode(firstObjID)
	if !ok {
		t.Fatalf("Child node not found")
	}
	assert.NotNil(t, childNode, "Child node should not be nil")
	childNode.SetKeyValue("C", "3", ClientID(clientID))

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`{
        "A": "1",
        "B": "2",
        "C": "3",
        "D": "2"
    }`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTAddToArrayAfterImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`[
		{"id": "A", "value": "1"},
		{"id": "B", "value": "2"},
		{"id": "D", "value": "2"}
	]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	nodeC := c.CreateNode("C", Map, clientID)
	//nodeC.IsMap = true
	nodeC.SetKeyValue("id", "C", ClientID(clientID))
	nodeC.SetKeyValue("value", "3", ClientID(clientID))

	arrayNodeID := c.Root.Edges[0].To
	leftTo, err := c.GetSibling(arrayNodeID, 1)
	assert.Nil(t, err, "GetSibling should not return an error")
	c.InsertEdgeRight(arrayNodeID, nodeC.ID, "", leftTo.ID, clientID) // Insert C

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`[
		{"id": "A", "value": "1"},
		{"id": "B", "value": "2"},
		{"id": "C", "value": "3"},
		{"id": "D", "value": "2"}
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTInsertStringAfterImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "D"]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	nodeIDC := generateRandomNodeID(string("C"))
	nodeC := c.getOrCreateNode(nodeIDC, Literal, clientID, 1)
	//nodeC.IsLiteral = true
	nodeC.LiteralValue = "C"

	arrayNodeID := c.Root.Edges[0].To
	sibling, err := c.GetSibling(arrayNodeID, 2) // Index 1 is D
	err = c.InsertEdgeLeft(arrayNodeID, nodeIDC, "", sibling.ID, clientID)
	assert.Nil(t, err, "InsertEdge should not return an error")

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`[
		"A",
		"B",
		"C",
		"D"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTInsertIntAfterImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`[1, 2, 4]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	nodeIDC := generateRandomNodeID(string("C"))
	nodeC := c.getOrCreateNode(nodeIDC, Literal, clientID, 1)
	//nodeC.IsLiteral = true
	nodeC.LiteralValue = 3

	sibling, err := c.GetSibling(c.Root.ID, 20)
	assert.NotNil(t, err, "Invalid index should not return an error")

	sibling, err = c.GetSibling(c.Root.ID, -1)
	assert.NotNil(t, err, "Invalid index should not return an error")

	arrayNodeID := c.Root.Edges[0].To

	sibling, err = c.GetSibling(arrayNodeID, 1) // Index 1 is B

	err = c.InsertEdgeRight(arrayNodeID, nodeIDC, "", sibling.ID, clientID)
	assert.Nil(t, err, "InsertEdge should not return an error")

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	expectedJSON := []byte(`[
		1,
		2,
		3,
		4
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestTreeCRDTSaveAndLoad(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`[
		{"id": "A", "value": "1"},
		{"id": "B", "value": "2"},
		{"id": "C", "value": "3"},
		{"id": "D", "value": "2"}
	]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := c.Save()
	assert.Nil(t, err, "ExportToRaw should not return an error")

	c2 := newTreeCRDT()
	err = c2.Load(rawJSON)
	assert.Nil(t, err, "ImportRawJSON should not return an error")
	assert.True(t, c.Equal(c2), "Trees should be equal after import/export")
}

// The prupose of this test is to check that the graph ID is correctly set
func TestTreeCRDTNodeIDAfterImport(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "B"]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── B <- Duplicate

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	// Correct expected JSON
	expectedJSON := []byte(`[
		"A",
		"B",
		"B"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}
