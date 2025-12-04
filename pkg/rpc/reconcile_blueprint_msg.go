package rpc

import (
	"encoding/json"
)

const ReconcileBlueprintPayloadType = "reconcileblueprintmsg"

type ReconcileBlueprintMsg struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Force     bool   `json:"force"`
	MsgType   string `json:"msgtype"`
}

func CreateReconcileBlueprintMsg(namespace, name string, force bool) *ReconcileBlueprintMsg {
	msg := &ReconcileBlueprintMsg{}
	msg.Namespace = namespace
	msg.Name = name
	msg.Force = force
	msg.MsgType = ReconcileBlueprintPayloadType

	return msg
}

func (msg *ReconcileBlueprintMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ReconcileBlueprintMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ReconcileBlueprintMsg) Equals(msg2 *ReconcileBlueprintMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Name == msg2.Name &&
		msg.Force == msg2.Force
}

func CreateReconcileBlueprintMsgFromJSON(jsonString string) (*ReconcileBlueprintMsg, error) {
	var msg *ReconcileBlueprintMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
