package rpc

import (
	"encoding/json"
)

const ResetDatabasePayloadType = "resetdatabasemsg"

type ResetDatabaseMsg struct {
	MsgType string `json:"msgtype"`
}

func CreateResetDatabaseMsg() *ResetDatabaseMsg {
	msg := &ResetDatabaseMsg{}
	msg.MsgType = ResetDatabasePayloadType

	return msg
}

func (msg *ResetDatabaseMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResetDatabaseMsg) Equals(msg2 *ResetDatabaseMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType {
		return true
	}

	return false
}

func (msg *ResetDatabaseMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateResetDatabaseMsgFromJSON(jsonString string) (*ResetDatabaseMsg, error) {
	var msg *ResetDatabaseMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
