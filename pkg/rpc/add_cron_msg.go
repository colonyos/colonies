package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddCronPayloadType = "addcronmsg"

type AddCronMsg struct {
	Cron    *core.Cron `json:"cron"`
	MsgType string     `json:"msgtype"`
}

func CreateAddCronMsg(cron *core.Cron) *AddCronMsg {
	msg := &AddCronMsg{}
	msg.Cron = cron
	msg.MsgType = AddCronPayloadType

	return msg
}

func (msg *AddCronMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddCronMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddCronMsg) Equals(msg2 *AddCronMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Cron.Equals(msg2.Cron) {
		return true
	}

	return false
}

func CreateAddCronMsgFromJSON(jsonString string) (*AddCronMsg, error) {
	var msg *AddCronMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
