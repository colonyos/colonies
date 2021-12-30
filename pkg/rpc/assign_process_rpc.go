package rpc

import (
	"encoding/json"
)

const AssignProcessMsgType = "AssignProcessSpec"

type AssignProcessMsg struct {
	RPC      RPC    `json:"rpc"`
	ColonyID string `json:"colonyid"`
}

func CreateAssignProcessMsg(colonyID string) *AssignProcessMsg {
	msg := &AssignProcessMsg{}
	msg.RPC.Method = AssignProcessMsgType
	msg.ColonyID = colonyID

	return msg
}

func (msg *AssignProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAssignProcessMsgFromJSON(jsonString string) (*AssignProcessMsg, error) {
	var msg *AssignProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}