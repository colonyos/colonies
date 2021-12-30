package rpc

import (
	"encoding/json"
)

const RPCMethodGetColonies = "GetColonies"

type GetColoniesRPC struct {
	RPC          RPC    `json:"rpc"`
	RootPassword string `json:"rootpassword"`
}

func CreateGetColoniesRPC(rootPassword string) *GetColoniesRPC {
	msg := &GetColoniesRPC{}
	msg.RPC.Method = RPCMethodGetColonies
	msg.RootPassword = rootPassword

	return msg
}

func (msg *GetColoniesRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetColoniesRPCFromJSON(jsonString string) (*GetColoniesRPC, error) {
	var msg *GetColoniesRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
