package rpc

import (
	"encoding/json"
)

const GetBlueprintPayloadType = "getblueprintmsg"

type GetBlueprintMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	MsgType   string `json:"msgtype"`
}

func CreateGetBlueprintMsg(namespace, name string) *GetBlueprintMsg {
	msg := &GetBlueprintMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.MsgType = GetBlueprintPayloadType

	return msg
}

func (msg *GetBlueprintMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintMsg) Equals(msg2 *GetBlueprintMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name
}

func CreateGetBlueprintMsgFromJSON(jsonString string) (*GetBlueprintMsg, error) {
	var msg *GetBlueprintMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
