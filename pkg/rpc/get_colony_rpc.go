package rpc

import (
	"encoding/json"
)

const RPCMethodGetColony = "GetColony"

type GetColonyRPC struct {
	RPC      RPC    `json:"rpc"`
	ColonyID string `json:"colonyid"`
}

func CreateGetColonyRPC(colonyID string) *GetColonyRPC {
	msg := &GetColonyRPC{}
	msg.RPC.Method = RPCMethodGetColony
	msg.ColonyID = colonyID

	return msg
}

func (msg *GetColonyRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetColonyRPCFromJSON(jsonString string) (*GetColonyRPC, error) {
	var msg *GetColonyRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
