package rpc

import (
	"encoding/json"
)

const CloseFailedPayloadType = "closefailedmsg"

type CloseFailedMsg struct {
	ProcessID string   `json:"processid"`
	MsgType   string   `json:"msgtype"`
	Errors    []string `json:"errors"`
}

func CreateCloseFailedMsg(processID string, errors []string) *CloseFailedMsg {
	msg := &CloseFailedMsg{}
	msg.ProcessID = processID
	msg.MsgType = CloseFailedPayloadType
	msg.Errors = errors

	return msg
}

func (msg *CloseFailedMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CloseFailedMsg) Equals(msg2 *CloseFailedMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID {
		return true
	}

	counter := 0
	for _, r1 := range msg.Errors {
		for _, r2 := range msg2.Errors {
			if r1 == r2 {
				counter++
			}
		}
	}
	if counter != len(msg.Errors) && counter != len(msg2.Errors) {
		return false
	}

	return false
}

func (msg *CloseFailedMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateCloseFailedMsgFromJSON(jsonString string) (*CloseFailedMsg, error) {
	var msg *CloseFailedMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
