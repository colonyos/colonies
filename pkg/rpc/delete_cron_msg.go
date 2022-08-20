package rpc

import (
	"encoding/json"
)

const DeleteCronPayloadType = "deletecronmsg"

type DeleteCronMsg struct {
	CronID  string `json:"cronid"`
	MsgType string `json:"msgtype"`
	All     bool   `json:"all"`
}

func CreateDeleteCronMsg(cronID string) *DeleteCronMsg {
	msg := &DeleteCronMsg{}
	msg.CronID = cronID
	msg.MsgType = DeleteCronPayloadType

	return msg
}

func (msg *DeleteCronMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteCronMsg) Equals(msg2 *DeleteCronMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.CronID == msg2.CronID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *DeleteCronMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteCronMsgFromJSON(jsonString string) (*DeleteCronMsg, error) {
	var msg *DeleteCronMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
