package rpc

import (
	"encoding/json"
)

const GetColonyPayloadType = "getcolonymsg"

type GetColonyMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetColonyMsg(colonyID string) *GetColonyMsg {
	msg := &GetColonyMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetColonyPayloadType

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

func (msg *GetColonyMsg) Equals(msg2 *GetColonyMsg) bool {
	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetColonyMsgFromJSON(jsonString string) (*GetColonyMsg, error) {
	var msg *GetColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
