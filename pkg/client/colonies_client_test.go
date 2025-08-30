package client

import (
	"testing"

	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestClientPauseResumeMessageConstruction(t *testing.T) {
	colonyName := "test_colony"
	
	// Test that pause assignments constructs the correct message
	msg := rpc.CreatePauseAssignmentsMsg(colonyName)
	assert.Equal(t, rpc.PauseAssignmentsPayloadType, msg.MsgType)
	assert.Equal(t, colonyName, msg.ColonyName)
	jsonString, err := msg.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, jsonString, "pauseassignmentsmsg")
	assert.Contains(t, jsonString, colonyName)

	// Test that resume assignments constructs the correct message  
	msg2 := rpc.CreateResumeAssignmentsMsg(colonyName)
	assert.Equal(t, rpc.ResumeAssignmentsPayloadType, msg2.MsgType)
	assert.Equal(t, colonyName, msg2.ColonyName)
	jsonString2, err := msg2.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, jsonString2, "resumeassignmentsmsg")
	assert.Contains(t, jsonString2, colonyName)
}