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

	crdt := NewTreeCRDT()
	_, err := crdt.ImportJSON(originalJSON, "", "", -1, false, clientID)
	if err != nil {
		t.Fatalf("Failed to import JSON: %v", err)
	}

	node, err := crdt.GetNodeByPath("/uid")
	assert.Nil(t, err)
	assert.NotNil(t, node)
	v, ok := node.GetValue("uid")
	assert.True(t, ok)
	vStr := v.(string)
	assert.Equal(t, "user_1", vStr)

	v, ok, err = crdt.GetValueByPath("/uid")
	assert.Nil(t, err)
	assert.True(t, ok)
	vStr = v.(string)
	assert.Equal(t, "user_1", vStr)

	s, ok, err := crdt.GetStringValueByPath("/uid")
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, "user_1", s)
}
