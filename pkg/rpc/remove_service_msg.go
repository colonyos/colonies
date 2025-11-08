package rpc

import (
	"encoding/json"
)

const RemoveResourcePayloadType = "removeresourcemsg"

type RemoveResourceMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveResourceMsg(namespace, name string) *RemoveResourceMsg {
	msg := &RemoveResourceMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveResourcePayloadType

	return msg
}

func (msg *RemoveResourceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveResourceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveResourceMsg) Equals(msg2 *RemoveResourceMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveResourceMsgFromJSON(jsonString string) (*RemoveResourceMsg, error) {
	var msg *RemoveResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
