package rpc

import (
	"encoding/json"
)

const CancelProcessPayloadType = "cancelprocessmsg"

type CancelProcessMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
}

func CreateCancelProcessMsg(processID string) *CancelProcessMsg {
	msg := &CancelProcessMsg{}
	msg.ProcessID = processID
	msg.MsgType = CancelProcessPayloadType

	return msg
}

func (msg *CancelProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CancelProcessMsg) Equals(msg2 *CancelProcessMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID {
		return true
	}

	return false
}

func (msg *CancelProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateCancelProcessMsgFromJSON(jsonString string) (*CancelProcessMsg, error) {
	var msg *CancelProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
