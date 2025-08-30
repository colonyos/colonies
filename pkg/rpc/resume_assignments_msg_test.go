package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCResumeAssignmentsMsg(t *testing.T) {
	colonyName := "test_colony"
	msg := CreateResumeAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateResumeAssignmentsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResumeAssignmentsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.Equal(t, colonyName, msg2.ColonyName)
}

func TestRPCResumeAssignmentsMsgIndent(t *testing.T) {
	colonyName := "test_colony"
	msg := CreateResumeAssignmentsMsg(colonyName)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateResumeAssignmentsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResumeAssignmentsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.Equal(t, colonyName, msg2.ColonyName)
}

func TestRPCResumeAssignmentsMsgEquals(t *testing.T) {
	colonyName := "test_colony"
	msg1 := CreateResumeAssignmentsMsg(colonyName)
	msg2 := CreateResumeAssignmentsMsg(colonyName)

	assert.True(t, msg1.Equals(msg2))
	assert.False(t, msg1.Equals(nil))

	// Test different message type
	msg3 := &ResumeAssignmentsMsg{MsgType: "different", ColonyName: colonyName}
	assert.False(t, msg1.Equals(msg3))

	// Test different colony name
	msg4 := CreateResumeAssignmentsMsg("different_colony")
	assert.False(t, msg1.Equals(msg4))
}