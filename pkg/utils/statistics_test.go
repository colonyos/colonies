package utils

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/go-playground/assert/v2"
)

func TestCalcAvgTimes(t *testing.T) {
	startTime := time.Now()

	colonyID := core.GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec1 := core.CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)

	var zeroProcesses []*core.Process
	var processes []*core.Process
	process1 := core.CreateProcess(processSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(2 * time.Second))
	process1.SetEndTime(startTime.Add(3 * time.Second))
	process1.SetState(core.SUCCESS)
	process1.Retries = 3
	processes = append(processes, process1)

	process2 := core.CreateProcess(processSpec1)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(10 * time.Second))
	process2.SetEndTime(startTime.Add(20 * time.Second))
	process2.SetState(core.SUCCESS)
	process2.Retries = 4
	processes = append(processes, process2)

	avgWaitTime := CalcAvgWaitingTime(processes)
	assert.Equal(t, avgWaitTime, 6.0)
	avgWaitTime = CalcAvgWaitingTime(zeroProcesses)
	assert.Equal(t, avgWaitTime, 0.0)

	avgProcessingTime := CalcAvgProcessingTime(processes)
	assert.Equal(t, avgProcessingTime, 5.5)
	avgProcessingTime = CalcAvgProcessingTime(zeroProcesses)
	assert.Equal(t, avgProcessingTime, 0.0)

	utilization := CalcUtilization(processes)
	assert.Equal(t, utilization > 0, true)
	utilization = CalcUtilization(zeroProcesses)
	assert.Equal(t, utilization, 0.0)

	retries := CalcRetries(processes)
	assert.Equal(t, retries, 7)
	retries = CalcRetries(zeroProcesses)
	assert.Equal(t, retries, 0)
}
