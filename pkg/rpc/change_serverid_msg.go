package rpc

import (
	"encoding/json"
)

const ChangeServerIDPayloadType = "changeserveridmsg"

type ChangeServerIDMsg struct {
	ServerID string `json:"serverid"`
	MsgType  string `json:"msgtype"`
}

func CreateChangeServerIDMsg(serverID string) *ChangeServerIDMsg {
	msg := &ChangeServerIDMsg{}
	msg.ServerID = serverID
	msg.MsgType = ChangeServerIDPayloadType

	return msg
}

func (msg *ChangeServerIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeServerIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeServerIDMsg) Equals(msg2 *ChangeServerIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ServerID == msg2.ServerID {
		return true
	}

	return false
}

func CreateChangeServerIDMsgFromJSON(jsonString string) (*ChangeServerIDMsg, error) {
	var msg *ChangeServerIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
