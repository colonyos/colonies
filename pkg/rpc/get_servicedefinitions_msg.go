package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const GetServiceDefinitionsPayloadType = "getservicedefinitionsmsg"

type GetServiceDefinitionsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateGetServiceDefinitionsMsg(colonyName string) *GetServiceDefinitionsMsg {
	msg := &GetServiceDefinitionsMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = GetServiceDefinitionsPayloadType

	return msg
}

func (msg *GetServiceDefinitionsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceDefinitionsMsg) Equals(msg2 *GetServiceDefinitionsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetServiceDefinitionsMsgFromJSON(jsonString string) (*GetServiceDefinitionsMsg, error) {
	var msg *GetServiceDefinitionsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type GetServiceDefinitionsReplyMsg struct {
	ServiceDefinitions []*core.ServiceDefinition `json:"servicedefinitions"`
	MsgType            string                    `json:"msgtype"`
}

func CreateGetServiceDefinitionsReplyMsg(sds []*core.ServiceDefinition) *GetServiceDefinitionsReplyMsg {
	msg := &GetServiceDefinitionsReplyMsg{}
	msg.ServiceDefinitions = sds
	msg.MsgType = GetServiceDefinitionsPayloadType

	return msg
}

func (msg *GetServiceDefinitionsReplyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceDefinitionsReplyMsg) Equals(msg2 *GetServiceDefinitionsReplyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if len(msg.ServiceDefinitions) != len(msg2.ServiceDefinitions) {
		return false
	}

	// Simple comparison - check IDs match
	for i := range msg.ServiceDefinitions {
		if msg.ServiceDefinitions[i].ID != msg2.ServiceDefinitions[i].ID {
			return false
		}
	}

	return true
}

func CreateGetServiceDefinitionsReplyMsgFromJSON(jsonString string) (*GetServiceDefinitionsReplyMsg, error) {
	var msg *GetServiceDefinitionsReplyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
