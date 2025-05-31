package crdt

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestTreeCRDTGetByPath(t *testing.T) {
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

	crdt := newTreeCRDT()
	_, err := crdt.ImportJSON(originalJSON, clientID)
	assert.Nil(t, err)

	node, err := crdt.GetNodeByPath("/uid")
	v, err := node.GetLiteral()
	assert.Nil(t, err)

	vStr := v.(string)
	assert.Equal(t, "user_1", vStr)

	v, err = crdt.GetValueByPath("/uid")
	assert.Nil(t, err)
	vStr = v.(string)
	assert.Equal(t, "user_1", vStr)

	s, err := crdt.GetStringValueByPath("/uid")
	assert.Nil(t, err)
	assert.Equal(t, "user_1", s)

	node, err = crdt.GetNodeByPath("/friends")
	assert.Nil(t, err)
	assert.True(t, node.IsArray)

	node, err = crdt.GetNodeByPath("/friends/0/uid")
	assert.Nil(t, err)
	v, err = node.GetLiteral()
	assert.Nil(t, err)
	vStr = v.(string)
	assert.Equal(t, "user_2", vStr)

	node, err = crdt.GetNodeByPath("/friends/1/friends/0/uid")
	assert.Nil(t, err)
	v, err = node.GetLiteral()
	assert.Nil(t, err)
	vStr = v.(string)
	assert.Equal(t, "user_4", vStr)

	node, err = crdt.GetNodeByPath("/friends/1/friends/0/name/doesnotexist")
	assert.NotNil(t, err)
	assert.Nil(t, node)

	node, err = crdt.GetNodeByPath("friends/1/friends/0/uid")
	assert.NotNil(t, err)
	assert.Nil(t, node)
}

func TestTreeCRDTGetByPathArray(t *testing.T) {
	clientID := ClientID(core.GenerateRandomID())
	initialJSON := []byte(`[2, 3, 4]`)

	c := newTreeCRDT()
	_, err := c.ImportJSON(initialJSON, clientID)
	assert.Nil(t, err, "AddNodeRecursively should not return an error")

	node3, err := c.GetNodeByPath("/1")
	assert.NoError(t, err, "GetNodeByPath should not return an error")
	assert.True(t, node3.IsLiteral, "Node should be a literal")
	assert.True(t, node3.LiteralValue.(float64) == 3, "Node value should be 3")
}
