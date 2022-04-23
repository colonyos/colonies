package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createProcessSpec() *core.ProcessSpec {
	colonyID := core.GenerateRandomID()
	runtimeType := "test_runtime_type"
	runtime1ID := core.GenerateRandomID()
	runtime2ID := core.GenerateRandomID()
	maxExecTime := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	env := make(map[string]string)
	env["test_key"] = "test_value"

	return core.CreateProcessSpec("test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, maxExecTime, maxRetries, mem, cores, gpus, env)
}

func TestRPCSubmitProcessSpecMsg(t *testing.T) {
	msg := CreateSubmitProcessSpecMsg(createProcessSpec())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubmitProcessSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitProcessSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitProcessSpecMsgIndent(t *testing.T) {
	msg := CreateSubmitProcessSpecMsg(createProcessSpec())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubmitProcessSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitProcessSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitProcessSpecMsgEquals(t *testing.T) {
	msg := CreateSubmitProcessSpecMsg(createProcessSpec())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
