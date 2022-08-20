package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddCronMsg(t *testing.T) {
	cron := core.CreateCron(core.GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	msg := CreateAddCronMsg(cron)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddCronMsgIndent(t *testing.T) {
	cron := core.CreateCron(core.GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	msg := CreateAddCronMsg(cron)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddCronMsgEquals(t *testing.T) {
	cron := core.CreateCron(core.GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	msg := CreateAddCronMsg(cron)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
