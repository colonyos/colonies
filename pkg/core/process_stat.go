package core

import (
	"encoding/json"
)

type ProcessStat struct {
	Waiting int `json:"waiting"`
	Running int `json:"running"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

func CreateProcessStat(waiting int, running int, success int, failed int) *ProcessStat {
	stat := &ProcessStat{Waiting: waiting, Running: running, Success: success, Failed: failed}

	return stat
}

func ConvertJSONToProcessStat(jsonString string) (*ProcessStat, error) {
	var stat *ProcessStat
	err := json.Unmarshal([]byte(jsonString), &stat)
	if err != nil {
		return nil, err
	}

	return stat, nil
}

func (stat *ProcessStat) Equals(stat2 *ProcessStat) bool {
	if stat.Waiting == stat2.Waiting && stat.Running == stat2.Running && stat.Success == stat2.Success && stat.Failed == stat2.Failed {
		return true
	}

	return false
}

func (stat *ProcessStat) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(stat, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
