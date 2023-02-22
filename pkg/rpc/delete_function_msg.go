package rpc

import (
	"encoding/json"
)

const DeleteFunctionPayloadType = "deletefunctionmsg"

type DeleteFunctionMsg struct {
	Name    string `json:"name"`
	MsgType string `json:"msgtype"`
}

func CreateDeleteFunctionMsg(name string) *DeleteFunctionMsg {
	msg := &DeleteFunctionMsg{}
	msg.Name = name
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

	if msg.MsgType == msg2.MsgType && msg.Name == msg2.Name {
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
