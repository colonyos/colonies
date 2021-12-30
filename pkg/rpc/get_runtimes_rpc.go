package rpc

import (
	"encoding/json"
)

const RPCMethodGetRuntimes = "GetRuntimes"

type GetRuntimesRPC struct {
	RPC      RPC    `json:"rpc"`
	ColonyID string `json:"colonyid"`
}

func CreateGetRuntimesRPC(colonyID string) *GetRuntimesRPC {
	msg := &GetRuntimesRPC{}
	msg.RPC.Method = RPCMethodGetRuntimes
	msg.ColonyID = colonyID

	return msg
}

func (msg *GetRuntimesRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetRuntimesRPCFromJSON(jsonString string) (*GetRuntimesRPC, error) {
	var msg *GetRuntimesRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
