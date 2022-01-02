package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const ApproveRuntimeMsgType = "approveruntime"

type ApproveRuntimeRPC struct {
	RPC       RPC    `json:"rpc"`
	RuntimeID string `json:"runtimeid"`
}

func CreateApproveRuntimeMsg(runtimeID string) *ApproveRuntimeRPC {
	msg := &ApproveRuntimeRPC{}
	msg.RPC.Method = ApproveRuntimeMsgType
	msg.RPC.Nonce = core.GenerateRandomID()
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

func (msg *ApproveRuntimeRPC) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateApproveRuntimeMsgFromJSON(jsonString string) (*ApproveRuntimeRPC, error) {
	var msg *ApproveRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
