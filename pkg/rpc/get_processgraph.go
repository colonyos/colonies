package rpc

import (
	"encoding/json"
)

const GetProcessGraphPayloadType = "getprocessgraphmsg"

type GetProcessGraphMsg struct {
	ProcessGraphID string `json:"processgraphid"`
	MsgType        string `json:"msgtype"`
}

func CreateGetProcessGraphMsg(processGraphID string) *GetProcessGraphMsg {
	msg := &GetProcessGraphMsg{}
	msg.ProcessGraphID = processGraphID
	msg.MsgType = GetProcessGraphPayloadType

	return msg
}

func (msg *GetProcessGraphMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessGraphMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessGraphMsg) Equals(msg2 *GetProcessGraphMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessGraphID == msg2.ProcessGraphID {
		return true
	}

	return false
}

func CreateGetProcessGraphMsgFromJSON(jsonString string) (*GetProcessGraphMsg, error) {
	var msg *GetProcessGraphMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
