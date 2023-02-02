package rpc

import (
	"encoding/json"
)

const RejectExecutorPayloadType = "rejectexecutormsg"

type RejectExecutorMsg struct {
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateRejectExecutorMsg(executorID string) *RejectExecutorMsg {
	msg := &RejectExecutorMsg{}
	msg.ExecutorID = executorID
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

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID {
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
