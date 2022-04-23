package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddRuntimePayloadType = "addruntimemsg"

type AddRuntimeMsg struct {
	Runtime *core.Runtime `json:"runtime"`
	MsgType string        `json:"msgtype"`
}

func CreateAddRuntimeMsg(runtime *core.Runtime) *AddRuntimeMsg {
	msg := &AddRuntimeMsg{}
	msg.Runtime = runtime
	msg.MsgType = AddRuntimePayloadType

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

func (msg *AddRuntimeMsg) Equals(msg2 *AddRuntimeMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Runtime.Equals(msg2.Runtime) {
		return true
	}

	return false
}

func CreateAddRuntimeMsgFromJSON(jsonString string) (*AddRuntimeMsg, error) {
	var msg *AddRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
