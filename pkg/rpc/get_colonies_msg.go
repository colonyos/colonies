package rpc

import (
	"encoding/json"
)

const GetColoniesPayloadType = "getcoloniesmsg"

type GetColoniesMsg struct {
	MsgType string `json:"msgtype"`
}

func CreateGetColoniesMsg() *GetColoniesMsg {
	msg := &GetColoniesMsg{}
	msg.MsgType = GetColoniesPayloadType

	return msg
}

func (msg *GetColoniesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetColoniesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetColoniesMsg) Equals(msg2 *GetColoniesMsg) bool {
	if msg.MsgType == msg2.MsgType {
		return true
	}

	return false
}

func CreateGetColoniesMsgFromJSON(jsonString string) (*GetColoniesMsg, error) {
	var msg *GetColoniesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
