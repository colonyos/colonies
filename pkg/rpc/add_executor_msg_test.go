package rpc

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createExecutor() *core.Executor {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyName := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	return core.CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
}

func TestRPCAddExecutorMsg(t *testing.T) {
	executor := createExecutor()

	msg := CreateAddExecutorMsg(executor)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddExecutorMsgIndent(t *testing.T) {
	executor := createExecutor()

	msg := CreateAddExecutorMsg(executor)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddExecutorMsgEquals(t *testing.T) {
	executor := createExecutor()

	msg := CreateAddExecutorMsg(executor)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
