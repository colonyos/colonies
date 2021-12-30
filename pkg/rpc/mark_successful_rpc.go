package rpc

import (
	"encoding/json"
)

const MarkSuccessfulMsgType = "MarkSuccessful"

type MarkSuccesfulMsg struct {
	RPC       RPC    `json:"rpc"`
	ProcessID string `json:"processid"`
}

func CreateMarkSuccessfulMsg(processID string) *MarkSuccesfulMsg {
	msg := &MarkSuccesfulMsg{}
	msg.RPC.Method = MarkSuccessfulMsgType
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

func CreateMarkSuccessfulMsgFromJSON(jsonString string) (*MarkSuccesfulMsg, error) {
	var msg *MarkSuccesfulMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
