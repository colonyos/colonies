package rpc

import (
	"encoding/json"
)

const DeleteProcessGraphPayloadType = "deleteprocessgraphmsg"

type DeleteProcessGraphMsg struct {
	ProcessGraphID string `json:"processgraphid"`
	MsgType        string `json:"msgtype"`
	All            bool   `json:"all"`
}

func CreateDeleteProcessGraphMsg(processGraphID string) *DeleteProcessGraphMsg {
	msg := &DeleteProcessGraphMsg{}
	msg.ProcessGraphID = processGraphID
	msg.MsgType = DeleteProcessGraphPayloadType

	return msg
}

func (msg *DeleteProcessGraphMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteProcessGraphMsg) Equals(msg2 *DeleteProcessGraphMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ProcessGraphID == msg2.ProcessGraphID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *DeleteProcessGraphMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteProcessGraphMsgFromJSON(jsonString string) (*DeleteProcessGraphMsg, error) {
	var msg *DeleteProcessGraphMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
