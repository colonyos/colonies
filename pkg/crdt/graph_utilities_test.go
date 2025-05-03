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

func TestImportGraph(t *testing.T) {
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

	g := NewGraph()
	_, err := g.ImportJSON(originalJSON, "", "", -1, false, clientID)
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))
	compareJSON(t, originalJSON, exportedJSON)
}

func TestGraphAddToObject(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	// Initial JSON with missing "C"
	initialJSON := []byte(`{
        "A": "1",
        "B": "2",
        "D": "2"
    }`)

	// We will create graph that looks like this:
	// Root
	// └── FirstChild
	//     ├── A
	//     ├── B
	//     └── D

	g := NewGraph()

	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	// Find the root child
	firstObjID := g.Root.Edges[0].To
	childNode, ok := g.GetNode(firstObjID)
	if !ok {
		t.Fatalf("Child node not found")
	}
	assert.NotNil(t, childNode, "Child node should not be nil")
	childNode.SetField("C", "3", ClientID(clientID), 1)

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))

	// Define the expected correct JSON
	expectedJSON := []byte(`{
        "A": "1",
        "B": "2",
        "C": "3",
        "D": "2"
    }`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestGraphAddToArray(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	// Initial JSON with missing "C"
	initialJSON := []byte(`[
		{"id": "A", "value": "1"},
		{"id": "B", "value": "2"},
		{"id": "D", "value": "2"}
	]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── D

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	if err != nil {
		t.Fatalf("Failed to add node recursively: %v", err)
	}

	// Create missing C
	nodeIDC := NodeID("C")
	nodeC := g.GetOrCreateNode(nodeIDC, true)
	nodeC.SetField("id", "C", ClientID(clientID), 1)
	nodeC.SetField("value", "3", ClientID(clientID), 1)

	g.InsertEdge(g.Root.ID, NodeID("C"), "", 2, clientID) // Insert C

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))

	// Correct expected JSON
	expectedJSON := []byte(`[
		{"id": "A", "value": "1"},
		{"id": "B", "value": "2"},
		{"id": "C", "value": "3"},
		{"id": "D", "value": "2"}
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestGraphStringArrayLitteral(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "D"]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── D

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	nodeIDC := generateRandomNodeID(string("C"))
	nodeC := g.GetOrCreateNode(nodeIDC, false)
	nodeC.Litteral = true
	nodeC.LitteralValue = "C"

	err = g.InsertEdge(g.Root.ID, nodeIDC, "", 2, clientID) // Position 2
	assert.Nil(t, err, "InsertEdge should not return an error")

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))

	// Correct expected JSON
	expectedJSON := []byte(`[
		"A",
		"B",
		"C",
		"D"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestGraphIntArrayLitteral(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`[1, 2, 4]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── D

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := g.ExportRawToJSON()
	assert.Nil(t, err, "ExportToRaw should not return an error")
	t.Logf("Exported Graph Raw JSON:\n%s", string(rawJSON))

	nodeIDC := generateRandomNodeID(string("C"))
	nodeC := g.GetOrCreateNode(nodeIDC, false)
	nodeC.Litteral = true
	nodeC.LitteralValue = 3

	err = g.InsertEdge(g.Root.ID, nodeIDC, "", 10, clientID) // Invalid position
	assert.NotNil(t, err, "InsertEdge should return an error for invalid position")
	err = g.InsertEdge(g.Root.ID, nodeIDC, "", -1, clientID) // Invalid position
	assert.NotNil(t, err, "InsertEdge should return an error for invalid position")
	err = g.InsertEdge(g.Root.ID, nodeIDC, "", 2, clientID) // Position 2
	assert.Nil(t, err, "InsertEdge should not return an error")

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))

	// Correct expected JSON
	expectedJSON := []byte(`[
		1,
		2,
		3,
		4
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestExportImportRaw(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`{
		"A": "1",
		"B": "2",
		"D": "2"
	}`)

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	rawJSON, err := g.ExportRawToJSON()
	assert.Nil(t, err, "ExportToRaw should not return an error")
	t.Logf("Exported Graph Raw JSON:\n%s", string(rawJSON))

	json, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(json))

	g2 := NewGraph()
	err = g2.ImportRawJSON(rawJSON)
	assert.Nil(t, err, "ImportRawJSON should not return an error")

	json2, err := g2.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(json2))

	compareJSON(t, json, json2)
	assert.True(t, g.Equal(g2), "Graphs should be equal after import/export")

}

// The prupose of this test is to check that the graph ID is correctly set
func TestGraphID(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())

	initialJSON := []byte(`["A", "B", "B"]`)

	// We will create graph that looks like this:
	// Root
	// ├── A
	// ├── B
	// └── B <- Duplicate

	g := NewGraph()
	_, err := g.ImportJSON(initialJSON, "", "", -1, false, ClientID(clientID))
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	exportedJSON, err := g.ExportToJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")
	t.Logf("Exported Graph JSON:\n%s", string(exportedJSON))

	// Correct expected JSON
	expectedJSON := []byte(`[
		"A",
		"B",
		"B"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)

	// Print raw JSON
	rawJSON, err := g.ExportRawToJSON()
	assert.Nil(t, err, "ExportToRaw should not return an error")
	t.Logf("Exported Graph Raw JSON:\n%s", string(rawJSON))
}
