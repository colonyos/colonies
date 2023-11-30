package rpc

import (
	"encoding/json"
)

const RemoveFunctionPayloadType = "removefunctionmsg"

type RemoveFunctionMsg struct {
	FunctionID string `json:"functionid"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveFunctionMsg(functionID string) *RemoveFunctionMsg {
	msg := &RemoveFunctionMsg{}
	msg.FunctionID = functionID
	msg.MsgType = RemoveFunctionPayloadType

	return msg
}

func (msg *RemoveFunctionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveFunctionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveFunctionMsg) Equals(msg2 *RemoveFunctionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.FunctionID == msg2.FunctionID {
		return true
	}

	return false
}

func CreateRemoveFunctionMsgFromJSON(jsonString string) (*RemoveFunctionMsg, error) {
	var msg *RemoveFunctionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
