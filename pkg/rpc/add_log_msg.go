package rpc

import (
	"encoding/json"
)

const AddLogPayloadType = "addlogmsg"

type AddLogMsg struct {
	ProcessID string `json:"processid"`
	Message   string `json:"message"`
	MsgType   string `json:"msgtype"`
}

func CreateAddLogMsg(processID string, logmsg string) *AddLogMsg {
	msg := &AddLogMsg{}
	msg.ProcessID = processID
	msg.Message = logmsg
	msg.MsgType = AddLogPayloadType

	return msg
}

func (msg *AddLogMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddLogMsg) Equals(msg2 *AddLogMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID && msg.Message == msg2.Message {
		return true
	}

	return false
}

func (msg *AddLogMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddLogMsgFromJSON(jsonString string) (*AddLogMsg, error) {
	var msg *AddLogMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
