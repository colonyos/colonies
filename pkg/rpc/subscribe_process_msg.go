package rpc

import (
	"encoding/json"
)

const SubscribeProcessMsgType = "subscribeprocess"

type SubscribeProcessMsg struct {
	ProcessID string `json:"processid"`
	State     int    `json:"state"`
	Timeout   int    `json:"timeout"`
}

func CreateSubscribeProcessMsg(processID string, state int, timeout int) *SubscribeProcessMsg {
	msg := &SubscribeProcessMsg{}
	msg.ProcessID = processID
	msg.State = state
	msg.Timeout = timeout

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

func CreateSubscribeProcessMsgFromJSON(jsonString string) (*SubscribeProcessMsg, error) {
	var msg *SubscribeProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
