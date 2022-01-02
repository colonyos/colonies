package rpc

import (
	"encoding/json"
)

const MarkSuccessfulMsgType = "marksuccessful"

type MarkSuccesfulMsg struct {
	ProcessID string `json:"processid"`
}

func CreateMarkSuccessfulMsg(processID string) *MarkSuccesfulMsg {
	msg := &MarkSuccesfulMsg{}
	msg.ProcessID = processID

	return msg
}

func (msg *MarkSuccesfulMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *MarkSuccesfulMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateMarkSuccessfulMsgFromJSON(jsonString string) (*MarkSuccesfulMsg, error) {
	var msg *MarkSuccesfulMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
