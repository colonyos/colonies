package crdt

import (
	"fmt"
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

	crdt := NewTreeCRDT()
	_, err := crdt.ImportJSON(originalJSON, "", "", -1, false, clientID)
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

	raw, err := crdt.Save()
	assert.Nil(t, err)
	fmt.Println(string(raw))

	// node, err = crdt.GetNodeByPath("/friends/0/uid")
	// assert.Nil(t, err)
	// v, err = node.GetLiteral()
	// assert.Nil(t, err)
	// vStr = v.(string)
	// assert.Equal(t, "user_2", vStr)
}
