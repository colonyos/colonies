package core

import (
	"encoding/json"
)

type Statistics struct {
	Waiting int `json:"waiting"`
	Running int `json:"running"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

func CreateStatistics(waiting int, running int, success int, failed int) *Statistics {
	stat := &Statistics{Waiting: waiting, Running: running, Success: success, Failed: failed}

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

	if stat.Waiting == stat2.Waiting && stat.Running == stat2.Running && stat.Success == stat2.Success && stat.Failed == stat2.Failed {
		return true
	}

	return false
}

func (stat *Statistics) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(stat, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
