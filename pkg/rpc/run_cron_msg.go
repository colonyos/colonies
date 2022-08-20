package rpc

import (
	"encoding/json"
)

const RunCronPayloadType = "runcronmsg"

type RunCronMsg struct {
	CronID  string `json:"cronid"`
	MsgType string `json:"msgtype"`
}

func CreateRunCronMsg(cronID string) *RunCronMsg {
	msg := &RunCronMsg{}
	msg.CronID = cronID
	msg.MsgType = RunCronPayloadType

	return msg
}

func (msg *RunCronMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RunCronMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RunCronMsg) Equals(msg2 *RunCronMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.CronID == msg2.CronID {
		return true
	}

	return false
}

func CreateRunCronMsgFromJSON(jsonString string) (*RunCronMsg, error) {
	var msg *RunCronMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
