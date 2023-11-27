package rpc

import (
	"encoding/json"
)

const GetUserPayloadType = "getusermsg"

type GetUserMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
}

func CreateGetUserMsg(colonyName string, name string) *GetUserMsg {
	msg := &GetUserMsg{}
	msg.MsgType = GetUserPayloadType
	msg.ColonyName = colonyName
	msg.Name = name

	return msg
}

func (msg *GetUserMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUserMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUserMsg) Equals(msg2 *GetUserMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Name == msg2.Name && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetUserMsgFromJSON(jsonString string) (*GetUserMsg, error) {
	var msg *GetUserMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
