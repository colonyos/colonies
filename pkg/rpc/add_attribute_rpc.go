package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const AddAttributeMsgType = "AddAttribute"

type AddAttributeMsg struct {
	RPC       RPC             `json:"rpc"`
	Attribute *core.Attribute `json:"attribute"`
}

func CreateAddAttributeMsg(attribute *core.Attribute) *AddAttributeMsg {
	msg := &AddAttributeMsg{}
	msg.RPC.Method = AddAttributeMsgType
	msg.Attribute = attribute

	return msg
}

func (msg *AddAttributeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
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
