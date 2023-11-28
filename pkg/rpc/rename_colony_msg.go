package rpc

import (
	"encoding/json"
)

const RenameColonyPayloadType = "renamecolonymsg"

type RenameColonyMsg struct {
	OldName string `json:"oldname"`
	NewName string `json:"newname"`
	MsgType string `json:"msgtype"`
}

func CreateRenameColonyMsg(oldName string, newName string) *RenameColonyMsg {
	msg := &RenameColonyMsg{}
	msg.OldName = oldName
	msg.NewName = newName
	msg.MsgType = RenameColonyPayloadType

	return msg
}

func (msg *RenameColonyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RenameColonyMsg) Equals(msg2 *RenameColonyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.OldName == msg2.OldName && msg.NewName == msg2.NewName {
		return true
	}

	return false
}

func (msg *RenameColonyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRenameColonyMsgFromJSON(jsonString string) (*RenameColonyMsg, error) {
	var msg *RenameColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
