package core

import (
	"encoding/json"
)

type Statistics struct {
	Colonies            int `json:"colonies"`
	Executors           int `json:"executors"`
	WaitingProcesses    int `json:"waitingprocesses"`
	RunningProcesses    int `json:"runningprocesses"`
	SuccessfulProcesses int `json:"successfulprocesses"`
	FailedProcesses     int `json:"failedprocesses"`
	WaitingWorkflows    int `json:"waitingworkflows"`
	RunningWorkflows    int `json:"runningworkflows"`
	SuccessfulWorkflows int `json:"successfulworkflows"`
	FailedWorkflows     int `json:"failedworkflows"`
}

func CreateStatistics(colonies int,
	executors int,
	waitingProcesses int,
	runningProcesses int,
	successfulProcesses int,
	failedProcesses int,
	waitingWorkflows int,
	runningWorkflows int,
	successfulWorkflows int,
	failedWorkflows int) *Statistics {
	stat := &Statistics{
		Colonies:            colonies,
		Executors:           executors,
		WaitingProcesses:    waitingProcesses,
		RunningProcesses:    runningProcesses,
		SuccessfulProcesses: successfulProcesses,
		FailedProcesses:     failedProcesses,
		WaitingWorkflows:    waitingWorkflows,
		RunningWorkflows:    runningWorkflows,
		SuccessfulWorkflows: successfulWorkflows,
		FailedWorkflows:     failedWorkflows}

	return stat
}

func ConvertJSONToStatistics(jsonString string) (*Statistics, error) {
	var stat *Statistics
	err := json.Unmarshal([]byte(jsonString), &stat)
	if err != nil {
		return nil, err
	}

	return stat, nil
}

func (stat *Statistics) Equals(stat2 *Statistics) bool {
	if stat2 == nil {
		return false
	}

	if stat.Colonies == stat2.Colonies &&
		stat.Executors == stat2.Executors &&
		stat.WaitingProcesses == stat2.WaitingProcesses &&
		stat.RunningProcesses == stat2.RunningProcesses &&
		stat.SuccessfulProcesses == stat2.SuccessfulProcesses &&
		stat.FailedProcesses == stat2.FailedProcesses &&
		stat.WaitingWorkflows == stat2.WaitingWorkflows &&
		stat.RunningWorkflows == stat2.RunningWorkflows &&
		stat.SuccessfulWorkflows == stat2.SuccessfulWorkflows &&
		stat.FailedWorkflows == stat2.FailedWorkflows {
		return true
	}

	return false
}

func (stat *Statistics) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(stat)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
