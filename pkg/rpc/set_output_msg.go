package rpc

import (
	"encoding/json"
)

const SetOutputPayloadType = "setoutputmsg"

type SetOutputMsg struct {
	ProcessID string        `json:"processid"`
	MsgType   string        `json:"msgtype"`
	Output    []interface{} `json:"out"`
}

func CreateSetOutputMsg(processID string, output []interface{}) *SetOutputMsg {
	msg := &SetOutputMsg{}
	msg.ProcessID = processID
	msg.MsgType = SetOutputPayloadType
	msg.Output = output

	return msg
}

func (msg *SetOutputMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SetOutputMsg) Equals(msg2 *SetOutputMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID {
		return true
	}

	counter := 0
	for _, r1 := range msg.Output {
		for _, r2 := range msg2.Output {
			if r1 == r2 {
				counter++
			}
		}
	}
	if counter != len(msg.Output) && counter != len(msg2.Output) {
		return false
	}

	return false
}

func (msg *SetOutputMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateSetOutputMsgFromJSON(jsonString string) (*SetOutputMsg, error) {
	var msg *SetOutputMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
