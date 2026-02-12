package core

import (
	"encoding/json"
)

type Statistics struct {
	Colonies            int `json:"colonies"`
	Executors           int `json:"executors"`
	ActiveExecutors     int `json:"activeexecutors"`
	UnregisteredExecutors int `json:"unregisteredexecutors"`
	WaitingProcesses    int `json:"waitingprocesses"`
	RunningProcesses    int `json:"runningprocesses"`
	SuccessfulProcesses int `json:"successfulprocesses"`
	FailedProcesses     int `json:"failedprocesses"`
	CancelledProcesses  int `json:"cancelledprocesses"`
	WaitingWorkflows    int `json:"waitingworkflows"`
	RunningWorkflows    int `json:"runningworkflows"`
	SuccessfulWorkflows int `json:"successfulworkflows"`
	FailedWorkflows     int `json:"failedworkflows"`
	CancelledWorkflows  int `json:"cancelledworkflows"`
}

func CreateStatistics(colonies int,
	executors int,
	activeExecutors int,
	unregisteredExecutors int,
	waitingProcesses int,
	runningProcesses int,
	successfulProcesses int,
	failedProcesses int,
	cancelledProcesses int,
	waitingWorkflows int,
	runningWorkflows int,
	successfulWorkflows int,
	failedWorkflows int,
	cancelledWorkflows int) *Statistics {
	stat := &Statistics{
		Colonies:            colonies,
		Executors:           executors,
		ActiveExecutors:     activeExecutors,
		UnregisteredExecutors: unregisteredExecutors,
		WaitingProcesses:    waitingProcesses,
		RunningProcesses:    runningProcesses,
		SuccessfulProcesses: successfulProcesses,
		FailedProcesses:     failedProcesses,
		CancelledProcesses:  cancelledProcesses,
		WaitingWorkflows:    waitingWorkflows,
		RunningWorkflows:    runningWorkflows,
		SuccessfulWorkflows: successfulWorkflows,
		FailedWorkflows:     failedWorkflows,
		CancelledWorkflows:  cancelledWorkflows}

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
		stat.ActiveExecutors == stat2.ActiveExecutors &&
		stat.UnregisteredExecutors == stat2.UnregisteredExecutors &&
		stat.WaitingProcesses == stat2.WaitingProcesses &&
		stat.RunningProcesses == stat2.RunningProcesses &&
		stat.SuccessfulProcesses == stat2.SuccessfulProcesses &&
		stat.FailedProcesses == stat2.FailedProcesses &&
		stat.CancelledProcesses == stat2.CancelledProcesses &&
		stat.WaitingWorkflows == stat2.WaitingWorkflows &&
		stat.RunningWorkflows == stat2.RunningWorkflows &&
		stat.SuccessfulWorkflows == stat2.SuccessfulWorkflows &&
		stat.FailedWorkflows == stat2.FailedWorkflows &&
		stat.CancelledWorkflows == stat2.CancelledWorkflows {
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
