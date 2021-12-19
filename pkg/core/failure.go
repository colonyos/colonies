package core

import (
	"encoding/json"
)

type FailureJSON struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type Failure struct {
	status  int
	message string
}

func CreateFailure(status int, message string) *Failure {
	return &Failure{status: status, message: message}
}

func CreateFailureFromJSON(jsonString string) (*Failure, error) {
	var failureJSON FailureJSON
	err := json.Unmarshal([]byte(jsonString), &failureJSON)
	if err != nil {
		return nil, err
	}

	return CreateFailure(failureJSON.Status, failureJSON.Message), nil
}

func (failure *Failure) Status() int {
	return failure.status
}

func (failure *Failure) Message() string {
	return failure.message
}

func (failure *Failure) ToJSON() (string, error) {
	failureJSON := &FailureJSON{Status: failure.status, Message: failure.message}

	jsonString, err := json.Marshal(failureJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
