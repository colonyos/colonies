package core

import (
	"encoding/json"
)

type Log struct {
	ProcessID  string `json:"processid"`
	ColonyID   string `json:"colonyid"`
	ExecutorID string `json:"executorid"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"` // UTC Unix time
}

func ConvertJSONToLog(jsonString string) (Log, error) {
	var log Log
	err := json.Unmarshal([]byte(jsonString), &log)
	if err != nil {
		return Log{}, err
	}

	return log, nil
}

func ConvertLogArrayToJSON(logs []Log) (string, error) {
	jsonBytes, err := json.MarshalIndent(logs, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToLogArray(jsonString string) ([]Log, error) {
	var logs []Log
	err := json.Unmarshal([]byte(jsonString), &logs)
	if err != nil {
		return logs, err
	}

	return logs, nil
}

func (log *Log) Equals(log2 Log) bool {
	same := true
	if log.ProcessID != log2.ProcessID ||
		log.ColonyID != log2.ColonyID ||
		log.ExecutorID != log2.ExecutorID ||
		log.Message != log2.Message ||
		log.Timestamp != log2.Timestamp {
		same = false
	}

	return same
}

func IsLogArraysEqual(logs1 []Log, logs2 []Log) bool {
	if logs1 == nil || logs2 == nil {
		return false
	}

	if len(logs1) != len(logs2) {
		return false
	}

	counter := 0
	for i := range logs1 {
		if logs1[i].Equals(logs2[i]) {
			counter++
		}
	}

	if counter == len(logs1) {
		return true
	}

	return false
}

func (log *Log) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(log)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
