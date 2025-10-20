package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const UpdateResourcePayloadType = "updateresourcemsg"

type UpdateResourceMsg struct {
	Resource *core.Resource `json:"resource"`
	MsgType  string         `json:"msgtype"`
}

func CreateUpdateResourceMsg(resource *core.Resource) *UpdateResourceMsg {
	msg := &UpdateResourceMsg{}
	msg.Resource = resource
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

	if msg.Resource == nil && msg2.Resource == nil {
		return true
	}

	if msg.Resource == nil || msg2.Resource == nil {
		return false
	}

	return msg.Resource.ID == msg2.Resource.ID
}

func CreateUpdateResourceMsgFromJSON(jsonString string) (*UpdateResourceMsg, error) {
	var msg *UpdateResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
