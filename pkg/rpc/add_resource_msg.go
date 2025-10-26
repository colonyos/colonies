package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddResourcePayloadType = "addresourcemsg"

type AddResourceMsg struct {
	Resource *core.Resource `json:"resource"`
	MsgType  string         `json:"msgtype"`
}

func CreateAddResourceMsg(resource *core.Resource) *AddResourceMsg {
	msg := &AddResourceMsg{}
	msg.Resource = resource
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

	if msg.Resource == nil && msg2.Resource == nil {
		return true
	}

	if msg.Resource == nil || msg2.Resource == nil {
		return false
	}

	return msg.Resource.ID == msg2.Resource.ID
}

func CreateAddResourceMsgFromJSON(jsonString string) (*AddResourceMsg, error) {
	var msg *AddResourceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
