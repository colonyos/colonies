package rpc

import (
	"encoding/json"
)

const GetServiceDefinitionPayloadType = "getservicedefinitionmsg"

type GetServiceDefinitionMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateGetServiceDefinitionMsg(colonyName, name string) *GetServiceDefinitionMsg {
	msg := &GetServiceDefinitionMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = GetServiceDefinitionPayloadType

	return msg
}

func (msg *GetServiceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceDefinitionMsg) Equals(msg2 *GetServiceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Name == msg2.Name
}

func CreateGetServiceDefinitionMsgFromJSON(jsonString string) (*GetServiceDefinitionMsg, error) {
	var msg *GetServiceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
