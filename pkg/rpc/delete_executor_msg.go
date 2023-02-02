package rpc

import (
	"encoding/json"
)

const DeleteExecutorPayloadType = "deleteexecutormsg"

type DeleteExecutorMsg struct {
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateDeleteExecutorMsg(executorID string) *DeleteExecutorMsg {
	msg := &DeleteExecutorMsg{}
	msg.ExecutorID = executorID
	msg.MsgType = DeleteExecutorPayloadType

	return msg
}

func (msg *DeleteExecutorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteExecutorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteExecutorMsg) Equals(msg2 *DeleteExecutorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID {
		return true
	}

	return false
}

func CreateDeleteExecutorMsgFromJSON(jsonString string) (*DeleteExecutorMsg, error) {
	var msg *DeleteExecutorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
