package rpc

import (
	"encoding/json"
)

const GetProcessPayloadType = "getprocessmsg"

type GetProcessMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
}

func CreateGetProcessMsg(processID string) *GetProcessMsg {
	msg := &GetProcessMsg{}
	msg.ProcessID = processID
	msg.MsgType = GetProcessPayloadType

	return msg
}

func (msg *GetProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessMsg) Equals(msg2 *GetProcessMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID {
		return true
	}

	return false
}

func CreateGetProcessMsgFromJSON(jsonString string) (*GetProcessMsg, error) {
	var msg *GetProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
