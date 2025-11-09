package rpc

import (
	"encoding/json"
)

const RemoveBlueprintPayloadType = "removeblueprintmsg"

type RemoveBlueprintMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveBlueprintMsg(namespace, name string) *RemoveBlueprintMsg {
	msg := &RemoveBlueprintMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveBlueprintPayloadType

	return msg
}

func (msg *RemoveBlueprintMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveBlueprintMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveBlueprintMsg) Equals(msg2 *RemoveBlueprintMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveBlueprintMsgFromJSON(jsonString string) (*RemoveBlueprintMsg, error) {
	var msg *RemoveBlueprintMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
