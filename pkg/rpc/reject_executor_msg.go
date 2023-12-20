package rpc

import (
	"encoding/json"
)

const RejectExecutorPayloadType = "rejectexecutormsg"

type RejectExecutorMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateRejectExecutorMsg(colonyName string, executorName string) *RejectExecutorMsg {
	msg := &RejectExecutorMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.MsgType = RejectExecutorPayloadType

	return msg
}

func (msg *RejectExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RejectExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RejectExecutorMsg) Equals(msg2 *RejectExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorName == msg2.ExecutorName && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateRejectExecutorMsgFromJSON(jsonString string) (*RejectExecutorMsg, error) {
	var msg *RejectExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
