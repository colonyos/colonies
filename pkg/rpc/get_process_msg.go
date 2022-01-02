package rpc

import (
	"encoding/json"
)

const GetProcessMsgType = "getprocess"

type GetProcessMsg struct {
	ProcessID string `json:"processid"`
}

func CreateGetProcessMsg(processID string) *GetProcessMsg {
	msg := &GetProcessMsg{}
	msg.ProcessID = processID

	return msg
}

func (msg *GetProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetProcessMsgFromJSON(jsonString string) (*GetProcessMsg, error) {
	var msg *GetProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
