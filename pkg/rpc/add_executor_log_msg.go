package rpc

import (
	"encoding/json"
)

const AddExecutorLogPayloadType = "addexecutorlogmsg"

type AddExecutorLogMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	Message      string `json:"message"`
	MsgType      string `json:"msgtype"`
}

func CreateAddExecutorLogMsg(colonyName, executorName, message string) *AddExecutorLogMsg {
	msg := &AddExecutorLogMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Message = message
	msg.MsgType = AddExecutorLogPayloadType

	return msg
}

func (msg *AddExecutorLogMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddExecutorLogMsg) Equals(msg2 *AddExecutorLogMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorName == msg2.ExecutorName &&
		msg.Message == msg2.Message {
		return true
	}

	return false
}

func (msg *AddExecutorLogMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddExecutorLogMsgFromJSON(jsonString string) (*AddExecutorLogMsg, error) {
	var msg *AddExecutorLogMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
