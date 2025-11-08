package rpc

import (
	"encoding/json"
)

const RemoveServicePayloadType = "removeservicemsg"

type RemoveServiceMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveServiceMsg(namespace, name string) *RemoveServiceMsg {
	msg := &RemoveServiceMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveServicePayloadType

	return msg
}

func (msg *RemoveServiceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveServiceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveServiceMsg) Equals(msg2 *RemoveServiceMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveServiceMsgFromJSON(jsonString string) (*RemoveServiceMsg, error) {
	var msg *RemoveServiceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
