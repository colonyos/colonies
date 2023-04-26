package rpc

import (
	"encoding/json"
)

const RenameColonyPayloadType = "renamecolonymsg"

type RenameColonyMsg struct {
	ColonyID string `json:"colonyid"`
	Name     string `json:"name"`
	MsgType  string `json:"msgtype"`
}

func CreateRenameColonyMsg(colonyID string, name string) *RenameColonyMsg {
	msg := &RenameColonyMsg{}
	msg.ColonyID = colonyID
	msg.Name = name
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

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID && msg.Name == msg2.Name {
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
