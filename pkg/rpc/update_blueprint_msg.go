package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const UpdateBlueprintPayloadType = "updateblueprintmsg"

type UpdateBlueprintMsg struct {
	Blueprint *core.Blueprint `json:"blueprint"`
	MsgType  string         `json:"msgtype"`
}

func CreateUpdateBlueprintMsg(blueprint *core.Blueprint) *UpdateBlueprintMsg {
	msg := &UpdateBlueprintMsg{}
	msg.Blueprint = blueprint
	msg.MsgType = UpdateBlueprintPayloadType

	return msg
}

func (msg *UpdateBlueprintMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateBlueprintMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateBlueprintMsg) Equals(msg2 *UpdateBlueprintMsg) bool {
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

func CreateUpdateBlueprintMsgFromJSON(jsonString string) (*UpdateBlueprintMsg, error) {
	var msg *UpdateBlueprintMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
