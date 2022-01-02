package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const AddRuntimeMsgType = "addruntime"

type AddRuntimeMsg struct {
	Runtime *core.Runtime `json:"runtime"`
}

func CreateAddRuntimeMsg(runtime *core.Runtime) *AddRuntimeMsg {
	msg := &AddRuntimeMsg{}
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
