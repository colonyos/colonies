package rpc

import (
	"encoding/json"
)

const RemoveColonyPayloadType = "removecolonymsg"

type RemoveColonyMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveColonyMsg(colonyName string) *RemoveColonyMsg {
	msg := &RemoveColonyMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = RemoveColonyPayloadType

	return msg
}

func (msg *RemoveColonyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveColonyMsg) Equals(msg2 *RemoveColonyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func (msg *RemoveColonyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveColonyMsgFromJSON(jsonString string) (*RemoveColonyMsg, error) {
	var msg *RemoveColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
