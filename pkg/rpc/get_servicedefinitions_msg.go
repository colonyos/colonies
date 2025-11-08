package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const GetResourceDefinitionsPayloadType = "getresourcedefinitionsmsg"

type GetResourceDefinitionsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateGetResourceDefinitionsMsg(colonyName string) *GetResourceDefinitionsMsg {
	msg := &GetResourceDefinitionsMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = GetResourceDefinitionsPayloadType

	return msg
}

func (msg *GetResourceDefinitionsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceDefinitionsMsg) Equals(msg2 *GetResourceDefinitionsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetResourceDefinitionsMsgFromJSON(jsonString string) (*GetResourceDefinitionsMsg, error) {
	var msg *GetResourceDefinitionsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type GetResourceDefinitionsReplyMsg struct {
	ResourceDefinitions []*core.ResourceDefinition `json:"resourcedefinitions"`
	MsgType             string                     `json:"msgtype"`
}

func CreateGetResourceDefinitionsReplyMsg(rds []*core.ResourceDefinition) *GetResourceDefinitionsReplyMsg {
	msg := &GetResourceDefinitionsReplyMsg{}
	msg.ResourceDefinitions = rds
	msg.MsgType = GetResourceDefinitionsPayloadType

	return msg
}

func (msg *GetResourceDefinitionsReplyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourceDefinitionsReplyMsg) Equals(msg2 *GetResourceDefinitionsReplyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if len(msg.ResourceDefinitions) != len(msg2.ResourceDefinitions) {
		return false
	}

	// Simple comparison - check IDs match
	for i := range msg.ResourceDefinitions {
		if msg.ResourceDefinitions[i].ID != msg2.ResourceDefinitions[i].ID {
			return false
		}
	}

	return true
}

func CreateGetResourceDefinitionsReplyMsgFromJSON(jsonString string) (*GetResourceDefinitionsReplyMsg, error) {
	var msg *GetResourceDefinitionsReplyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
