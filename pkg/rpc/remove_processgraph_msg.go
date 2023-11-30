package rpc

import (
	"encoding/json"
)

const RemoveProcessGraphPayloadType = "removeprocessgraphmsg"

type RemoveProcessGraphMsg struct {
	ProcessGraphID string `json:"processgraphid"`
	MsgType        string `json:"msgtype"`
	All            bool   `json:"all"`
}

func CreateRemoveProcessGraphMsg(processGraphID string) *RemoveProcessGraphMsg {
	msg := &RemoveProcessGraphMsg{}
	msg.ProcessGraphID = processGraphID
	msg.MsgType = RemoveProcessGraphPayloadType

	return msg
}

func (msg *RemoveProcessGraphMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveProcessGraphMsg) Equals(msg2 *RemoveProcessGraphMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessGraphID == msg2.ProcessGraphID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *RemoveProcessGraphMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveProcessGraphMsgFromJSON(jsonString string) (*RemoveProcessGraphMsg, error) {
	var msg *RemoveProcessGraphMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
