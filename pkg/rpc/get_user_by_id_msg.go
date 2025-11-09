package rpc

import (
	"encoding/json"
)

const GetUserByIDPayloadType = "getuserbyidmsg"

type GetUserByIDMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	UserID     string `json:"userid"`
}

func CreateGetUserByIDMsg(colonyName string, userID string) *GetUserByIDMsg {
	msg := &GetUserByIDMsg{}
	msg.MsgType = GetUserByIDPayloadType
	msg.ColonyName = colonyName
	msg.UserID = userID

	return msg
}

func (msg *GetUserByIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUserByIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetUserByIDMsg) Equals(msg2 *GetUserByIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.UserID == msg2.UserID && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetUserByIDMsgFromJSON(jsonString string) (*GetUserByIDMsg, error) {
	var msg *GetUserByIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
