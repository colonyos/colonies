package rpc

import (
	"encoding/json"
)

const RemoveServiceDefinitionPayloadType = "removeservicedefinitionmsg"

type RemoveServiceDefinitionMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveServiceDefinitionMsg(namespace, name string) *RemoveServiceDefinitionMsg {
	msg := &RemoveServiceDefinitionMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveServiceDefinitionPayloadType

	return msg
}

func (msg *RemoveServiceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveServiceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveServiceDefinitionMsg) Equals(msg2 *RemoveServiceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveServiceDefinitionMsgFromJSON(jsonString string) (*RemoveServiceDefinitionMsg, error) {
	var msg *RemoveServiceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
