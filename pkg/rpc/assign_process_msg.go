package rpc

import (
	"encoding/json"
)

const AssignProcessPayloadType = "assignprocessmsg"

type AssignProcessMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
	Timeout  int    `json:"timeout"`
}

func CreateAssignProcessMsg(colonyID string) *AssignProcessMsg {
	msg := &AssignProcessMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = AssignProcessPayloadType
	msg.Timeout = -1 // Not implemented yet

	return msg
}

func (msg *AssignProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AssignProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AssignProcessMsg) Equals(msg2 *AssignProcessMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateAssignProcessMsgFromJSON(jsonString string) (*AssignProcessMsg, error) {
	var msg *AssignProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
