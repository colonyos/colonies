package rpc

import (
	"encoding/json"
)

const GetAttributeMsgType = "GetAttribute"

type GetAttributeMsg struct {
	RPC         RPC    `json:"rpc"`
	AttributeID string `json:"attributeid"`
}

func CreateGetAttributeMsg(attributeID string) *GetAttributeMsg {
	msg := &GetAttributeMsg{}
	msg.RPC.Method = GetAttributeMsgType
	msg.AttributeID = attributeID

	return msg
}

func (msg *GetAttributeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetAttributeMsgFromJSON(jsonString string) (*GetAttributeMsg, error) {
	var msg *GetAttributeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
