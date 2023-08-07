package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddFilePayloadType = "addfilemsg"

type AddFileMsg struct {
	File    *core.File `json:"file"`
	MsgType string     `json:"msgtype"`
}

func CreateAddFileMsg(file *core.File) *AddFileMsg {
	msg := &AddFileMsg{}
	msg.File = file
	msg.MsgType = AddFilePayloadType

	return msg
}

func (msg *AddFileMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddFileMsg) Equals(msg2 *AddFileMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.File.Equals(msg2.File) {
		return true
	}

	return false
}

func (msg *AddFileMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddFileMsgFromJSON(jsonString string) (*AddFileMsg, error) {
	var msg *AddFileMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
