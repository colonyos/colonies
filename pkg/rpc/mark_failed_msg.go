package rpc

import (
	"encoding/json"
)

const MarkFailedMsgType = "markfailed"

type MarkFailedMsg struct {
	ProcessID string `json:"processid"`
}

func CreateMarkFailedMsg(processID string) *MarkFailedMsg {
	msg := &MarkFailedMsg{}
	msg.ProcessID = processID

	return msg
}

func (msg *MarkFailedMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *MarkFailedMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateMarkFailedMsgFromJSON(jsonString string) (*MarkFailedMsg, error) {
	var msg *MarkFailedMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
