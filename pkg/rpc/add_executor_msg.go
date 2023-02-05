package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddExecutorPayloadType = "addexecutormsg"

type AddExecutorMsg struct {
	Executor *core.Executor `json:"executor"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddExecutorMsg(executor *core.Executor) *AddExecutorMsg {
	msg := &AddExecutorMsg{}
	msg.Executor = executor
	msg.MsgType = AddExecutorPayloadType

	return msg
}

func (msg *AddExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddExecutorMsg) Equals(msg2 *AddExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Executor.Equals(msg2.Executor) {
		return true
	}

	return false
}

func CreateAddExecutorMsgFromJSON(jsonString string) (*AddExecutorMsg, error) {
	var msg *AddExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
