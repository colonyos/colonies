package rpc

import (
	"encoding/json"
)

const RPCMethodGetRuntime = "GetRuntime"

type GetRuntimeRPC struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateGetRuntimeRPC(runtimeID string) *GetRuntimeRPC {
	msg := &GetRuntimeRPC{}
	msg.RPC.Method = RPCMethodGetRuntime
	msg.RuntimeID = runtimeID

	return msg
}

func (msg *GetRuntimeRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetRuntimeRPCFromJSON(jsonString string) (*GetRuntimeRPC, error) {
	var msg *GetRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
