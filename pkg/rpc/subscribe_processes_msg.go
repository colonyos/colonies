package rpc

import (
	"encoding/json"
)

const SubscribeProcessesMsgType = "subscribeprocesses"

type SubscribeProcessesMsg struct {
	RuntimeType string `json:"runtimetype"`
	State       int    `json:"state"`
	Timeout     int    `json:"timeout"`
}

func CreateSubscribeProcessesMsg(runtimeType string, state int, timeout int) *SubscribeProcessesMsg {
	msg := &SubscribeProcessesMsg{}
	msg.RuntimeType = runtimeType
	msg.State = state
	msg.Timeout = timeout

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

func CreateSubscribeProcessesMsgFromJSON(jsonString string) (*SubscribeProcessesMsg, error) {
	var msg *SubscribeProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
