package rpc

import (
	"encoding/json"
)

const GetServerInfoPayloadType = "getserverinfomsg"

type GetServerInfoMsg struct {
	MsgType string `json:"msgtype"`
}

func CreateGetServerInfoMsg() *GetServerInfoMsg {
	msg := &GetServerInfoMsg{}
	msg.MsgType = GetServerInfoPayloadType
	return msg
}

func (msg *GetServerInfoMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func CreateGetServerInfoMsgFromJSON(jsonString string) (*GetServerInfoMsg, error) {
	var msg *GetServerInfoMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
