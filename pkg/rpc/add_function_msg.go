package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddFunctionPayloadType = "addfunctionmsg"

type AddFunctionMsg struct {
	Function *core.Function `json:"fun"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddFunctionMsg(function *core.Function) *AddFunctionMsg {
	msg := &AddFunctionMsg{}
	msg.Function = function
	msg.MsgType = AddFunctionPayloadType

	return msg
}

func (msg *AddFunctionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddFunctionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddFunctionMsg) Equals(msg2 *AddFunctionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Function.Equals(msg2.Function) {
		return true
	}

	return false
}

func CreateAddFunctionMsgFromJSON(jsonString string) (*AddFunctionMsg, error) {
	var msg *AddFunctionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
