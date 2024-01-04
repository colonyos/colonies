package rpc

import (
	"encoding/json"
)

const GetCronPayloadType = "getcronmsg"

type GetCronMsg struct {
	CronID  string `json:"cronid"`
	MsgType string `json:"msgtype"`
}

func CreateGetCronMsg(cronName string) *GetCronMsg {
	msg := &GetCronMsg{}
	msg.CronID = cronName
	msg.MsgType = GetCronPayloadType

	return msg
}

func (msg *GetCronMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetCronMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetCronMsg) Equals(msg2 *GetCronMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.CronID == msg2.CronID {
		return true
	}

	return false
}

func CreateGetCronMsgFromJSON(jsonString string) (*GetCronMsg, error) {
	var msg *GetCronMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
