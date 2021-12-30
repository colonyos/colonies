package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const RPCMethodAddRuntime = "AddRuntime"

type AddRuntimeRPC struct {
	RPC     RPC           `json:"rpc"`
	Runtime *core.Runtime `json:"runtime"`
}

func CreateAddRuntimeRPC(runtime *core.Runtime) *AddRuntimeRPC {
	msg := &AddRuntimeRPC{}
	msg.RPC.Method = RPCMethodAddRuntime
	msg.Runtime = runtime

	return msg
}

func (msg *AddRuntimeRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddRuntimeRPCFromJSON(jsonString string) (*AddRuntimeRPC, error) {
	var msg *AddRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
