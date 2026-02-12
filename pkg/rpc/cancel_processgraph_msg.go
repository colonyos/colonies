package rpc

import (
	"encoding/json"
)

const CancelProcessGraphPayloadType = "cancelprocessgraphmsg"

type CancelProcessGraphMsg struct {
	ProcessGraphID string `json:"processgraphid"`
	MsgType        string `json:"msgtype"`
}

func CreateCancelProcessGraphMsg(processGraphID string) *CancelProcessGraphMsg {
	msg := &CancelProcessGraphMsg{}
	msg.ProcessGraphID = processGraphID
	msg.MsgType = CancelProcessGraphPayloadType

	return msg
}

func (msg *CancelProcessGraphMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CancelProcessGraphMsg) Equals(msg2 *CancelProcessGraphMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessGraphID == msg2.ProcessGraphID {
		return true
	}

	return false
}

func (msg *CancelProcessGraphMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateCancelProcessGraphMsgFromJSON(jsonString string) (*CancelProcessGraphMsg, error) {
	var msg *CancelProcessGraphMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
