package rpc

import (
	"encoding/json"
)

const RemoveProcessPayloadType = "removeprocessmsg"

type RemoveProcessMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
	All       bool   `json:"all"`
}

func CreateRemoveProcessMsg(processID string) *RemoveProcessMsg {
	msg := &RemoveProcessMsg{}
	msg.ProcessID = processID
	msg.MsgType = RemoveProcessPayloadType

	return msg
}

func (msg *RemoveProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveProcessMsg) Equals(msg2 *RemoveProcessMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *RemoveProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveProcessMsgFromJSON(jsonString string) (*RemoveProcessMsg, error) {
	var msg *RemoveProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
