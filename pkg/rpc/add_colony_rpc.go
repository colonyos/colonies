package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const RPCMethodAddColony = "AddColony"

type AddColonyRPC struct {
	RPC          RPC          `json:"rpc"`
	RootPassword string       `json:"rootpassword"`
	Colony       *core.Colony `json:"colony"`
}

func CreateAddColonyRPC(rootPassword string, colony *core.Colony) *AddColonyRPC {
	msg := &AddColonyRPC{}
	msg.RPC.Method = RPCMethodAddColony
	msg.RootPassword = rootPassword
	msg.Colony = colony

	return msg
}

func (msg *AddColonyRPC) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddColonyRPCFromJSON(jsonString string) (*AddColonyRPC, error) {
	var msg *AddColonyRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
