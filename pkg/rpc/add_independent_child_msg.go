package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddIndependentChildPayloadType = "addindependentchildmsg"

type AddIndependentChildMsg struct {
	ProcessGraphID  string             `json:"processgraphid"`
	ParentProcessID string             `json:"parentprocessid"`
	FunctionSpec    *core.FunctionSpec `json:"spec"`
	MsgType         string             `json:"msgtype"`
}

func CreateAddIndependentChildMsg(processGraphID string, parentProcessID string, funcSpec *core.FunctionSpec) *AddIndependentChildMsg {
	msg := &AddIndependentChildMsg{}
	msg.ProcessGraphID = processGraphID
	msg.ParentProcessID = parentProcessID
	msg.FunctionSpec = funcSpec
	msg.MsgType = AddIndependentChildPayloadType

	return msg
}

func (msg *AddIndependentChildMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddIndependentChildMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddIndependentChildMsg) Equals(msg2 *AddIndependentChildMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessGraphID == msg2.ProcessGraphID &&
		msg.ParentProcessID == msg2.ParentProcessID &&
		msg.FunctionSpec.Equals(msg2.FunctionSpec) {
		return true
	}

	return false
}

func CreateAddIndependentChildMsgFromJSON(jsonString string) (*AddIndependentChildMsg, error) {
	var msg *AddIndependentChildMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
