package rpc

import (
	"encoding/json"
)

const RemoveResourceDefinitionPayloadType = "removeresourcedefinitionmsg"

type RemoveResourceDefinitionMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveResourceDefinitionMsg(namespace, name string) *RemoveResourceDefinitionMsg {
	msg := &RemoveResourceDefinitionMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveResourceDefinitionPayloadType

	return msg
}

func (msg *RemoveResourceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveResourceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveResourceDefinitionMsg) Equals(msg2 *RemoveResourceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveResourceDefinitionMsgFromJSON(jsonString string) (*RemoveResourceDefinitionMsg, error) {
	var msg *RemoveResourceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
