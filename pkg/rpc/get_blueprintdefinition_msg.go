package rpc

import (
	"encoding/json"
)

const GetBlueprintDefinitionPayloadType = "getblueprintdefinitionmsg"

type GetBlueprintDefinitionMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateGetBlueprintDefinitionMsg(colonyName, name string) *GetBlueprintDefinitionMsg {
	msg := &GetBlueprintDefinitionMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = GetBlueprintDefinitionPayloadType

	return msg
}

func (msg *GetBlueprintDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintDefinitionMsg) Equals(msg2 *GetBlueprintDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Name == msg2.Name
}

func CreateGetBlueprintDefinitionMsgFromJSON(jsonString string) (*GetBlueprintDefinitionMsg, error) {
	var msg *GetBlueprintDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
