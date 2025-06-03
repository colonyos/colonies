package crdt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecureTreeAdapterSubscribe(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	json := []byte(`{
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

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(json, prvKey)
	assert.Nil(t, err, "ImportJSON should not return an error")

	// / Map
	// /friends Array
	// /friends/0 Map
	// /friends/1 Map
	// /friends/1/uid Literal
	// /friends/1/name Literal
	// /friends/1/friends Array
	// /friends/1/friends/0 Map

	events := make(chan NodeEvent, 10)
	c.Subscribe("/", events)

	fmt.Println("----------------------")
	node, err := c.GetNodeByPath("/friends/1/name")
	assert.Nil(t, err, "GetNodeByPath should not return an error")
	err = node.SetLiteral("Robert", prvKey)
	assert.Nil(t, err, "SetLiteral should not return an error")

	event := <-events
	assert.Equal(t, "/friends/1/name", event.Path, "Event path should match")
}

func TestSecureTreeAdapterSubscribe_ArrayElement(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	json := []byte(`{
		"uid": "user_1",
		"name": "Alice",
		"friends": [
			{
				"uid": "user_2",
				"name": "Bob"
			},
			{
				"uid": "user_3",
				"name": "Charlie"
			}
		]
	}`)

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(json, prvKey)
	assert.Nil(t, err)

	events := make(chan NodeEvent, 10)
	c.Subscribe("/", events)

	node, err := c.GetNodeByPath("/friends/0/name")
	assert.Nil(t, err)
	err = node.SetLiteral("Bobby", prvKey)
	assert.Nil(t, err)

	event := <-events
	assert.Equal(t, "/friends/0/name", event.Path, "Event path should match")
}

func TestSecureTreeAdapterSubscribe_DeepNested(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	json := []byte(`{
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

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(json, prvKey)
	assert.Nil(t, err)

	events := make(chan NodeEvent, 10)
	c.Subscribe("/", events)

	node, err := c.GetNodeByPath("/friends/1/friends/0/name")
	assert.Nil(t, err)
	err = node.SetLiteral("Daniela", prvKey)
	assert.Nil(t, err)

	event := <-events
	assert.Equal(t, "/friends/1/friends/0/name", event.Path, "Event path should match")
}

func TestSecureTreeAdapterSubscribe_ExactPath(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	json := []byte(`{
		"uid": "user_1",
		"name": "Alice"
	}`)

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(json, prvKey)
	assert.Nil(t, err)

	events := make(chan NodeEvent, 10)
	c.Subscribe("/name", events)

	node, err := c.GetNodeByPath("/name")
	assert.Nil(t, err)
	err = node.SetLiteral("Alicia", prvKey)
	assert.Nil(t, err)

	event := <-events
	assert.Equal(t, "/name", event.Path, "Event path should match")
}

func TestSecureTreeAdapterSubscribe_Subpath_MultipleEvents(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"

	json := []byte(`{
		"uid": "user_1",
		"name": "Alice",
		"friends": [
			{
				"uid": "user_2",
				"name": "Bob"
			},
			{
				"uid": "user_3",
				"name": "Charlie"
			}
		]
	}`)

	c, err := NewSecureTree(prvKey)
	assert.Nil(t, err)

	_, err = c.ImportJSON(json, prvKey)
	assert.Nil(t, err)

	events := make(chan NodeEvent, 10)
	c.Subscribe("/friends/1", events)

	node1, err := c.GetNodeByPath("/friends/1/name")
	assert.Nil(t, err)
	err = node1.SetLiteral("Charles", prvKey)
	assert.Nil(t, err)

	event := <-events
	assert.Equal(t, "/friends/1/name", event.Path, "Event path should match")
}
