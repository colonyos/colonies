package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCReportAllocationMsg(t *testing.T) {
	project1 := core.Project{AllocatedCPU: 1, UsedCPU: 1, AllocatedGPU: 1, UsedGPU: 1, AllocatedStorage: 1, UsedStorage: 1}
	projects := make(map[string]core.Project)
	projects["test_project1"] = project1
	alloc := core.Allocations{Projects: projects}
	msg := CreateReportAllocationsMsg("test_colony_name", "test_executor_name", alloc)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateReportAllocationsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateReportAllocationsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCReportAllocationMsgIndent(t *testing.T) {
	project1 := core.Project{AllocatedCPU: 1, UsedCPU: 1, AllocatedGPU: 1, UsedGPU: 1, AllocatedStorage: 1, UsedStorage: 1}
	projects := make(map[string]core.Project)
	projects["test_project1"] = project1
	alloc := core.Allocations{Projects: projects}
	msg := CreateReportAllocationsMsg("test_colony_name", "test_executor_name", alloc)

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateReportAllocationsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateReportAllocationsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCReportAllocationMsgEquals(t *testing.T) {
	project1 := core.Project{AllocatedCPU: 1, UsedCPU: 1, AllocatedGPU: 1, UsedGPU: 1, AllocatedStorage: 1, UsedStorage: 1}
	projects := make(map[string]core.Project)
	projects["test_project1"] = project1
	alloc := core.Allocations{Projects: projects}
	msg := CreateReportAllocationsMsg("test_colony_name", "test_executor_name", alloc)

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
