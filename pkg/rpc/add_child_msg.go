package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddChildPayloadType = "addchildmsg"

type AddChildMsg struct {
	ProcessGraphID  string             `json:"processgraphid"`
	ParentProcessID string             `json:"parentprocessid"`
	ChildProcessID  string             `json:"childprocessid"`
	FunctionSpec    *core.FunctionSpec `json:"spec"`
	Insert          bool               `json:"insert"`
	MsgType         string             `json:"msgtype"`
}

func CreateAddChildMsg(processGraphID string, parentProcessID string, childProcessID string, funcSpec *core.FunctionSpec, insert bool) *AddChildMsg {
	msg := &AddChildMsg{}
	msg.ProcessGraphID = processGraphID
	msg.ParentProcessID = parentProcessID
	msg.ChildProcessID = childProcessID
	msg.FunctionSpec = funcSpec
	msg.Insert = insert
	msg.MsgType = AddChildPayloadType

	return msg
}

func (msg *AddChildMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddChildMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddChildMsg) Equals(msg2 *AddChildMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessGraphID == msg2.ProcessGraphID &&
		msg.ParentProcessID == msg2.ParentProcessID &&
		msg.ChildProcessID == msg2.ChildProcessID &&
		msg.FunctionSpec.Equals(msg2.FunctionSpec) {
		return true
	}

	return false
}

func CreateAddChildMsgFromJSON(jsonString string) (*AddChildMsg, error) {
	var msg *AddChildMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
