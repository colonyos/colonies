package rpc

import (
	"encoding/json"
)

const DeleteRuntimePayloadType = "deleteruntimemsg"

type DeleteRuntimeMsg struct {
	RuntimeID string `json:"runtimeid"`
	MsgType   string `json:"msgtype"`
}

func CreateDeleteRuntimeMsg(runtimeID string) *DeleteRuntimeMsg {
	msg := &DeleteRuntimeMsg{}
	msg.RuntimeID = runtimeID
	msg.MsgType = DeleteRuntimePayloadType

	return msg
}

func (msg *DeleteRuntimeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteRuntimeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteRuntimeMsg) Equals(msg2 *DeleteRuntimeMsg) bool {
	if msg.MsgType == msg2.MsgType && msg.RuntimeID == msg2.RuntimeID {
		return true
	}

	return false
}

func CreateDeleteRuntimeMsgFromJSON(jsonString string) (*DeleteRuntimeMsg, error) {
	var msg *DeleteRuntimeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
