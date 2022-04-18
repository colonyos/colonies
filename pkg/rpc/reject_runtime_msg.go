package rpc

import (
	"encoding/json"
)

const RejectRuntimePayloadType = "rejectruntimemsg"

type RejectRuntimeMsg struct {
	RuntimeID string `json:"runtimeid"`
	MsgType   string `json:"msgtype"`
}

func CreateRejectRuntimeMsg(runtimeID string) *RejectRuntimeMsg {
	msg := &RejectRuntimeMsg{}
	msg.RuntimeID = runtimeID
	msg.MsgType = RejectRuntimePayloadType

	return msg
}

func (msg *RejectRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RejectRuntimeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RejectRuntimeMsg) Equals(msg2 *RejectRuntimeMsg) bool {
	if msg.MsgType == msg2.MsgType && msg.RuntimeID == msg2.RuntimeID {
		return true
	}

	return false
}

func CreateRejectRuntimeMsgFromJSON(jsonString string) (*RejectRuntimeMsg, error) {
	var msg *RejectRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
