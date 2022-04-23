package rpc

import (
	"encoding/json"
)

const GetRuntimePayloadType = "getruntimemsg"

type GetRuntimeMsg struct {
	RuntimeID string `json:"runtimeid"`
	MsgType   string `json:"msgtype"`
}

func CreateGetRuntimeMsg(runtimeID string) *GetRuntimeMsg {
	msg := &GetRuntimeMsg{}
	msg.RuntimeID = runtimeID
	msg.MsgType = GetRuntimePayloadType

	return msg
}

func (msg *GetRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetRuntimeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetRuntimeMsg) Equals(msg2 *GetRuntimeMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.RuntimeID == msg2.RuntimeID {
		return true
	}

	return false
}

func CreateGetRuntimeMsgFromJSON(jsonString string) (*GetRuntimeMsg, error) {
	var msg *GetRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
