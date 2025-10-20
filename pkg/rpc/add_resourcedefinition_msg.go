package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddResourceDefinitionPayloadType = "addresourcedefinitionmsg"

type AddResourceDefinitionMsg struct {
	ResourceDefinition *core.ResourceDefinition `json:"resourcedefinition"`
	MsgType            string                   `json:"msgtype"`
}

func CreateAddResourceDefinitionMsg(rd *core.ResourceDefinition) *AddResourceDefinitionMsg {
	msg := &AddResourceDefinitionMsg{}
	msg.ResourceDefinition = rd
	msg.MsgType = AddResourceDefinitionPayloadType

	return msg
}

func (msg *AddResourceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddResourceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddResourceDefinitionMsg) Equals(msg2 *AddResourceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.ResourceDefinition == nil && msg2.ResourceDefinition == nil {
		return true
	}

	if msg.ResourceDefinition == nil || msg2.ResourceDefinition == nil {
		return false
	}

	return msg.ResourceDefinition.ID == msg2.ResourceDefinition.ID
}

func CreateAddResourceDefinitionMsgFromJSON(jsonString string) (*AddResourceDefinitionMsg, error) {
	var msg *AddResourceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
