package rpc

import (
	"encoding/json"
)

const GetResourceDefinitionPayloadType = "getresourcedefinitionmsg"

type GetResourceDefinitionMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateGetResourceDefinitionMsg(colonyName, name string) *GetResourceDefinitionMsg {
	msg := &GetResourceDefinitionMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = GetResourceDefinitionPayloadType

	return msg
}

func (msg *GetResourceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceDefinitionMsg) Equals(msg2 *GetResourceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Name == msg2.Name
}

func CreateGetResourceDefinitionMsgFromJSON(jsonString string) (*GetResourceDefinitionMsg, error) {
	var msg *GetResourceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
