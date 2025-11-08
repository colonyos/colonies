package rpc

import (
	"encoding/json"
)

const GetServicePayloadType = "getservicemsg"

type GetServiceMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateGetServiceMsg(namespace, name string) *GetServiceMsg {
	msg := &GetServiceMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = GetServicePayloadType

	return msg
}

func (msg *GetServiceMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServiceMsg) Equals(msg2 *GetServiceMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateGetServiceMsgFromJSON(jsonString string) (*GetServiceMsg, error) {
	var msg *GetServiceMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
