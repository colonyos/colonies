package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

type RPC struct {
	Method string `json:"method"`
}

const MethodSubmitProcessSpec = "SubmitProcessSpec"

type SubmitProcessSpec struct {
	RPC         RPC               `json:"rpc"`
	ProcessSpec *core.ProcessSpec `json:"spec"`
}

func CreateSubmitProcessSpec(processSpec *core.ProcessSpec) *SubmitProcessSpec {
	msg := &SubmitProcessSpec{}
	msg.RPC.Method = "SubmitProcessSpec"
	msg.ProcessSpec = processSpec
	return msg
}

func CreateSubmitProcessSpecFromJSON(jsonString string) (*SubmitProcessSpec, error) {
	var msg *SubmitProcessSpec

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func DetermineRPCMethod(jsonString string) string {
	var msgMap map[string]interface{}
	json.Unmarshal([]byte(jsonString), &msgMap)
	rpcMap := msgMap["rpc"].(map[string]interface{})

	return rpcMap["method"].(string)
}

func (msg *SubmitProcessSpec) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
