package rpc

import (
	"encoding/json"
)

const GetResourcesPayloadType = "getresourcesmsg"

type GetResourcesMsg struct {
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	MsgType   string `json:"msgtype"`
}

func CreateGetResourcesMsg(namespace, kind string) *GetResourcesMsg {
	msg := &GetResourcesMsg{}
	msg.Namespace = namespace
	msg.Kind = kind
	msg.MsgType = GetResourcesPayloadType

	return msg
}

func (msg *GetResourcesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourcesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetResourcesMsg) Equals(msg2 *GetResourcesMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Kind == msg2.Kind
}

func CreateGetResourcesMsgFromJSON(jsonString string) (*GetResourcesMsg, error) {
	var msg *GetResourcesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
