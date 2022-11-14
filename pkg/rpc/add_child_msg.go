package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddChildPayloadType = "addchildmsg"

type AddChildMsg struct {
	ProcessGraphID string            `json:"processgraphid"`
	ProcessID      string            `json:"processid"`
	ProcessSpec    *core.ProcessSpec `json:"spec"`
	MsgType        string            `json:"msgtype"`
}

func CreateAddChildMsg(processGraphID string, processID string, processSpec *core.ProcessSpec) *AddChildMsg {
	msg := &AddChildMsg{}
	msg.ProcessGraphID = processGraphID
	msg.ProcessID = processID
	msg.ProcessSpec = processSpec
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

	if msg.MsgType == msg2.MsgType && msg.ProcessGraphID == msg2.ProcessGraphID && msg.ProcessID == msg2.ProcessID && msg.ProcessSpec.Equals(msg2.ProcessSpec) {
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
