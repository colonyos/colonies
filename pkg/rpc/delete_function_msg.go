package rpc

import (
	"encoding/json"
)

const DeleteFunctionPayloadType = "deletefunctionmsg"

type DeleteFunctionMsg struct {
	FunctionID string `json:"functionid"`
	MsgType    string `json:"msgtype"`
}

func CreateDeleteFunctionMsg(functionID string) *DeleteFunctionMsg {
	msg := &DeleteFunctionMsg{}
	msg.FunctionID = functionID
	msg.MsgType = DeleteFunctionPayloadType

	return msg
}

func (msg *DeleteFunctionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteFunctionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteFunctionMsg) Equals(msg2 *DeleteFunctionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.FunctionID == msg2.FunctionID {
		return true
	}

	return false
}

func CreateDeleteFunctionMsgFromJSON(jsonString string) (*DeleteFunctionMsg, error) {
	var msg *DeleteFunctionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
