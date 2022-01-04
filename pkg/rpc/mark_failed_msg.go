package rpc

import (
	"encoding/json"
)

const MarkFailedPayloadType = "markfailedmsg"

type MarkFailedMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
}

func CreateMarkFailedMsg(processID string) *MarkFailedMsg {
	msg := &MarkFailedMsg{}
	msg.ProcessID = processID
	msg.MsgType = MarkFailedPayloadType

	return msg
}

func (msg *MarkFailedMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *MarkFailedMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateMarkFailedMsgFromJSON(jsonString string) (*MarkFailedMsg, error) {
	var msg *MarkFailedMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
