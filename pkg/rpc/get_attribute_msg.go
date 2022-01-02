package rpc

import (
	"encoding/json"
)

const GetAttributeMsgType = "getattribute"

type GetAttributeMsg struct {
	AttributeID string `json:"attributeid"`
}

func CreateGetAttributeMsg(attributeID string) *GetAttributeMsg {
	msg := &GetAttributeMsg{}
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

func (msg *GetAttributeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
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
