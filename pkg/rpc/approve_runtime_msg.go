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

func (msg *ApproveRuntimeRPC) Equals(msg2 *ApproveRuntimeRPC) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.RuntimeID == msg2.RuntimeID {
		return true
	}

	return false
}

func CreateApproveRuntimeMsgFromJSON(jsonString string) (*ApproveRuntimeRPC, error) {
	var msg *ApproveRuntimeRPC

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
