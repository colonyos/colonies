package rpc

import (
	"encoding/json"
)

const RejectRuntimeMsgType = "RejectRuntime"

type RejectRuntimeMsg struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateRejectRuntimeMsg(runtimeID string) *RejectRuntimeMsg {
	msg := &RejectRuntimeMsg{}
	msg.RPC.Method = RejectRuntimeMsgType
	msg.RuntimeID = runtimeID

	return msg
}

func (msg *RejectRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
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
