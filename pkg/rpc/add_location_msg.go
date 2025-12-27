package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddLocationPayloadType = "addlocationmsg"

type AddLocationMsg struct {
	Location *core.Location `json:"location"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddLocationMsg(location *core.Location) *AddLocationMsg {
	msg := &AddLocationMsg{}
	msg.Location = location
	msg.MsgType = AddLocationPayloadType

	return msg
}

func (msg *AddLocationMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddLocationMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddLocationMsg) Equals(msg2 *AddLocationMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Location.Equals(msg2.Location) {
		return true
	}

	return false
}

func CreateAddLocationMsgFromJSON(jsonString string) (*AddLocationMsg, error) {
	var msg *AddLocationMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
