package rpc

import (
	"encoding/json"
)

const RemoveBlueprintDefinitionPayloadType = "removeblueprintdefinitionmsg"

type RemoveBlueprintDefinitionMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateRemoveBlueprintDefinitionMsg(namespace, name string) *RemoveBlueprintDefinitionMsg {
	msg := &RemoveBlueprintDefinitionMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = RemoveBlueprintDefinitionPayloadType

	return msg
}

func (msg *RemoveBlueprintDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveBlueprintDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveBlueprintDefinitionMsg) Equals(msg2 *RemoveBlueprintDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateRemoveBlueprintDefinitionMsgFromJSON(jsonString string) (*RemoveBlueprintDefinitionMsg, error) {
	var msg *RemoveBlueprintDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
