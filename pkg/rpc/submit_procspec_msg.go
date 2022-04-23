package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const SubmitProcessSpecPayloadType = "submitprocessespecmsg"

type SubmitProcessSpecMsg struct {
	ProcessSpec *core.ProcessSpec `json:"spec"`
	MsgType     string            `json:"msgtype"`
}

func CreateSubmitProcessSpecMsg(processSpec *core.ProcessSpec) *SubmitProcessSpecMsg {
	msg := &SubmitProcessSpecMsg{}
	msg.ProcessSpec = processSpec
	msg.MsgType = SubmitProcessSpecPayloadType

	return msg
}

func (msg *SubmitProcessSpecMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitProcessSpecMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitProcessSpecMsg) Equals(msg2 *SubmitProcessSpecMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessSpec.Equals(msg2.ProcessSpec) {
		return true
	}

	return false
}

func CreateSubmitProcessSpecMsgFromJSON(jsonString string) (*SubmitProcessSpecMsg, error) {
	var msg *SubmitProcessSpecMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
