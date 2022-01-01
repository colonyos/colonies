package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const SubscribeProcessesType = "SubscribeProcesses"

type SubscribeProcessesMsg struct {
	RPC     RPC           `json:"rpc"`
	Runtime *core.Runtime `json:"runtime"`
}

func CreateSubscribeProcessesMsg(runtime *core.Runtime) *SubscribeProcessesMsg {
	msg := &SubscribeProcessesMsg{}
	msg.RPC.Method = SubscribeProcessesType
	msg.Runtime = runtime

	return msg
}

func (msg *SubscribeProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateSubscribeProcessesMsgFromJSON(jsonString string) (*SubscribeProcessesMsg, error) {
	var msg *SubscribeProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
