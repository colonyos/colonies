package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddBlueprintDefinitionPayloadType = "addblueprintdefinitionmsg"

type AddBlueprintDefinitionMsg struct {
	BlueprintDefinition *core.BlueprintDefinition `json:"blueprintdefinition"`
	MsgType           string                  `json:"msgtype"`
}

func CreateAddBlueprintDefinitionMsg(sd *core.BlueprintDefinition) *AddBlueprintDefinitionMsg {
	msg := &AddBlueprintDefinitionMsg{}
	msg.BlueprintDefinition = sd
	msg.MsgType = AddBlueprintDefinitionPayloadType

	return msg
}

func (msg *AddBlueprintDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddBlueprintDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddBlueprintDefinitionMsg) Equals(msg2 *AddBlueprintDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.BlueprintDefinition == nil && msg2.BlueprintDefinition == nil {
		return true
	}

	if msg.BlueprintDefinition == nil || msg2.BlueprintDefinition == nil {
		return false
	}

	return msg.BlueprintDefinition.ID == msg2.BlueprintDefinition.ID
}

func CreateAddBlueprintDefinitionMsgFromJSON(jsonString string) (*AddBlueprintDefinitionMsg, error) {
	var msg *AddBlueprintDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
