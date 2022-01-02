package rpc

import (
	"colonies/pkg/core"
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
	msg.RPC.Nonce = core.GenerateRandomID()
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

func (msg *GetColonyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
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
