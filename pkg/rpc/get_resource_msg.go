package rpc

import (
	"encoding/json"
)

const GetResourcePayloadType = "getresourcemsg"

type GetResourceMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateGetResourceMsg(namespace, name string) *GetResourceMsg {
	msg := &GetResourceMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = GetResourcePayloadType

	return msg
}

func (msg *GetResourceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceMsg) Equals(msg2 *GetResourceMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateGetResourceMsgFromJSON(jsonString string) (*GetResourceMsg, error) {
	var msg *GetResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
