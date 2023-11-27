package rpc

import (
	"encoding/json"
)

const GetUsersPayloadType = "getusersmsg"

type GetUsersMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
}

func CreateGetUsersMsg(colonyName string) *GetUsersMsg {
	msg := &GetUsersMsg{}
	msg.MsgType = GetUsersPayloadType
	msg.ColonyName = colonyName

	return msg
}

func (msg *GetUsersMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUsersMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUsersMsg) Equals(msg2 *GetUsersMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType {
		return true
	}

	return false
}

func CreateGetUsersMsgFromJSON(jsonString string) (*GetUsersMsg, error) {
	var msg *GetUsersMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
