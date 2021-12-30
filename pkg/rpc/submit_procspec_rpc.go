package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const RPCMethodSubmitProcessSpec = "SubmitProcessSpec"

type SubmitProcessSpecRPC struct {
	RPC         RPC               `json:"rpc"`
	ProcessSpec *core.ProcessSpec `json:"spec"`
}

func CreateSubmitProcessSpecRPC(processSpec *core.ProcessSpec) *SubmitProcessSpecRPC {
	msg := &SubmitProcessSpecRPC{}
	msg.RPC.Method = RPCMethodSubmitProcessSpec
	msg.ProcessSpec = processSpec

	return msg
}

func (msg *SubmitProcessSpecRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateSubmitProcessSpecRPCFromJSON(jsonString string) (*SubmitProcessSpecRPC, error) {
	var msg *SubmitProcessSpecRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
