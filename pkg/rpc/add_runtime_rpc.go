package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const AddRuntimeMsgType = "addruntime"

type AddRuntimeMsg struct {
	RPC     RPC           `json:"rpc"`
	Runtime *core.Runtime `json:"runtime"`
}

func CreateAddRuntimeMsg(runtime *core.Runtime) *AddRuntimeMsg {
	msg := &AddRuntimeMsg{}
	msg.RPC.Method = AddRuntimeMsgType
	msg.RPC.Nonce = core.GenerateRandomID()
	msg.Runtime = runtime

	return msg
}

func (msg *AddRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddRuntimeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddRuntimeMsgFromJSON(jsonString string) (*AddRuntimeMsg, error) {
	var msg *AddRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
