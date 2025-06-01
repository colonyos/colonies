package crdt

import (
	"testing"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestSecureTreeAdapterBasic(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"
	initialJSON := []byte(`["A", "B", "B"]`)

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err, "NewSecureTree should not return an error")

	_, err = c.ImportJSON(initialJSON, prvKey)
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	aNode, err := c.GetNodeByPath("/0")
	assert.Nil(t, err, "GetNodeByPath should not return an error")
	err = aNode.SetLiteral("AA", prvKeyInvalid)
	assert.NotNil(t, err, "SetLiteral should return an error for invalid private key")
	err = aNode.SetLiteral("AA", prvKey)
	assert.Nil(t, err, "SetLiteral should not return an error")

	exportedJSON, err := c.ExportJSON()
	assert.Nil(t, err, "ExportToJSON should not return an error")

	// Correct expected JSON
	expectedJSON := []byte(`[
		"AA",
		"B",
		"B"
	]`)

	compareJSON(t, expectedJSON, exportedJSON)
}

func TestSecureTreeAdapterSetLiteral(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"
	initialJSON := []byte(`["A", "B", "B"]`)

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(initialJSON, prvKey)
	assert.Nil(t, err)

	t.Run("Reject SetLiteral with invalid key", func(t *testing.T) {
		aNode, err := c.GetNodeByPath("/0")
		assert.Nil(t, err)

		err = aNode.SetLiteral("AA", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow SetLiteral with valid key", func(t *testing.T) {
		aNode, err := c.GetNodeByPath("/0")
		assert.Nil(t, err)

		err = aNode.SetLiteral("AA", prvKey)
		assert.Nil(t, err)

		secureNode := aNode.(*AdapterSecureNodeCRDT)
		assert.NotEmpty(t, secureNode.nodeCrdt.Nounce)
		assert.NotEmpty(t, secureNode.nodeCrdt.Signature)
	})
}

func TestSecureTreeAdapterCreateMapNode(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)
	secureNode := root.(*AdapterSecureNodeCRDT)

	t.Run("Reject CreateMapNode with invalid key", func(t *testing.T) {
		_, err := secureNode.CreateMapNode(prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow CreateMapNode with valid key", func(t *testing.T) {
		mapNode, err := secureNode.CreateMapNode(prvKey)
		assert.Nil(t, err)

		_, ok := mapNode.(*AdapterSecureNodeCRDT)
		assert.True(t, ok, "returned node should be of type *AdapterSecureNodeCRDT")
	})
}

func TestSecureTreeAdapterSetKeyValue(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Get the root node and create a map node under it
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)
	mapNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)

	t.Run("Reject SetKeyValue on map node with invalid key", func(t *testing.T) {
		_, err := mapNode.SetKeyValue("someKey", "someValue", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow SetKeyValue on map node with valid key", func(t *testing.T) {
		nodeID, err := mapNode.SetKeyValue("someKey", "someValue", prvKey)
		assert.Nil(t, err)
		assert.NotEmpty(t, nodeID)

		// Optionally verify the key is accessible
		childNode, err := c.GetNodeByPath("/someKey")
		assert.Nil(t, err)
		assert.NotNil(t, childNode)
	})
}

func TestSecureTreeAdapterRemoveKeyValue(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create map node under root and add a key-value
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)
	mapNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)

	_, err = mapNode.SetKeyValue("keyToRemove", "value", prvKey)
	assert.Nil(t, err)

	t.Run("Reject RemoveKeyValue with invalid key", func(t *testing.T) {
		err := mapNode.RemoveKeyValue("keyToRemove", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow RemoveKeyValue with valid key", func(t *testing.T) {
		err := mapNode.RemoveKeyValue("keyToRemove", prvKey)
		assert.Nil(t, err)

		// Confirm the key no longer exists
		_, err = c.GetNodeByPath("/keyToRemove")
		assert.NotNil(t, err)
	})
}

func TestSecureTreeAdapterCreateAttachedNode(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create a parent map node under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)
	parentNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	parentID := parentNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject CreateAttachedNode with invalid key", func(t *testing.T) {
		_, err := c.CreateAttachedNode("child", Literal, parentID, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow CreateAttachedNode with valid key", func(t *testing.T) {
		childNode, err := c.CreateAttachedNode("child", Map, parentID, prvKey)
		assert.Nil(t, err)
		assert.NotNil(t, childNode)
	})
}

func TestSecureTreeAdapterCreateNode(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	t.Run("Reject CreateNode with invalid key", func(t *testing.T) {
		_, err := c.CreateNode("myNode", Map, prvKeyInvalid)
		assert.Nil(t, err) // This is actually ok, as long as the node is not attached to the tree
	})

	t.Run("Allow CreateNode with valid key", func(t *testing.T) {
		node, err := c.CreateNode("myNode", Map, prvKey)
		assert.Nil(t, err)
		assert.NotNil(t, node)
	})
}

