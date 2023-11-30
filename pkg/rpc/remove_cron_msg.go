package rpc

import (
	"encoding/json"
)

const RemoveCronPayloadType = "removecronmsg"

type RemoveCronMsg struct {
	CronID  string `json:"cronid"`
	MsgType string `json:"msgtype"`
	All     bool   `json:"all"`
}

func CreateRemoveCronMsg(cronID string) *RemoveCronMsg {
	msg := &RemoveCronMsg{}
	msg.CronID = cronID
	msg.MsgType = RemoveCronPayloadType

	return msg
}

func (msg *RemoveCronMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveCronMsg) Equals(msg2 *RemoveCronMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.CronID == msg2.CronID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *RemoveCronMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveCronMsgFromJSON(jsonString string) (*RemoveCronMsg, error) {
	var msg *RemoveCronMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
