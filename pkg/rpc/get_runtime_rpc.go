package rpc

import (
	"encoding/json"
)

const GetRuntimeMsgType = "GetRuntime"

type GetRuntimeMsg struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateGetRuntimeMsg(runtimeID string) *GetRuntimeMsg {
	msg := &GetRuntimeMsg{}
	msg.RPC.Method = GetRuntimeMsgType
	msg.RuntimeID = runtimeID

	return msg
}

func (msg *GetRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetRuntimeMsgFromJSON(jsonString string) (*GetRuntimeMsg, error) {
	var msg *GetRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
