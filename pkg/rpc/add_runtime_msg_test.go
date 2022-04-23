package rpc

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createRuntime() *core.Runtime {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	return core.CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
}

func TestRPCAddRuntimeMsg(t *testing.T) {
	runtime := createRuntime()

	msg := CreateAddRuntimeMsg(runtime)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddRuntimeMsgIndent(t *testing.T) {
	runtime := createRuntime()

	msg := CreateAddRuntimeMsg(runtime)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddRuntimeMsgEquals(t *testing.T) {
	runtime := createRuntime()

	msg := CreateAddRuntimeMsg(runtime)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
