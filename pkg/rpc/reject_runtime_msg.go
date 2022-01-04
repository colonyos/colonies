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

func CreateRejectRuntimeMsgFromJSON(jsonString string) (*RejectRuntimeMsg, error) {
	var msg *RejectRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
