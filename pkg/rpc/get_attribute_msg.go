package rpc

import (
	"encoding/json"
)

const GetAttributePayloadType = "getattributemsg"

type GetAttributeMsg struct {
	AttributeID string `json:"attributeid"`
	MsgType     string `json:"msgtype"`
}

func CreateGetAttributeMsg(attributeID string) *GetAttributeMsg {
	msg := &GetAttributeMsg{}
	msg.AttributeID = attributeID
	msg.MsgType = GetAttributePayloadType

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

func (msg *GetAttributeMsg) Equals(msg2 *GetAttributeMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.AttributeID == msg2.AttributeID {
		return true
	}

	return false
}

func CreateGetAttributeMsgFromJSON(jsonString string) (*GetAttributeMsg, error) {
	var msg *GetAttributeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
