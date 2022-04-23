package core

import (
	"encoding/json"
)

type Failure struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func CreateFailure(status int, message string) *Failure {
	return &Failure{Status: status, Message: message}
}

func ConvertJSONToFailure(jsonString string) (*Failure, error) {
	var failure *Failure
	err := json.Unmarshal([]byte(jsonString), &failure)
	if err != nil {
		return nil, err
	}

	return failure, nil
}

func (failure *Failure) Equals(failure2 *Failure) bool {
	if failure2 == nil {
		return false
	}

	if failure.Status == failure2.Status &&
		failure.Message == failure2.Message {
		return true
	}

	return false
}

func (failure *Failure) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(failure)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
