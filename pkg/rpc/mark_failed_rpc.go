package rpc

import (
	"encoding/json"
)

const MarkFailedMsgType = "MarkFailed"

type MarkFailedMsg struct {
	RPC       RPC    `json:"rpc"`
	ProcessID string `json:"processid"`
}

func CreateMarkFailedMsg(processID string) *MarkFailedMsg {
	msg := &MarkFailedMsg{}
	msg.RPC.Method = MarkFailedMsgType
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

func CreateMarkFailedMsgFromJSON(jsonString string) (*MarkFailedMsg, error) {
	var msg *MarkFailedMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
