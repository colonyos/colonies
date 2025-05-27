package crdt

import (
	"testing"

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
