package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const GetBlueprintDefinitionsPayloadType = "getblueprintdefinitionsmsg"

type GetBlueprintDefinitionsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateGetBlueprintDefinitionsMsg(colonyName string) *GetBlueprintDefinitionsMsg {
	msg := &GetBlueprintDefinitionsMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = GetBlueprintDefinitionsPayloadType

	return msg
}

func (msg *GetBlueprintDefinitionsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintDefinitionsMsg) Equals(msg2 *GetBlueprintDefinitionsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetBlueprintDefinitionsMsgFromJSON(jsonString string) (*GetBlueprintDefinitionsMsg, error) {
	var msg *GetBlueprintDefinitionsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

type GetBlueprintDefinitionsReplyMsg struct {
	BlueprintDefinitions []*core.BlueprintDefinition `json:"blueprintdefinitions"`
	MsgType            string                    `json:"msgtype"`
}

func CreateGetBlueprintDefinitionsReplyMsg(sds []*core.BlueprintDefinition) *GetBlueprintDefinitionsReplyMsg {
	msg := &GetBlueprintDefinitionsReplyMsg{}
	msg.BlueprintDefinitions = sds
	msg.MsgType = GetBlueprintDefinitionsPayloadType

	return msg
}

func (msg *GetBlueprintDefinitionsReplyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintDefinitionsReplyMsg) Equals(msg2 *GetBlueprintDefinitionsReplyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if len(msg.BlueprintDefinitions) != len(msg2.BlueprintDefinitions) {
		return false
	}

	// Simple comparison - check IDs match
	for i := range msg.BlueprintDefinitions {
		if msg.BlueprintDefinitions[i].ID != msg2.BlueprintDefinitions[i].ID {
			return false
		}
	}

	return true
}

func CreateGetBlueprintDefinitionsReplyMsgFromJSON(jsonString string) (*GetBlueprintDefinitionsReplyMsg, error) {
	var msg *GetBlueprintDefinitionsReplyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
