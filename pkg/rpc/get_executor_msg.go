package rpc

import (
	"encoding/json"
)

const GetExecutorPayloadType = "getexecutormsg"

type GetExecutorMsg struct {
	ExecutorID string `json:"executorname"`
	MsgType    string `json:"msgtype"`
}

func CreateGetExecutorMsg(executorID string) *GetExecutorMsg {
	msg := &GetExecutorMsg{}
	msg.ExecutorID = executorID
	msg.MsgType = GetExecutorPayloadType

	return msg
}

func (msg *GetExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorMsg) Equals(msg2 *GetExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID {
		return true
	}

	return false
}

func CreateGetExecutorMsgFromJSON(jsonString string) (*GetExecutorMsg, error) {
	var msg *GetExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
