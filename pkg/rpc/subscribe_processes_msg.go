package rpc

import (
	"encoding/json"
)

const SubscribeProcessesPayloadType = "subscribeprocessesmsg"

type SubscribeProcessesMsg struct {
	RuntimeType string `json:"runtimetype"`
	State       int    `json:"state"`
	Timeout     int    `json:"timeout"`
	MsgType     string `json:"msgtype"`
}

func CreateSubscribeProcessesMsg(runtimeType string, state int, timeout int) *SubscribeProcessesMsg {
	msg := &SubscribeProcessesMsg{}
	msg.RuntimeType = runtimeType
	msg.State = state
	msg.Timeout = timeout
	msg.MsgType = SubscribeProcessesPayloadType

	return msg
}

func (msg *SubscribeProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeProcessesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeProcessesMsg) Equals(msg2 *SubscribeProcessesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.RuntimeType == msg2.RuntimeType &&
		msg.State == msg2.State &&
		msg.Timeout == msg2.Timeout {
		return true
	}

	return false
}

func CreateSubscribeProcessesMsgFromJSON(jsonString string) (*SubscribeProcessesMsg, error) {
	var msg *SubscribeProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
