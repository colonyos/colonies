package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddServicePayloadType = "addservicemsg"

type AddServiceMsg struct {
	Service *core.Service `json:"service"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddServiceMsg(service *core.Service) *AddServiceMsg {
	msg := &AddServiceMsg{}
	msg.Service = service
	msg.MsgType = AddServicePayloadType

	return msg
}

func (msg *AddServiceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddServiceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddServiceMsg) Equals(msg2 *AddServiceMsg) bool {
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

func CreateAddServiceMsgFromJSON(jsonString string) (*AddServiceMsg, error) {
	var msg *AddServiceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
