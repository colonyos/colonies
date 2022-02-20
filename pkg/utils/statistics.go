package utils

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func CalcAvgWaitingTime(processes []*core.Process) float64 {
	if len(processes) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, process := range processes {
		sum += process.WaitingTime().Seconds()
	}
	return sum / float64(len(processes))
}

func CalcAvgProcessingTime(processes []*core.Process) float64 {
	if len(processes) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, process := range processes {
		sum += process.ProcessingTime().Seconds()
	}
	return sum / float64(len(processes))
}

// This function assumes that the processes are sorted so that the oldest process is at index 0
func CalcUtilization(processes []*core.Process) float64 {
	sum := 0.0
	for _, process := range processes {
		sum += process.ProcessingTime().Seconds()
	}

	if len(processes) == 0 {
		return 0.0
	}

	startTime := processes[0].SubmissionTime
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime).Seconds()

	return sum / elapsedTime
}
