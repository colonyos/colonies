package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddResourcePayloadType = "addresourcemsg"

type AddResourceMsg struct {
	Service *core.Service `json:"service"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddResourceMsg(service *core.Service) *AddResourceMsg {
	msg := &AddResourceMsg{}
	msg.Service = service
	msg.MsgType = AddResourcePayloadType

	return msg
}

func (msg *AddResourceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddResourceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddResourceMsg) Equals(msg2 *AddResourceMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.Service == nil && msg2.Service == nil {
		return true
	}

	if msg.Service == nil || msg2.Service == nil {
		return false
	}

	return msg.Service.ID == msg2.Service.ID
}

func CreateAddResourceMsgFromJSON(jsonString string) (*AddResourceMsg, error) {
	var msg *AddResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
