package rpc

import (
	"encoding/json"
)

const RPCMethodApproveRuntime = "ApproveRuntime"

type ApproveRuntimeRPC struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateApproveRuntimeRPC(runtimeID string) *ApproveRuntimeRPC {
	msg := &ApproveRuntimeRPC{}
	msg.RPC.Method = RPCMethodApproveRuntime
	msg.RuntimeID = runtimeID

	return msg
}

func (msg *ApproveRuntimeRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateApproveRuntimeRPCFromJSON(jsonString string) (*ApproveRuntimeRPC, error) {
	var msg *ApproveRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
