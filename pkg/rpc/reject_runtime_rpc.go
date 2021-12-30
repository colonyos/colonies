package rpc

import (
	"encoding/json"
)

const RPCMethodRejectRuntime = "RejectRuntime"

type RejectRuntimeRPC struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateRejectRuntimeRPC(runtimeID string) *RejectRuntimeRPC {
	msg := &RejectRuntimeRPC{}
	msg.RPC.Method = RPCMethodRejectRuntime
	msg.RuntimeID = runtimeID

	return msg
}

func (msg *RejectRuntimeRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRejectRuntimeRPCFromJSON(jsonString string) (*RejectRuntimeRPC, error) {
	var msg *RejectRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
