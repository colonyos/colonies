package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddBlueprintPayloadType = "addblueprintmsg"

type AddBlueprintMsg struct {
	Blueprint *core.Blueprint `json:"blueprint"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddBlueprintMsg(blueprint *core.Blueprint) *AddBlueprintMsg {
	msg := &AddBlueprintMsg{}
	msg.Blueprint = blueprint
	msg.MsgType = AddBlueprintPayloadType

	return msg
}

func (msg *AddBlueprintMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddBlueprintMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddBlueprintMsg) Equals(msg2 *AddBlueprintMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.Blueprint == nil && msg2.Blueprint == nil {
		return true
	}

	if msg.Blueprint == nil || msg2.Blueprint == nil {
		return false
	}

	return msg.Blueprint.ID == msg2.Blueprint.ID
}

func CreateAddBlueprintMsgFromJSON(jsonString string) (*AddBlueprintMsg, error) {
	var msg *AddBlueprintMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
