package rpc

import (
	"encoding/json"
)

const GetExecutorPayloadType = "getexecutormsg"

type GetExecutorMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateGetExecutorMsg(colonyName string, executorName string) *GetExecutorMsg {
	msg := &GetExecutorMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
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

	if msg.MsgType == msg2.MsgType && msg.ExecutorName == msg2.ExecutorName {
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
