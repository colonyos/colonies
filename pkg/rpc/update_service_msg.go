package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const UpdateServicePayloadType = "updateservicemsg"

type UpdateServiceMsg struct {
	Service *core.Service `json:"service"`
	MsgType  string         `json:"msgtype"`
}

func CreateUpdateServiceMsg(service *core.Service) *UpdateServiceMsg {
	msg := &UpdateServiceMsg{}
	msg.Service = service
	msg.MsgType = UpdateServicePayloadType

	return msg
}

func (msg *UpdateServiceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateServiceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateServiceMsg) Equals(msg2 *UpdateServiceMsg) bool {
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

func CreateUpdateServiceMsgFromJSON(jsonString string) (*UpdateServiceMsg, error) {
	var msg *UpdateServiceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
