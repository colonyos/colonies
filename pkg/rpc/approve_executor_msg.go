package rpc

import (
	"encoding/json"
)

const ApproveExecutorPayloadType = "approveexecutormsg"

type ApproveExecutorRPC struct {
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateApproveExecutorMsg(executorID string) *ApproveExecutorRPC {
	msg := &ApproveExecutorRPC{}
	msg.ExecutorID = executorID
	msg.MsgType = ApproveExecutorPayloadType

	return msg
}

func (msg *ApproveExecutorRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ApproveExecutorRPC) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ApproveExecutorRPC) Equals(msg2 *ApproveExecutorRPC) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID {
		return true
	}

	return false
}

func CreateApproveExecutorMsgFromJSON(jsonString string) (*ApproveExecutorRPC, error) {
	var msg *ApproveExecutorRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
