package rpc

import (
	"encoding/json"
)

const SubscribeProcessPayloadType = "subscribeprocessmsg"

type SubscribeProcessMsg struct {
	ProcessID   string `json:"processid"`
	RuntimeType string `json:"runtimetype"`
	State       int    `json:"state"`
	Timeout     int    `json:"timeout"`
	MsgType     string `json:"msgtype"`
}

func CreateSubscribeProcessMsg(processID string, runtimeType string, state int, timeout int) *SubscribeProcessMsg {
	msg := &SubscribeProcessMsg{}
	msg.ProcessID = processID
	msg.RuntimeType = runtimeType
	msg.State = state
	msg.Timeout = timeout
	msg.MsgType = SubscribeProcessPayloadType

	return msg
}

func (msg *SubscribeProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeProcessMsg) Equals(msg2 *SubscribeProcessMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessID == msg2.ProcessID &&
		msg.RuntimeType == msg2.RuntimeType &&
		msg.State == msg2.State &&
		msg.Timeout == msg2.Timeout {
		return true
	}

	return false
}

func CreateSubscribeProcessMsgFromJSON(jsonString string) (*SubscribeProcessMsg, error) {
	var msg *SubscribeProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
