package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddAttributePayloadType = "addattributemsg"

type AddAttributeMsg struct {
	Attribute *core.Attribute `json:"attribute"`
	MsgType   string          `json:"msgtype"`
}

func CreateAddAttributeMsg(attribute *core.Attribute) *AddAttributeMsg {
	msg := &AddAttributeMsg{}
	msg.Attribute = attribute
	msg.MsgType = AddAttributePayloadType

	return msg
}

func (msg *AddAttributeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddAttributeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddAttributeMsgFromJSON(jsonString string) (*AddAttributeMsg, error) {
	var msg *AddAttributeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
