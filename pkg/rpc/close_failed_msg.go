package rpc

import (
	"encoding/json"
)

const CloseFailedPayloadType = "closefailedmsg"

type CloseFailedMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
}

func CreateCloseFailedMsg(processID string) *CloseFailedMsg {
	msg := &CloseFailedMsg{}
	msg.ProcessID = processID
	msg.MsgType = CloseFailedPayloadType

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
	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID {
		return true
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
