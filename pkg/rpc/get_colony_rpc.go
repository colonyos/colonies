package rpc

import (
	"encoding/json"
)

const GetColonyMsgType = "GetColony"

type GetColonyMsg struct {
	RPC      RPC    `json:"rpc"`
	ColonyID string `json:"colonyid"`
}

func CreateGetColonyMsg(colonyID string) *GetColonyMsg {
	msg := &GetColonyMsg{}
	msg.RPC.Method = GetColonyMsgType
	msg.ColonyID = colonyID

	return msg
}

func (msg *GetColonyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetColonyMsgFromJSON(jsonString string) (*GetColonyMsg, error) {
	var msg *GetColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
