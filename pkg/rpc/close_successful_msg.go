package rpc

import (
	"encoding/json"
)

const CloseSuccessfulPayloadType = "closesuccessfulmsg"

type CloseSuccessfulMsg struct {
	ProcessID string   `json:"processid"`
	MsgType   string   `json:"msgtype"`
	Output    []string `json:"out"`
}

func CreateCloseSuccessfulMsg(processID string) *CloseSuccessfulMsg {
	msg := &CloseSuccessfulMsg{}
	msg.ProcessID = processID
	msg.MsgType = CloseSuccessfulPayloadType

	return msg
}

func (msg *CloseSuccessfulMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CloseSuccessfulMsg) Equals(msg2 *CloseSuccessfulMsg) bool {
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

func (msg *CloseSuccessfulMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateCloseSuccessfulMsgFromJSON(jsonString string) (*CloseSuccessfulMsg, error) {
	var msg *CloseSuccessfulMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
