package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const SubmitProcessSpecMsgType = "SubmitProcessSpec"

type SubmitProcessSpecMsg struct {
	RPC         RPC               `json:"rpc"`
	ProcessSpec *core.ProcessSpec `json:"spec"`
}

func CreateSubmitProcessSpecMsg(processSpec *core.ProcessSpec) *SubmitProcessSpecMsg {
	msg := &SubmitProcessSpecMsg{}
	msg.RPC.Method = SubmitProcessSpecMsgType
	msg.ProcessSpec = processSpec

	return msg
}

func (msg *SubmitProcessSpecMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateSubmitProcessSpecMsgFromJSON(jsonString string) (*SubmitProcessSpecMsg, error) {
	var msg *SubmitProcessSpecMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
