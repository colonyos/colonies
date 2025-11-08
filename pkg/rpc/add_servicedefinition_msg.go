package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddServiceDefinitionPayloadType = "addservicedefinitionmsg"

type AddServiceDefinitionMsg struct {
	ServiceDefinition *core.ServiceDefinition `json:"servicedefinition"`
	MsgType           string                  `json:"msgtype"`
}

func CreateAddServiceDefinitionMsg(sd *core.ServiceDefinition) *AddServiceDefinitionMsg {
	msg := &AddServiceDefinitionMsg{}
	msg.ServiceDefinition = sd
	msg.MsgType = AddServiceDefinitionPayloadType

	return msg
}

func (msg *AddServiceDefinitionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddServiceDefinitionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddServiceDefinitionMsg) Equals(msg2 *AddServiceDefinitionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.ServiceDefinition == nil && msg2.ServiceDefinition == nil {
		return true
	}

	if msg.ServiceDefinition == nil || msg2.ServiceDefinition == nil {
		return false
	}

	return msg.ServiceDefinition.ID == msg2.ServiceDefinition.ID
}

func CreateAddServiceDefinitionMsgFromJSON(jsonString string) (*AddServiceDefinitionMsg, error) {
	var msg *AddServiceDefinitionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
