package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddUserPayloadType = "addusermsg"

type AddUserMsg struct {
	User    *core.User `json:"user"`
	MsgType string     `json:"msgtype"`
}

func CreateAddUserMsg(user *core.User) *AddUserMsg {
	msg := &AddUserMsg{}
	msg.User = user
	msg.MsgType = AddUserPayloadType

	return msg
}

func (msg *AddUserMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddUserMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddUserMsg) Equals(msg2 *AddUserMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.User.Equals(msg2.User) {
		return true
	}

	return false
}

func CreateAddUserMsgFromJSON(jsonString string) (*AddUserMsg, error) {
	var msg *AddUserMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
