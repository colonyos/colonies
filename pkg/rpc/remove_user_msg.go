package rpc

import (
	"encoding/json"
)

const RemoveUserPayloadType = "removeusermsg"

type RemoveUserMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveUserMsg(colonyName string, name string) *RemoveUserMsg {
	msg := &RemoveUserMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = RemoveUserPayloadType

	return msg
}

func (msg *RemoveUserMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveUserMsg) Equals(msg2 *RemoveUserMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Name == msg2.Name && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func (msg *RemoveUserMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveUserMsgFromJSON(jsonString string) (*RemoveUserMsg, error) {
	var msg *RemoveUserMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
