package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCPauseAssignmentsMsg(t *testing.T) {
	colonyName := "test_colony"
	msg := CreatePauseAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreatePauseAssignmentsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreatePauseAssignmentsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.Equal(t, colonyName, msg2.ColonyName)
}

func TestRPCPauseAssignmentsMsgIndent(t *testing.T) {
	colonyName := "test_colony"
	msg := CreatePauseAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreatePauseAssignmentsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreatePauseAssignmentsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.Equal(t, colonyName, msg2.ColonyName)
}

func TestRPCPauseAssignmentsMsgEquals(t *testing.T) {
	colonyName := "test_colony"
	msg1 := CreatePauseAssignmentsMsg(colonyName)
	msg2 := CreatePauseAssignmentsMsg(colonyName)

	assert.True(t, msg1.Equals(msg2))
	assert.False(t, msg1.Equals(nil))

	// Test different message type
	msg3 := &PauseAssignmentsMsg{MsgType: "different", ColonyName: colonyName}
	assert.False(t, msg1.Equals(msg3))

	// Test different colony name
	msg4 := CreatePauseAssignmentsMsg("different_colony")
	assert.False(t, msg1.Equals(msg4))
}