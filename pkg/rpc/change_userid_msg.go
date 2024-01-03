package rpc

import (
	"encoding/json"
)

const ChangeUserIDPayloadType = "changeuseridmsg"

type ChangeUserIDMsg struct {
	ColonyName string `json:"colonyname"`
	UserID     string `json:"userid"`
	MsgType    string `json:"msgtype"`
}

func CreateChangeUserIDMsg(colonyName, userID string) *ChangeUserIDMsg {
	msg := &ChangeUserIDMsg{}
	msg.UserID = userID
	msg.ColonyName = colonyName
	msg.MsgType = ChangeUserIDPayloadType

	return msg
}

func (msg *ChangeUserIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeUserIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeUserIDMsg) Equals(msg2 *ChangeUserIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.UserID == msg2.UserID && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateChangeUserIDMsgFromJSON(jsonString string) (*ChangeUserIDMsg, error) {
	var msg *ChangeUserIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