func TestSecureTreeAdapterAddEdge(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create toNode as detached node (not attached to root)
	toNode, err := c.CreateNode("detachedNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject AddEdge with invalid key", func(t *testing.T) {
		err := c.AddEdge(fromNodeID, toNodeID, "edgeLabel", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow AddEdge with valid key", func(t *testing.T) {
		err := c.AddEdge(fromNodeID, toNodeID, "edgeLabel", prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterRemoveEdge(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create toNode as detached node
	toNode, err := c.CreateNode("detachedNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// First: Add the edge (valid)
	err = c.AddEdge(fromNodeID, toNodeID, "edgeLabel", prvKey)
	assert.Nil(t, err)

	t.Run("Reject RemoveEdge with invalid key", func(t *testing.T) {
		err := c.RemoveEdge(fromNodeID, toNodeID, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow RemoveEdge with valid key", func(t *testing.T) {
		err := c.RemoveEdge(fromNodeID, toNodeID, prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterAppendEdge(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create toNode as detached node
	toNode, err := c.CreateNode("detachedNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject AppendEdge with invalid key", func(t *testing.T) {
		err := c.AppendEdge(fromNodeID, toNodeID, "edgeLabel", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow AppendEdge with valid key", func(t *testing.T) {
		err := c.AppendEdge(fromNodeID, toNodeID, "edgeLabel", prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterPrependEdge(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create toNode as detached node
	toNode, err := c.CreateNode("detachedNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject PrependEdge with invalid key", func(t *testing.T) {
		err := c.PrependEdge(fromNodeID, toNodeID, "edgeLabel", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow PrependEdge with valid key", func(t *testing.T) {
		err := c.PrependEdge(fromNodeID, toNodeID, "edgeLabel", prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterInsertEdgeLeft(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create sibling node (first edge)
	siblingNode, err := c.CreateNode("siblingNode", Map, prvKey)
	assert.Nil(t, err)
	siblingNodeID := siblingNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Add sibling edge first
	err = c.AppendEdge(fromNodeID, siblingNodeID, "edgeLabel", prvKey)
	assert.Nil(t, err)

	// Create toNode (node we want to insert to the left of sibling)
	toNode, err := c.CreateNode("toNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject InsertEdgeLeft with invalid key", func(t *testing.T) {
		err := c.InsertEdgeLeft(fromNodeID, toNodeID, "edgeLabel", siblingNodeID, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow InsertEdgeLeft with valid key", func(t *testing.T) {
		err := c.InsertEdgeLeft(fromNodeID, toNodeID, "edgeLabel", siblingNodeID, prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterInsertEdgeRight(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create fromNode under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	fromNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	fromNodeID := fromNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Create sibling node (first edge)
	siblingNode, err := c.CreateNode("siblingNode", Map, prvKey)
	assert.Nil(t, err)
	siblingNodeID := siblingNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Add sibling edge first
	err = c.AppendEdge(fromNodeID, siblingNodeID, "edgeLabel", prvKey)
	assert.Nil(t, err)

	// Create toNode (node we want to insert to the right of sibling)
	toNode, err := c.CreateNode("toNode", Map, prvKey)
	assert.Nil(t, err)
	toNodeID := toNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	t.Run("Reject InsertEdgeRight with invalid key", func(t *testing.T) {
		err := c.InsertEdgeRight(fromNodeID, toNodeID, "edgeLabel", siblingNodeID, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow InsertEdgeRight with valid key", func(t *testing.T) {
		err := c.InsertEdgeRight(fromNodeID, toNodeID, "edgeLabel", siblingNodeID, prvKey)
		assert.Nil(t, err)
	})
}

func TestSecureTreeAdapterImportJSON(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Example JSON structure
	jsonData := []byte(`{
		"foo": "bar",
		"baz": 123
	}`)

	t.Run("Reject ImportJSON with invalid key", func(t *testing.T) {
		_, err := c.ImportJSON(jsonData, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow ImportJSON with valid key", func(t *testing.T) {
		nodeID, err := c.ImportJSON(jsonData, prvKey)
		assert.Nil(t, err)
		assert.NotEmpty(t, nodeID)

		// OPTIONAL: Verify that keys are accessible
		nodeFoo, err := c.GetNodeByPath("/foo")
		assert.Nil(t, err)
		assert.NotNil(t, nodeFoo)

		nodeBaz, err := c.GetNodeByPath("/baz")
		assert.Nil(t, err)
		assert.NotNil(t, nodeBaz)
	})
}

func TestSecureTreeAdapterImportJSONToMap(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create parent map node under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	parentMapNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)
	parentID := parentMapNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Example JSON to import
	jsonData := []byte(`{
		"nestedFoo": "value1",
		"nestedBar": 42
	}`)

	t.Run("Reject ImportJSONToMap with invalid key", func(t *testing.T) {
		_, err := c.ImportJSONToMap(jsonData, parentID, "childKey", prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow ImportJSONToMap with valid key", func(t *testing.T) {
		nodeID, err := c.ImportJSONToMap(jsonData, parentID, "childKey", prvKey)
		assert.Nil(t, err)
		assert.NotEmpty(t, nodeID)
	})
}

func TestSecureTreeAdapterImportJSONToArray(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKeyInvalid := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	// Create parent array node under root
	root, err := c.GetNodeByPath("/")
	assert.Nil(t, err)

	parentArrayNode, err := root.(*AdapterSecureNodeCRDT).CreateMapNode(prvKey)
	assert.Nil(t, err)

	// Now under parentArrayNode, add an array key
	parentID := parentArrayNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	arrayNode, err := c.CreateNode("arrayKey", Array, prvKey)
	assert.Nil(t, err)
	arrayNodeID := arrayNode.(*AdapterSecureNodeCRDT).nodeCrdt.ID

	// Link the array node under parent map node
	err = c.AppendEdge(parentID, arrayNodeID, "arrayKey", prvKey)
	assert.Nil(t, err)

	// Example array JSON
	jsonData := []byte(`[
		"elem1",
		"elem2",
		"elem3"
	]`)

	t.Run("Reject ImportJSONToArray with invalid key", func(t *testing.T) {
		_, err := c.ImportJSONToArray(jsonData, arrayNodeID, prvKeyInvalid)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Allow ImportJSONToArray with valid key", func(t *testing.T) {
		nodeID, err := c.ImportJSONToArray(jsonData, arrayNodeID, prvKey)
		assert.Nil(t, err)
		assert.NotEmpty(t, nodeID)
	})
}

func TestSecureTreeAdapterMerge(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	c1, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	jsonData := []byte(`{
		"foo": "bar",
		"baz": 123
	}`)

	c1.ImportJSON(jsonData, prvKey)

	c2, err := c1.Clone()

	mapNode, err := c2.GetNodeByPath("/")
	assert.Nil(t, err)

	valueNodeID, err := mapNode.SetKeyValue("newKey", "newValue", prvKey)
	valueNode, ok := c2.GetNode(valueNodeID)
	assert.True(t, ok, "GetNode should return the node")
	assert.NotNil(t, valueNode, "valueNode should not be nil")
	oldSignature := valueNode.(*AdapterSecureNodeCRDT).nodeCrdt.Signature
	valueNode.(*AdapterSecureNodeCRDT).nodeCrdt.Signature = "e713a1bb015fecabb5a084b0fe6d6e7271fca6f79525a634183cfdb175fe69241f4da161779d8e6b761200e1cf93766010a19072fa778f9643363e2cfadd640900" // Invalid signature for testing
	assert.Nil(t, err, "SetKeyValue should return an error for invalid private key")

	err = c1.Merge(c2, prvKey)
	assert.NotNil(t, err, "Merge should return an error since c2 has a node with an invalid signature")

	// Restore the original signature for a valid merge
	valueNode.(*AdapterSecureNodeCRDT).nodeCrdt.Signature = oldSignature

	err = c1.Merge(c2, prvKey)
	assert.Nil(t, err, "Merge should not return an error after restoring the signature")
}

func TestCloneABACPolicyIsolation(t *testing.T) {
	tree := newTreeCRDT()
	tree.ABACPolicy = NewABACPolicy(tree)
	tree.ABACPolicy.Allow("client1", ActionModify, "root", true)

	clone, err := tree.Clone()
	assert.Nil(t, err)

	// Modify clone ABACPolicy
	clone.ABACPolicy.Allow("client2", ActionModify, "root", true)

	// Original should not have client2
	allowed := tree.ABACPolicy.IsAllowed("client2", ActionModify, "root")
	assert.False(t, allowed, "Original tree ABACPolicy should not be affected by clone")
}

func TestSecureTreeAdapterMergeABAC(t *testing.T) {
	prvKey1 := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKey2 := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	identity2, err := crypto.CreateIdendityFromString(prvKey2)
	assert.Nil(t, err)

	c1, err := NewSecureTree(prvKey1)
	assert.Nil(t, err)

	jsonData := []byte(`{
		"foo": "bar",
		"baz": 123
	}`)

	c1.ImportJSON(jsonData, prvKey1)

	c2, err := c1.Clone()

	mapNode, err := c2.GetNodeByPath("/")
	assert.Nil(t, err)

	valueNodeID, err := mapNode.SetKeyValue("newKey", "newValue", prvKey2)
	assert.Error(t, err, "SetKeyValue should return an error for prvKey2 since identity2 is not allowed to modify the root node")

	c2.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	valueNodeID, err = mapNode.SetKeyValue("newKey", "newValue", prvKey2)
	assert.NoError(t, err, "SetKeyValue should not return an error for prvKey2")

	valueNode, ok := c2.GetNode(valueNodeID)
	assert.True(t, ok, "GetNode should return the node")
	assert.NotNil(t, valueNode, "valueNode should not be nil")

	err = c1.Merge(c2, prvKey1)
	assert.NotNil(t, err, "Merge should return an error since identity2 is not allowed to modify the root node")

	c1.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	err = c1.Merge(c2, prvKey1)
	assert.Nil(t, err, "Merge should not return an error after restoring the signature")
}

func TestSecureTreeAdapterSyncABAC(t *testing.T) {
	prvKey1 := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKey2 := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	identity2, err := crypto.CreateIdendityFromString(prvKey2)
	assert.Nil(t, err)

	c1, err := NewSecureTree(prvKey1)
	assert.Nil(t, err)

	jsonData := []byte(`{
		"foo": "bar",
		"baz": 123
	}`)

	c1.ImportJSON(jsonData, prvKey1)

	c2, err := c1.Clone()

	mapNode, err := c2.GetNodeByPath("/")
	assert.Nil(t, err)

	valueNodeID, err := mapNode.SetKeyValue("newKey", "newValue", prvKey2)
	assert.Error(t, err, "SetKeyValue should return an error for prvKey2 since identity2 is not allowed to modify the root node")

	c2.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	valueNodeID, err = mapNode.SetKeyValue("newKey", "newValue", prvKey2)
	assert.NoError(t, err, "SetKeyValue should not return an error for prvKey2")

	valueNode, ok := c2.GetNode(valueNodeID)
	assert.True(t, ok, "GetNode should return the node")
	assert.NotNil(t, valueNode, "valueNode should not be nil")

	c1.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	err = c1.Sync(c2, prvKey1)
	assert.NotNil(t, err, "Merge should return an error since identity2 is not allowed to modify the root node")

	exportedC1JSON, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")
	exportedC2JSON, err := c2.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")
	compareJSON(t, exportedC1JSON, exportedC2JSON)
}

func TestSecureTreeAdapterArraySyncABAC(t *testing.T) {
	prvKey1 := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKey2 := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	identity2, err := crypto.CreateIdendityFromString(prvKey2)
	assert.Nil(t, err)

	c1, err := NewSecureTree(prvKey1)
	assert.Nil(t, err)

	c1.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	jsonData := []byte(`["A", "B", "C"]`)

	c1.ImportJSON(jsonData, prvKey1)

	c2, err := c1.Clone()
	assert.NoError(t, err, "Clone should not return an error")

	c2.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	arrayNode, err := c2.GetNodeByPath("/")
	assert.Nil(t, err)

	node, err := c2.CreateNode("5", Literal, prvKey2)
	node.SetLiteral("D", prvKey2)
	err = c2.AppendEdge(arrayNode.ID(), node.ID(), "", prvKey2)
	assert.Nil(t, err, "AppendEdge should not return an error")

	err = c1.Sync(c2, prvKey1)
	assert.NotNil(t, err, "Merge should return an error since identity2 is not allowed to modify the root node")

	exportedC1JSON, err := c1.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")
	exportedC2JSON, err := c2.ExportJSON()
	assert.Nil(t, err, "ExportJSON should not return an error")
	compareJSON(t, exportedC1JSON, exportedC2JSON)
}

func TestSecureTreeAdapterMergeComplexJSONABAC(t *testing.T) {
	prvKey1 := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	prvKey2 := "ed26531bac1838e519c2c6562ac717b22aac041730f0d753d3ad35b76b5f4924"

	json := []byte(`{
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

	identity2, err := crypto.CreateIdendityFromString(prvKey2)
	assert.Nil(t, err)

	c1, err := NewSecureTree(prvKey1)
	assert.Nil(t, err)

	_, err = c1.ImportJSON(json, prvKey1)
	assert.Nil(t, err, "ImportJSON should not return an error")

	c2, err := NewSecureTree(prvKey2)
	assert.Nil(t, err)
	_, err = c2.ImportJSON(json, prvKey2)
	assert.Nil(t, err, "ImportJSON should not return an error")

	err = c1.Merge(c2, prvKey1)
	assert.NotNil(t, err, "Merge should return an error since identity2 is not allowed to modify the root node")

	c1.ABAC().Allow(identity2.ID(), ActionModify, "root", true)

	err = c1.Merge(c2, prvKey1)
	assert.Nil(t, err, "Merge should not return an error after restoring the signature")
}
