package rpc

import (
	"encoding/json"
)

const CloseSuccessfulPayloadType = "closesuccessfulmsg"

type CloseSuccesfulMsg struct {
	ProcessID string `json:"processid"`
	MsgType   string `json:"msgtype"`
}

func CreateCloseSuccessfulMsg(processID string) *CloseSuccesfulMsg {
	msg := &CloseSuccesfulMsg{}
	msg.ProcessID = processID
	msg.MsgType = CloseSuccessfulPayloadType

	return msg
}

func (msg *CloseSuccesfulMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CloseSuccesfulMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateCloseSuccessfulMsgFromJSON(jsonString string) (*CloseSuccesfulMsg, error) {
	var msg *CloseSuccesfulMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
