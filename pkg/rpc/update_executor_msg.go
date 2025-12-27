package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const UpdateExecutorPayloadType = "updateexecutormsg"

type UpdateExecutorMsg struct {
	ColonyName   string           `json:"colonyname"`
	ExecutorName string           `json:"executorname"`
	Capabilities core.Capabilities `json:"capabilities"`
	MsgType      string           `json:"msgtype"`
}

func CreateUpdateExecutorMsg(colonyName string, executorName string, capabilities core.Capabilities) *UpdateExecutorMsg {
	msg := &UpdateExecutorMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Capabilities = capabilities
	msg.MsgType = UpdateExecutorPayloadType

	return msg
}

func (msg *UpdateExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateExecutorMsg) Equals(msg2 *UpdateExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorName == msg2.ExecutorName {
		return true
	}

	return false
}

func CreateUpdateExecutorMsgFromJSON(jsonString string) (*UpdateExecutorMsg, error) {
	var msg *UpdateExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
