package rpc

import (
	"encoding/json"
)

const RemoveExecutorPayloadType = "removeexecutormsg"

type RemoveExecutorMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateRemoveExecutorMsg(colonyName string, executorName string) *RemoveExecutorMsg {
	msg := &RemoveExecutorMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.MsgType = RemoveExecutorPayloadType

	return msg
}

func (msg *RemoveExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveExecutorMsg) Equals(msg2 *RemoveExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorName == msg2.ExecutorName && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateRemoveExecutorMsgFromJSON(jsonString string) (*RemoveExecutorMsg, error) {
	var msg *RemoveExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
