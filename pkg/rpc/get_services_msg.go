package rpc

import (
	"encoding/json"
)

const GetServicesPayloadType = "getservicesmsg"

type GetServicesMsg struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	MsgType   string `json:"msgtype"`
}

func CreateGetServicesMsg(namespace, kind string) *GetServicesMsg {
	msg := &GetServicesMsg{}
	msg.Namespace = namespace
	msg.Kind = kind
	msg.MsgType = GetServicesPayloadType

	return msg
}

func (msg *GetServicesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServicesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetServicesMsg) Equals(msg2 *GetServicesMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Kind == msg2.Kind
}

func CreateGetServicesMsgFromJSON(jsonString string) (*GetServicesMsg, error) {
	var msg *GetServicesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
