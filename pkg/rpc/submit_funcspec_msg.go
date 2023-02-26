package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const SubmitFunctionSpecPayloadType = "submitfuncspecmsg"

type SubmitFunctionSpecMsg struct {
	FunctionSpec *core.FunctionSpec `json:"spec"`
	MsgType      string             `json:"msgtype"`
}

func CreateSubmitFunctionSpecMsg(funcSpec *core.FunctionSpec) *SubmitFunctionSpecMsg {
	msg := &SubmitFunctionSpecMsg{}
	msg.FunctionSpec = funcSpec
	msg.MsgType = SubmitFunctionSpecPayloadType

	return msg
}

func (msg *SubmitFunctionSpecMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitFunctionSpecMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitFunctionSpecMsg) Equals(msg2 *SubmitFunctionSpecMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.FunctionSpec.Equals(msg2.FunctionSpec) {
		return true
	}

	return false
}

func CreateSubmitFunctionSpecMsgFromJSON(jsonString string) (*SubmitFunctionSpecMsg, error) {
	var msg *SubmitFunctionSpecMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
