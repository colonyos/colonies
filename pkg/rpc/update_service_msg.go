package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const UpdateResourcePayloadType = "updateresourcemsg"

type UpdateResourceMsg struct {
	Service *core.Service `json:"service"`
	MsgType  string         `json:"msgtype"`
}

func CreateUpdateResourceMsg(service *core.Service) *UpdateResourceMsg {
	msg := &UpdateResourceMsg{}
	msg.Service = service
	msg.MsgType = UpdateResourcePayloadType

	return msg
}

func (msg *UpdateResourceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateResourceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateResourceMsg) Equals(msg2 *UpdateResourceMsg) bool {
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

func CreateUpdateResourceMsgFromJSON(jsonString string) (*UpdateResourceMsg, error) {
	var msg *UpdateResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
