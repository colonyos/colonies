package rpc

import (
	"encoding/json"
)

const ApproveRuntimePayloadType = "approveruntimemsg"

type ApproveRuntimeRPC struct {
	RuntimeID string `json:"runtimeid"`
	MsgType   string `json:"msgtype"`
}

func CreateApproveRuntimeMsg(runtimeID string) *ApproveRuntimeRPC {
	msg := &ApproveRuntimeRPC{}
	msg.RuntimeID = runtimeID
	msg.MsgType = ApproveRuntimePayloadType

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
