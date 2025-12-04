package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createTestBlueprint() *core.Blueprint {
	blueprint := core.CreateBlueprint("ExecutorDeployment", "test-deployment", "test-colony")
	blueprint.SetSpec("image", "nginx:latest")
	blueprint.SetSpec("replicas", 3)
	return blueprint
}

func TestRPCUpdateBlueprintMsg(t *testing.T) {
	blueprint := createTestBlueprint()

	msg := CreateUpdateBlueprintMsg(blueprint)
	assert.False(t, msg.ForceGeneration, "ForceGeneration should be false by default")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateUpdateBlueprintMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateUpdateBlueprintMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.False(t, msg2.ForceGeneration)
}

func TestRPCUpdateBlueprintMsgWithForce(t *testing.T) {
	blueprint := createTestBlueprint()

	// Test with ForceGeneration = true
	msg := CreateUpdateBlueprintMsgWithForce(blueprint, true)
	assert.True(t, msg.ForceGeneration, "ForceGeneration should be true")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateUpdateBlueprintMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg2.ForceGeneration, "ForceGeneration should be preserved after JSON round-trip")
	assert.True(t, msg.Equals(msg2))

	// Test with ForceGeneration = false
	msg3 := CreateUpdateBlueprintMsgWithForce(blueprint, false)
	assert.False(t, msg3.ForceGeneration, "ForceGeneration should be false")

	jsonString3, err := msg3.ToJSON()
	assert.Nil(t, err)

	msg4, err := CreateUpdateBlueprintMsgFromJSON(jsonString3)
	assert.Nil(t, err)
	assert.False(t, msg4.ForceGeneration)
}

func TestRPCUpdateBlueprintMsgIndent(t *testing.T) {
	blueprint := createTestBlueprint()

	msg := CreateUpdateBlueprintMsgWithForce(blueprint, true)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateUpdateBlueprintMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateUpdateBlueprintMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
	assert.True(t, msg2.ForceGeneration)
}

func TestRPCUpdateBlueprintMsgEquals(t *testing.T) {
	blueprint := createTestBlueprint()

	msg := CreateUpdateBlueprintMsg(blueprint)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))

	// Test equals with different blueprints
	blueprint2 := createTestBlueprint()
	blueprint2.ID = "different-id"
	msg2 := CreateUpdateBlueprintMsg(blueprint2)
	assert.False(t, msg.Equals(msg2))
}

func TestRPCUpdateBlueprintMsgNilBlueprint(t *testing.T) {
	msg1 := &UpdateBlueprintMsg{Blueprint: nil, MsgType: UpdateBlueprintPayloadType}
	msg2 := &UpdateBlueprintMsg{Blueprint: nil, MsgType: UpdateBlueprintPayloadType}

	assert.True(t, msg1.Equals(msg2), "Two messages with nil blueprints should be equal")

	blueprint := createTestBlueprint()
	msg3 := CreateUpdateBlueprintMsg(blueprint)
	assert.False(t, msg1.Equals(msg3), "Message with nil blueprint should not equal message with blueprint")
	assert.False(t, msg3.Equals(msg1), "Message with blueprint should not equal message with nil blueprint")
}
