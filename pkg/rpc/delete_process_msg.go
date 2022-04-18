package rpc

import (
	"encoding/json"
)

const DeleteProcessPayloadType = "deleteprocessmsg"

type DeleteProcessMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
	All       bool   `json:"all"`
}

func CreateDeleteProcessMsg(processID string) *DeleteProcessMsg {
	msg := &DeleteProcessMsg{}
	msg.ProcessID = processID
	msg.MsgType = DeleteProcessPayloadType

	return msg
}

func (msg *DeleteProcessMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteProcessMsg) Equals(msg2 *DeleteProcessMsg) bool {
	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *DeleteProcessMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteProcessMsgFromJSON(jsonString string) (*DeleteProcessMsg, error) {
	var msg *DeleteProcessMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
